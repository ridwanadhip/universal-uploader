package config

import (
	"fmt"
)

type (
	ValueType  string
	TargetType string
	ConfigType string
	TargetMode string
)

const (
	// supported config types
	ConfigTypeYAML ConfigType = "yaml"

	// known value types
	ValueTypeInteger ValueType = "integer"
	ValueTypeString  ValueType = "string"
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeDecimal ValueType = "decimal"

	// known target types
	TargetTypeMySQL TargetType = "mysql"
	TargetTypeRedis TargetType = "redis"

	// known target operation mode
	TargetModeInsert TargetMode = "insert"
	TargetModeUpsert TargetMode = "upsert"
	TargetModeUpdate TargetMode = "update"
)

// constants
const (
	// defaults
	DefaultTargetType     = TargetTypeMySQL
	DefaultConfigType     = ConfigTypeYAML
	DefaultInputType      = "csv"
	DefaultOutputType     = "zip"
	DefaultHost           = "localhost"
	DefaultEnvarToken     = '$'
	DefaultReferenceToken = '^'
	DefaultBatchSize      = 250
	DefaultDelay          = 1000

	// ports
	DefaultMySQLPort = 3306
	DefaultRedisPort = 6379

	// paths
	DefaultCheckPointPath = ".checkpoint"
)

// map and list constants
var (
	validValueType = map[ValueType]bool{
		ValueTypeInteger: true,
		ValueTypeString:  true,
		ValueTypeBoolean: true,
		ValueTypeDecimal: true,
	}
)

func (val ValueType) Validate() error {
	if ok := validValueType[val]; !ok {
		return fmt.Errorf("unknown value type: %s", val)
	}

	return nil
}

// interface for data parser

type Parser struct {
	args         *Args
	configParser ConfigParser
}

type ConfigParser interface {
	Parse(path string) (*Config, error)
}

func NewParser(args *Args) (Parser, error) {
	switch ConfigType(args.ConfigType) {
	case ConfigTypeYAML:
		return Parser{args, &yamlParser{}}, nil
	}

	return Parser{}, fmt.Errorf("unknown config parser type: %s", args.ConfigType)
}

func (parser *Parser) Parse() (*Config, error) {
	cfg, err := parser.configParser.Parse(parser.args.ConfigPath)
	if err != nil {
		return nil, err
	}

	cfg.Args = parser.args
	err = cfg.SetDefaults()

	return cfg, err
}
