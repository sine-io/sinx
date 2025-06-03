package main

import (
	"testing"

	"github.com/sine-io/sinx/plugin"
	"github.com/sine-io/sinx/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProcess(t *testing.T) {
	now := timestamppb.Now()

	pa := &plugin.ProcessorArgs{
		Execution: types.Execution{
			StartedAt: now,
			NodeName:  "testNode",
			Output:    []byte("test"),
		},
		Config: plugin.Config{
			"forward": "false",
		},
	}

	fo := &LogOutput{}
	ex := fo.Process(pa)

	assert.Equal(t, "Output in dkron log", string(ex.Output))
}
