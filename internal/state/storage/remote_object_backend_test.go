package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseRemoteURI(t *testing.T) {
	tests := []struct {
		name           string
		uri            string
		expectedBucket string
		expectedObject string
		description    string
	}{
		{
			name:           "S3 URI with path prefix",
			uri:            "s3://my-bucket/conduktor/state",
			expectedBucket: "s3://my-bucket",
			expectedObject: "conduktor/state/cli-state.json",
			description:    "Should append cli-state.json to path",
		},
		{
			name:           "S3 URI with custom .json filename",
			uri:            "s3://my-bucket/conduktor/my-state.json",
			expectedBucket: "s3://my-bucket",
			expectedObject: "conduktor/my-state.json",
			description:    "Should use existing .json filename",
		},
		{
			name:           "S3 URI without path",
			uri:            "s3://my-bucket",
			expectedBucket: "s3://my-bucket",
			expectedObject: "cli-state.json",
			description:    "Should use cli-state.json at root",
		},
		{
			name:           "S3 URI with query params and path",
			uri:            "s3://my-bucket/path/to/state?region=us-east-1",
			expectedBucket: "s3://my-bucket?region=us-east-1",
			expectedObject: "path/to/state/cli-state.json",
			description:    "Should handle query params and append filename",
		},
		{
			name:           "S3 URI with query params and .json file",
			uri:            "s3://my-bucket/my-custom.json?region=us-east-1",
			expectedBucket: "s3://my-bucket?region=us-east-1",
			expectedObject: "my-custom.json",
			description:    "Should preserve custom .json filename with query params",
		},
		{
			name:           "GCS URI with path",
			uri:            "gs://bucket-name/path/prefix/",
			expectedBucket: "gs://bucket-name",
			expectedObject: "path/prefix/cli-state.json",
			description:    "Should work with GCS URIs",
		},
		{
			name:           "GCS URI with custom json file",
			uri:            "gs://bucket-name/my-state.json",
			expectedBucket: "gs://bucket-name",
			expectedObject: "my-state.json",
			description:    "Should preserve custom .json with GCS",
		},
		{
			name:           "Azure Blob URI with path",
			uri:            "azblob://container/path/",
			expectedBucket: "azblob://container",
			expectedObject: "path/cli-state.json",
			description:    "Should work with Azure Blob URIs",
		},
		{
			name:           "Azure Blob URI with custom json",
			uri:            "azblob://container/custom-state.json",
			expectedBucket: "azblob://container",
			expectedObject: "custom-state.json",
			description:    "Should preserve custom .json with Azure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucketURI, objectPath := parseRemoteURI(tt.uri)
			assert.Equal(t, tt.expectedBucket, bucketURI, "Bucket URI mismatch: %s", tt.description)
			assert.Equal(t, tt.expectedObject, objectPath, "Object path mismatch: %s", tt.description)
		})
	}
}
