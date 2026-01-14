package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Config_Defaults(t *testing.T) {
	// Load config without passing any pointers or environment variables
	config := NewStorageConfig(nil, nil)

	assert.False(t, config.Enabled)
	assert.Nil(t, config.FilePath)
}

func Test_Config_Load_From_Env(t *testing.T) {
	// Set environment variables
	os.Setenv("CDK_STATE_ENABLED", "true")
	os.Setenv("CDK_STATE_FILE", "/env/path/to/state.json")

	// Load config without passing any pointers
	config := NewStorageConfig(nil, nil)

	assert.True(t, config.Enabled)
	assert.NotNil(t, config.FilePath)
	assert.Equal(t, "/env/path/to/state.json", *config.FilePath)

	// Clean up environment variables
	os.Unsetenv("CDK_STATE_ENABLED")
	os.Unsetenv("CDK_STATE_FILE")
}

func Test_Config_Load_From_Env_When_Empty(t *testing.T) {
	// Set environment variables
	os.Setenv("CDK_STATE_ENABLED", "true")
	os.Setenv("CDK_STATE_FILE", "/env/path/to/state.json")

	// Load config without passing any pointers

	enabled := false
	filePath := ""
	config := NewStorageConfig(&enabled, &filePath)

	assert.True(t, config.Enabled)
	assert.NotNil(t, config.FilePath)
	assert.Equal(t, "/env/path/to/state.json", *config.FilePath)

	// Clean up environment variables
	os.Unsetenv("CDK_STATE_ENABLED")
	os.Unsetenv("CDK_STATE_FILE")
}

func Test_Config_Load_From_Params(t *testing.T) {
	enabled := true
	filePath := "/param/path/to/state.json"

	// Load config by passing parameters
	config := NewStorageConfig(&enabled, &filePath)

	assert.True(t, config.Enabled)
	assert.NotNil(t, config.FilePath)
	assert.Equal(t, "/param/path/to/state.json", *config.FilePath)
}

func Test_Config_Params_Override_Env(t *testing.T) {
	// Set environment variables
	os.Setenv("CDK_STATE_ENABLED", "false")
	os.Setenv("CDK_STATE_FILE", "/env/path/to/state.json")

	enabled := true
	filePath := "/param/path/to/state.json"

	// Load config by passing parameters
	config := NewStorageConfig(&enabled, &filePath)

	assert.True(t, config.Enabled)
	assert.NotNil(t, config.FilePath)
	assert.Equal(t, "/param/path/to/state.json", *config.FilePath)

	// Clean up environment variables
	os.Unsetenv("CDK_STATE_ENABLED")
	os.Unsetenv("CDK_STATE_FILE")
}
