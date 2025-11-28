package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func tmpStateLocation(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return filepath.Join(tempDir, StateFileName)
}

func TestNewLocalFileBackend(t *testing.T) {
	// Test with nil path (should use default)
	backend := NewLocalFileBackend(nil, true)
	assert.NotNil(t, backend)
	switch runtime.GOOS {
	case "linux":
		assert.Contains(t, backend.FilePath, ".local/share/conduktor/cli-state.json")
	case "windows":
		assert.Contains(t, backend.FilePath, "AppData/Roaming/conduktor/cli-state.json")
	case "darwin":
		assert.Contains(t, backend.FilePath, "Library/Application Support/conduktor/cli-state.json")
	}

	// Test with empty string path (should use default)
	emptyPath := ""
	backend2 := NewLocalFileBackend(&emptyPath, true)
	assert.NotNil(t, backend2)
	assert.Equal(t, backend.FilePath, backend2.FilePath)

	// Test with custom path
	customPath := "/tmp/custom/model.json"
	backend3 := NewLocalFileBackend(&customPath, true)
	assert.NotNil(t, backend3)
	assert.Equal(t, customPath, backend3.FilePath)
}

func TestLocalFileBackend_LoadState_NewFile(t *testing.T) {
	stateFilePath := tmpStateLocation(t)
	backend := NewLocalFileBackend(&stateFilePath, true)

	// Load state from non-existent file
	loadedState, err := backend.LoadState(true)
	assert.NoError(t, err)
	assert.NotNil(t, loadedState)
	assert.Equal(t, model.StateFileVersion, loadedState.Version)
	assert.NotEmpty(t, loadedState.LastUpdated)
	assert.Empty(t, loadedState.Resources)
}

func TestLocalFileBackend_LoadState_ExistingFile(t *testing.T) {
	stateFilePath := tmpStateLocation(t)
	backend := NewLocalFileBackend(&stateFilePath, true)

	// Create initial state with resources
	initialState := &model.State{
		Version:     model.StateFileVersion,
		LastUpdated: "2024-01-01T00:00:00Z",
		Resources: []model.ResourceState{
			{
				APIVersion: "v1",
				Kind:       "TestKind",
				Metadata:   &map[string]any{"name": "TestLocalFileBackend_LoadState_ExistingFile-test-resource"},
			},
		},
	}

	// Save the initial state to file
	err := backend.SaveState(initialState, true)
	assert.NoError(t, err)

	// Load state from existing file
	loadedState, err := backend.LoadState(true)
	assert.NoError(t, err)
	assert.NotNil(t, loadedState)
	assert.Equal(t, model.StateFileVersion, loadedState.Version)
	assert.Equal(t, "2024-01-01T00:00:00Z", loadedState.LastUpdated)
	assert.Len(t, loadedState.Resources, 1)
	assert.Equal(t, "TestKind", loadedState.Resources[0].Kind)
	assert.Equal(t, "TestLocalFileBackend_LoadState_ExistingFile-test-resource", (*loadedState.Resources[0].Metadata)["name"])
}

func TestLocalFileBackend_SaveState(t *testing.T) {
	stateFilePath := tmpStateLocation(t)
	backend := NewLocalFileBackend(&stateFilePath, true)

	// Create state with resources
	testState := &model.State{
		Version:     model.StateFileVersion,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Resources: []model.ResourceState{
			{
				APIVersion: "v1",
				Kind:       "TestKind1",
				Metadata:   &map[string]any{"name": "TestLocalFileBackend_SaveState-resource1"},
			},
			{
				APIVersion: "v2",
				Kind:       "TestKind2",
				Metadata:   &map[string]any{"name": "TestLocalFileBackend_SaveState-resource2"},
			},
		},
	}

	beforeSave := time.Now()
	err := backend.SaveState(testState, true)
	assert.NoError(t, err)

	// Check that file exists
	assert.FileExists(t, stateFilePath)

	// Read the file and verify content
	data, err := os.ReadFile(stateFilePath)
	assert.NoError(t, err)

	var savedState model.State
	err = json.Unmarshal(data, &savedState)
	assert.NoError(t, err)
	assert.Equal(t, testState.Version, savedState.Version)
	assert.Equal(t, testState.LastUpdated, savedState.LastUpdated)
	assert.Len(t, savedState.Resources, 2)
	assert.Equal(t, "TestKind1", savedState.Resources[0].Kind)
	assert.Equal(t, "TestKind2", savedState.Resources[1].Kind)

	// Verify that the save operation completed within reasonable time
	savedTime, err := time.Parse(time.RFC3339, savedState.LastUpdated)
	assert.NoError(t, err)
	assert.True(t, savedTime.After(beforeSave.Add(-time.Second)) && savedTime.Before(beforeSave.Add(5*time.Second)))
}

func TestLocalFileBackend_SaveState_NestedDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test_nested_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	stateFilePath := filepath.Join(tempDir, "nested", "dir", StateFileName)
	backend := NewLocalFileBackend(&stateFilePath, true)

	testState := &model.State{
		Version:     model.StateFileVersion,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Resources:   []model.ResourceState{},
	}

	err = backend.SaveState(testState, true)
	assert.NoError(t, err)
	assert.FileExists(t, stateFilePath)
}

func TestLocalFileBackend_LoadState_CorruptedFile(t *testing.T) {
	stateFilePath := tmpStateLocation(t)
	backend := NewLocalFileBackend(&stateFilePath, true)

	// Write corrupted JSON to file
	corruptedJSON := `{"version": "v1", "lastUpdated": "2024-01-01", "resources": [`
	err := os.WriteFile(stateFilePath, []byte(corruptedJSON), 0644)
	assert.NoError(t, err)

	// Try to load corrupted file
	_, err = backend.LoadState(true)
	assert.Error(t, err)
	expectedError := "file storage error: failed to unmarshal state JSON."
	expectedError += "\n  Cause: unexpected end of JSON input"
	expectedError += fmt.Sprintf("\nThe state file may be corrupted or not in the expected format. You can try deleting or backing up the state file located at %s and rerun the command to generate a new state file.", stateFilePath)
	assert.Equal(t, expectedError, err.Error())
}

func TestLocalFileBackend_SaveState_ReadOnlyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test_readonly_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		err := os.Chmod(tempDir, 0755) // Restore permissions for cleanup
		assert.NoError(t, err)
		err = os.RemoveAll(tempDir)
		assert.NoError(t, err)
	})

	// Make directory read-only
	err = os.Chmod(tempDir, 0444)
	assert.NoError(t, err)

	stateFilePath := filepath.Join(tempDir, StateFileName)
	backend := NewLocalFileBackend(&stateFilePath, true)

	testState := &model.State{
		Version:     model.StateFileVersion,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Resources:   []model.ResourceState{},
	}

	// Should fail to save to read-only directory
	err = backend.SaveState(testState, true)
	assert.Error(t, err)
	expectedError := "file storage error: failed to write state."
	expectedError += fmt.Sprintf("\n  Cause: open %s: permission denied", stateFilePath)
	expectedError += fmt.Sprintf("\nEnsure that the file path %s is correct and that you have the necessary permissions to write to it.", stateFilePath)
	assert.Equal(t, expectedError, err.Error())
}

func TestLocalFileBackend_IntegrationWithState(t *testing.T) {
	stateFilePath := tmpStateLocation(t)
	backend := NewLocalFileBackend(&stateFilePath, true)

	// Create and populate state
	testState := model.NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestLocalFileBackend_IntegrationWithState-resource1"},
	}

	testState.AddManagedResource(resource1)
	err := backend.SaveState(testState, true)
	assert.NoError(t, err)

	// Load state again and verify
	reloadedState, err := backend.LoadState(true)
	assert.NoError(t, err)
	assert.Len(t, reloadedState.Resources, 1)
	assert.True(t, reloadedState.IsResourceManaged(resource1))

	// Add another resource and save
	resource2 := resource.Resource{
		Kind:     "TestKind2",
		Version:  "v2",
		Metadata: map[string]any{"name": "TestLocalFileBackend_IntegrationWithState-resource2"},
	}

	reloadedState.AddManagedResource(resource2)
	err = backend.SaveState(reloadedState, true)
	assert.NoError(t, err)

	// Final reload and verification
	finalState, err := backend.LoadState(true)
	assert.NoError(t, err)
	assert.Len(t, finalState.Resources, 2)
	assert.True(t, finalState.IsResourceManaged(resource1))
	assert.True(t, finalState.IsResourceManaged(resource2))
}

func TestLocalFileBackend_ConcurrentOperations(t *testing.T) {
	stateFilePath := tmpStateLocation(t)

	// Create multiple backends pointing to same file
	backend1 := NewLocalFileBackend(&stateFilePath, true)
	backend2 := NewLocalFileBackend(&stateFilePath, true)

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestLocalFileBackend_ConcurrentOperations-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestLocalFileBackend_ConcurrentOperations-resource2"},
	}

	// Create different states and save them
	state1 := model.NewState()
	state1.AddManagedResource(resource1)

	state2 := model.NewState()
	state2.AddManagedResource(resource2)

	// Save both (state2 will overwrite state1's changes)
	err := backend1.SaveState(state1, true)
	assert.NoError(t, err)

	err = backend2.SaveState(state2, true)
	assert.NoError(t, err)

	// Reload and verify that only state2's changes persisted
	finalState, err := backend1.LoadState(true)
	assert.NoError(t, err)
	assert.Len(t, finalState.Resources, 1)
	assert.True(t, finalState.IsResourceManaged(resource2))
	assert.False(t, finalState.IsResourceManaged(resource1))
}

func TestStateDefaultLocation(t *testing.T) {
	location := stateDefaultLocation()
	assert.NotEmpty(t, location)

	// Check that it contains expected path elements
	switch runtime.GOOS {
	case "linux":
		assert.Contains(t, location, ".local/share/conduktor/cli-state.json")
	case "windows":
		assert.Contains(t, location, "AppData/Roaming/conduktor/cli-state.json")
	case "darwin":
		assert.Contains(t, location, "Library/Application Support/conduktor/cli-state.json")
	}
}
