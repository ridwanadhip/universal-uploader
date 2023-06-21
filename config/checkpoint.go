package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"time"

	"github.com/ridwanadhip/universal-uploader/util"
)

type CheckPoint struct {
	ConfigFile string
	InputFile  string
	Timestamp  time.Time
	Error      string
	BatchSize  int
	Progress   map[string]int // map of target id to last executed data index, 0 based index
}

func (cp *CheckPoint) Save(path string, reason error) error {
	cp.Timestamp = time.Now()
	cp.Error = reason.Error()

	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	newData := util.Jsonify(cp)
	_, err = f.WriteString(newData)
	return err
}

func (cp *CheckPoint) Load(path string) (exists bool, err error) {
	if path == "" {
		path = DefaultCheckPointPath
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	err = json.Unmarshal(content, cp)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (cp *CheckPoint) IsLoaded() bool {
	return !cp.Timestamp.IsZero() || cp.Error != ""
}
