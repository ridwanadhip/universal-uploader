package main

import (
	"github.com/ridwanadhip/universal-uploader/config"
	"github.com/ridwanadhip/universal-uploader/uploader"
)

func main() {
	args := &config.Args{}
	err := args.Parse()
	if err != nil {
		panic(err)
	}

	up, err := uploader.NewUploader(args, nil)
	if err != nil {
		panic(err)
	}

	defer up.Close()

	err = up.Run()
	if err != nil {
		panic(err)
	}
}
