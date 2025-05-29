package main

import (
	"testing"

	dktypes "github.com/distribworks/dkron/v4/types"
)

func TestPublishExecute(t *testing.T) {
	pa := &dktypes.ExecuteRequest{
		JobName: "testJob",
		Config: map[string]string{
			"bucket":     "bkt01",
			"key":        "obj001",
			"access_key": "2T7ORJ33KZM4LKNRA2XK",
			"secret_key": "sZFnXcfnKvAojAO84pzMONFPoDLCUpdGpdPUj2Up",
			"endpoint":   "http://10.155.31.145:7480",
			"region":     "us-east-1",
		},
	}
	s3 := &S3{}
	_, err := s3.Execute(pa, nil)
	if err != nil {
		t.Fatal(err)
	}
}
