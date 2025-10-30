package utils

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/conduktor/ctl/internal/resource"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

// PrintDiff dispatches the diff operation for supported resource types
// and prints the resulting diff to stdout.
func PrintDiff(currentResource *resource.Resource, modifiedResource *resource.Resource) error {
	txt, err := DiffResources(currentResource, modifiedResource)

	if err != nil {
		return err
	}
	fmt.Printf("%s\n", txt)
	return nil
}

// DiffResources compares two Resources objects and returns a unified diff in git-like format.
func DiffResources(curRes, newRes *resource.Resource) (string, error) {
	var curResObj, newResObj interface{}
	var err error

	isNotYetCreated := false

	// unmarshall the JSON data into a generic interface{}
	err = json.Unmarshal(curRes.Json, &curResObj)
	if err != nil {
		isNotYetCreated = true
	}
	err = json.Unmarshal(newRes.Json, &newResObj)
	if err != nil {
		emptyRes := resource.Resource{}
		_ = json.Unmarshal(emptyRes.Json, &newResObj)
	}

	// recursively sort the maps and slices in generic interface{}
	sortedCurRes := sortInterface(curResObj)
	sortedNewRes := sortInterface(newResObj)

	// Marshal both structs back to YAML
	yamlCurRes, err := yaml.Marshal(sortedCurRes)
	if err != nil {
		return "", err
	}
	yamlNewRes, err := yaml.Marshal(sortedNewRes)
	if err != nil {
		return "", err
	}

	// Create a new diff instance
	dmp := diffmatchpatch.New()

	if isNotYetCreated {
		yamlCurRes = []byte{0}
	}
	// Generate unified diff
	diffs := dmp.DiffMain(string(yamlCurRes), string(yamlNewRes), false)

	// Check if there are any actual differences
	hasChanges := false
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			hasChanges = true
			break
		}
	}

	// Return empty string if no changes
	if !hasChanges {
		return "", nil
	}

	// Format the diff nicely
	diffText := dmp.DiffPrettyText(diffs)
	return "\n" + diffText, nil
}

func sortInterface(input interface{}) interface{} {
	switch v := input.(type) {
	case map[string]interface{}:
		// Sort map keys
		sortedMap := make(map[string]interface{})
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sortedMap[key] = sortInterface(v[key])
		}
		return sortedMap
	case []interface{}:
		// Sort slices
		for i := range v {
			v[i] = sortInterface(v[i])
		}
		sort.SliceStable(v, func(i, j int) bool {
			return fmt.Sprintf("%v", v[i]) < fmt.Sprintf("%v", v[j])
		})
		return v
	default:
		// Return other types as-is
		return v
	}
}
