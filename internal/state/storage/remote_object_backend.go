package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/conduktor/ctl/internal/state/model"
	"github.com/go-kit/log"
	"github.com/thanos-io/objstore"
	"github.com/thanos-io/objstore/client"
)

const RemoteStateFileName = "cli-state.json"

type RemoteFileBackend struct {
	Bucket     objstore.Bucket
	ObjectName string
}

// NewRemoteFileBackend creates a new remote file backend from bucket configuration YAML.
// The bucketConfigPath should point to a YAML file containing the bucket configuration.
// Example configuration for S3:
//
//	type: S3
//	config:
//	  bucket: "my-bucket"
//	  endpoint: "s3.amazonaws.com"
//	  access_key: "xxx"
//	  secret_key: "xxx"
//	prefix: "conduktor/state/"
//
// The objectName parameter specifies the name of the state file in the bucket.
// If nil or empty, defaults to "cli-state.json".
func NewRemoteFileBackend(bucketConfigPath *string, objectName *string, debug bool) (*RemoteFileBackend, error) {
	if bucketConfigPath == nil || *bucketConfigPath == "" {
		return nil, NewStorageError(RemoteBackend, "bucket configuration path is required", nil, "Provide a valid path to a bucket configuration YAML file using the --state-backend-config flag.")
	}

	// Read bucket configuration from file
	configData, err := os.ReadFile(*bucketConfigPath)
	if err != nil {
		return nil, NewStorageError(RemoteBackend, "failed to read bucket configuration file", err, fmt.Sprintf("Ensure that the file at %s exists and is readable.", *bucketConfigPath))
	}

	// Create logger
	logger := log.NewNopLogger()
	if debug {
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	}

	// Create bucket client from configuration
	bucket, err := client.NewBucket(logger, configData, "conduktor-cli", nil)
	if err != nil {
		return nil, NewStorageError(RemoteBackend, "failed to create bucket client", err, "Check that your bucket configuration is valid and credentials are correct.")
	}

	// Determine object name
	stateObjectName := RemoteStateFileName
	if objectName != nil && *objectName != "" {
		stateObjectName = *objectName
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Remote backend initialized: bucket=%s, object=%s\n", bucket.Name(), stateObjectName)
	}

	return &RemoteFileBackend{
		Bucket:     bucket,
		ObjectName: stateObjectName,
	}, nil
}

func (b RemoteFileBackend) Type() StorageBackendType {
	return RemoteBackend
}

func (b RemoteFileBackend) LoadState(debug bool) (*model.State, error) {
	ctx := context.Background()

	// Check if state file exists
	exists, err := b.Bucket.Exists(ctx, b.ObjectName)
	if err != nil {
		return nil, NewStorageError(RemoteBackend, "failed to check if state file exists", err, "Verify your bucket permissions and network connectivity.")
	}

	if !exists {
		if debug {
			fmt.Fprintf(os.Stderr, "State file does not exist in remote storage, creating a new one\n")
		}
		return model.NewState(), nil
	}

	// Get the state file from remote storage
	reader, err := b.Bucket.Get(ctx, b.ObjectName)
	if err != nil {
		if b.Bucket.IsObjNotFoundErr(err) {
			if debug {
				fmt.Fprintf(os.Stderr, "State file not found in remote storage, creating a new one\n")
			}
			return model.NewState(), nil
		}
		return nil, NewStorageError(RemoteBackend, "failed to read state file from remote storage", err, "Verify your bucket permissions and network connectivity.")
	}
	defer reader.Close()

	// Read the entire content
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, NewStorageError(RemoteBackend, "failed to read state file content", err, "Network error or corrupted data. Try again.")
	}

	// Unmarshal JSON
	var state *model.State
	err = json.Unmarshal(data, &state)
	if err != nil {
		tip := fmt.Sprintf("The state file in remote storage may be corrupted or not in the expected format. You can try deleting the object %s from bucket %s and rerun the command to generate a new state file.", b.ObjectName, b.Bucket.Name())
		return nil, NewStorageError(RemoteBackend, "failed to unmarshal state JSON", err, tip)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Loaded state from remote storage: %d resources\n", len(state.Resources))
	}

	return state, nil
}

func (b RemoteFileBackend) SaveState(state *model.State, debug bool) error {
	ctx := context.Background()

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		tip := "Something went wrong while converting the state to JSON format. Check cause error and contact support if the issue persists."
		return NewStorageError(RemoteBackend, "failed to marshal state to JSON", err, tip)
	}

	// Upload to remote storage
	// The Upload method should be idempotent according to the objstore.Bucket interface
	reader := bytes.NewReader(data)
	err = b.Bucket.Upload(ctx, b.ObjectName, reader)
	if err != nil {
		tip := fmt.Sprintf("Failed to upload state file to bucket %s. Verify your bucket permissions and network connectivity.", b.Bucket.Name())
		return NewStorageError(RemoteBackend, "failed to upload state file to remote storage", err, tip)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Saved state to remote storage: %d resources\n", len(state.Resources))
	}

	return nil
}

func (b RemoteFileBackend) DebugString() string {
	return fmt.Sprintf("Remote Storage: bucket=%s, object=%s", b.Bucket.Name(), b.ObjectName)
}
