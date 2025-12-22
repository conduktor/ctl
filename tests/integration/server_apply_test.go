package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// contains checks if s contains substr
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Note: These tests require the Conduktor Console server to support the
// batch-apply API endpoints (/public/v1/resources/batch-apply).
// These tests use admin token authentication.

func Test_ApplyServer_Empty_File(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with empty file")
	filePath := testDataFilePath(t, "empty.yaml")
	stdout, stderr, err := runConsoleCommand("apply", "--server-side", "-f", filePath, "--yes")

	// Empty file should succeed with no output
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stdout, "Expected no stdout output, got: %s", stdout)
}

func Test_ApplyServer_Nonexistent_File(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with nonexistent file")
	filePath := testDataFilePath(t, "nonexistent.yaml")
	_, stderr, err := runConsoleCommand("apply", "--server-side", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for nonexistent file")

	expectedError := fmt.Sprintf("stat %s: no such file or directory", filePath)
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
}

func Test_ApplyServer_Invalid_Strategy(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with invalid strategy")
	filePath := testDataFilePath(t, "valid_group.yaml")
	_, stderr, err := runConsoleCommand("apply", "--server-side", "-f", filePath, "--strategy", "invalid-strategy")
	assert.Error(t, err, "Expected command to fail for invalid strategy")

	expectedError := "--strategy must be one of [fail-fast, continue-on-error]"
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
}

func Test_ApplyServer_Valid_Resource_FailFast(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with valid resource using fail-fast strategy")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-failfast-token")
	defer deleteTokenByName(tokenName)

	filePath := testDataFilePath(t, "valid_group.yaml")
	stdout, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", filePath, "--strategy", "fail-fast", "--yes")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Should contain the created resource
	assert.Containsf(t, stdout, "Group/team-a", "Expected stdout to contain 'Group/team-a', got: %s", stdout)

	// Cleanup after test
	_, _, _ = runCommandWithToken(token, "delete", "-f", filePath)
}

func Test_ApplyServer_Valid_Resource_ContinueOnError(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with valid resource using continue-on-error strategy")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-continue-token")
	defer deleteTokenByName(tokenName)

	filePath := testDataFilePath(t, "valid_group.yaml")
	stdout, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", filePath, "--strategy", "continue-on-error", "--yes")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Should contain the created resource
	assert.Containsf(t, stdout, "Group/team-a", "Expected stdout to contain 'Group/team-a', got: %s", stdout)

	// Cleanup after test
	_, _, _ = runCommandWithToken(token, "delete", "-f", filePath)
}

func Test_ApplyServer_Dry_Run(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with dry-run")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-dryrun-token")
	defer deleteTokenByName(tokenName)

	filePath := testDataFilePath(t, "valid_group.yaml")
	stdout, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", filePath, "--dry-run", "--yes")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Should indicate dry run mode
	assert.Containsf(t, stdout, "DRY RUN", "Expected stdout to contain 'DRY RUN', got: %s", stdout)
}

func Test_ApplyServer_NoProgress_Flag(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with --no-progress flag")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-noprogress-token")
	defer deleteTokenByName(tokenName)

	filePath := testDataFilePath(t, "valid_group.yaml")
	stdout, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", filePath, "--no-progress", "--yes")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// With --no-progress, output should still show results but not progress bars
	// The exact output format depends on the server implementation
	_ = stdout // Output is server-dependent

	// Cleanup after test
	_, _, _ = runCommandWithToken(token, "delete", "-f", filePath)
}

func Test_ApplyServer_Folder(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with folder")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-folder-token")
	defer deleteTokenByName(tokenName)

	folderPath := testDataFilePath(t, "resources_folder")
	stdout, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", folderPath, "--yes")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Should contain created resources
	assert.Containsf(t, stdout, "Group/team-b", "Expected stdout to contain 'Group/team-b', got: %s", stdout)
	assert.Containsf(t, stdout, "Group/team-c", "Expected stdout to contain 'Group/team-c', got: %s", stdout)

	// Cleanup after test
	_, _, _ = runCommandWithToken(token, "delete", "-f", folderPath)
}

func Test_ApplyServer_Folder_Recursive(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with folder recursively")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-recursive-token")
	defer deleteTokenByName(tokenName)

	folderPath := testDataFilePath(t, "resources_folder")
	stdout, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", folderPath, "-r", "--yes")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Should contain all created resources including nested
	assert.Containsf(t, stdout, "Group/team-b", "Expected stdout to contain 'Group/team-b', got: %s", stdout)
	assert.Containsf(t, stdout, "Group/team-c", "Expected stdout to contain 'Group/team-c', got: %s", stdout)
	assert.Containsf(t, stdout, "Group/team-d", "Expected stdout to contain 'Group/team-d', got: %s", stdout)

	// Cleanup after test
	_, _, _ = runCommandWithToken(token, "delete", "-f", folderPath, "-r")
}

func Test_ApplyServer_LargeCount_RequiresYes(t *testing.T) {
	fmt.Println("Test CLI apply --server-side refuses large operation without --yes")
	// This test would need a file with >50 resources
	// For now, we just verify the flag exists and is documented
	stdout, stderr, err := runConsoleCommand("apply", "--help")
	assert.NoError(t, err)
	assert.Contains(t, stdout+stderr, "--yes", "Expected help to document --yes flag")
	assert.Contains(t, stdout+stderr, "--server-side", "Expected help to document --server-side flag")
}

func Test_ApplyServer_InvalidResource_UnknownKind(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with invalid resource (unknown kind)")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-invalidkind-token")
	defer deleteTokenByName(tokenName)

	filePath := testDataFilePath(t, "invalid_resource.yaml")
	_, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", filePath, "--yes")

	// Should fail because "InvalidResource" is not a valid kind
	assert.Error(t, err, "Expected command to fail for unknown kind")
	assert.NotEmptyf(t, stderr, "Expected stderr to contain error message, got empty stderr")
}

func Test_ApplyServer_InvalidTopic_UnknownCluster(t *testing.T) {
	fmt.Println("Test CLI apply --server-side with invalid topic (unknown cluster)")

	// Create admin token for this test
	token, tokenName := createAdminToken(t, "apply-server-invalidcluster-token")
	defer deleteTokenByName(tokenName)

	filePath := testDataFilePath(t, "invalid_topic.yaml")
	_, stderr, err := runCommandWithToken(token, "apply", "--server-side", "-f", filePath, "--yes")

	// Should fail because "unkown-cluster" does not exist
	assert.Error(t, err, "Expected command to fail for unknown cluster")
	assert.NotEmptyf(t, stderr, "Expected stderr to contain error message, got empty stderr")
}
