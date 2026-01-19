package schema

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"

	"github.com/conduktor/ctl/internal/utils"
	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"gopkg.in/yaml.v3"
)

type OpenAPIParser struct {
	doc *libopenapi.DocumentModel[v3high.Document]
}

func NewOpenAPIParser(schema []byte) (*OpenAPIParser, error) {
	doc, err := libopenapi.NewDocument(schema)
	if err != nil {
		return nil, err
	}
	v3Model, err := doc.BuildV3Model()
	if err != nil {
		return nil, err
	}

	return &OpenAPIParser{
		doc: v3Model,
	}, nil
}

func getKinds[T KindVersion](s *OpenAPIParser, strict bool, buildKindVersion func(s *OpenAPIParser, path, kind string, order int, put *v3high.Operation, get *v3high.Operation, strict bool) (T, error)) (map[string]Kind, error) {
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

func (s *OpenAPIParser) getRuns(backendType BackendType) (RunCatalog, error) {
	result := make(RunCatalog, 0)
	for path := s.doc.Model.Paths.PathItems.First(); path != nil; path = path.Next() {
		err := handleExecuteOperation(backendType, path.Key(), path.Value().Get, resty.MethodGet, result)
		if err != nil {
			return nil, err
		}
		err = handleExecuteOperation(backendType, path.Key(), path.Value().Post, resty.MethodPost, result)
		if err != nil {
			return nil, err
		}
		err = handleExecuteOperation(backendType, path.Key(), path.Value().Put, resty.MethodPut, result)
		if err != nil {
			return nil, err
		}
		err = handleExecuteOperation(backendType, path.Key(), path.Value().Delete, resty.MethodDelete, result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

//nolint:unparam
func handleExecuteOperation(backendType BackendType, path string, operation *v3high.Operation, method string, result RunCatalog) error {
	if operation == nil {
		return nil
	}

	nameYaml, present := operation.Extensions.Get("x-cdk-run-name")
	if !present {
		return nil
	}
	name := nameYaml.Value
	docYaml, docPresent := operation.Extensions.Get("x-cdk-run-doc")
	var doc string
	if docPresent {
		doc = docYaml.Value
	} else {
		doc = ""
	}
	run := Run{
		BackendType:    backendType,
		Path:           path,
		Name:           name,
		Doc:            doc,
		QueryParameter: make(map[string]FlagParameterOption, len(operation.Parameters)),
		PathParameter:  make([]string, 0, len(operation.Parameters)),
		Method:         method,
	}
	for _, parameter := range operation.Parameters {
		if parameter.In == "path" && *parameter.Required {
			run.PathParameter = append(run.PathParameter, parameter.Name)
		}
		if parameter.In == "query" {
			schemaTypes := parameter.Schema.Schema().Type
			if len(schemaTypes) == 1 {
				queryName := parameter.Name
				run.QueryParameter[queryName] = FlagParameterOption{
					FlagName: computeFlagName(queryName),
					Required: *parameter.Required,
					Type:     schemaTypes[0],
				}
			}
		}
	}
	run.BodyFields = computeBodyFields(operation.RequestBody)
	result[name] = run
	return nil
}

func computeBodyFields(body *v3high.RequestBody) map[string]FlagParameterOption {
	var result = make(map[string]FlagParameterOption)
	if body == nil {
		return result
	}
	jsonMediaType, present := body.Content.Get("application/json")
	if present && jsonMediaType.Schema.Schema() != nil && jsonMediaType.Schema.Schema().Properties != nil {
		bodySchema := jsonMediaType.Schema.Schema()
		for propertiesPair := bodySchema.Properties.First(); propertiesPair != nil; propertiesPair = propertiesPair.Next() {
			key := propertiesPair.Key()
			value := propertiesPair.Value()
			if value != nil && value.Schema() != nil {
				valueType := value.Schema().Type[0]
				if valueType == "string" || valueType == "boolean" || valueType == "integer" {
					result[key] = FlagParameterOption{
						FlagName: computeFlagName(key),
						Type:     valueType,
						Required: slices.Contains(bodySchema.Required, key),
					}
				}
			}
		}
	}
	return result
}

func (s *OpenAPIParser) GetConsoleCatalog(strict bool) (*Catalog, error) {
	kinds, err := s.GetConsoleKinds(strict)
	if err != nil {
		return nil, err
	}
	runs, err := s.getRuns(CONSOLE)
	if err != nil {
		return nil, err
	}
	return &Catalog{
		Kind: kinds,
		Run:  runs,
	}, nil
}

func (s *OpenAPIParser) GetConsoleKinds(strict bool) (KindCatalog, error) {
	return getKinds(s, strict, buildConsoleKindVersion)
}

func (s *OpenAPIParser) GetGatewayKinds(strict bool) (KindCatalog, error) {
	return getKinds(s, strict, buildGatewayKindVersion)
}

func (s *OpenAPIParser) GetGatewayCatalog(strict bool) (*Catalog, error) {
	kinds, err := s.GetGatewayKinds(strict)
	if err != nil {
		return nil, err
	}
	runs, err := s.getRuns(GATEWAY)
	if err != nil {
		return nil, err
	}

	return &Catalog{
		Kind: kinds,
		Run:  runs,
	}, nil
}

func buildConsoleKindVersion(s *OpenAPIParser, path, kind string, order int, put *v3high.Operation, get *v3high.Operation, strict bool) (*ConsoleKindVersion, error) {
	newKind := &ConsoleKindVersion{
		Name:               kind,
		ListPath:           path,
		ParentPathParam:    make([]string, 0, len(put.Parameters)),
		ListQueryParameter: make(map[string]FlagParameterOption, len(get.Parameters)),
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
				newKind.ListQueryParameter[name] = FlagParameterOption{
					FlagName: computeFlagName(name),
					Required: *getParameter.Required,
					Type:     schemaType,
				}
			}
		}
	}
	schemaJSON, ok := put.RequestBody.Content.Get("application/json")
	if ok && schemaJSON.Example != nil {
		// Example is a *yaml.Node, we need to decode it first then marshal
		var example interface{}
		if err := schemaJSON.Example.Decode(&example); err == nil {
			data, err := yaml.Marshal(example)
			if err == nil {
				newKind.ApplyExample = string(data)
			}
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

// two logics: uniformize flag name and kebab case.
func computeFlagName(name string) string {
	kebab := utils.CamelToKebab(name)
	kebab = strings.TrimPrefix(kebab, "filter-by-")
	return strings.Replace(kebab, "app-instance", "application-instance", 1)
}

func buildGatewayKindVersion(s *OpenAPIParser, path, kind string, order int, put *v3high.Operation, get *v3high.Operation, strict bool) (*GatewayKindVersion, error) {
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
		return tagParseResult{}, fmt.Errorf("invalid tag format: %s", tag)
	}

	kind := matches[1]
	version, err := strconv.Atoi(matches[2])
	if err != nil {
		return tagParseResult{}, fmt.Errorf("invalid version number in tag: %s", matches[2])
	}
	orderStr := matches[3]
	var order int
	if orderStr == "" {
		order = DefaultPriority
	} else {
		order, err = strconv.Atoi(orderStr)
	}
	if err != nil {
		return tagParseResult{}, fmt.Errorf("invalid order number in tag: %s", orderStr)
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
		return fmt.Errorf("no application/json content for kind %s", kind.Name)
	}
	schema := jsonContent.Schema.Schema()
	metadata, ok := schema.Properties.Get("metadata")
	if !ok {
		return fmt.Errorf("no metadata in schema for kind %s", kind.Name)
	}
	for _, parentPath := range kind.ParentPathParam {
		schema := metadata.Schema()
		_, ok := schema.Properties.Get(parentPath)
		if !ok {
			return fmt.Errorf("parent path param %s not found in metadata for kind %s", parentPath, kind.Name)
		}
		if !slices.Contains(schema.Required, parentPath) {
			return fmt.Errorf("parent path param %s in metadata for kind %s not required", parentPath, kind.Name)
		}

	}
	return nil
}

func checkThatOrderArePresent(kind *ConsoleKindVersion) error {
	if kind.Order == DefaultPriority {
		return fmt.Errorf("no priority set in schema for kind %s", kind.Name)
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
