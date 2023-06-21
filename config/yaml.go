package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ridwanadhip/universal-uploader/util"
	"gopkg.in/yaml.v3"
)

type yamlParser struct{}

func (parser *yamlParser) Parse(path string) (*Config, error) {
	res := &Config{}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return res, err
	}

	fileBytes, err := os.ReadFile(absPath)
	if err != nil {
		return res, err
	}

	fileBytes = injectEnvar(fileBytes)

	err = yaml.Unmarshal(fileBytes, res)
	return res, err
}

func injectEnvar(data []byte) []byte {
	strData := string(data)

	envars := util.FindSurroundedWords(strData, DefaultEnvarToken)
	for _, key := range envars {
		val := os.Getenv(util.RemoveToken(key))
		strData = strings.ReplaceAll(strData, key, val)
	}

	return []byte(strData)
}
