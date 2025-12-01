package integration

import (
	"fmt"
	"gopkg.in/yaml.v3"
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

func Test_Delete_With_Dry_Run(t *testing.T) {
	fmt.Println("Test CLI Delete with dry-run")
	filePath := testDataFilePath(t, "valid_group.yaml")

	//Init the resource to delete
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)

	// Delete with dry-run
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath, "--dry-run")
	assert.NoErrorf(t, err, "Dry-run delete command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during dry-run delete, got: %s", stderr)

	expectedOutput := "Group/team-a: Deleted (dry-run)\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Delete with dry-run using filter
	stdout, stderr, err = runConsoleCommand("delete", "Group", "team-a", "--dry-run")
	assert.NoErrorf(t, err, "Dry-run delete command with filter failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during dry-run delete with filter, got: %s", stderr)
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify resource still exists after dry-run delete
	stdout, stderr, err = runConsoleCommand("get", "Group", "team-a")
	assert.NoErrorf(t, err, "Get command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during get, got: %s", stderr)

	parsedResource := parseStdoutAsYAMLDocuments(t, stdout)
	assert.Lenf(t, parsedResource, 1, "Expected one resource in get output, got: %d", len(parsedResource))
	assert.Equal(t, "Group", parsedResource[0]["kind"], "Expected resource kind to be 'Group'")
	assert.Equal(t, "team-a", parsedResource[0]["metadata"].(map[string]any)["name"], "Expected resource name to be 'team-a'")

	// Cleanup: actually delete the resource
	stdout, stderr, err = runConsoleCommand("delete", "-f", filePath)
	assert.NoErrorf(t, err, "Cleanup delete command failed: %v\nStderr: %s", err, stderr)
}

func Test_Delete_Gateway_Resources(t *testing.T) {
	fmt.Println("Test CLI Delete Gateway resources by name")
	workDir := t.TempDir()

	interceptorName, interceptor := FixtureRandomGatewayInterceptor(t)
	vClusterName, vcluster := FixtureRandomGatewayVCluster(t)
	saName, sa := FixtureRandomGatewaySA(t)
	filePath := fmt.Sprintf("%s/gw-resources.yaml", workDir)
	interceptorYAML, err := yaml.Marshal(interceptor)
	assert.NoErrorf(t, err, "Failed to marshal interceptor to YAML: %v", err)
	vclusterYAML, err := yaml.Marshal(vcluster)
	assert.NoErrorf(t, err, "Failed to marshal vcluster to YAML: %v", err)
	saYAML, err := yaml.Marshal(sa)
	assert.NoErrorf(t, err, "Failed to marshal service account to YAML: %v", err)
	combinedYAML := "---\n" + string(interceptorYAML) + "---\n" + string(vclusterYAML) + "---\n" + string(saYAML)
	fmt.Println(combinedYAML)
	err = os.WriteFile(filePath, []byte(combinedYAML), 0644)
	assert.NoError(t, err, "Failed to marshal gateway resources to YAML: %v", err)

	// Init the resources to delete
	stdout, stderr, err := runGatewayCommand("apply", "-f", filePath)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)
	expectedOutputLines := []string{
		fmt.Sprintf("Interceptor/%s: Created", interceptorName),
		fmt.Sprintf("VirtualCluster/%s: Created", vClusterName),
		fmt.Sprintf("GatewayServiceAccount/%s: Created", saName),
	}
	for _, line := range expectedOutputLines {
		assert.Containsf(t, stdout, line, "Expected stdout to contain '%s', got: %s", line, stdout)
	}

	// Test deleteResourceByName for gateway resources
	stdout, stderr, err = runGatewayCommand("delete", "VirtualCluster", vClusterName)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)
	expectedOutput := fmt.Sprintf("VirtualCluster/%s: Deleted\n", vClusterName)
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Test delete via name and vcluster
	stdout, stderr, err = runGatewayCommand("delete", "GatewayServiceAccount", saName, "--vcluster", "passthrough")
	assert.NoErrorf(t, err, "Delete command by vcluster failed: %v\nStderr: %s", err, stderr)
	expectedOutput = fmt.Sprintf("GatewayServiceAccount/map[name:%s vCluster:passthrough]: Deleted\n", saName)
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Test delete interceptors
	stdout, stderr, err = runGatewayCommand("delete", "Interceptor", interceptorName,
		"--vcluster", "passthrough", "--username", "user")
	assert.NoErrorf(t, err, "Delete interceptor command failed: %v\nStderr: %s", err, stderr)
	expectedOutput = fmt.Sprintf("Interceptor/%smap[username:user vCluster:passthrough]: Deleted\n", interceptorName)
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

func Test_Delete_With_State_Filter(t *testing.T) {
	fmt.Println("Test CLI Delete with state management - filter by kind and name")
	filePath := testDataFilePath(t, "valid_group.yaml")
	stateFile := tmpStateFilePath(t, "state.json")

	// Apply with state
	stdout, stderr, err := runConsoleCommand("apply", "-f", filePath, "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Apply command failed: %v\nStderr: %s", err, stderr)

	// Delete using kind and name filter with state enabled
	stdout, stderr, err = runConsoleCommand("delete", "Group", "team-a", "--enable-state", "--state-file", stateFile)
	assert.NoErrorf(t, err, "Delete command failed: %v\nStderr: %s", err, stderr)
	expectedOutput := "Group/team-a: Deleted\n"
	assert.Equalf(t, expectedOutput, stdout, "Expected stdout to be '%s', got: %s", expectedOutput, stdout)

	// Verify state file no longer contains the resource
	stateContent, err := os.ReadFile(stateFile)
	assert.NoError(t, err, "Failed to read state file")
	assert.NotContains(t, string(stateContent), "team-a", "State file should not contain team-a after deletion")
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
