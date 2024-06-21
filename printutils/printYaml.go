package printutils

import (
	"fmt"
	"io"
	"slices"

	"github.com/conduktor/ctl/orderedjson"
	orderedmap "github.com/wk8/go-ordered-map/v2"
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

// TODO: delete once backend properly send resource fields in correct order
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
	asMap, isMap := data.(map[string]interface{})
	orderedData, isOrderedData := data.(orderedjson.OrderedData)
	isOrderedMap := false
	var asOrderedMap *orderedmap.OrderedMap[string, orderedjson.OrderedData]
	if isOrderedData {
		asOrderedMap = orderedData.GetMapOrNil()
		isOrderedMap = asOrderedMap != nil
	}
	if isOrderedMap {
		printResourceOrderedMapInCorrectOrder(w, *asOrderedMap)
	} else if isMap {
		printResourceMapInCorrectOrder(w, asMap)
	} else {
		fmt.Fprint(w, string(yamlBytes))
	}
	return err
}

func printResourceMapInCorrectOrder(w io.Writer, dataAsMap map[string]interface{}) {
	wantedKeys := []string{"apiVersion", "kind", "metadata", "spec"}
	for _, wantedKey := range wantedKeys {
		printKeyYaml(w, wantedKey, dataAsMap[wantedKey])
	}
	for otherKey, data := range dataAsMap {
		if !slices.Contains(wantedKeys, otherKey) {
			printKeyYaml(w, otherKey, data)
		}
	}
}

func printResourceOrderedMapInCorrectOrder(w io.Writer, dataAsMap orderedmap.OrderedMap[string, orderedjson.OrderedData]) {
	wantedKeys := []string{"apiVersion", "kind", "metadata", "spec"}
	for _, wantedKey := range wantedKeys {
		value, ok := dataAsMap.Get(wantedKey)
		if ok {
			printKeyYaml(w, wantedKey, value)
		}
	}
	for pair := dataAsMap.Oldest(); pair != nil; pair = pair.Next() {
		if !slices.Contains(wantedKeys, pair.Key) {
			printKeyYaml(w, pair.Key, pair.Value)
		}
	}
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
	case orderedjson.OrderedData:
		array := dataType.GetArrayOrNil()
		if array == nil {
			return printResource(w, data)
		}

		for _, d := range *array {
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
