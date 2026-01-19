package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/conduktor/ctl/internal/state/model"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

const RemoteStateFileName = "cli-state.json"

type RemoteFileBackend struct {
	Bucket     *blob.Bucket
	BucketURI  string
	ObjectPath string
}

// NewRemoteFileBackend creates a new remote file backend from a bucket URI.
// The URI format depends on the provider:
//
//	S3:    s3://bucket-name/path/prefix?region=us-east-1
//	GCS:   gs://bucket-name/path/prefix
//	Azure: azblob://container-name/path/prefix
//
// The state file will be stored at: <bucket>/<path/prefix>/cli-state.json
// The remoteURI parameter must not be nil or empty.
//
// Authentication is handled through provider-specific mechanisms:
//   - S3: AWS credentials (env vars, IAM role, ~/.aws/credentials)
//   - GCS: Google Application Default Credentials or GOOGLE_APPLICATION_CREDENTIALS
//   - Azure: AZURE_STORAGE_ACCOUNT + AZURE_STORAGE_KEY or AZURE_STORAGE_SAS_TOKEN or managed identity
func NewRemoteFileBackend(remoteURI string, debug bool) (*RemoteFileBackend, error) {
	if remoteURI == "" {
		return nil, NewStorageError(RemoteBackend, "remote URI is required", nil, "Remote URI cannot be empty. This should be provided through StorageConfig.")
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Opening remote bucket with URI: %s\n", remoteURI)
	}

	bucketURI, objectPath := parseRemoteURI(remoteURI)

	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, bucketURI)
	if err != nil {
		return nil, NewStorageError(RemoteBackend, "failed to open bucket", err, fmt.Sprintf("Ensure the URI is valid and credentials are configured. URI: %s", bucketURI))
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Remote backend initialized: URI=%s, object path=%s\n", bucketURI, objectPath)
	}

	return &RemoteFileBackend{
		Bucket:     bucket,
		BucketURI:  bucketURI,
		ObjectPath: objectPath,
	}, nil
}

// parseRemoteURI extracts the bucket URI and object path from the full URI.
// Example: s3://bucket/path/to/state -> (s3://bucket, path/to/state/cli-state.json).
// Example: s3://bucket/path?region=us-east-1 -> (s3://bucket?region=us-east-1, path/cli-state.json).
func parseRemoteURI(uri string) (string, string) {
	// Find the scheme (e.g., s3://, gs://, azblob://)
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		// Invalid URI, return as-is
		return uri, RemoteStateFileName
	}

	scheme := parts[0]
	remainder := parts[1]

	// Split remainder into bucket and path
	pathParts := strings.SplitN(remainder, "/", 2)
	bucketNameWithQuery := pathParts[0]

	var pathPrefix string
	if len(pathParts) > 1 {
		pathPrefix = pathParts[1]
	}

	// Extract query parameters from path if present
	var queryParams string
	if strings.Contains(pathPrefix, "?") {
		queryIdx := strings.Index(pathPrefix, "?")
		queryParams = pathPrefix[queryIdx:]
		pathPrefix = pathPrefix[:queryIdx]
	}

	// If query params weren't in the path, they might be in the bucket name
	if queryParams == "" && strings.Contains(bucketNameWithQuery, "?") {
		queryIdx := strings.Index(bucketNameWithQuery, "?")
		queryParams = bucketNameWithQuery[queryIdx:]
		bucketNameWithQuery = bucketNameWithQuery[:queryIdx]
	}

	// Remove trailing slash from path
	pathPrefix = strings.TrimSuffix(pathPrefix, "/")

	// Reconstruct bucket URI (scheme://bucket with any query params)
	bucketURI := fmt.Sprintf("%s://%s%s", scheme, bucketNameWithQuery, queryParams)

	// Construct object path
	var objectPath string
	if pathPrefix != "" {
		// If the path already ends with .json, use it as-is
		if strings.HasSuffix(pathPrefix, ".json") {
			objectPath = pathPrefix
		} else {
			objectPath = path.Join(pathPrefix, RemoteStateFileName)
		}
	} else {
		objectPath = RemoteStateFileName
	}

	return bucketURI, objectPath
}

func (b RemoteFileBackend) Type() StorageBackendType {
	return RemoteBackend
}

func (b RemoteFileBackend) LoadState(debug bool) (*model.State, error) {
	ctx := context.Background()

	// Check if state file exists
	exists, err := b.Bucket.Exists(ctx, b.ObjectPath)
	if err != nil {
		return nil, NewStorageError(RemoteBackend, "failed to check if state file exists", err, "Verify your bucket permissions and network connectivity.")
	}

	if !exists {
		if debug {
			fmt.Fprintf(os.Stderr, "State file does not exist in remote storage, creating a new one\n")
		}
		return model.NewState(), nil
	}

	// Read the state file from remote storage
	reader, err := b.Bucket.NewReader(ctx, b.ObjectPath, nil)
	if err != nil {
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
		tip := fmt.Sprintf("The state file in remote storage may be corrupted or not in the expected format. You can try deleting the object %s and rerun the command to generate a new state file or restore a previous version.", b.ObjectPath)
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

	// Write to remote storage
	writer, err := b.Bucket.NewWriter(ctx, b.ObjectPath, nil)
	if err != nil {
		return NewStorageError(RemoteBackend, "failed to create writer for remote storage", err, "Verify your bucket permissions and network connectivity.")
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return NewStorageError(RemoteBackend, "failed to write state data to remote storage", err, "Network error or insufficient storage. Try again.")
	}

	// Close the writer (this actually uploads the data)
	err = writer.Close()
	if err != nil {
		return NewStorageError(RemoteBackend, "failed to finalize upload to remote storage", err, "Network error during upload finalization. Try again.")
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Saved state to remote storage: %d resources\n", len(state.Resources))
	}

	return nil
}

func (b RemoteFileBackend) DebugString() string {
	return fmt.Sprintf("remote storage: URI=%s, object=%s", b.BucketURI, b.ObjectPath)
}

// Close closes the bucket connection.
func (b RemoteFileBackend) Close() error {
	if b.Bucket != nil {
		return b.Bucket.Close()
	}
	return nil
}
