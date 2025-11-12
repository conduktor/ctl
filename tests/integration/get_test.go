package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Get_All(t *testing.T) {
	fmt.Println("Test CLI Get All resources")
	stdout, stderr, err := runConsoleCommand("get", "all")
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	yamlDocuments := parseStdoutAsYAMLDocuments(t, stdout)
	assert.GreaterOrEqualf(t, len(yamlDocuments), 2, "Expected multiple YAML documents, got: %s", stdout)
}

func Test_Get_All_Kind(t *testing.T) {
	fmt.Println("Test CLI Get All User resource")
	stdout, stderr, err := runConsoleCommand("get", "User")
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	yamlDocuments := parseStdoutAsYAMLDocuments(t, stdout)
	assert.Lenf(t, yamlDocuments, 1, "Expected one YAML document, got: %s", stdout)

	user := yamlDocuments[0]
	kind := user["kind"].(string)
	assert.Equal(t, "User", kind)

	metadata, ok := user["metadata"].(map[string]any)
	assert.True(t, ok, "Metadata is not a map")

	name := metadata["name"].(string)
	assert.Equal(t, consoleAdminEmail, name)
}

func Test_Get_User(t *testing.T) {
	fmt.Println("Test CLI Get User resource")
	stdout, stderr, err := runConsoleCommand("get", "User", consoleAdminEmail)
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	yamlDocuments := parseStdoutAsYAMLDocuments(t, stdout)
	assert.Lenf(t, yamlDocuments, 1, "Expected one YAML document, got: %s", stdout)

	user := yamlDocuments[0]
	kind := user["kind"].(string)
	assert.Equal(t, "User", kind)

	metadata, ok := user["metadata"].(map[string]any)
	assert.True(t, ok, "Metadata is not a map")

	name := metadata["name"].(string)
	assert.Equal(t, consoleAdminEmail, name)
}

func Test_Get_User_Output_Json(t *testing.T) {
	fmt.Println("Test CLI Get User resource")
	stdout, stderr, err := runConsoleCommand("get", "User", consoleAdminEmail, "--output", "json")
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	jsonResources := parseStdoutAsJsonArray(t, stdout)
	assert.Lenf(t, jsonResources, 1, "Expected one JSON document, got: %s", stdout)

	user := jsonResources[0]
	kind := user["kind"].(string)
	assert.Equal(t, "User", kind)

	metadata, ok := user["metadata"].(map[string]any)
	assert.True(t, ok, "Metadata is not a map")

	name := metadata["name"].(string)
	assert.Equal(t, consoleAdminEmail, name)
}

func Test_Get_User_Output_Name(t *testing.T) {
	fmt.Println("Test CLI Get User resource")
	stdout, stderr, err := runConsoleCommand("get", "User", consoleAdminEmail, "--output", "name")
	assert.NoErrorf(t, err, "Command failed: %v\nStderr: %s", err, stderr)
	assert.Emptyf(t, stderr, "Expected no stderr output, got: %s", stderr)

	expectedOutput := fmt.Sprintf("User/%s\n", consoleAdminEmail)
	assert.Equalf(t, expectedOutput, stdout, "Expected output '%s', got: %s", expectedOutput, stdout)
}
