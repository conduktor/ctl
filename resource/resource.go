package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	yamlJson "github.com/ghodss/yaml"
	yaml "gopkg.in/yaml.v3"
)

type Resource struct {
	Json    []byte
	Kind    string
	Name    string
	Version string
}

func (r Resource) String() string {
	return fmt.Sprintf(`version: %s, kind: %s, name: %s, json: '%s'`, r.Version, r.Kind, r.Name, string(r.Json))
}

type yamlRoot struct {
	Version  string
	Kind     string
	Metadata metadata
}

type metadata struct {
	Name string
}

func FromFile(path string) ([]Resource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return FromByte(data)
}

func FromFolder(path string) ([]Resource, error) {
	dirEntry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var result = make([]Resource, 0, 0)
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

func FromByte(data []byte) ([]Resource, error) {
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

func yamlByteToResource(data []byte) (Resource, error) {
	jsonByte, err := yamlJson.YAMLToJSON(data)
	if err != nil {
		return Resource{}, nil
	}

	var yamlRoot yamlRoot
	err = json.Unmarshal(jsonByte, &yamlRoot)
	if err != nil {
		return Resource{}, nil
	}

	return Resource{Json: jsonByte, Kind: yamlRoot.Kind, Name: yamlRoot.Metadata.Name, Version: yamlRoot.Version}, nil
}
