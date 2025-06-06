package main

import (
	sxplugin "github.com/sine-io/sinx/plugin"
)

func main() {
	sxplugin.Serve(&sxplugin.ServeOpts{
		Processor: new(FilesOutput),
	})
}
