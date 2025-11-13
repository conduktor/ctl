package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var consoleVersion = "1.38.0"
var gatewayVersion = "3.14.0"

var consoleURL = "http://localhost:8080"
var consoleAdminEmail = "admin@conduktor.io"
var consoleAdminPassword = "testP4ss!"
var gatewayURL = "http://localhost:8888"
var gatewayAdmin = "admin"
var gatewayAdminPassword = "conduktor"

func TestMain(m *testing.M) {
	// Start Docker Compose
	if err := setupDocker(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Wait 30s for compose to be up and running")
	time.Sleep(30 * time.Second)

	// Wait for API to be ready
	if err := waitForConsole(); err != nil {
		teardownDocker()
		fmt.Fprintf(os.Stderr, "API not ready: %v\n", err)
		os.Exit(1)
	}

	if err := waitForGateway(); err != nil {
		teardownDocker()
		fmt.Fprintf(os.Stderr, "API not ready: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	teardownDocker()
	os.Exit(code)
}

func setupDocker() error {
	fmt.Println("Start integration docker stack")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	setComposeEnv()
	cmd := exec.CommandContext(ctx, "docker", "compose",
		"-f", "docker-compose.test.yml",
		"up", "-d", "--build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func teardownDocker() {
	// setComposeEnv()
	cmd := exec.Command("docker", "compose",
		"-f", "docker-compose.test.yml",
		"down", "-v")
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to teardown docker compose: %v\n", err)
	}
}

func setComposeEnv() {
	fmt.Printf("Setting up environment variables for Docker Compose\n")
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
	fmt.Println("Setting up environment variables for CLI Console")
	logAndSetEnv("CDK_BASE_URL", consoleURL)
	logAndSetEnv("CDK_USER", consoleAdminEmail)
	logAndSetEnv("CDK_PASSWORD", consoleAdminPassword)
}

func setCLIGatewayEnv() {
	fmt.Println("Setting up environment variables for CLI Gateway")
	logAndSetEnv("CDK_GATEWAY_BASE_URL", gatewayURL)
	logAndSetEnv("CDK_GATEWAY_USER", gatewayAdmin)
	logAndSetEnv("CDK_GATEWAY_PASSWORD", gatewayAdminPassword)
}

func logAndSetEnv(key, value string) {
	fmt.Printf("Set env %s=%s\n", key, value)
	os.Setenv(key, value)
}

func waitForConsole() error {
	fmt.Printf("Wait for Console API to be ready on %s\n", consoleURL)
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(consoleURL + "/api/health/ready")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("Console API ready !")
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
	fmt.Printf("Wait for Gateway API to be ready on %s\n", gatewayURL)
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(gatewayURL + "/health/ready")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("Gateway API ready !")
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

	baseCmd := []string{"go", "run", "../../main.go"}
	command := append(baseCmd, args...)
	cmd := exec.Command(command[0], command[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	fmt.Printf("Run command : %v\n", command)
	fmt.Printf("####stdout \n%s\n", stdout.String())
	fmt.Printf("####stderr \n%s\n", stderr.String())

	return stdout.String(), stderr.String(), err
}

func runGatewayCommand(args ...string) (string, string, error) {
	setCLIGatewayEnv()

	baseCmd := []string{"go", "run", "../../main.go"}
	command := append(baseCmd, args...)
	cmd := exec.Command(command[0], command[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	fmt.Printf("Run command : %v\n", command)
	fmt.Printf("	stdout %s", stdout.String())
	fmt.Printf("	stderr %s", stderr.String())

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
	fmt.Printf("Current working directory: %s\n", workDir)
	return fmt.Sprintf("%s/testdata/resources/%s", workDir, fileName)
}
