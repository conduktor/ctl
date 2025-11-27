package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Delete_Empty_File(t *testing.T) {
	fmt.Println("Test CLI Delete with empty file")
	filePath := testDataFilePath(t, "empty.yaml")
	stdout, stderr, err := runConsoleCommand("delete", "-f", filePath)

	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)
	assert.Emptyf(t, stdout, "Expected no stdout output, got: %s", stdout)
}

func Test_Delete_Nonexistent_File(t *testing.T) {
	fmt.Println("Test CLI Delete with nonexistent file")
	filePath := testDataFilePath(t, "nonexistent.yaml")
	_, stderr, err := runConsoleCommand("delete", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for nonexistent file")

	expectedError := fmt.Sprintf("stat %s: no such file or directory", filePath)
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
}

func Test_Delete_Invalid_Resource(t *testing.T) {
	fmt.Println("Test CLI Delete with invalid resource")
	filePath := testDataFilePath(t, "invalid_resource.yaml")
	_, stderr, err := runConsoleCommand("delete", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for invalid resource")

	expectedError := "Could not delete resource InvalidResource/invalid-resource: kind InvalidResource not found"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
}

func Test_Relete_Invalid_ClientConfig(t *testing.T) {
	fmt.Println("Test CLI Delete with invalid resource due to invalid client config")
	filePath := testDataFilePath(t, "valid_group.yaml")

	UnsetCLIConsoleEnv()
	_, stderr, err := runGatewayCommand("delete", "-f", filePath)
	assert.Error(t, err, "Expected command to fail for invalid resource")

	expectedError := "Error: fail to delete: cannot delete Console API resource Group/team-a: Cannot create client: Please set CDK_BASE_URL"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)
	UnsetCLIConsoleEnv()
}

func Test_Delete_Failure(t *testing.T) {
	fmt.Println("Test CLI Delete failed due to  used as a dependency by another resource")
	filePath := testDataFilePath(t, "self-serve.yaml")

	//init the resource to delete (create cluster and topic)
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Try to delete the application which is used as a dependency by the application instance
	stdout, stderr, err = runConsoleCommand("delete", "Application", "clickstream-app")
	assert.Error(t, err, "Expected command to fail when deleting resource used as a dependency")
	expectedError := "Cannot delete clickstream-app because it has instances"
	assert.NotEmptyf(t, stderr, "Expected stderr to contain '%s', got empty stderr", expectedError)
	assert.Containsf(t, stderr, expectedError, "Expected stderr to contain '%s', got: %s", expectedError, stderr)

	// Cleanup : delete the application instance, the application, the cluster and the group
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
	expectedOutput := "ApplicationInstance/clickstream-app-instance: Deleted\nApplication/clickstream-app: Deleted\nKafkaCluster/kafka-cluster: Deleted\nGroup/app-team: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)
}

func Test_Delete_Valid_Resource(t *testing.T) {
	fmt.Println("Test CLI Delete with valid resource")

	//Init the resource to delete
	filePath := testDataFilePath(t, "valid_group.yaml")
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Delete
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during cleanup, got: %s", stderr)
	expectedOutput := "Group/team-a: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)
}

func Test_Delete_Folder_Non_Recursive(t *testing.T) {
	fmt.Println("Test CLI Apply with folder")
	folderPath := testDataFilePath(t, "resources_folder")

	//Init the resources to delete
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath)
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Test delete non-recursive
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath)
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedOutput := "Group/team-b: Deleted\nGroup/team-c: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)
}

func Test_Delete_Folder_Recursive(t *testing.T) {
	fmt.Println("Test CLI Apply with folder recursively")
	folderPath := testDataFilePath(t, "resources_folder")

	//Init the resources to delete
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath, "-r")
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Test delete recursive
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath, "-r")
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedOutput := "Group/team-b: Deleted\nGroup/team-c: Deleted\nGroup/team-d: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)
}

// ======================================
// State Management Tests
// ======================================

func Test_Delete_With_State_Removes_From_State(t *testing.T) {
	fmt.Println("Test CLI Delete with state management - removes from state")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// First apply to create the resource and state
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Initial apply command failed: %v\nStderr: %s", err, stderr)

	// Verify state file contains the resource
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-a", "State file should contain team-a before deletion")

	// Delete with state enabled
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-a: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file no longer contains the resource
	stateContent, err = os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.NotContains(t, string(stateContent), "team-a", "State file should not contain team-a after deletion")
}

func Test_Delete_With_State_Multiple_Resources(t *testing.T) {
	fmt.Println("Test CLI Delete with state management - multiple resources")
	folderPath := testDataFilePath(t, "resources_folder")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply multiple resources with state
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)

	// Verify state file contains both resources
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-b", "State file should contain team-b")
	assert.Contains(t, string(stateContent), "team-c", "State file should contain team-c")

	// Delete with state enabled
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-b: Deleted\nGroup/team-c: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file no longer contains the resources
	stateContent, err = os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.NotContains(t, string(stateContent), "team-b", "State file should not contain team-b after deletion")
	assert.NotContains(t, string(stateContent), "team-c", "State file should not contain team-c after deletion")
}

func Test_Delete_With_State_Via_Env_Var(t *testing.T) {
	fmt.Println("Test CLI Delete with state management via environment variables")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply with state
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)

	// Set environment variables for delete
	os.Setenv("CDK_STATE_ENABLED", "true")
	os.Setenv("CDK_STATE_FILE", stateFile)
	defer func() {
		os.Unsetenv("CDK_STATE_ENABLED")
		os.Unsetenv("CDK_STATE_FILE")
	}()

	// Delete using env vars
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-a: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file no longer contains the resource
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.NotContains(t, string(stateContent), "team-a", "State file should not contain team-a after deletion")
}

// TODO fix behavior so that partial failures are properly handled
func Test_Delete_With_State_Partial_Failure(t *testing.T) {
	fmt.Println("Test CLI Delete with state management - partial failure")
	validFilePath := testDataFilePath(t, "valid_user.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply resource with state
	_, stderr, err := runConsoleCommand("apply", "-f", validFilePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)

	// Manually delete the resource from API (simulating partial failure scenario)
	_, stderr, err = runConsoleCommand("delete", "-f", validFilePath)
	assert.NoErrorf(t, err, "Manual delete command failed: %v\nStderr: %s", err, stderr)

	// Try to delete again with state enabled (resource already deleted from API)
	// This should fail but still update the state
	_, stderr, err = runConsoleCommand("delete", "-f", validFilePath, "--enable-state", "--state-file", stateFile)
	assert.Error(t, err, "Expected error when deleting non-existent resource")

	// Verify state file is updated (resource removed from state even though API delete failed)
	_, err = os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
}

func Test_Delete_Without_State_Does_Not_Update_State(t *testing.T) {
	fmt.Println("Test CLI Delete without state management - does not update state")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply with state to create resource and state
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)

	// Get initial state content
	initialStateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read initial state file")
	assert.Contains(t, string(initialStateContent), "team-a", "State file should contain team-a")

	// Delete WITHOUT state management
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-a: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file still contains the resource (not updated)
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-a", "State file should still contain team-a when state is disabled")
}

func Test_Delete_With_State_Recursive(t *testing.T) {
	fmt.Println("Test CLI Delete with state management - recursive folder")
	folderPath := testDataFilePath(t, "resources_folder")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply recursively with state
	stdout, stderr, err := runConsoleCommand("apply", "-f", folderPath, "-r", "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)

	// Verify state file contains all resources
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.Contains(t, string(stateContent), "team-b", "State file should contain team-b")
	assert.Contains(t, string(stateContent), "team-c", "State file should contain team-c")
	assert.Contains(t, string(stateContent), "team-d", "State file should contain team-d")

	// Delete recursively with state enabled
	stdout, stderr, err = runConsoleCommand("delete", "-f", folderPath, "-r", "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-b: Deleted\nGroup/team-c: Deleted\nGroup/team-d: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file no longer contains the resources
	stateContent, err = os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.NotContains(t, string(stateContent), "team-b", "State file should not contain team-b")
	assert.NotContains(t, string(stateContent), "team-c", "State file should not contain team-c")
	assert.NotContains(t, string(stateContent), "team-d", "State file should not contain team-d")
}

func Test_Delete_With_State_Custom_Location(t *testing.T) {
	fmt.Println("Test CLI Delete with state management - custom state file location")
	filePath := testDataFilePath(t, "valid_group.yaml")
	customStateFile := tmpStateFilePath(t, "custom_state.json")

	// Apply with custom state file location
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--enable-state", "--state-file", customStateFile)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)

	// Verify state file exists at custom location
	_, err = os.Stat(customStateFile)
	assert.NoError(t, err, "State file should exist at custom location")

	// Delete with custom state file location
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath, "--enable-state", "--state-file", customStateFile)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)

	expectedOutput := "Group/team-a: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file at custom location is updated
	stateContent, err := os.ReadFile(customStateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.NotContains(t, string(stateContent), "team-a", "State file should not contain team-a")
}
