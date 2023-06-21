package input

import (
	"fmt"
	"strings"

	"github.com/ridwanadhip/universal-uploader/config"
)

type (
	InputType string
)

// known types
const (
	InputTypeCSV InputType = "csv"
)

type Parser struct {
	cfg         *config.Config
	inputParser InputParser
	fieldNames  []string
}

type Batch struct {
	Data  [][]string
	Index int
}

type InputParser interface {
	NextBatch(path string) (batch *Batch, exists bool, err error)
	GetFieldNames(path string) ([]string, error)
	GetCurrentIndex() int
	Close()
}

func NewParser(cfg *config.Config) (parser Parser, err error) {
	switch InputType(cfg.Input.Type) {
	case InputTypeCSV:
		parser = Parser{cfg: cfg, inputParser: &csvParser{batchSize: cfg.BatchSize}}
	default:
		return Parser{}, fmt.Errorf("unknown input parser type: %s", cfg.Input.Type)
	}

	// inject fields if user didn't define field spec in config file
	fieldNames, err := parser.inputParser.GetFieldNames(cfg.Args.InputPath)
	if err != nil {
		return parser, err
	}

	parser.fieldNames = fieldNames
	cfg.InjectFieldsWithDefaultValue(fieldNames)

	return parser, nil
}

func (parser *Parser) Validate() error {
	totalInputFields := len(parser.cfg.Input.Fields)
	if totalInputFields > len(parser.fieldNames) {
		return fmt.Errorf("the total fields in input file is less than total fields in config")
	}

	for i := range parser.cfg.Input.Fields {
		f := &parser.cfg.Input.Fields[i]

		if f.Order == nil {
			return fmt.Errorf("unable to determine order of config field '%s' from input file", f.Name)
		}

		fieldName := parser.fieldNames[*f.Order]
		if f.Name != fieldName {
			return fmt.Errorf("config field '%s' is referencing to wrong input field '%s'", f.Name, fieldName)
		}
	}

	return nil
}

func (parser *Parser) NextBatch() (batch *Batch, exists bool, err error) {
	batch, exists, err = parser.inputParser.NextBatch(parser.cfg.Args.InputPath)
	if err != nil || !exists {
		return batch, exists, err
	}

	parser.preProcessBatchData(batch)

	return batch, exists, nil
}

func (parser *Parser) GetCurrentIndex() int {
	return parser.inputParser.GetCurrentIndex()
}

func (parser *Parser) Close() {
	if parser.inputParser != nil {
		parser.inputParser.Close()
	}
}

func (parser *Parser) preProcessBatchData(batch *Batch) {
	fields := parser.cfg.Input.Fields

	preProcessedData := [][]string{}
	for i := range batch.Data {
		data := []string{}

		for j := range fields {
			f := &fields[j]

			fieldOrder := 0
			if f.Order != nil {
				fieldOrder = *f.Order
			}

			val := batch.Data[i][fieldOrder]
			if f.TrimSpaces || parser.cfg.Input.TrimSpaces {
				val = strings.TrimSpace(val)
			}

			data = append(data, val)
		}

		preProcessedData = append(preProcessedData, data)
	}

	batch.Data = preProcessedData
}
