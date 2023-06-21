package processor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ridwanadhip/universal-uploader/config"
	"github.com/ridwanadhip/universal-uploader/hook"
	"github.com/ridwanadhip/universal-uploader/util"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const NullExpr = "NULL"

type mySQLImplementation struct {
	input       *config.Input
	target      *config.Target
	db          *gorm.DB
	verboseMode bool
	procHook    hook.ProcessorHook
}

func NewMySQLImplementation(input *config.Input, target *config.Target, verboseMode bool, procHook hook.ProcessorHook) (*mySQLImplementation, error) {
	connStr := "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	connStr = fmt.Sprintf(connStr, target.Username, target.Password, target.Host, target.Port, target.Name)
	db, err := gorm.Open(mysql.Open(connStr))
	if err != nil {
		return nil, err
	}

	return &mySQLImplementation{input, target, db, verboseMode, procHook}, nil
}

func (impl *mySQLImplementation) Process(data [][]string) error {
	switch impl.target.Mode {
	case config.TargetModeUpdate:
		return impl.performUpdate(data)
	default:
		return impl.performUpsert(data)
	}
}

func (impl *mySQLImplementation) DryRun(data [][]string) error {
	// TODO: implement dry run
	return fmt.Errorf("not implemented")
}

func (impl *mySQLImplementation) Close() {
	db, _ := impl.db.DB()
	if db != nil {
		db.Close()
	}
}

func (impl *mySQLImplementation) getColumnNames() []string {
	fields := impl.target.Fields

	columnNames := []string{}
	for i := range fields {
		f := &fields[i]
		columnNames = append(columnNames, f.Name)
	}

	return columnNames
}

func (impl *mySQLImplementation) performUpdate(data [][]string) error {
	updatedRows := []map[string]any{}
	whereParams := [][]any{}

	for i := range data {
		row := map[string]any{}
		param := []any{}

		for j := range impl.target.Fields {
			f := &impl.target.Fields[j]

			cval, err := impl.parseColumnValue(f, data[i])
			if err != nil {
				return err
			}

			if f.FilterQuery {
				param = append(param, cval)
			} else {
				row[f.Name] = cval
			}
		}

		updatedRows = append(updatedRows, row)
		whereParams = append(whereParams, param)
	}

	// build where clause and updated column names manually
	columnNames := []string{}
	clausePlaceholders := []string{}
	for i := range impl.target.Fields {
		f := &impl.target.Fields[i]

		if f.FilterQuery {
			placeholder := fmt.Sprintf("`%s` = ?", f.Name)
			clausePlaceholders = append(clausePlaceholders, placeholder)
		} else {
			columnNames = append(columnNames, f.Name)
		}
	}

	// creating a string like '`col1` = ? AND `col2` = ?'
	whereClause := strings.Join(clausePlaceholders, " AND ")

	for i := range data {
		err := impl.db.
			Table(impl.target.DataName).
			Select(columnNames).
			Where(whereClause, whereParams[i]...).
			Updates(updatedRows[i]).
			Error

		if err != nil {
			return err
		}
	}

	return nil
}

func (impl *mySQLImplementation) performUpsert(data [][]string) error {
	newRows := []map[string]any{}
	for i := range data {
		row := map[string]any{}
		for j := range impl.target.Fields {
			f := &impl.target.Fields[j]

			cval, err := impl.parseColumnValue(f, data[i])
			if err != nil {
				return err
			}

			row[f.Name] = cval
		}

		newRows = append(newRows, row)
	}

	// log generated values
	if impl.verboseMode {
		fmt.Printf("[Target MySQL ID: %s] %s\n", impl.target.ID, util.Jsonify(newRows))
	}

	query := impl.db.
		Table(impl.target.DataName).
		Select(impl.getColumnNames())

	// upsert mode is deprecated, will be removed later
	if impl.target.Upsert || impl.target.Mode == config.TargetModeUpsert {
		upsertHandler := impl.constructUpsertHandler()
		query = query.Clauses(upsertHandler)
	}

	return query.Create(newRows).Error
}

func (impl *mySQLImplementation) parseColumnValue(field *config.TargetField, rowData []string) (any, error) {
	if field == nil {
		return nil, fmt.Errorf("missing target field config")
	}

	fieldValue := ""
	if field.Value != nil {
		fieldValue = *field.Value
	}

	md := hook.NewProcessorHookMetadataFromTarget(impl.target)

	// if hook available then override pre-formatted field value with new value from hook
	if impl.procHook != nil {
		newVal, err := impl.procHook.OverrideBaseFieldValue(md, field.ID, fieldValue)
		if err != nil {
			return nil, err
		}

		fieldValue = newVal
	}

	// handle null sql
	if val, ok := toSQLNull(fieldValue); ok {
		return val, nil
	}

	// handle current timestamp
	if val, ok := toSQLCurrentTimestamp(fieldValue); ok {
		return val, nil
	}

	for _, ref := range field.References {
		refID := util.RemoveToken(ref)

		// if not exists then treat missing reference value as empty string
		replacer := ""
		if i, ok := impl.input.FieldsIndexMap[refID]; ok {
			replacer = rowData[i]
		}

		fieldValue = strings.ReplaceAll(fieldValue, ref, replacer)
	}

	// if the original value or formatting result is empty string then convert it to defined value from YAML
	if fieldValue == "" {
		if field.EmptyAsNil {
			fieldValue = string(NilValue)
		} else if field.ValueIfEmpty != nil {
			fieldValue = *field.ValueIfEmpty
		}

		// handle SQL null if predefined default value in YAML is null
		if val, ok := toSQLNull(fieldValue); ok {
			return val, nil
		}
	}

	// if hook available then override formatted field value with new value from hook
	if impl.procHook != nil {
		newVal, err := impl.procHook.OverrideFormattedFieldValue(md, field.ID, fieldValue)
		if err != nil {
			return nil, err
		}

		fieldValue = newVal
	}

	return parseStringToType(field.Type, fieldValue)
}

func (impl *mySQLImplementation) constructUpsertHandler() (res clause.OnConflict) {
	cols := []clause.Column{}
	for _, f := range impl.target.UniqueConstraints {
		cols = append(cols, clause.Column{Name: f.Name})
	}

	updates := []string{}
	for _, f := range impl.target.OldValueReplacers {
		updates = append(updates, f.Name)
	}

	res.Columns = cols
	res.DoUpdates = clause.AssignmentColumns(updates)

	return res
}

func toSQLNull(val string) (any, bool) {
	if Function(val) == NilValue {
		return gorm.Expr(NullExpr), true
	}

	return val, false
}

func toSQLCurrentTimestamp(val string) (any, bool) {
	if Function(val) == CurrentTimestamp {
		return time.Now(), true
	}

	return val, false
}

func parseStringToType(valType config.ValueType, stringVal string) (any, error) {
	switch valType {
	case config.ValueTypeBoolean:
		return strconv.ParseBool(stringVal)
	case config.ValueTypeInteger:
		return strconv.ParseInt(stringVal, 10, 64)
	case config.ValueTypeDecimal:
		return strconv.ParseFloat(stringVal, 64)
	default:
		return stringVal, nil
	}
}
