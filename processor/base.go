package processor

import (
	"fmt"

	"github.com/ridwanadhip/universal-uploader/config"
	"github.com/ridwanadhip/universal-uploader/hook"
)

type (
	ProcessorType string
	Function      string
)

// supported types
const (
	ProcessorTypeMySQL ProcessorType = "mysql"
	ProcessorTypeRedis ProcessorType = "redis"
)

// supported functions
const (
	NilValue         Function = "NIL"
	CurrentTimestamp Function = "CURRENT_TIMESTAMP"
)

type Processor struct {
	ID       string
	cfg      *config.Config
	target   *config.Target
	impl     Implementation
	procHook hook.ProcessorHook
}

type Implementation interface {
	Process(data [][]string) error
	DryRun(data [][]string) error
	Close()
}

func NewProcessor(cfg *config.Config, id string, procHook hook.ProcessorHook) (processor Processor, err error) {
	target, exists := cfg.TargetMap[id]
	if !exists {
		return processor, fmt.Errorf("unknown target id: %s", id)
	}

	// TODO: decouple hook processing from each processor implementation
	var impl Implementation
	switch ProcessorType(target.Type) {
	case ProcessorTypeMySQL:
		impl, err = NewMySQLImplementation(&cfg.Input, target, cfg.Args.VerboseModeFlag, procHook)
	case ProcessorTypeRedis:
		impl, err = NewRedisImplementation(&cfg.Input, target, cfg.Args.VerboseModeFlag)
	default:
		err = fmt.Errorf("unknown processor implementation type: %s", cfg.Input.Type)
	}

	if err != nil {
		return Processor{}, err
	}

	return Processor{id, cfg, target, impl, procHook}, nil
}

func (proc *Processor) Process(data [][]string, index int) error {
	md := hook.NewProcessorHookMetadataFromTarget(proc.target)

	// perform batch prepartion here via hook
	if proc.procHook != nil {
		if err := proc.procHook.PrepareBatch(md, index, len(data)); err != nil {
			return err
		}
	}

	err := proc.impl.Process(data)

	// perform batch clean up here via hook
	if proc.procHook != nil {
		if hookErr := proc.procHook.CleanUpBatch(md, index, len(data), err); hookErr != nil {
			return hookErr
		}
	}

	return err
}

func (proc *Processor) DryRun(data [][]string) error {
	return proc.impl.DryRun(data)
}

func (proc *Processor) Close() {
	if proc.impl != nil {
		proc.impl.Close()
	}
}
