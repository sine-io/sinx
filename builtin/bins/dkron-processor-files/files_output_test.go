package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	sxplugin "github.com/sine-io/sinx/plugin"
	sxproto "github.com/sine-io/sinx/types"
)

func TestProcess(t *testing.T) {
	now := timestamppb.Now()

	pa := &sxplugin.ProcessorArgs{
		Execution: sxproto.Execution{
			StartedAt: now,
			NodeName:  "testNode",
			Output:    []byte("test"),
		},
		Config: sxplugin.Config{
			"forward": "false",
			"log_dir": "/tmp",
		},
	}

	fo := &FilesOutput{}
	ex := fo.Process(pa)

	assert.Equal(t, fmt.Sprintf("/tmp/%s.log", ex.Key()), string(ex.Output))
}
