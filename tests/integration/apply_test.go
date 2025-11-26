package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Apply_Empty_File(t *testing.T) {
	fmt.Println("Test CLI Apply with empty file")
	filePath := testDataFilePath(t, "empty.yaml")
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath)

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)
	assert.Emptyf(t, stdout, "Expected no stdout output, got: %s", stdout)
}

func Test_Apply_Nonexistent_File(t *testing.T) {
	fmt.Println("Test CLI Apply with nonexistent file")
	filePath := testDataFilePath(t, "nonexistent.yaml")
	_, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for nonexistent file")

	expectedError := fmt.Sprintf("stat %s: no such file or directory", filePath)
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
}

func Test_Apply_Invalid_Resource(t *testing.T) {
	fmt.Println("Test CLI Apply with invalid resource")
	filePath := testDataFilePath(t, "invalid_resource.yaml")
	_, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for invalid resource")

	expectedError := "Could not apply resource InvalidResource/invalid-resource: kind InvalidResource not found"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
}

func Test_Apply_Invalid_ClientConfig(t *testing.T) {
	fmt.Println("Test CLI Apply with invalid client config for resource")
	filePath := testDataFilePath(t, "valid_group.yaml")
	UnsetCLIConsoleEnv()
	_, stderr, err := runGatewayCommand("apply", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for invalid resource")

	expectedError := "Error: failed to run apply: cannot apply ConsoleAPI resources Group: Cannot create client: Please set CDK_BASE_URL"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
	UnsetCLIGatewayEnv()
}

func Test_Apply_Failure(t *testing.T) {
	fmt.Println("Test CLI Apply with resource missing a dependency")
	filePath := testDataFilePath(t, "invalid_topic.yaml")
	_, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for invalid resource")

	expectedError := "Could not apply resource Topic/invalid-topic: Cluster unknown-cluster not found"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
}

func Test_Apply_Dry_Run(t *testing.T) {
	fmt.Println("Test CLI Apply with dry-run")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--dry-run")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)
}

func Test_Apply_Valid_Resource(t *testing.T) {
	fmt.Println("Test CLI Apply with valid resource")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath)

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_Folder_Non_Recursive(t *testing.T) {
	fmt.Println("Test CLI Apply with folder")
	folderPath := testDataFilePath(t, "resources_folder")
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath)

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedOutput := "Group/team-b: Created\nGroup/team-c: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_Folder_Recursive(t *testing.T) {
	fmt.Println("Test CLI Apply with folder recursively")
	folderPath := testDataFilePath(t, "resources_folder")
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath, "-r")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedOutput := "Group/team-b: Created\nGroup/team-c: Created\nGroup/team-d: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath, "-r")
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_Bad_Parallelism(t *testing.T) {
	fmt.Println("Test CLI Apply with bad parallelism value")
	filePath := testDataFilePath(t, "valid_group.yaml")
	_, stderr, err := runConsoleCommand("apply", "-f", filePath, "--parallelism", "0")
	assert.Error(t, err, "Expected command to fail for bad parallelism value")
	expectedError := "argument --parallelism must be between 1 and 100 (got 0)"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
}

func Test_Apply_Diff_Output(t *testing.T) {
	fmt.Println("Test CLI Apply with diff output")
	filePath := testDataFilePath(t, "valid_group.yaml")

	// First apply to create the resource
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.NoErrorf(t, err, "Initial apply command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during initial apply, got: %s", stderr)

	// Modify the resource file to trigger a diff
	modifiedFilePath := testDataFilePath(t, "valid_group_updated.yaml")
	stdout, stderr, err = runConsoleCommand("apply", "-f", modifiedFilePath, "--print-diff")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedDiffOutput := `
apiVersion: v2
kind: Group
metadata:
    name: team-a
spec:
    description: [32mUpdated [0mGroup for Team A members
    displayName: Team A[32m update[0m
Group/team-a: Updated
`
	assert.Equalf(t, expectedDiffOutput, stdout, "Expected stdout to be '%s', got: %s", expectedDiffOutput, stdout)

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

// ======================================
// State Management Tests
// ======================================

func Test_Apply_With_State_Fail_Write_Permissions(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - fail due to write permissions")
	filePath := testDataFilePath(t, "valid_group.yaml")
	tmpDir := t.TempDir()
	// remove write permissions
	err := os.Chmod(tmpDir, 0555)
	stateFile := fmt.Sprintf("%s/state.json", tmpDir)

	// Apply with state enabled should fail due to write permission error
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.Error(t, err, "Expected command to fail due to write permission error")
	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	expectedError := fmt.Sprintf("Error: could not save state, file storage error: failed to write state.\n  Cause: open %s: permission denied", stateFile)
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_With_State_Fail_Read_Permissions(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - fail due to read permissions")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// First apply with state enabled
	stdout, stderr, err := runConsoleCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Unexpected command failed: %v\nStderr: %s", err, stderr)
	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// remove read permissions
	err = os.Chmod(stateFile, 0222)
	assert.NoError(t, err, "Failed to change state file permissions")

	// Second apply should fail due to read permission error
	_, stderr, err = runConsoleCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.Error(t, err, "Expected command to fail due to read permission error")
	expectedError := fmt.Sprintf("Error: could not load state, file storage error: failed to read state file.\n  Cause: open %s: permission denied", stateFile)
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_With_State_Fail_Corrupted_State(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - fail due to corrupted state file")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// write corrupted content to state file
	err := os.WriteFile(stateFile, []byte("{invalid_json: true,"), 0644)
	assert.NoError(t, err, "Failed to write corrupted state file")

	// Apply with state enabled should fail due to corrupted state file
	_, stderr, err := runConsoleCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.Error(t, err, "Expected command to fail due to corrupted state file")
	expectedError := fmt.Sprintf("Error: failed to run apply: cannot load state from %s: failed to unmarshal state JSON", stateFile)
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
}

func Test_Apply_With_State_Fail_API_Unreachable(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - fail due to unreachable API")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// First apply with state enabled
	stdout, stderr, err := runConsoleCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Unexpected command failed: %v\nStderr: %s", err, stderr)
	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Simulate unreachable API by setting wrong api URL
	os.Setenv("CDK_BASE_URL", "http://localhost:9999")

	// Second apply should fail due to unreachable API
	_, stderr, err = RunCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.Error(t, err, "Expected command to fail due to unreachable API")
	expectedError := "Error: failed to run apply: cannot apply ConsoleAPI resources Group: Cannot create client: dial tcp [::1]:9999: connect: connection refused"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_With_State_First_Run(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - first run")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// First apply with state enabled
	stdout, stderr, err := runConsoleCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)

	assert.NoErrorf(t, err, "Unexpected command failed: %v\nStderr: %s", err, stderr)
	assert.Contains(t, stderr, fmt.Sprintf("Loading state from : Local File %s", stateFile), "Expected loading log")
	assert.Contains(t, stderr, "State file does not exist, creating a new one", "Expected new empty state log")
	assert.Contains(t, stderr, fmt.Sprintf("Saving state into : Local File %s", stateFile), "Expected saving log")

	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file was created and contains the resource
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-a", "State file should contain the resource name")
	assert.Contains(t, string(stateContent), "Group", "State file should contain the resource kind")

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_With_State_Subsequent_Run_No_Changes(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - subsequent run with no changes")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// First apply to create the resource and state
	stdout, stderr, err := runConsoleCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Initial apply command failed: %v\nStderr: %s", err, stderr)
	assert.Contains(t, stderr, fmt.Sprintf("Loading state from : Local File %s", stateFile), "Expected loading log")
	assert.Contains(t, stderr, "State file does not exist, creating a new one", "Expected new empty state log")
	assert.Contains(t, stderr, fmt.Sprintf("Saving state into : Local File %s", stateFile), "Expected saving log")

	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")

	// Second apply with the same file (should update, not create)
	stdout, stderr, err = runConsoleCommand("apply", "-v", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Second apply command failed: %v\nStderr: %s", err, stderr)
	assert.Contains(t, stderr, fmt.Sprintf("Loading state from : Local File %s", stateFile), "Expected loading log")
	// no creating new state log
	assert.Contains(t, stderr, fmt.Sprintf("Saving state into : Local File %s", stateFile), "Expected saving log")

	expectedOutput := "Group/team-a: NotChanged\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	stateUpdate, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Equalf(t, string(stateContent), string(stateUpdate), "State file should remain unchanged when no resource changes")

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
}

func Test_Apply_With_State_Resource_Deletion(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - automatic resource deletion")
	folderPath := testDataFilePath(t, "resources_folder")
	stateFile := tmpStateFilePath(t, "state.json")

	// First apply create group B.C.D from recursive resource_folder
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath, "-r", "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Initial apply command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-b: Created\nGroup/team-c: Created\nGroup/team-d: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file contains both resources
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-b", "State file should contain team-b")
	assert.Contains(t, string(stateContent), "team-c", "State file should contain team-c")
	assert.Contains(t, string(stateContent), "team-d", "State file should contain team-d")

	// Second apply remove recursion to remove nested group D from state
	stdout, stderr, err = runConsoleCommand("apply", "-f", folderPath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Second apply command failed: %v\nStderr: %s", err, stderr)

	// Should delete team-b and team-c, and create team-a
	assert.Contains(t, stderr, "Deleting resources missing from state", "Expected deletion message in stderr")
	assert.Contains(t, stdout, "Group/team-d: Deleted", "Expected team-d deletion in stdout")
	assert.Contains(t, stdout, "Group/team-b: NotChanged", "Expected team-b creation in stdout")
	assert.Contains(t, stdout, "Group/team-c: NotChanged", "Expected team-c creation in stdout")

	// Verify state file contains both resources
	stateContent, err = os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-b", "State file should contain team-b")
	assert.Contains(t, string(stateContent), "team-c", "State file should contain team-c")

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath, "-r")
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}

func Test_Apply_With_State_Dry_Run(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - dry run")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// Dry run with state enabled
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--enable-state", "--state-file", stateFile, "--dry-run")

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file was NOT created (dry-run should not persist state)
	_, err = os.Stat(stateFile)
	assert.True(t, os.IsNotExist(err), "State file should not exist after dry-run")
}

func Test_Apply_With_State_Via_Env_Var(t *testing.T) {
	fmt.Println("Test CLI Apply with state management via environment variables")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// Set environment variables
	os.Setenv("CDK_STATE_ENABLED", "true")
	os.Setenv("CDK_STATE_FILE", stateFile)
	defer func() {
		os.Unsetenv("CDK_STATE_ENABLED")
		os.Unsetenv("CDK_STATE_FILE")
	}()

	// Apply with state enabled via env vars
	stdout, stderr, err := runConsoleCommand("apply", "-v", "-f", filePath)

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Contains(t, stderr, fmt.Sprintf("Loading state from : Local File %s", stateFile), "Expected loading log")
	assert.Contains(t, stderr, "State file does not exist, creating a new one", "Expected new empty state log")
	assert.Contains(t, stderr, fmt.Sprintf("Saving state into : Local File %s", stateFile), "Expected saving log")

	expectedOutput := "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file was created
	_, err = os.Stat(stateFile)
	assert.NoError(t, err, "State file should exist")

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}

func Test_Apply_With_State_Multiple_Resources(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - multiple resources")
	folderPath := testDataFilePath(t, "resources_folder")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply multiple resources with state enabled
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath, "--enable-state", "--state-file", stateFile)

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-b: Created\nGroup/team-c: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file contains both resources
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-b", "State file should contain team-b")
	assert.Contains(t, string(stateContent), "team-c", "State file should contain team-c")

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}

func Test_Apply_With_State_Partial_Failure(t *testing.T) {
	fmt.Println("Test CLI Apply with state management - partial failure")
	validFilePath := testDataFilePath(t, "valid_group.yaml")
	invalidFilePath := testDataFilePath(t, "invalid_topic.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply both valid and invalid resources
	stdout, stderr, err := runConsoleCommand("apply", "-f", validFilePath, "-f", invalidFilePath, "--enable-state", "--state-file", stateFile)

	// Command should fail due to invalid resource
	assert.Error(t, err, "Expected command to fail due to invalid resource")

	// Valid resource should have been created
	assert.Contains(t, stdout, "Group/team-a: Created", "Valid resource should be created")

	// Invalid resource should have error
	assert.Contains(t, stderr, "Could not apply resource Topic/invalid-topic", "Expected error for invalid topic")

	// Verify state file only contains successful resource
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-a", "State file should contain successful resource")
	assert.NotContains(t, string(stateContent), "invalid-topic", "State file should not contain failed resource")

	// Cleanup after test
	stdout, stderr, err = runConsoleCommand("delete", "-f", validFilePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}

func Test_Apply_Without_State_Does_Not_Delete(t *testing.T) {
	fmt.Println("Test CLI Apply without state management - no automatic deletion")
	filePath1 := testDataFilePath(t, "valid_group.yaml")
	folderPath := testDataFilePath(t, "resources_folder")

	// First apply multiple resources without state
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath)
	assert.NoErrorf(t, err, "Initial apply command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-b: Created\nGroup/team-c: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Second apply with different resources (should NOT delete existing ones)
	stdout, stderr, err = runConsoleCommand("apply", "-f", filePath1)
	assert.NoErrorf(t, err, "Second apply command failed: %v\nStderr: %s", err, stderr)

	expectedOutput = "Group/team-a: Created\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)
	assert.NotContains(t, stdout, "Deleted", "Should not delete resources when state is disabled")

	// Cleanup after test - delete all resources
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath1)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
}
