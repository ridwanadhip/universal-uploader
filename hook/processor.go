package hook

import "github.com/ridwanadhip/universal-uploader/config"

type ProcessorHookMetadata struct {
	TargetID   string
	TargetType string
	TargetMode string
}

type ProcessorHook interface {
	PrepareBatch(metadata *ProcessorHookMetadata, index, batchSize int) error
	CleanUpBatch(metadata *ProcessorHookMetadata, index, batchSize int, batchError error) error
	OverrideBaseFieldValue(metadata *ProcessorHookMetadata, fieldID, originalValue string) (newValue string, err error)
	OverrideFormattedFieldValue(metadata *ProcessorHookMetadata, fieldID, originalValue string) (newValue string, err error)
}

func NewProcessorHookMetadataFromTarget(target *config.Target) *ProcessorHookMetadata {
	return &ProcessorHookMetadata{
		TargetID:   target.ID,
		TargetMode: string(target.Mode),
		TargetType: string(target.Type),
	}
}
