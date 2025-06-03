package main

import (
	"github.com/sine-io/sinx/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Processor: new(FilesOutput),
	})
}
