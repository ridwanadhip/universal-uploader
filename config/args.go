package config

import (
	"flag"
	"fmt"
	"strings"
)

type Args struct {
	HelpFlag        bool
	DryRunFlag      bool
	VerboseModeFlag bool
	ResumeFlag      bool
	ConfigType      string
	ConfigPath      string
	InputPath       string
	CheckPointPath  string
}

func (args *Args) Parse() error {
	flag.BoolVar(&args.HelpFlag, "help", false, "")
	flag.BoolVar(&args.DryRunFlag, "dry-run", false, "")
	flag.BoolVar(&args.VerboseModeFlag, "verbose", false, "")
	flag.BoolVar(&args.ResumeFlag, "resume", false, "")
	flag.StringVar(&args.ConfigType, "config-type", string(ConfigTypeYAML), "")
	flag.StringVar(&args.CheckPointPath, "check-point-path", DefaultCheckPointPath, "")

	flag.Parse()
	tailArgs := flag.Args()

	if len(tailArgs) != 2 {
		return fmt.Errorf("require config and input file path")
	}

	args.ConfigPath = strings.TrimSpace(tailArgs[0])
	args.InputPath = strings.TrimSpace(tailArgs[1])

	return nil
}
