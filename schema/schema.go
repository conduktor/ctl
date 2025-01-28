package schema

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/conduktor/ctl/utils"
	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"gopkg.in/yaml.v3"
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

func getKinds[T KindVersion](s *Schema, strict bool, buildKindVersion func(s *Schema, path, kind string, order int, put *v3high.Operation, get *v3high.Operation, strict bool) (T, error)) (map[string]Kind, error) {
	result := make(map[string]Kind, 0)
	for path := s.doc.Model.Paths.PathItems.First(); path != nil; path = path.Next() {
		put := path.Value().Put
		get := path.Value().Get
		if put != nil && get != nil {
			cliTag := findCliTag(path.Value().Put.Tags)
			if cliTag != "" {
				tagParsed, err := parseTag(cliTag)
				if err != nil {
					return nil, err
				}
				newKind, err := buildKindVersion(s, path.Key(), tagParsed.kind, tagParsed.order, put, get, strict)
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

func (s *Schema) GetConsoleKinds(strict bool) (map[string]Kind, error) {
	return getKinds(s, strict, buildConsoleKindVersion)
}

func (s *Schema) GetGatewayKinds(strict bool) (map[string]Kind, error) {
	return getKinds(s, strict, buildGatewayKindVersion)
}

func buildConsoleKindVersion(s *Schema, path, kind string, order int, put *v3high.Operation, get *v3high.Operation, strict bool) (*ConsoleKindVersion, error) {
	newKind := &ConsoleKindVersion{
		Name:               kind,
		ListPath:           path,
		ParentPathParam:    make([]string, 0, len(put.Parameters)),
		ListQueryParameter: make(map[string]QueryParameterOption, len(get.Parameters)),
		Order:              order,
	}
	for _, putParameter := range put.Parameters {
		if putParameter.In == "path" && *putParameter.Required {
			newKind.ParentPathParam = append(newKind.ParentPathParam, putParameter.Name)
		}
		if putParameter.In == "query" && putParameter.Name != "dryMode" {
			newKind.ParentQueryParam = append(newKind.ParentQueryParam, putParameter.Name)
		}
	}
	for _, getParameter := range get.Parameters {
		if getParameter.In == "query" {
			schemaTypes := getParameter.Schema.Schema().Type
			if len(schemaTypes) == 1 {
				schemaType := schemaTypes[0]
				name := getParameter.Name
				newKind.ListQueryParameter[name] = QueryParameterOption{
					FlagName: ComputeFlagName(name),
					Required: *getParameter.Required,
					Type:     schemaType,
				}
			}
		}
	}
	schemaJson, ok := put.RequestBody.Content.Get("application/json")
	if ok && schemaJson.Example != nil {
		data, err := yaml.Marshal(schemaJson.Example)
		if err == nil {
			newKind.ApplyExample = string(data)
		}
	}
	if strict {
		err := checkThatPathParamAreInSpec(newKind, put.RequestBody)
		if err != nil {
			return nil, err
		}

		err = checkThatOrderArePresent(newKind)
		if err != nil {
			return nil, err
		}
	}
	return newKind, nil
}

func buildGatewayKindVersion(s *Schema, path, kind string, order int, put *v3high.Operation, get *v3high.Operation, strict bool) (*GatewayKindVersion, error) {
	//for the moment there is the same but this might evolve latter
	consoleKind, err := buildConsoleKindVersion(s, path, kind, order, put, get, strict)
	if err != nil {
		return nil, err
	}
	var getAvailable = false
	for path := s.doc.Model.Paths.PathItems.First(); path != nil; path = path.Next() {
		get := path.Value().Get
		if get != nil && strings.HasPrefix(path.Key(), consoleKind.ListPath+"/{") {
			getAvailable = true
		}
	}
	return &GatewayKindVersion{
		Name:               consoleKind.Name,
		ListPath:           consoleKind.ListPath,
		ParentPathParam:    consoleKind.ParentPathParam,
		ListQueryParameter: consoleKind.ListQueryParameter,
		ApplyExample:       consoleKind.ApplyExample,
		GetAvailable:       getAvailable,
		Order:              consoleKind.Order,
	}, nil
}

type tagParseResult struct {
	kind    string
	version int
	order   int
}

func parseTag(tag string) (tagParseResult, error) {
	// we extract kind and version from the tag
	// e.g. cli_cluster_kafka_v1 -> kind: Cluster, version: 1
	// e.g. cli_topic-policy_self-serve_v2 -> kind: TopicPolicy, version: 2
	re := regexp.MustCompile(`cli_([^_]+)_[^_]+_v(\d+)(?:_(\d+))?`)
	matches := re.FindStringSubmatch(tag)

	if len(matches) < 4 {
		return tagParseResult{}, fmt.Errorf("Invalid tag format: %s", tag)
	}

	kind := matches[1]
	version, err := strconv.Atoi(matches[2])
	if err != nil {
		return tagParseResult{}, fmt.Errorf("Invalid version number in tag: %s", matches[2])
	}
	orderStr := matches[3]
	var order int
	if orderStr == "" {
		order = DefaultPriority
	} else {
		order, err = strconv.Atoi(orderStr)
	}
	if err != nil {
		return tagParseResult{}, fmt.Errorf("Invalid order number in tag: %s", orderStr)
	}

	finalKind := utils.KebabToUpperCamel(kind)
	if finalKind == "Vclusters" {
		finalKind = "VClusters"
	}
	return tagParseResult{kind: finalKind, version: version, order: order}, nil
}

func checkThatPathParamAreInSpec(kind *ConsoleKindVersion, requestBody *v3high.RequestBody) error {
	if len(kind.ParentPathParam) == 0 {
		return nil
	}
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

func checkThatOrderArePresent(kind *ConsoleKindVersion) error {
	if kind.Order == DefaultPriority {
		return fmt.Errorf("No priority set in schema for kind %s", kind.Name)
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
