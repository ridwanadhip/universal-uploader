package config

import (
	"fmt"

	"github.com/ridwanadhip/universal-uploader/util"
)

type Config struct {
	Args      *Args        `yaml:"-"`
	Parser    ParserConfig `yaml:"-"`
	Input     Input
	Output    Output
	Targets   []Target
	TargetMap map[string]*Target `yaml:"-"`

	// general configurations
	BatchSize int `yaml:"batchSize"`
	Delay     int

	// unparsed data
	EnvarTokenRaw     string `yaml:"envarToken"`
	ReferenceTokenRaw string `yaml:"referenceToken"`
}

type ParserConfig struct {
	EnvarToken     rune
	ReferenceToken rune
}

type Input struct {
	Type            string
	Fields          []InputField
	TrimSpaces      bool                   `yaml:"trimSpaces"`
	InjectFields    bool                   `yaml:"-"`
	FieldsIDMap     map[string]*InputField `yaml:"-"`
	FieldsNameIDMap map[string]string      `yaml:"-"`
	FieldsIndexMap  map[string]int         `yaml:"-"`
}

type InputField struct {
	ID         string
	Name       string
	Order      *int
	TrimSpaces bool `yaml:"trimSpaces"`
}

type Target struct {
	Type              TargetType
	ID                string
	Name              string
	DataName          string `yaml:"dataName"`
	Host              string
	Port              int
	Username          string
	Password          string
	Upsert            bool
	Mode              TargetMode
	Fields            []TargetField
	InjectFields      bool                    `yaml:"-"`
	FieldsIDMap       map[string]*TargetField `yaml:"-"`
	FieldsNameIDMap   map[string]string       `yaml:"-"`
	UniqueConstraints []*TargetField          `yaml:"-"`
	OldValueReplacers []*TargetField          `yaml:"-"`
	FilterQueries     []*TargetField          `yaml:"-"`
}

type TargetField struct {
	ID              string
	Name            string
	Type            ValueType
	Value           *string
	References      []string
	UniqueValue     bool    `yaml:"uniqueValue"`
	ReplaceOldValue bool    `yaml:"replaceOldValue"`
	FilterQuery     bool    `yaml:"filterQuery"`
	EmptyAsNil      bool    `yaml:"emptyAsNil"`
	ValueIfEmpty    *string `yaml:"valueIfEmpty"`
}

type Output struct {
	Type   string
	Enable bool
}

// TODO: research library: https://github.com/creasty/defaults
func (cfg *Config) SetDefaults() error {
	err := cfg.setDefaultValues()
	if err != nil {
		return err
	}

	err = cfg.setParserDefaults()
	if err != nil {
		return err
	}

	err = cfg.setInputDefaults()
	if err != nil {
		return err
	}

	err = cfg.setOutputDefaults()
	if err != nil {
		return err
	}

	err = cfg.setTargetDefaults()
	if err != nil {
		return err
	}

	cfg.assignFieldsOrder()
	cfg.constructHelperMaps()
	cfg.constructHelperArrays()

	return nil
}

func (cfg *Config) InjectFieldsWithDefaultValue(fieldNames []string) {
	injected := false

	if cfg.Input.InjectFields {
		injected = true

		for _, f := range fieldNames {
			cfg.Input.Fields = append(cfg.Input.Fields, InputField{
				ID:   f,
				Name: f,
			})
		}
	}

	for i := range cfg.Targets {
		t := &cfg.Targets[i]

		if t.InjectFields {
			injected = true

			for j := range cfg.Input.Fields {
				inf := &cfg.Input.Fields[j]

				value, refs := constuctReferenceValue(inf.ID, cfg.Parser.ReferenceToken)
				t.Fields = append(t.Fields, TargetField{
					ID:         inf.ID,
					Name:       inf.Name,
					Type:       ValueTypeString,
					Value:      &value,
					References: refs,
				})
			}
		}
	}

	// reset generated data if new config injected
	if injected {
		cfg.assignFieldsOrder()
		cfg.constructHelperMaps()
		cfg.constructHelperArrays()
	}
}

func (cfg *Config) NewCheckPoint() *CheckPoint {
	cp := &CheckPoint{
		ConfigFile: cfg.Args.ConfigPath,
		InputFile:  cfg.Args.InputPath,
		BatchSize:  cfg.BatchSize,
		Progress:   map[string]int{},
	}

	for i := range cfg.Targets {
		cp.Progress[cfg.Targets[i].ID] = 0
	}

	return cp
}

func (cfg *Config) setDefaultValues() error {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = DefaultBatchSize
	}

	if cfg.Delay == 0 {
		cfg.Delay = DefaultDelay
	}

	return nil
}

func (cfg *Config) setParserDefaults() error {
	cfg.Parser = ParserConfig{
		EnvarToken:     DefaultEnvarToken,
		ReferenceToken: DefaultReferenceToken,
	}

	if cfg.EnvarTokenRaw != "" {
		if len(cfg.EnvarTokenRaw) > 1 {
			return fmt.Errorf("envar token must be a single character")
		}

		cfg.Parser.EnvarToken = []rune(cfg.EnvarTokenRaw)[0]
	}

	if cfg.ReferenceTokenRaw != "" {
		if len(cfg.ReferenceTokenRaw) > 1 {
			return fmt.Errorf("reference token must be a single character")
		}

		cfg.Parser.ReferenceToken = []rune(cfg.ReferenceTokenRaw)[0]
	}

	return nil
}

func (cfg *Config) setInputDefaults() error {
	if cfg.Input.Type == "" {
		cfg.Input.Type = DefaultInputType
	}

	for i := range cfg.Input.Fields {
		f := &cfg.Input.Fields[i]

		if f.Name == "" {
			return fmt.Errorf("input field name is required")
		}

		if f.ID == "" {
			f.ID = f.Name
		}
	}

	if len(cfg.Input.Fields) == 0 {
		cfg.Input.InjectFields = true
	}

	return nil
}

func (cfg *Config) setOutputDefaults() error {
	if cfg.Output.Type == "" {
		cfg.Output.Type = DefaultOutputType
	}

	return nil
}

func (cfg *Config) setTargetDefaults() error {
	for i := range cfg.Targets {
		t := &cfg.Targets[i]

		if t.Type == "" {
			t.Type = DefaultTargetType
		}

		if t.Name == "" {
			return fmt.Errorf("target name is required")
		}

		if t.ID == "" {
			t.ID = t.Name
		}

		if t.Host == "" {
			t.Host = DefaultHost
		}

		if t.Port == 0 {
			t.Port = getDefaultPort(t.Type)
		}

		if t.Mode == "" {
			t.Mode = TargetModeInsert
		}

		for j := range t.Fields {
			f := &t.Fields[j]

			if f.Name == "" {
				return fmt.Errorf("target field name is required")
			}

			if f.ID == "" {
				f.ID = f.Name
			}

			if f.Type == "" {
				f.Type = ValueTypeString
			}

			if err := f.Type.Validate(); err != nil {
				return err
			}

			if f.UniqueValue && f.ReplaceOldValue {
				f.ReplaceOldValue = false
			}

			if f.Value == nil {
				token := string(cfg.Parser.ReferenceToken)
				value := token + f.ID + token
				f.Value = &value
			}

			f.References = util.FindSurroundedWords(*f.Value, cfg.Parser.ReferenceToken)
		}

		if len(t.Fields) == 0 {
			t.InjectFields = true
		}
	}

	return nil
}

// order start from 1
func (cfg *Config) assignFieldsOrder() {
	existingOrder := map[int]bool{}
	for i := range cfg.Input.Fields {
		f := &cfg.Input.Fields[i]

		if f.Order != nil {
			existingOrder[*f.Order] = true
		}
	}

	newOrder := -1
	for i := range cfg.Input.Fields {
		f := &cfg.Input.Fields[i]

		if f.Order == nil {
			for {
				newOrder += 1
				if exists := existingOrder[newOrder]; !exists {
					break
				}
			}

			order := newOrder // prevent race condition
			f.Order = &order
			existingOrder[order] = true
		}
	}
}

func (cfg *Config) constructHelperMaps() {
	cfg.Input.FieldsIDMap = map[string]*InputField{}
	cfg.Input.FieldsNameIDMap = map[string]string{}
	cfg.Input.FieldsIndexMap = map[string]int{}
	cfg.TargetMap = map[string]*Target{}

	for i := range cfg.Input.Fields {
		f := &cfg.Input.Fields[i]
		cfg.Input.FieldsIDMap[f.ID] = f
		cfg.Input.FieldsNameIDMap[f.Name] = f.ID
		cfg.Input.FieldsIndexMap[f.ID] = i
	}

	for i := range cfg.Targets {
		t := &cfg.Targets[i]
		t.FieldsIDMap = map[string]*TargetField{}
		t.FieldsNameIDMap = map[string]string{}

		for j := range t.Fields {
			f := &t.Fields[j]
			t.FieldsIDMap[f.ID] = f
			t.FieldsNameIDMap[f.Name] = f.ID
		}
	}

	for i := range cfg.Targets {
		t := &cfg.Targets[i]
		cfg.TargetMap[t.ID] = t
	}
}

func (cfg *Config) constructHelperArrays() {
	for i := range cfg.Targets {
		t := &cfg.Targets[i]

		t.UniqueConstraints = []*TargetField{}
		t.OldValueReplacers = []*TargetField{}
		t.FilterQueries = []*TargetField{}

		for j := range t.Fields {
			f := &t.Fields[j]

			if f.UniqueValue {
				t.UniqueConstraints = append(t.UniqueConstraints, f)
			}

			if f.ReplaceOldValue {
				t.OldValueReplacers = append(t.OldValueReplacers, f)
			}

			if f.FilterQuery {
				t.FilterQueries = append(t.FilterQueries, f)
			}
		}
	}
}

func getDefaultPort(targetType TargetType) int {
	switch TargetType(targetType) {
	case TargetTypeMySQL:
		return DefaultMySQLPort
	case TargetTypeRedis:
		return DefaultRedisPort
	}

	return 0
}

func constuctReferenceValue(original string, token rune) (value string, references []string) {
	refToken := string(token)
	value = refToken + original + refToken
	references = util.FindSurroundedWords(value, token)
	return value, references
}
