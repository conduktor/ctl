package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testDataFilePath(t *testing.T, fileName string) string {
	workDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")
	fmt.Printf("Current working directory: %s\n", workDir)
	return fmt.Sprintf("%s/testdata/apply/%s", workDir, fileName)
}

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

func Test_Apply_Diff_Output(t *testing.T) {
	fmt.Println("Test CLI Apply with diff output")
	filePath := testDataFilePath(t, "valid_group.yaml")

	// First apply to create the resource
	_, stderr, err := runConsoleCommand("apply", "-f", filePath)
	assert.NoErrorf(t, err, "Initial apply command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output during initial apply, got: %s", stderr)

	// Modify the resource file to trigger a diff
	modifiedFilePath := testDataFilePath(t, "valid_group_updated.yaml")
	stdout, stderr, err := runConsoleCommand("apply", "-f", modifiedFilePath, "--print-diff")

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
