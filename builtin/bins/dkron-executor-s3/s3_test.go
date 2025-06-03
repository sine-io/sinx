package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	dktypes "github.com/sine-io/sinx/types"
)

func TestPublishExecute(t *testing.T) {
	pa := &dktypes.ExecuteRequest{
		JobName: "testJob",
		Config: map[string]string{
			"bucket":     "bkt01",
			"key":        fmt.Sprintf("TestPublishExecute%s", uuid.NewString()),
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

func TestPublishExecuteConcurrent(t *testing.T) {

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s3 := &S3{}

			pa := &dktypes.ExecuteRequest{
				JobName: "testJob",
				Config: map[string]string{
					"bucket":     "bkt01",
					"key":        fmt.Sprintf("TestPublishExecuteConcurrent%s", uuid.NewString()),
					"access_key": "2T7ORJ33KZM4LKNRA2XK",
					"secret_key": "sZFnXcfnKvAojAO84pzMONFPoDLCUpdGpdPUj2Up",
					"endpoint":   "http://10.155.31.145:7480",
					"region":     "us-east-1",
				},
			}
			resp, err := s3.Execute(pa, nil)
			if err != nil {
				t.Error(err)
			}

			fmt.Printf("No.: %d, resp: %v\n", i+1, resp)
		}()
	}
	wg.Wait()
}
