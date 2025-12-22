package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var composeFilePath = "./testdata/docker-compose.integration-test.yml"
var consoleVersion = "1.38.0"
var gatewayVersion = "3.14.0"

var consoleURL = "http://localhost:8080"
var consoleAdminEmail = "admin@conduktor.io"
var consoleAdminPassword = "testP4ss!"
var gatewayURL = "http://localhost:8888"
var gatewayAdmin = "admin"
var gatewayAdminPassword = "conduktor"

var debugLogger = log.New(os.Stderr, "", 1)

func TestMain(m *testing.M) {
	if !strings.EqualFold(os.Getenv("INTEGRATION_TESTS"), "true") {
		fmt.Println("Skipping integration tests. Set INTEGRATION_TESTS=true to enable.")
		return
	}

	// Start Docker Compose
	if err := setupDocker(); err != nil {
		debugLogger.Printf("Failed to setup: %v\n", err)
		os.Exit(1)
	}
	debugLogger.Println("Wait 30s for compose to be up and running")
	time.Sleep(30 * time.Second)

	// Wait for API to be ready
	if err := waitForConsole(); err != nil {
		teardownDocker()
		debugLogger.Printf("API not ready: %v\n", err)
		os.Exit(1)
	}

	if err := waitForGateway(); err != nil {
		teardownDocker()
		debugLogger.Printf("API not ready: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	teardownDocker()
	os.Exit(code)
}

func setupDocker() error {
	debugLogger.Println("Start integration docker stack")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	setComposeEnv()
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", composeFilePath,
		"up", "-d", "--build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func teardownDocker() {
	// setComposeEnv()
	cmd := exec.Command("docker", "compose",
		"-f", composeFilePath,
		"down", "-v")
	err := cmd.Run()
	if err != nil {
		debugLogger.Printf("Failed to teardown docker compose: %v\n", err)
	}
}

func setComposeEnv() {
	debugLogger.Println("Setting up environment variables for Docker Compose")
	logAndSetEnv("CONDUKTOR_CONSOLE_IMAGE", fmt.Sprintf("conduktor/conduktor-console:%s", consoleVersion))
	logAndSetEnv("CONDUKTOR_CONSOLE_CORTEX_IMAGE", fmt.Sprintf("conduktor/conduktor-console-cortex:%s", consoleVersion))
	logAndSetEnv("CONDUKTOR_GATEWAY_IMAGE", fmt.Sprintf("conduktor/conduktor-gateway:%s", gatewayVersion))
	logAndSetEnv("CDK_BASE_URL", consoleURL)
	logAndSetEnv("CDK_ADMIN_EMAIL", consoleAdminEmail)
	logAndSetEnv("CDK_ADMIN_PASSWORD", consoleAdminPassword)
	logAndSetEnv("CDK_GATEWAY_BASE_URL", gatewayURL)
	logAndSetEnv("CDK_GATEWAY_USER", gatewayAdmin)
	logAndSetEnv("CDK_GATEWAY_PASSWORD", gatewayAdminPassword)
}

func setCLIConsoleEnv() {
	debugLogger.Println("Setting up environment variables for CLI Console")
	logAndSetEnv("CDK_BASE_URL", consoleURL)
	logAndSetEnv("CDK_USER", consoleAdminEmail)
	logAndSetEnv("CDK_PASSWORD", consoleAdminPassword)
}

func setCLIGatewayEnv() {
	debugLogger.Println("Setting up environment variables for CLI Gateway")
	logAndSetEnv("CDK_GATEWAY_BASE_URL", gatewayURL)
	logAndSetEnv("CDK_GATEWAY_USER", gatewayAdmin)
	logAndSetEnv("CDK_GATEWAY_PASSWORD", gatewayAdminPassword)
}

func logAndSetEnv(key, value string) {
	debugLogger.Printf("Set env %s=%s\n", key, value)
	os.Setenv(key, value)
}

func waitForConsole() error {
	debugLogger.Printf("Wait for Console API to be ready on %s\n", consoleURL)
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(consoleURL + "/api/health/ready")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			debugLogger.Println("Console API ready !")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("API not ready after 60 seconds")
}

func waitForGateway() error {
	debugLogger.Printf("Wait for Gateway API to be ready on %s\n", gatewayURL)
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(gatewayURL + "/health/ready")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			debugLogger.Println("Gateway API ready !")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("API not ready after 60 seconds")
}

func runConsoleCommand(args ...string) (string, string, error) {
	setCLIConsoleEnv()
	return runCommand(args...)
}

func runGatewayCommand(args ...string) (string, string, error) {
	setCLIGatewayEnv()
	return runCommand(args...)
}

func runCommand(args ...string) (string, string, error) {

	baseCmd := []string{"go", "run", "../../main.go"}
	command := append(baseCmd, args...)
	cmd := exec.Command(command[0], command[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	debugLogger.Printf("Run command : %v\n", command)
	debugLogger.Printf("####stdout %s", stdout.String())
	debugLogger.Printf("####stderr %s", stderr.String())

	return stdout.String(), stderr.String(), err
}

func parseStdoutAsYAMLDocuments(t *testing.T, stdout string) []map[string]any {
	yamlDocuments := strings.Split(stdout, "---")
	var results []map[string]any
	for _, doc := range yamlDocuments {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		var result map[string]any
		err := yaml.Unmarshal([]byte(doc), &result)
		assert.NoErrorf(t, err, "Failed to parse YAML document: %v\nDocument: %s", err, doc)
		results = append(results, result)
	}
	return results
}

func parseStdoutAsJsonArray(t *testing.T, stdout string) []map[string]any {
	var results []map[string]any
	err := json.Unmarshal([]byte(stdout), &results)
	assert.NoErrorf(t, err, "Failed to parse JSON array: %v\nOutput: %s", err, stdout)
	return results
}

func parseStdoutAsJsonObject(t *testing.T, stdout string) map[string]any {
	var results map[string]any
	err := json.Unmarshal([]byte(stdout), &results)
	assert.NoErrorf(t, err, "Failed to parse JSON object: %v\nOutput: %s", err, stdout)
	return results
}

func testDataFilePath(t *testing.T, fileName string) string {
	workDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")
	debugLogger.Printf("Current working directory: %s\n", workDir)
	return fmt.Sprintf("%s/testdata/resources/%s", workDir, fileName)
}

func createAdminToken(t *testing.T, name string) (token string, tokenName string) {
	stdout, stderr, err := runConsoleCommand("token", "create", "admin", name)
	assert.NoErrorf(t, err, "Failed to create admin token: %s", stderr)

	token = strings.TrimSpace(stdout)
	assert.NotEmpty(t, token, "Expected token in response")

	debugLogger.Printf("Created admin token: %s\n", name)
	return token, name
}

func deleteTokenByName(tokenName string) {
	// First, list admin tokens to find the ID
	stdout, stderr, err := runConsoleCommand("token", "list", "admin")
	if err != nil {
		debugLogger.Printf("Failed to list admin tokens: %s\n", stderr)
		return
	}

	// Parse output format: "name:\tUUID"
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "No tokens found" {
			continue
		}
		parts := strings.Split(line, ":\t")
		if len(parts) == 2 && parts[0] == tokenName {
			tokenId := parts[1]
			_, stderr, err := runConsoleCommand("token", "delete", tokenId)
			if err != nil {
				debugLogger.Printf("Failed to delete token %s (id: %s): %s\n", tokenName, tokenId, stderr)
			} else {
				debugLogger.Printf("Deleted token: %s (id: %s)\n", tokenName, tokenId)
			}
			return
		}
	}
	debugLogger.Printf("Token not found for deletion: %s\n", tokenName)
}

func runCommandWithToken(token string, args ...string) (string, string, error) {
	// Save original env
	originalApiKey := os.Getenv("CDK_API_KEY")
	originalUser := os.Getenv("CDK_USER")
	originalPassword := os.Getenv("CDK_PASSWORD")

	// Set token auth
	logAndSetEnv("CDK_BASE_URL", consoleURL)
	logAndSetEnv("CDK_API_KEY", token)
	os.Unsetenv("CDK_USER")
	os.Unsetenv("CDK_PASSWORD")

	defer func() {
		// Restore original env
		if originalApiKey != "" {
			os.Setenv("CDK_API_KEY", originalApiKey)
		} else {
			os.Unsetenv("CDK_API_KEY")
		}
		if originalUser != "" {
			os.Setenv("CDK_USER", originalUser)
		}
		if originalPassword != "" {
			os.Setenv("CDK_PASSWORD", originalPassword)
		}
	}()

	return runCommand(args...)
}
