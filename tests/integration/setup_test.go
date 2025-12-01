package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"golang.org/x/exp/rand"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var composeFilePath = "./testdata/docker-compose.integration-test.yml"
var consoleVersion = "1.40.0"
var gatewayVersion = "3.15.0"

var consoleURL = "http://localhost:8080"
var consoleAdminEmail = "admin@conduktor.io"
var consoleAdminPassword = "testP4ss!"
var gatewayURL = "http://localhost:8888"
var gatewayAdmin = "admin"
var gatewayAdminPassword = "conduktor"

func TestMain(m *testing.M) {
	if !strings.EqualFold(os.Getenv("INTEGRATION_TESTS"), "true") {
		fmt.Println("Skipping integration tests. Set INTEGRATION_TESTS=true to enable.")
		return
	}

	// Start Docker Compose
	if shouldManageComposeStack() {
		if err := setupDocker(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to setup: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "Wait 30s for compose to be up and running")
		time.Sleep(30 * time.Second)
	}
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

// By default, management of the tests compose stack is part of tests lifecycle.
// For quicker feedback loop, tests compose stack can be run outside of tests.
// In this case set INTEGRATION_MANAGE_COMPOSE=false.
func shouldManageComposeStack() bool {
	return os.Getenv("INTEGRATION_MANAGE_COMPOSE") != "false"
}

func setupDocker() error {
	fmt.Fprintln(os.Stderr, "Start integration docker stack")
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
	if shouldManageComposeStack() {
		setComposeEnv()
		cmd := exec.Command("docker", "compose",
			"-f", composeFilePath,
			"down", "-v")
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to teardown docker compose: %v\n", err)
		}
	}
}

func setComposeEnv() {
	fmt.Fprintln(os.Stderr, "Setting up environment variables for Docker Compose")
	logAndSetEnv("CONDUKTOR_CONSOLE_IMAGE", fmt.Sprintf("conduktor/conduktor-console:%s", consoleVersion))
	logAndSetEnv("CONDUKTOR_CONSOLE_CORTEX_IMAGE", fmt.Sprintf("conduktor/conduktor-console-cortex:%s", consoleVersion))
	logAndSetEnv("CONDUKTOR_GATEWAY_IMAGE", fmt.Sprintf("conduktor/conduktor-gateway:%s", gatewayVersion))
	logAndSetEnv("CDK_BASE_URL", consoleURL)
	logAndSetEnv("CDK_USER", consoleAdminEmail)
	logAndSetEnv("CDK_PASSWORD", consoleAdminPassword)
	logAndSetEnv("CDK_GATEWAY_BASE_URL", gatewayURL)
	logAndSetEnv("CDK_GATEWAY_USER", gatewayAdmin)
	logAndSetEnv("CDK_GATEWAY_PASSWORD", gatewayAdminPassword)
}

func SetCLIConsoleEnv() {
	fmt.Fprintln(os.Stderr, "Setting up environment variables for CLI Console")
	os.Setenv("CDK_BASE_URL", consoleURL)
	os.Setenv("CDK_USER", consoleAdminEmail)
	os.Setenv("CDK_PASSWORD", consoleAdminPassword)
}

func UnsetCLIConsoleEnv() {
	fmt.Fprintln(os.Stderr, "Unsetting environment variables for CLI Console")
	os.Unsetenv("CDK_BASE_URL")
	os.Unsetenv("CDK_USER")
	os.Unsetenv("CDK_PASSWORD")
}

func SetCLIGatewayEnv() {
	fmt.Fprintln(os.Stderr, "Setting up environment variables for CLI Gateway")
	os.Setenv("CDK_GATEWAY_BASE_URL", gatewayURL)
	os.Setenv("CDK_GATEWAY_USER", gatewayAdmin)
	os.Setenv("CDK_GATEWAY_PASSWORD", gatewayAdminPassword)
}

func UnsetCLIGatewayEnv() {
	fmt.Fprintln(os.Stderr, "Unsetting environment variables for CLI Gateway")
	os.Unsetenv("CDK_GATEWAY_BASE_URL")
	os.Unsetenv("CDK_GATEWAY_USER")
	os.Unsetenv("CDK_GATEWAY_PASSWORD")
}

func logAndSetEnv(key, value string) {
	fmt.Fprintf(os.Stderr, "Set env %s=%s\n", key, value)
	os.Setenv(key, value)
}

func waitForConsole() error {
	fmt.Fprintf(os.Stderr, "Wait for Console API to be ready on %s\n", consoleURL)
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(consoleURL + "/api/health/ready")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Fprintln(os.Stderr, "Console API ready !")
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
	fmt.Fprintf(os.Stderr, "Wait for Gateway API to be ready on %s\n", gatewayURL)
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(gatewayURL + "/health/ready")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Fprintln(os.Stderr, "Gateway API ready !")
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
	SetCLIConsoleEnv()
	return RunCommand(args...)
}

func runGatewayCommand(args ...string) (string, string, error) {
	SetCLIGatewayEnv()
	return RunCommand(args...)
}

func RunCommand(args ...string) (string, string, error) {

	baseCmd := []string{"go", "run", "../../main.go"}
	command := append(baseCmd, args...)
	cmd := exec.Command(command[0], command[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	fmt.Fprintf(os.Stderr, "## Run command : %v\n", command)
	fmt.Fprintf(os.Stderr, "#### stdout\n%s", stdout.String())
	fmt.Fprintf(os.Stderr, "#### stderr\n%s", stderr.String())
	fmt.Fprintln(os.Stderr, "##")

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
	fmt.Fprintf(os.Stderr, "Current working directory: %s\n", workDir)
	return fmt.Sprintf("%s/testdata/resources/%s", workDir, fileName)
}

func tmpStateFilePath(t *testing.T, fileName string) string {
	tmpDir := t.TempDir()
	return fmt.Sprintf("%s/%s", tmpDir, fileName)
}

func FixtureRandomConsoleUser(t *testing.T) (string, any) {
	workDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")
	templatePath := fmt.Sprintf("%s/testdata/fixtures/console_user.yaml.tmpl", workDir)

	randomSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := fmt.Sprintf("user-%s@company.io", randomSuffix)
	data := map[string]string{
		"name":      name,
		"lastname":  "Doe",
		"firstname": "John",
	}
	return name, TemplateFixtureYAML(t, templatePath, data)
}

func FixtureRandomConsoleGroup(t *testing.T) (string, any) {
	workDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")
	templatePath := fmt.Sprintf("%s/testdata/fixtures/console_group.yaml.tmpl", workDir)

	randomSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := fmt.Sprintf("group-%s", randomSuffix)
	data := map[string]string{
		"name":         name,
		"display_name": fmt.Sprintf("Group %s", randomSuffix),
	}
	return name, TemplateFixtureYAML(t, templatePath, data)
}

func FixtureRandomGatewayInterceptor(t *testing.T) (string, any) {
	workDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")
	templatePath := fmt.Sprintf("%s/testdata/fixtures/gateway_interceptor.yaml.tmpl", workDir)

	randomSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := fmt.Sprintf("interceptor-%s", randomSuffix)
	min := 1 + rand.Intn(5)
	max := min + rand.Intn(5)
	data := map[string]string{
		"name":              name,
		"vCluster":          "passthrough",
		"username":          "user",
		"priority":          strconv.Itoa(rand.Intn(100)),
		"topic":             fmt.Sprintf("topic-%s", randomSuffix),
		"min_num_partition": strconv.Itoa(min),
		"max_num_partition": strconv.Itoa(max),
	}
	return name, TemplateFixtureYAML(t, templatePath, data)
}

func FixtureRandomGatewaySA(t *testing.T) (string, any) {
	workDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")
	templatePath := fmt.Sprintf("%s/testdata/fixtures/gateway_service_account.yaml.tmpl", workDir)

	randomSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := fmt.Sprintf("sa-%s", randomSuffix)
	data := map[string]string{
		"name":     name,
		"vCluster": "passthrough",
	}
	return name, TemplateFixtureYAML(t, templatePath, data)
}

func FixtureRandomGatewayVCluster(t *testing.T) (string, any) {
	workDir, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")
	templatePath := fmt.Sprintf("%s/testdata/fixtures/gateway_vcluster.yaml.tmpl", workDir)

	randomSuffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := fmt.Sprintf("vcluster-%s", randomSuffix)
	data := map[string]string{
		"name": name,
	}
	return name, TemplateFixtureYAML(t, templatePath, data)
}

func TemplateFixtureYAML(t *testing.T, templateFilePath string, data map[string]string) any {
	tpl, err := template.ParseFiles(templateFilePath)
	assert.NoError(t, err, "Failed to parse template file")

	// write the template output to a buffer
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err, "Failed to execute template")

	var result any
	err = yaml.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err, "Failed to unmarshal YAML")

	return result
}
