package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/armon/circbuf"
	dkplugin "github.com/distribworks/dkron/v4/plugin"
	dktypes "github.com/distribworks/dkron/v4/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	// maxBufSize limits how much data we collect from a handler
	maxBufSize = 256000
)

// S3 plugin uploads files to Amazon S3
type S3 struct{}

// Execute Process method of the plugin
// "executor": "s3",
//
//	"executor_config": {
//	    "bucket": "my-bucket",                // S3 bucket name
//	    "key": "path/to/object.txt",          // S3 object key (path)
//	    "region": "us-east-1",                // AWS region
//	    "access_key": "ACCESS_KEY",           // AWS access key ID
//	    "secret_key": "SECRET_KEY",           // AWS secret access key
//	    "endpoint": "https://custom-s3.com",  // Custom S3 endpoint (optional)
//	}
func (s *S3) Execute(args *dktypes.ExecuteRequest, cb dkplugin.StatusHelper) (*dktypes.ExecuteResponse, error) {
	out, err := s.ExecuteImpl(args)
	resp := &dktypes.ExecuteResponse{Output: out}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

// ExecuteImpl uploads a file to Amazon S3
func (s *S3) ExecuteImpl(args *dktypes.ExecuteRequest) ([]byte, error) {
	output, _ := circbuf.NewBuffer(maxBufSize)

	// Parse config
	bucket := args.Config["bucket"]
	key := args.Config["key"]
	region := args.Config["region"]
	accessKey := args.Config["access_key"]
	secretKey := args.Config["secret_key"]
	endpoint := args.Config["endpoint"]

	// Validate required parameters
	if bucket == "" {
		return output.Bytes(), errors.New("bucket is required")
	}
	if key == "" {
		return output.Bytes(), errors.New("key is required")
	}

	// Set defaults
	if region == "" {
		region = "us-east-1"
	}

	// Configure AWS client
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	// Add custom endpoint if specified
	if endpoint != "" {
		opts = append(opts, config.WithBaseEndpoint(endpoint))
	} else {
		return output.Bytes(), errors.New("endpoint is required")
	}

	// Add credentials if specified
	if accessKey != "" && secretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	} else {
		return output.Bytes(), errors.New("access_key and secret_key are required")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return output.Bytes(), fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	fileContent := make([]byte, 32)
	rand.Read(fileContent) // Simulate file content with random bytes
	// Prepare upload input
	uploadInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileContent), // Placeholder for file content
		ContentType: aws.String("application/octet-stream"),
	}

	// Upload to S3
	start := time.Now()
	_, err = client.PutObject(context.TODO(), uploadInput)
	elapsed := time.Since(start)

	if err != nil {
		errMsg := fmt.Sprintf("failed to upload to S3: %v", err)
		output.Write([]byte(errMsg))
		return output.Bytes(), err
	}

	successMsg := fmt.Sprintf("Successfully uploaded %d bytes to s3://%s/%s in %s",
		len(fileContent), bucket, key, elapsed)
	output.Write([]byte(successMsg))

	return output.Bytes(), nil
}
