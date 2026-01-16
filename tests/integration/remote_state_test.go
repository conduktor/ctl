package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/s3blob"
	"gopkg.in/yaml.v3"
)

const (
	minioEndpoint   = "localhost:9000"
	minioAccessKey  = "minioadmin"
	minioSecretKey  = "minioadmin"
	minioBucket     = "conduktor-state"
	minioRegion     = "us-east-1"
	testStatePrefix = "integration-test"
)

// Test apply with remote state backend using MinIO
func Test_Apply_With_Remote_State_MinIO(t *testing.T) {
	fmt.Println("Test CLI Apply with remote state backend (MinIO)")

	// Generate unique state path for this test
	statePath := fmt.Sprintf("%s/test-apply-%d.json", testStatePrefix, os.Getpid())
	stateURI := buildMinioURI(statePath)

	// Clean up any existing state file before test
	defer cleanupMinioState(t, statePath)
	cleanupMinioState(t, statePath)

	// Create test resources
	userName, userYAML := FixtureRandomConsoleUser(t)
	groupName, groupYAML := FixtureRandomConsoleGroup(t)

	// Write resources to temporary file
	tmpFile := writeTempYAMLFile(t, []any{userYAML, groupYAML})
	defer os.Remove(tmpFile)

	// Apply with remote state
	SetCLIConsoleEnv()
	stdout, stderr, err := RunCommand("apply", "-f", tmpFile, "--enable-state", "--state-remote-uri", stateURI)

	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)
	assert.Containsf(t, stdout, "User/"+userName, "Expected user to be created")
	assert.Containsf(t, stdout, "Group/"+groupName, "Expected group to be created")
	assert.Containsf(t, stderr, "Saving state into : Remote Storage", "Expected state to be saved to remote storage")

	// Verify state file exists in MinIO
	stateExists := checkMinioStateExists(t, statePath)
	assert.True(t, stateExists, "State file should exist in MinIO")

	// Read state file from MinIO and verify it contains our resources
	stateContent := readMinioState(t, statePath)
	assert.Containsf(t, stateContent, userName, "State should contain user name")
	assert.Containsf(t, stateContent, groupName, "State should contain group name")

	// Cleanup resources
	stdout, stderr, err = RunCommand("delete", "-f", tmpFile)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}

// Test delete with remote state backend
func Test_Delete_With_Remote_State_MinIO(t *testing.T) {
	fmt.Println("Test CLI Delete with remote state backend (MinIO)")

	// Generate unique state path for this test
	statePath := fmt.Sprintf("%s/test-delete-%d.json", testStatePrefix, os.Getpid())
	stateURI := buildMinioURI(statePath)

	// Clean up any existing state file
	defer cleanupMinioState(t, statePath)
	cleanupMinioState(t, statePath)

	// Create test resources
	groupName, groupYAML := FixtureRandomConsoleGroup(t)

	// Write resource to temporary file
	tmpFile := writeTempYAMLFile(t, []any{groupYAML})
	defer os.Remove(tmpFile)

	// Apply with remote state
	SetCLIConsoleEnv()
	stdout, stderr, err := RunCommand("apply", "-f", tmpFile, "--enable-state", "--state-remote-uri", stateURI)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)
	assert.Containsf(t, stdout, "Group/"+groupName, "Expected group to be created")

	// Verify state file exists
	assert.True(t, checkMinioStateExists(t, statePath), "State file should exist after apply")

	// Delete with remote state
	stdout, stderr, err = RunCommand("delete", "-f", tmpFile, "--enable-state", "--state-remote-uri", stateURI)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)
	assert.Containsf(t, stdout, "Group/"+groupName, "Expected group to be deleted")

	// State file should still exist but should be updated
	assert.True(t, checkMinioStateExists(t, statePath), "State file should still exist after delete")

	// Read state and verify resource is no longer tracked
	stateContent := readMinioState(t, statePath)
	assert.NotContainsf(t, stateContent, groupName, "State should not contain deleted group name")
}

// Test automatic cleanup of removed resources with remote state
func Test_Apply_Removed_Resources_With_Remote_State_MinIO(t *testing.T) {
	fmt.Println("Test CLI Apply with removed resources using remote state (MinIO)")

	// Generate unique state path for this test
	statePath := fmt.Sprintf("%s/test-removed-%d.json", testStatePrefix, os.Getpid())
	stateURI := buildMinioURI(statePath)

	// Clean up
	defer cleanupMinioState(t, statePath)
	cleanupMinioState(t, statePath)

	// Create two resources
	user1Name, user1YAML := FixtureRandomConsoleUser(t)
	user2Name, user2YAML := FixtureRandomConsoleUser(t)

	// First apply: both resources
	tmpFile1 := writeTempYAMLFile(t, []any{user1YAML, user2YAML})
	defer os.Remove(tmpFile1)

	SetCLIConsoleEnv()
	stdout, stderr, err := RunCommand("apply", "-f", tmpFile1, "--enable-state", "--state-remote-uri", stateURI)
	assert.NoErrorf(t, err, "First apply failed: %v\nStderr: %s", err, stderr)
	assert.Containsf(t, stdout, "User/"+user1Name, "Expected user1 to be created")
	assert.Containsf(t, stdout, "User/"+user2Name, "Expected user2 to be created")

	// Second apply: only first resource (second should be auto-deleted)
	tmpFile2 := writeTempYAMLFile(t, []any{user1YAML})
	defer os.Remove(tmpFile2)

	stdout, stderr, err = RunCommand("apply", "-f", tmpFile2, "--enable-state", "--state-remote-uri", stateURI)
	assert.NoErrorf(t, err, "Second apply failed: %v\nStderr: %s", err, stderr)

	// Should show that user2 was deleted automatically
	assert.Containsf(t, stderr, "Deleting removed resources", "Expected automatic deletion message")
	assert.Containsf(t, stdout, "User/"+user2Name, "Expected user2 to be auto-deleted")

	// Verify only user1 remains in state
	stateContent := readMinioState(t, statePath)
	assert.Containsf(t, stateContent, user1Name, "State should still contain user1")
	assert.NotContainsf(t, stateContent, user2Name, "State should not contain removed user2")

	// Cleanup remaining resource
	stdout, stderr, err = RunCommand("delete", "-f", tmpFile2)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}

// Test with custom .json file name in URI
func Test_Apply_With_Custom_State_Filename(t *testing.T) {
	fmt.Println("Test CLI Apply with custom state filename in URI")

	// Use a custom filename ending in .json
	statePath := fmt.Sprintf("%s/my-custom-state-%d.json", testStatePrefix, os.Getpid())
	stateURI := buildMinioURI(statePath)

	// Clean up
	defer cleanupMinioState(t, statePath)
	cleanupMinioState(t, statePath)

	// Create test resource
	groupName, groupYAML := FixtureRandomConsoleGroup(t)
	tmpFile := writeTempYAMLFile(t, []any{groupYAML})
	defer os.Remove(tmpFile)

	// Apply with custom state filename
	SetCLIConsoleEnv()
	stdout, stderr, err := RunCommand("apply", "-f", tmpFile, "--enable-state", "--state-remote-uri", stateURI)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)
	assert.Containsf(t, stdout, "Group/"+groupName, "Expected group to be created")

	// Verify state file exists with exact custom name
	assert.True(t, checkMinioStateExists(t, statePath), "State file should exist with custom name")

	// Cleanup
	stdout, stderr, err = RunCommand("delete", "-f", tmpFile)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}

// Helper functions

func buildMinioURI(statePath string) string {
	// Build S3-compatible URI for MinIO
	return fmt.Sprintf("s3://%s/%s?region=%s&endpoint=http://%s&disable_https=true&s3ForcePathStyle=true",
		minioBucket, statePath, minioRegion, minioEndpoint)
}

func checkMinioStateExists(t *testing.T, statePath string) bool {
	ctx := context.Background()
	bucketURL := fmt.Sprintf("s3://%s?region=%s&endpoint=http://%s&disable_https=true&s3ForcePathStyle=true",
		minioBucket, minioRegion, minioEndpoint)

	// Set AWS credentials for MinIO
	os.Setenv("AWS_ACCESS_KEY_ID", minioAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", minioSecretKey)
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")
	defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		t.Logf("Failed to open bucket: %v", err)
		return false
	}
	defer bucket.Close()

	exists, err := bucket.Exists(ctx, statePath)
	if err != nil {
		t.Logf("Failed to check if state exists: %v", err)
		return false
	}

	return exists
}

func readMinioState(t *testing.T, statePath string) string {
	ctx := context.Background()
	bucketURL := fmt.Sprintf("s3://%s?region=%s&endpoint=http://%s&disable_https=true&s3ForcePathStyle=true",
		minioBucket, minioRegion, minioEndpoint)

	// Set AWS credentials for MinIO
	os.Setenv("AWS_ACCESS_KEY_ID", minioAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", minioSecretKey)
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")
	defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	require.NoError(t, err, "Failed to open bucket")
	defer bucket.Close()

	reader, err := bucket.NewReader(ctx, statePath, nil)
	require.NoError(t, err, "Failed to read state file")
	defer reader.Close()

	content, err := io.ReadAll(reader)
	require.NoError(t, err, "Failed to read state content")

	return string(content)
}

func cleanupMinioState(t *testing.T, statePath string) {
	ctx := context.Background()
	bucketURL := fmt.Sprintf("s3://%s?region=%s&endpoint=http://%s&disable_https=true&s3ForcePathStyle=true",
		minioBucket, minioRegion, minioEndpoint)

	// Set AWS credentials for MinIO
	os.Setenv("AWS_ACCESS_KEY_ID", minioAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", minioSecretKey)
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")
	defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		t.Logf("Failed to open bucket for cleanup: %v", err)
		return
	}
	defer bucket.Close()

	// Check if file exists before attempting to delete
	exists, err := bucket.Exists(ctx, statePath)
	if err != nil || !exists {
		return
	}

	err = bucket.Delete(ctx, statePath)
	if err != nil {
		t.Logf("Failed to cleanup state file: %v", err)
	}
}

func writeTempYAMLFile(t *testing.T, resources []any) string {
	tmpFile, err := os.CreateTemp("", "conduktor-test-*.yaml")
	require.NoError(t, err, "Failed to create temp file")

	for i, resource := range resources {
		if i > 0 {
			_, err = tmpFile.WriteString("\n---\n")
			require.NoError(t, err, "Failed to write YAML separator")
		}

		// Marshal the resource to YAML
		encoder := yaml.NewEncoder(tmpFile)
		encoder.SetIndent(2)
		err = encoder.Encode(resource)
		require.NoError(t, err, "Failed to encode YAML")
		encoder.Close()
	}

	tmpFile.Close()
	return tmpFile.Name()
}
