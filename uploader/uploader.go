package uploader

import (
	"fmt"
	"time"

	"github.com/ridwanadhip/universal-uploader/config"
	"github.com/ridwanadhip/universal-uploader/hook"
	"github.com/ridwanadhip/universal-uploader/input"
	"github.com/ridwanadhip/universal-uploader/processor"
	"github.com/ridwanadhip/universal-uploader/util"
)

type Uploader struct {
	Args        *config.Args
	Config      *config.Config
	InputParser input.Parser
	Processors  []processor.Processor
	CheckPoint  *config.CheckPoint
}

func NewUploader(args *config.Args, procHook hook.ProcessorHook) (res *Uploader, err error) {
	if args.ConfigType == "" {
		args.ConfigType = string(config.DefaultConfigType)
	}

	cfgParser, err := config.NewParser(args)
	if err != nil {
		return nil, err
	}

	cfg, err := cfgParser.Parse()
	if err != nil {
		return nil, err
	}

	inputParser, err := input.NewParser(cfg)
	if err != nil {
		return nil, err
	}

	if err := inputParser.Validate(); err != nil {
		return nil, err
	}

	procs := []processor.Processor{}
	for i := range cfg.Targets {
		targetID := cfg.Targets[i].ID
		proc, err := processor.NewProcessor(cfg, targetID, procHook)
		if err != nil {
			return nil, fmt.Errorf("[Target ID: %s] error: %s", targetID, err)
		}

		procs = append(procs, proc)
	}

	res = &Uploader{
		Args:        args,
		Config:      cfg,
		InputParser: inputParser,
		Processors:  procs,
		CheckPoint:  cfg.NewCheckPoint(),
	}

	if cfg.Args.ResumeFlag {
		_, err := res.CheckPoint.Load(args.CheckPointPath)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (up *Uploader) Run() error {
	// log here to record injected fields
	if up.Config.Args.VerboseModeFlag {
		fmt.Printf("[Config] %s\n", util.Jsonify(up.Config))
	}

	if up.CheckPoint.IsLoaded() {
		fmt.Printf("[Check Point] load last checkpoint from file %s\n", up.Config.Args.CheckPointPath)

		if up.Config.Args.VerboseModeFlag {
			fmt.Printf("[Check Point] %s\n", util.Jsonify(up.CheckPoint))
		}
	}

	for {
		batch, exists, err := up.InputParser.NextBatch()
		if err != nil {
			return err
		}

		// end of file
		if !exists {
			break
		}

		if up.Config.Args.VerboseModeFlag {
			fmt.Printf("[Config] %s\n", util.Jsonify(batch))
		}

		for _, proc := range up.Processors {
			targetID := proc.ID
			start := batch.Index + 1
			stop := batch.Index + len(batch.Data)

			// skip if already processed in previous sesssion
			// TODO: handle changed batch size value when resume
			// TODO: handle changed target file configuration
			if up.CheckPoint.IsLoaded() && batch.Index < up.CheckPoint.Progress[targetID] {
				fmt.Printf("[Target ID: %s] line %d to %d already processed in previous session\n", targetID, start, stop)
				continue
			}

			up.CheckPoint.Progress[targetID] = batch.Index

			err := proc.Process(batch.Data, batch.Index)
			if err != nil {
				cpErr := up.CheckPoint.Save(up.Config.Args.CheckPointPath, err)
				if cpErr != nil {
					fmt.Printf("[Target ID: %s] unable to save checkpoint: %s\n", targetID, cpErr)
				}

				return fmt.Errorf("[Target ID: %s] error: %s", targetID, err)
			}

			fmt.Printf("[Target ID: %s] successfully uploaded line %d to %d\n", targetID, start, stop)
		}

		if up.Config.Args.VerboseModeFlag {
			fmt.Printf("[Delay] %d ms\n", up.Config.Delay)
		}

		time.Sleep(time.Duration(up.Config.Delay) * time.Millisecond)
	}

	// TODO: implement exporting result file

	return nil
}

func (up *Uploader) Close() {
	up.InputParser.Close()

	for _, proc := range up.Processors {
		proc.Close()
	}
}
