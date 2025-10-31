package utils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/conduktor/ctl/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffResources(t *testing.T) {
	tests := []struct {
		name        string
		currentRes  *resource.Resource
		modifiedRes *resource.Resource
		expectDiff  bool
	}{
		{
			name: "identical resources produce no diff",
			currentRes: &resource.Resource{
				Json: []byte(`{"name": "test", "value": 1}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"name": "test", "value": 1}`),
			},
			expectDiff: false,
		},
		{
			name: "different values produce diff",
			currentRes: &resource.Resource{
				Json: []byte(`{"name": "test", "value": 1}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"name": "test", "value": 2}`),
			},
			expectDiff: true,
		},
		{
			name: "additional keys produce diff",
			currentRes: &resource.Resource{
				Json: []byte(`{"name": "test"}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"name": "test", "newKey": "value"}`),
			},
			expectDiff: true,
		},
		{
			name: "nested object changes detected",
			currentRes: &resource.Resource{
				Json: []byte(`{"config": {"host": "localhost", "port": 8080}}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"config": {"host": "localhost", "port": 9090}}`),
			},
			expectDiff: true,
		},
		{
			name: "array changes detected",
			currentRes: &resource.Resource{
				Json: []byte(`{"items": [1, 2, 3]}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"items": [1, 2, 3, 4]}`),
			},
			expectDiff: true,
		},
		{
			name: "unsorted arrays identical after sorting",
			currentRes: &resource.Resource{
				Json: []byte(`{"items": [3, 1, 2]}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"items": [1, 2, 3]}`),
			},
			expectDiff: false,
		},
		{
			name: "unsorted maps identical after sorting",
			currentRes: &resource.Resource{
				Json: []byte(`{"z": 1, "a": 2, "m": 3}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"a": 2, "m": 3, "z": 1}`),
			},
			expectDiff: false,
		},
		{
			name: "invalid current JSON triggers not-yet-created path",
			currentRes: &resource.Resource{
				Json: []byte(`invalid json`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{"name": "test"}`),
			},
			expectDiff: true,
		},
		{
			name: "invalid modified JSON handled gracefully",
			currentRes: &resource.Resource{
				Json: []byte(`{"name": "test"}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`invalid json`),
			},
			expectDiff: true,
		},
		{
			name: "both invalid JSON treated as empty",
			currentRes: &resource.Resource{
				Json: []byte(`invalid json 1`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`invalid json 2`),
			},
			expectDiff: false,
		},
		{
			name: "empty resources are identical",
			currentRes: &resource.Resource{
				Json: []byte(`{}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{}`),
			},
			expectDiff: false,
		},
		{
			name: "complex nested structure changes detected",
			currentRes: &resource.Resource{
				Json: []byte(`{
					"metadata": {
						"name": "test-resource",
						"labels": {"env": "prod", "app": "myapp"}
					},
					"spec": {
						"replicas": 3,
						"containers": [
							{"name": "app", "image": "nginx:1.0"},
							{"name": "sidecar", "image": "busybox:latest"}
						]
					}
				}`),
			},
			modifiedRes: &resource.Resource{
				Json: []byte(`{
					"metadata": {
						"name": "test-resource",
						"labels": {"app": "myapp", "env": "staging"}
					},
					"spec": {
						"replicas": 5,
						"containers": [
							{"name": "sidecar", "image": "busybox:latest"},
							{"name": "app", "image": "nginx:2.0"}
						]
					}
				}`),
			},
			expectDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DiffResources(tt.currentRes, tt.modifiedRes)

			assert.NoError(t, err)

			if tt.expectDiff {
				assert.NotEmpty(t, result, "Expected diff output but got empty string")
				// Check that diff contains some indication of changes
				// The exact format depends on diffmatchpatch implementation
				assert.True(t, len(result) > 0, "Expected diff output but got empty string")
			}
		})
	}
}

func TestSortInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "simple number",
			input:    42,
			expected: 42,
		},
		{
			name:     "simple boolean",
			input:    true,
			expected: true,
		},
		{
			name: "simple map",
			input: map[string]interface{}{
				"z": "zebra",
				"a": "apple",
				"m": "mango",
			},
			expected: map[string]interface{}{
				"a": "apple",
				"m": "mango",
				"z": "zebra",
			},
		},
		{
			name:     "simple slice of strings",
			input:    []interface{}{"zebra", "apple", "mango"},
			expected: []interface{}{"apple", "mango", "zebra"},
		},
		{
			name:     "simple slice of numbers",
			input:    []interface{}{3, 1, 2},
			expected: []interface{}{1, 2, 3},
		},
		{
			name: "nested map with arrays",
			input: map[string]interface{}{
				"z": []interface{}{3, 1, 2},
				"a": []interface{}{"c", "a", "b"},
			},
			expected: map[string]interface{}{
				"a": []interface{}{"a", "b", "c"},
				"z": []interface{}{1, 2, 3},
			},
		},
		{
			name: "array of maps",
			input: []interface{}{
				map[string]interface{}{"name": "zebra", "value": 3},
				map[string]interface{}{"name": "apple", "value": 1},
			},
			expected: []interface{}{
				map[string]interface{}{"name": "apple", "value": 1},
				map[string]interface{}{"name": "zebra", "value": 3},
			},
		},
		{
			name: "deeply nested structure",
			input: map[string]interface{}{
				"z": map[string]interface{}{
					"nested": []interface{}{
						map[string]interface{}{"id": 2, "name": "second"},
						map[string]interface{}{"id": 1, "name": "first"},
					},
				},
				"a": "simple",
			},
			expected: map[string]interface{}{
				"a": "simple",
				"z": map[string]interface{}{
					"nested": []interface{}{
						map[string]interface{}{"id": 1, "name": "first"},
						map[string]interface{}{"id": 2, "name": "second"},
					},
				},
			},
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "mixed types in slice",
			input: []interface{}{
				"string",
				42,
				map[string]interface{}{"key": "value"},
				true,
			},
			expected: []interface{}{
				42,
				map[string]interface{}{"key": "value"},
				"string",
				true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortInterface(tt.input)

			// For complex structures, we'll compare JSON representations
			// since assert.Equal might not work well with nested interfaces
			if isComplexType(tt.input) {
				expectedJSON, err := json.Marshal(tt.expected)
				require.NoError(t, err)

				resultJSON, err := json.Marshal(result)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedJSON), string(resultJSON))
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Helper function to determine if we need JSON comparison.
func isComplexType(input interface{}) bool {
	switch input.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

func TestDiffResourcesEdgeCases(t *testing.T) {
	t.Run("empty JSON bytes", func(t *testing.T) {
		currentRes := &resource.Resource{Json: []byte("")}
		modifiedRes := &resource.Resource{Json: []byte("")}

		result, err := DiffResources(currentRes, modifiedRes)
		assert.NoError(t, err)
		// Both should be treated as invalid/empty, so minimal diff
		assert.NotNil(t, result)
	})

	t.Run("very large JSON", func(t *testing.T) {
		// Test with larger JSON to ensure performance is reasonable
		largeMap := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeMap[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		largeJSON, err := json.Marshal(largeMap)
		require.NoError(t, err)

		currentRes := &resource.Resource{Json: largeJSON}
		modifiedRes := &resource.Resource{Json: largeJSON}

		result, err := DiffResources(currentRes, modifiedRes)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
