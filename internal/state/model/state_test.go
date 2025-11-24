package model

import (
	"testing"

	"github.com/conduktor/ctl/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	state := NewState()

	assert.NotNil(t, state)
	assert.Equal(t, StateFileVersion, state.Version)
	assert.NotEmpty(t, state.LastUpdated)
	assert.Empty(t, state.Resources)
}

func TestState_AddManagedResource(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind1",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestState_AddManagedResource-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind2",
		Version:  "v2",
		Metadata: map[string]any{"name": "TestState_AddManagedResource-resource2"},
	}

	state.AddManagedResource(resource1)
	assert.Len(t, state.Resources, 1)

	// adding and already managed resource don't change the state
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
		Metadata: map[string]any{"name": "TestState_IsResourceManaged-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestState_IsResourceManaged-resource2"},
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
		Metadata: map[string]any{"name": "TestState_RemoveManagedResource-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestState_RemoveManagedResource-resource2"},
	}

	// Add both resources
	state.AddManagedResource(resource1)
	state.AddManagedResource(resource2)
	assert.Len(t, state.Resources, 2)

	// Remove resource1
	metadata1 := resource1.Metadata
	state.RemoveManagedResourceVKM(resource1.Version, resource1.Kind, &metadata1)
	assert.Len(t, state.Resources, 1)

	// Check that resource2 is still there
	assert.Equal(t, "TestState_RemoveManagedResource-resource2", (*state.Resources[0].Metadata)["name"])

	// Try to remove non-existent resource
	nonExistentMetadata := map[string]any{"name": "TestState_RemoveManagedResource-non-existent"}
	state.RemoveManagedResourceVKM("v1", "TestKind", &nonExistentMetadata)
	assert.Len(t, state.Resources, 1)
}

func TestState_GetRemovedResources(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestState_GetRemovedResources-resource1"},
	}

	resource2 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestState_GetRemovedResources-resource2"},
	}

	resource3 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestState_GetRemovedResources-resource3"},
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
		assert.Equal(t, "TestState_GetRemovedResources-resource2", removedResources[0].Name)
	}
}

func TestState_GetRemovedResources_EmptyActive(t *testing.T) {
	state := NewState()

	resource1 := resource.Resource{
		Kind:     "TestKind",
		Version:  "v1",
		Metadata: map[string]any{"name": "TestState_GetRemovedResources_EmptyActive-resource1"},
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
		Metadata: map[string]any{"name": "TestState_GetRemovedResources_EmptyState-resource1"},
	}

	// State is empty, but we have active resources
	activeResources := []resource.Resource{resource1}

	removedResources := state.GetRemovedResources(activeResources)

	assert.Empty(t, removedResources)
}

func TestState_ResourceOperationsWithComplexMetadata(t *testing.T) {
	state := NewState()

	// Resource with complex metadata including labels (which should be ignored in comparison)
	complexResource := resource.Resource{
		Kind:    "ComplexKind",
		Version: "v1",
		Metadata: map[string]any{
			"name":      "TestState_ResourceOperationsWithComplexMetadata-complex-resource",
			"namespace": "test-namespace",
			"labels": map[string]any{
				"app":     "test-app",
				"version": "1.0.0",
			},
			"annotations": map[string]any{
				"description": "A complex test resource",
			},
		},
	}

	// Add resource
	state.AddManagedResource(complexResource)

	// Should be managed
	assert.True(t, state.IsResourceManaged(complexResource))

	// Test that labels are ignored in comparison by creating resource with different labels
	resourceWithDifferentLabels := resource.Resource{
		Kind:    "ComplexKind",
		Version: "v1",
		Metadata: map[string]any{
			"name":      "TestState_ResourceOperationsWithComplexMetadata-complex-resource",
			"namespace": "test-namespace",
			"labels": map[string]any{
				"app":     "different-app",
				"version": "2.0.0",
			},
			"annotations": map[string]any{
				"description": "A complex test resource",
			},
		},
	}

	// Should still be considered managed (labels are ignored)
	assert.True(t, state.IsResourceManaged(resourceWithDifferentLabels))
}
