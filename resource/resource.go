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

func expendIncludeFiles(r *Resource) error {
	// Expend spec.schemaFile into spec.schema and remove spec.schemaFile if kind is Subject
	if r.Kind == "Subject" {
		jsonData, err := gabs.ParseJSON(r.Json)
		if err != nil {
			return err
		}
		schemaFileExist := jsonData.ExistsP("spec.schemaFile")
		if schemaFileExist {
			filePath := jsonData.Path("spec.schemaFile").Data().(string)
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			jsonData.SetP(string(fileContent), "spec.schema")
			jsonData.DeleteP("spec.schemaFile")
			r.Json = []byte(jsonData.String())
		}
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
