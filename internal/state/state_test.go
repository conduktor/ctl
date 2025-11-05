package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/conduktor/ctl/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func tmpStateLocation(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "state_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return filepath.Join(tempDir, "state.json")
}

func TestNewState(t *testing.T) {
	state := NewState()

	assert.NotNil(t, state)
	assert.Equal(t, StateVersion, state.Version)
	assert.Empty(t, state.LastUpdated)
	assert.Empty(t, state.Resources)
	assert.Equal(t, StateDefaultLocation(), state.StateLocation)
}

func TestLoadStateFromFile_NewFile(t *testing.T) {
	// Create temporary directory
	stateFilePath := tmpStateLocation(t)

	// Load state from non-existent file
	state, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, stateFilePath, state.StateLocation)
	assert.Equal(t, StateVersion, state.Version)
	assert.Empty(t, state.LastUpdated)
	assert.Empty(t, state.Resources)
}

func TestLoadStateFromFile_ExistingFile(t *testing.T) {
	// Create temporary directory
	stateFilePath := tmpStateLocation(t)

	// Create initial state and save it
	initialState := &State{
		StateLocation: stateFilePath,
		Version:       StateVersion,
		LastUpdated:   "2024-01-01T00:00:00Z",
		Resources: []ResourceState{
			{
				APIVersion: "v1",
				Kind:       "TestKind",
				Metadata:   &map[string]any{"name": "TestLoadStateFromFile_ExistingFile-test-resource"},
			},
		},
	}

	err := initialState.SaveToFile()
	assert.NoError(t, err)

	// Load state from existing file
	loadedState, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedState)
	assert.Equal(t, StateVersion, loadedState.Version)
	assert.NotEmpty(t, loadedState.LastUpdated)
	assert.Len(t, loadedState.Resources, 1)
	assert.Equal(t, "TestKind", loadedState.Resources[0].Kind)
}

func TestState_SaveToFile(t *testing.T) {
	// Create temporary directory
	stateFilePath := tmpStateLocation(t)

	state := &State{
		StateLocation: stateFilePath,
		Version:       StateVersion,
		Resources:     []ResourceState{},
	}

	beforeSave := time.Now()
	err := state.SaveToFile()
	assert.NoError(t, err)

	// Check that file exists
	assert.FileExists(t, stateFilePath)

	// Check that LastUpdated was set
	assert.NotEmpty(t, state.LastUpdated)

	// Parse the saved time and check it's within reasonable bounds
	savedTime, err := time.Parse(time.RFC3339, state.LastUpdated)
	assert.NoError(t, err)

	// Check that the saved time is within a reasonable window (allow 5 second difference)
	timeDiff := savedTime.Sub(beforeSave.UTC())
	assert.True(t, timeDiff >= -time.Second && timeDiff <= 5*time.Second,
		"LastUpdated time %v is not within reasonable bounds relative to test start %v (diff: %v)",
		savedTime, beforeSave.UTC(), timeDiff)

	// Read the file and verify content
	data, err := os.ReadFile(stateFilePath)
	assert.NoError(t, err)

	var savedState State
	err = json.Unmarshal(data, &savedState)
	assert.NoError(t, err)
	assert.Equal(t, StateVersion, savedState.Version)
}

func TestState_AddManagedResource(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind1",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_AddManagedResource-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind2",
		Version:  "v2",
		Metadata: map[string]interface{}{"name": "TestState_AddManagedResource-resource2"},
	}

	state.AddManagedResource(resource1)
	assert.Len(t, state.Resources, 1)

	state.AddManagedResource(resource2)
	assert.Len(t, state.Resources, 2)

	// Check first resource
	assert.Equal(t, "TestKind1", state.Resources[0].Kind)

	// Check second resource
	assert.Equal(t, "TestKind2", state.Resources[1].Kind)
}

func TestState_IsResourceManaged(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_IsResourceManaged-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_IsResourceManaged-resource2"},
	}

	// Resource not managed yet
	assert.False(t, state.IsResourceManaged(resource1))

	// Add resource1
	state.AddManagedResource(resource1)

	// Resource1 should now be managed
	assert.True(t, state.IsResourceManaged(resource1))

	// Resource2 should not be managed
	assert.False(t, state.IsResourceManaged(resource2))
}

func TestState_RemoveManagedResource(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_RemoveManagedResource-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_RemoveManagedResource-resource2"},
	}

	// Add both resources
	state.AddManagedResource(resource1)
	state.AddManagedResource(resource2)
	assert.Len(t, state.Resources, 2)

	// Remove resource1
	metadata1 := resource1.Metadata
	state.RemoveManagedResource(resource1.Version, resource1.Kind, &metadata1)
	assert.Len(t, state.Resources, 1)

	// Check that resource2 is still there
	assert.Equal(t, "TestState_RemoveManagedResource-resource2", (*state.Resources[0].Metadata)["name"])

	// Try to remove non-existent resource
	nonExistentMetadata := map[string]interface{}{"name": "TestState_RemoveManagedResource-non-existent"}
	state.RemoveManagedResource("v1", "TestKind", &nonExistentMetadata)
	assert.Len(t, state.Resources, 1)
}

func TestState_GetRemovedResources(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_GetRemovedResources-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_GetRemovedResources-resource2"},
	}

	resource3 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_GetRemovedResources-resource3"},
	}

	// Add all resources to state
	state.AddManagedResource(resource1)
	state.AddManagedResource(resource2)
	state.AddManagedResource(resource3)

	// Current active resources only include resource1 and resource3
	activeResources := []resource.Resource{resource1, resource3}

	removedResources := state.GetRemovedResources(activeResources)

	assert.Len(t, removedResources, 1)
	if len(removedResources) > 0 {
		assert.Equal(t, "TestState_GetRemovedResources-resource2", (*removedResources[0].Metadata)["name"])
	}
}

func TestState_GetRemovedResources_EmptyActive(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_GetRemovedResources_EmptyActive-resource1"},
	}

	state.AddManagedResource(resource1)

	// No active resources
	activeResources := []resource.Resource{}

	removedResources := state.GetRemovedResources(activeResources)

	assert.Len(t, removedResources, 1)
}

func TestState_GetRemovedResources_EmptyState(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_GetRemovedResources_EmptyState-resource1"},
	}

	// State is empty, but we have active resources
	activeResources := []resource.Resource{resource1}

	removedResources := state.GetRemovedResources(activeResources)

	assert.Empty(t, removedResources)
}

func TestState_IntegrationWithTempDir(t *testing.T) {
	stateFilePath := tmpStateLocation(t)

	// Create and populate state
	state, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_IntegrationWithTempDir-resource1"},
	}

	state.AddManagedResource(resource1)
	err = state.SaveToFile()
	assert.NoError(t, err)

	// Load state again and verify
	reloadedState, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)
	assert.Len(t, reloadedState.Resources, 1)
	assert.True(t, reloadedState.IsResourceManaged(resource1))

	// Add another resource and save
	resource2 := resource.Resource{
		Kind:     "TestKind2",
		Version:  "v2",
		Metadata: map[string]interface{}{"name": "TestState_IntegrationWithTempDir-resource2"},
	}

	reloadedState.AddManagedResource(resource2)
	err = reloadedState.SaveToFile()
	assert.NoError(t, err)

	// Final reload and verification
	finalState, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)
	assert.Len(t, finalState.Resources, 2)
	assert.True(t, finalState.IsResourceManaged(resource1))
	assert.True(t, finalState.IsResourceManaged(resource2))
}

func TestStateDefaultLocation(t *testing.T) {
	location := StateDefaultLocation()
	assert.NotEmpty(t, location)

	// Check that it contains expected path elements
	assert.Contains(t, location, ".conduktor")
	assert.Contains(t, location, "ctl")
	assert.Contains(t, location, StateFileName)
}

func TestLoadStateFromFile_EmptyPath(t *testing.T) {
	emptyStateLocation := ""
	state, err := LoadStateFromFile(&emptyStateLocation)
	assert.NoError(t, err)
	assert.Equal(t, StateDefaultLocation(), state.StateLocation)

	state2, err := LoadStateFromFile(nil)
	assert.NoError(t, err)
	assert.Equal(t, StateDefaultLocation(), state2.StateLocation)
}

func TestLoadStateFromFile_CorruptedFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "state_test_corrupted_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	stateFilePath := filepath.Join(tempDir, "state.json")

	// Write corrupted JSON to file
	corruptedJSON := `{"version": "v1", "lastUpdated": "2024-01-01", "resources": [`
	err = os.WriteFile(stateFilePath, []byte(corruptedJSON), 0644)
	assert.NoError(t, err)

	// Try to load corrupted file
	_, err = LoadStateFromFile(&stateFilePath)
	assert.Error(t, err)
}

func TestState_SaveToFile_ReadOnlyDirectory(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "state_test_readonly_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Make directory read-only
	err = os.Chmod(tempDir, 0444)
	assert.NoError(t, err)

	// Restore permissions for cleanup
	defer func() {
		err := os.Chmod(tempDir, 0755)
		assert.NoError(t, err)
	}()

	stateFilePath := filepath.Join(tempDir, "state.json")

	state := &State{
		StateLocation: stateFilePath,
		Version:       StateVersion,
		Resources:     []ResourceState{},
	}

	// Should fail to save to read-only directory
	err = state.SaveToFile()
	assert.Error(t, err)
}

func TestState_ResourceOperationsWithComplexMetadata(t *testing.T) {
	stateFilePath := tmpStateLocation(t)

	state, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)

	// Resource with complex metadata including labels (which should be ignored in comparison)
	complexResource := resource.Resource{
		Kind:    "ComplexKind",
		Version: "v1",
		Metadata: map[string]interface{}{
			"name":      "TestState_ResourceOperationsWithComplexMetadata-complex-resource",
			"namespace": "test-namespace",
			"labels": map[string]interface{}{
				"app":     "test-app",
				"version": "1.0.0",
			},
			"annotations": map[string]interface{}{
				"description": "A complex test resource",
			},
		},
	}

	// Add resource
	state.AddManagedResource(complexResource)

	// Save and reload
	err = state.SaveToFile()
	assert.NoError(t, err)

	reloadedState, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)

	// Should still be managed
	assert.True(t, reloadedState.IsResourceManaged(complexResource))

	// Test that labels are ignored in comparison by creating resource with different labels
	resourceWithDifferentLabels := resource.Resource{
		Kind:    "ComplexKind",
		Version: "v1",
		Metadata: map[string]interface{}{
			"name":      "TestState_ResourceOperationsWithComplexMetadata-complex-resource",
			"namespace": "test-namespace",
			"labels": map[string]interface{}{
				"app":     "different-app",
				"version": "2.0.0",
			},
			"annotations": map[string]interface{}{
				"description": "A complex test resource",
			},
		},
	}

	// Should still be considered managed (labels are ignored)
	assert.True(t, reloadedState.IsResourceManaged(resourceWithDifferentLabels))
}

func TestState_ConcurrentStateOperations(t *testing.T) {
	// Create temporary directory
	stateFilePath := tmpStateLocation(t)

	// Create multiple states pointing to same file
	state1, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)
	state1.StateLocation = stateFilePath

	state2, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)
	state2.StateLocation = stateFilePath

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_ConcurrentStateOperations-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]interface{}{"name": "TestState_ConcurrentStateOperations-resource2"},
	}

	// Add different resources to each state
	state1.AddManagedResource(resource1)
	state2.AddManagedResource(resource2)

	// Save both (state2 will overwrite state1's changes)
	err = state1.SaveToFile()
	assert.NoError(t, err)

	err = state2.SaveToFile()
	assert.NoError(t, err)

	// Reload and verify that only state2's changes persisted
	finalState, err := LoadStateFromFile(&stateFilePath)
	assert.NoError(t, err)
	assert.Len(t, finalState.Resources, 1)
	assert.True(t, finalState.IsResourceManaged(resource2))
	assert.False(t, finalState.IsResourceManaged(resource1))
}
