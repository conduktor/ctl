package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/conduktor/ctl/orderedjson"
	"github.com/conduktor/ctl/printutils"
	yamlJson "github.com/ghodss/yaml"
	yaml "gopkg.in/yaml.v3"
)

type Resource struct {
	Json     []byte
	Kind     string
	Name     string //duplicate data from Metadata.Name extracted for convenience
	Version  string
	Metadata map[string]interface{}
	Spec     map[string]interface{}
}

func (r Resource) MarshalJSON() ([]byte, error) {
	return r.Json, nil
}

func (r *Resource) UnmarshalJSON(data []byte) error {
	var forParsingStruct forParsingStruct
	err := json.Unmarshal(data, &forParsingStruct)
	if err != nil {
		return err
	}

	name, err := extractKeyFromMetadataMap(forParsingStruct.Metadata, "name")
	if err != nil {
		return err
	}

	r.Json = data
	r.Kind = forParsingStruct.Kind
	r.Name = name
	r.Version = forParsingStruct.ApiVersion
	r.Metadata = forParsingStruct.Metadata
	r.Spec = forParsingStruct.Spec
	return expendIncludeFiles(r)
}

func (r Resource) String() string {
	return fmt.Sprintf(`version: %s, kind: %s, name: %s, json: '%s'`, r.Version, r.Kind, r.Name, string(r.Json))
}

func (r Resource) StringFromMetadata(key string) (string, error) {
	return extractKeyFromMetadataMap(r.Metadata, key)
}

type forParsingStruct struct {
	ApiVersion string
	Kind       string
	Metadata   map[string]interface{}
	Spec       map[string]interface{}
}

func FromFile(path string) ([]Resource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return FromYamlByte(data)
}

func FromFolder(path string) ([]Resource, error) {
	dirEntry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var result = make([]Resource, 0)
	for _, entry := range dirEntry {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yml") || strings.HasSuffix(entry.Name(), ".yaml")) {
			resources, err := FromFile(filepath.Join(path, entry.Name()))
			result = append(result, resources...)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func FromYamlByte(data []byte) ([]Resource, error) {
	data = expandEnv(data)
	reader := bytes.NewReader(data)
	var yamlData interface{}
	results := make([]Resource, 0, 2)
	d := yaml.NewDecoder(reader)
	for {
		err := d.Decode(&yamlData)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		yamlByte, err := yaml.Marshal(yamlData)
		if err != nil {
			return nil, err
		}
		result, err := yamlByteToResource([]byte(yamlByte))
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// expandEnv replaces ${var} or $var in config according to the values of the current environment variables.
// The replacement is case-sensitive. References to undefined variables are replaced by the empty string.
// A default value can be given by using the form ${var:-default value}.
func expandEnv(data []byte) []byte {
	return []byte(os.Expand(string(data), func(key string) string {
		keyAndDefault := strings.SplitN(key, ":-", 2)
		key = keyAndDefault[0]

		v := os.Getenv(key)
		if v == "" && len(keyAndDefault) == 2 {
			v = keyAndDefault[1] // Set value to the default.
		}
		return v
	}))
}

func extractKeyFromMetadataMap(m map[string]interface{}, key string) (string, error) {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("key %s in metadata is not a string", key)
	}
	return "", fmt.Errorf("key %s not found in metadata", key)

}

func yamlByteToResource(data []byte) (Resource, error) {
	jsonByte, err := yamlJson.YAMLToJSON(data)
	if err != nil {
		return Resource{}, err
	}

	var result Resource
	err = json.Unmarshal(jsonByte, &result)
	return result, err
}

func loadTextFromFile(r *Resource, jsonPathForFilePath string, destJsonPath string) error {
	jsonData, err := gabs.ParseJSON(r.Json)
	if err != nil {
		return err
	}

	pathExist := jsonData.ExistsP(jsonPathForFilePath)
	if pathExist {
		filePath, ok := jsonData.Path(jsonPathForFilePath).Data().(string)
		if !ok {
			return fmt.Errorf("%s is not a string", filePath)
		}
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: \"%w\"", filePath, err)
		}
		jsonData.SetP(string(fileContent), destJsonPath)
		jsonData.DeleteP(jsonPathForFilePath)
		finalJson := []byte(jsonData.String())
		return r.UnmarshalJSON(finalJson)
	}
	return nil
}

func expendIncludeFiles(r *Resource) error {
	// Expend spec.schemaFile into spec.schema and remove spec.schemaFile if kind is Subject
	if r.Kind == "Subject" {
		return loadTextFromFile(r, "spec.schemaFile", "spec.schema")
	} else if r.Kind == "Topic" {
		return loadTextFromFile(r, "metadata.labels.conduktor~1io/descriptionFile", "metadata.labels.conduktor~1io/description")
	}
	return nil
}

func (r *Resource) PrintPreservingOriginalFieldOrder() error {
	var data orderedjson.OrderedData //using this instead of interface{} keep json order
	var finalData interface{}        // in case it does not work we will failback to deserializing directly to interface{}
	err := json.Unmarshal(r.Json, &data)
	if err != nil {
		err = json.Unmarshal(r.Json, &finalData)
		if err != nil {
			return err
		}
	} else {
		finalData = data
	}
	return printutils.PrintResourceLikeYamlFile(os.Stdout, finalData)
}
