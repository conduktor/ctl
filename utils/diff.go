package utils

import (
	"fmt"
	"github.com/conduktor/ctl/resource"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
	"slices"
	"sort"
)

// List of resource kinds that are supported for diffing
var supportedTypes = []string{
	"ServiceAccount",
}

// DiffIsSupported checks if diffing is supported for the given resource type.
func DiffIsSupported(res *resource.Resource) bool {
	if slices.Contains(supportedTypes, res.Kind) {
		return true
	} else {
		return false
	}
}

// PrintDiff dispatches the diff operation for supported resource types
// and prints the resulting diff to stdout.
func PrintDiff(currentResource *resource.Resource, modifiedResource *resource.Resource) error {
	var txt string
	var err error

	switch currentResource.Kind {
	case "ServiceAccount":
		txt, err = DiffServiceAccount(currentResource, modifiedResource)
	default:
		// Unsupported resource kind; nothing to diff
		return nil
	}

	if err != nil {
		return err
	}
	fmt.Printf("%s\n", txt)
	return nil
}

// DiffServiceAccount compares two ServiceAccount objects and returns a unified diff in git-like format.
func DiffServiceAccount(curSa, newSa *resource.Resource) (string, error) {

	// Convert generic resource representations to ServiceAccount-specific structs
	a, err := curSa.ConvertToServiceAccount()
	if err != nil {
		return "", err
	}
	b, err := newSa.ConvertToServiceAccount()
	if err != nil {
		return "", err
	}

	// Sort ACL slices by Name for consistent ordering
	sort.Slice(a.Spec.Authorization.Acls, func(i, j int) bool {
		return a.Spec.Authorization.Acls[i].Name < a.Spec.Authorization.Acls[j].Name
	})
	sort.Slice(b.Spec.Authorization.Acls, func(i, j int) bool {
		return b.Spec.Authorization.Acls[i].Name < b.Spec.Authorization.Acls[j].Name
	})

	// Marshal both structs back to YAML
	yamlA, err := yaml.Marshal(a)
	if err != nil {
		return "", err
	}
	yamlB, err := yaml.Marshal(b)
	if err != nil {
		return "", err
	}

	// Create a new diff instance
	dmp := diffmatchpatch.New()

	// Generate unified diff
	diffs := dmp.DiffMain(string(yamlA), string(yamlB), false)

	// Format the diff nicely
	diffText := dmp.DiffPrettyText(diffs)
	return diffText, nil
}
