package printutils

import (
	"fmt"
	"io"
	"slices"

	yaml "gopkg.in/yaml.v3"
)

func printKeyYaml(w io.Writer, key string, data interface{}) error {
	if data != nil {
		yamlBytes, err := yaml.Marshal(map[string]interface{}{
			key: data,
		})
		if err != nil {
			return err
		}
		fmt.Fprint(w, string(yamlBytes))
	}
	return nil
}

// this print a interface that is expected to a be a resource
// with the following field "version", "kind", "spec", "metadata"
// wit the field in a defined order.
// But in case the given interface is not a map or is a map with more or less field
// than expected we still properly write it
func printResource(w io.Writer, data interface{}) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	asMap, ok := data.(map[string]interface{})
	if !ok {
		fmt.Fprint(w, string(yamlBytes))
	} else {
		wantedKeys := []string{"version", "kind", "metadata", "spec"}
		for _, wantedKey := range wantedKeys {
			printKeyYaml(w, wantedKey, asMap[wantedKey])
		}
		for otherKey, data := range asMap {
			if !slices.Contains(wantedKeys, otherKey) {
				printKeyYaml(w, otherKey, data)
			}
		}
	}
	return err
}

// take a interface that can be a resource or multiple resource
// and print it as the content of a file we could use for an apply
func PrintResourceLikeYamlFile(w io.Writer, data interface{}) error {
	switch dataType := data.(type) {
	case []interface{}:
		for _, d := range dataType {
			fmt.Fprintln(w, "---")
			err := printResource(w, d)
			if err != nil {
				return err
			}
		}
	default:
		return printResource(w, data)
	}
	return nil
}
