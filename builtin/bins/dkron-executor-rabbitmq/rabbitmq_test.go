package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sxproto "github.com/sine-io/sinx/types"
)

func TestPublishExecute(t *testing.T) {
	pa := &sxproto.ExecuteRequest{
		JobName: "testJob",
		Config: map[string]string{
			"queue": "test",
			"url":   "amqp://guest:guest@localhost:5672",
			"text":  "{\"hello\":11}",
		},
	}
	rabbitmq := &RabbitMQ{}
	_, err := rabbitmq.Execute(pa, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublishExecute_V2(t *testing.T) {
	tests := []struct {
		name        string
		args        *sxproto.ExecuteRequest
		expectedErr string
	}{

		{
			name: "No url provided",
			args: &sxproto.ExecuteRequest{
				Config: map[string]string{},
			},
			expectedErr: "RabbitMQ url is empty",
		},
		{
			name: "No queue provided",
			args: &sxproto.ExecuteRequest{
				Config: map[string]string{
					"url": "amqp://guest:guest@localhost:5672",
				},
			},
			expectedErr: "RabbitMQ queue name is empty",
		},
		{
			name: "Body and base64Body provided",
			args: &sxproto.ExecuteRequest{
				Config: map[string]string{
					"url":                "amqp://guest:guest@localhost:5672",
					"queue.name":         "test",
					"message.body":       "body",
					"message.base64Body": "base64",
				},
			},
			expectedErr: "RabbitMQ message.body and message.base64Body are both set",
		},
		{
			name: "All good",
			args: &sxproto.ExecuteRequest{
				Config: map[string]string{
					"url":          "amqp://guest:guest@localhost:5672",
					"message.body": "{\"key\":\"value\"}",
					"queue.name":   "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Spin up the rabbitmq server
			// todo setup a consumer and check if the message is sent
			r := &RabbitMQ{}
			output, err := r.Execute(tt.args, nil)
			assert.NoError(t, err)
			if tt.expectedErr != "" {
				assert.Equal(t, tt.expectedErr, output.Error)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
