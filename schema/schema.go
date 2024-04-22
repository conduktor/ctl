package schema

import (
	"fmt"
	"github.com/conduktor/ctl/utils"
	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Schema struct {
	doc *libopenapi.DocumentModel[v3high.Document]
}

func New(schema []byte) (*Schema, error) {
	doc, err := libopenapi.NewDocument(schema)
	if err != nil {
		return nil, err
	}
	v3Model, errors := doc.BuildV3Model()
	if len(errors) > 0 {
		return nil, errors[0]
	}

	return &Schema{
		doc: v3Model,
	}, nil
}

func (s *Schema) GetKinds(strict bool) (map[string]Kind, error) {
	result := make(map[string]Kind, 0)
	for path := s.doc.Model.Paths.PathItems.First(); path != nil; path = path.Next() {
		put := path.Value().Put
		if put != nil {
			cliTag := findCliTag(path.Value().Put.Tags)
			if cliTag != "" {
				tagParsed, err := parseTag(cliTag)
				if err != nil {
					return nil, err
				}
				newKind, err := buildKindVersion(path.Key(), tagParsed.kind, put, strict)
				if err != nil {
					return nil, err
				}
				prec, kindAlreadyFound := result[tagParsed.kind]
				if kindAlreadyFound {
					err = prec.AddVersion(tagParsed.version, newKind)
					if err != nil {
						return nil, err
					}
				} else {
					result[tagParsed.kind] = NewKind(tagParsed.version, newKind)
				}
			}
		}
	}
	return result, nil
}

func buildKindVersion(path, kind string, put *v3high.Operation, strict bool) (*KindVersion, error) {
	newKind := &KindVersion{
		Name:            kind,
		ListPath:        path,
		ParentPathParam: make([]string, 0, len(put.Parameters)),
	}
	for _, parameter := range put.Parameters {
		if parameter.In == "path" && *parameter.Required {
			newKind.ParentPathParam = append(newKind.ParentPathParam, parameter.Name)

		}
	}
	if strict {
		err := checkThatPathParamAreInSpec(newKind, put.RequestBody)
		if err != nil {
			return nil, err
		}
	}
	return newKind, nil
}

type tagParseResult struct {
	kind    string
	version int
}

func parseTag(tag string) (tagParseResult, error) {
	// we extract kind and version from the tag
	// e.g. cli_cluster_kafka_v1 -> kind: Cluster, version: 1
	// e.g. cli_topic-policy_self-serve_v2 -> kind: TopicPolicy, version: 2
	re := regexp.MustCompile(`cli_([^_]+)_[^_]+_v(\d+)`)
	matches := re.FindStringSubmatch(tag)

	if len(matches) < 3 {
		return tagParseResult{}, fmt.Errorf("Invalid tag format: %s", tag)
	}

	kind := matches[1]
	version, err := strconv.Atoi(matches[2])
	if err != nil {
		return tagParseResult{}, fmt.Errorf("Invalid version number in tag: %s", matches[2])
	}

	return tagParseResult{kind: utils.KebabToUpperCamel(kind), version: version}, nil
}

func checkThatPathParamAreInSpec(kind *KindVersion, requestBody *v3high.RequestBody) error {
	jsonContent, ok := requestBody.Content.Get("application/json")
	if !ok {
		return fmt.Errorf("No application/json content for kind %s", kind.Name)
	}
	schema := jsonContent.Schema.Schema()
	metadata, ok := schema.Properties.Get("metadata")
	if !ok {
		return fmt.Errorf("No metadata in schema for kind %s", kind.Name)
	}
	for _, parentPath := range kind.ParentPathParam {
		schema := metadata.Schema()
		_, ok := schema.Properties.Get(parentPath)
		if !ok {
			return fmt.Errorf("Parent path param %s not found in metadata for kind %s", parentPath, kind.Name)
		}
		if !slices.Contains(schema.Required, parentPath) {
			return fmt.Errorf("Parent path param %s in metadata for kind %s not required", parentPath, kind.Name)
		}

	}
	return nil
}

func findCliTag(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "cli_") {
			return tag
		}
	}
	return ""
}
