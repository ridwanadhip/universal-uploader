package hook_example

import (
	"math/rand"

	"github.com/ridwanadhip/universal-uploader/config"
	"github.com/ridwanadhip/universal-uploader/hook"
	"github.com/ridwanadhip/universal-uploader/uploader"
)

// generate random alphanum string
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// hook implementation
type injectRandomValueHook struct{}

func (h *injectRandomValueHook) OverrideBaseFieldValue(metadata *hook.ProcessorHookMetadata, fieldID, originalValue string) (newValue string, err error) {
	if metadata.TargetID == "test" && fieldID == "col4" {
		originalValue = "RND-" + generateRandomString(8)
	}

	return originalValue, nil
}

func (h *injectRandomValueHook) OverrideFormattedFieldValue(metadata *hook.ProcessorHookMetadata, fieldID, originalValue string) (newValue string, err error) {
	return originalValue, nil
}

func (h *injectRandomValueHook) PrepareBatch(metadata *hook.ProcessorHookMetadata, index, batchSize int) error {
	return nil
}

func (h *injectRandomValueHook) CleanUpBatch(metadata *hook.ProcessorHookMetadata, index, batchSize int, batchError error) error {
	return nil
}

func (h *injectRandomValueHook) Start() error {
	return nil
}

func (h *injectRandomValueHook) Finish() error {
	return nil
}

// the script part, this example script inject a column value with random string (with hardcoded prefix) before inserting the row to mysql
func main() {
	args := &config.Args{
		ConfigPath: "./inject-random-value.yaml",
		InputPath:  "./inject-random-value.csv",
	}

	up, err := uploader.NewUploader(args, &injectRandomValueHook{})
	if err != nil {
		panic(err)
	}

	defer up.Close()

	err = up.Run()
	if err != nil {
		panic(err)
	}
}
