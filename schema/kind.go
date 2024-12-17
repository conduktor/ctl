package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/utils"
)

type KindVersion interface {
	GetListPath() string
	GetName() string
	GetParentPathParam() []string
	GetOrder() int
	GetListQueryParamter() map[string]QueryParameterOption
	GetApplyExample() string
}

// two logics: uniformize flag name and kebab case
func ComputeFlagName(name string) string {
	kebab := utils.UpperCamelToKebab(name)
	kebab = strings.TrimPrefix(kebab, "filter-by-")
	return strings.Replace(kebab, "app-instance", "application-instance", 1)
}

type QueryParameterOption struct {
	FlagName string
	Required bool
	Type     string
}
type ConsoleKindVersion struct {
	ListPath          string
	Name              string
	ParentPathParam   []string
	ListQueryParamter map[string]QueryParameterOption
	ApplyExample      string
	Order             int `json:1000` //same value DefaultPriority
}

func (c *ConsoleKindVersion) GetListPath() string {
	return c.ListPath
}

func (c *ConsoleKindVersion) GetApplyExample() string {
	return c.ApplyExample
}

func (c *ConsoleKindVersion) GetName() string {
	return c.Name
}

func (c *ConsoleKindVersion) GetParentPathParam() []string {
	return c.ParentPathParam
}

func (c *ConsoleKindVersion) GetOrder() int {
	return c.Order
}

func (c *ConsoleKindVersion) GetListQueryParamter() map[string]QueryParameterOption {
	return c.ListQueryParamter
}

type GetParameter struct {
	Name      string
	Mandatory bool
}

type GatewayKindVersion struct {
	ListPath           string
	Name               string
	ParentPathParam    []string
	ListQueryParameter map[string]QueryParameterOption
	GetAvailable       bool
	ApplyExample       string
	Order              int `json:1000` //same value DefaultPriority
}

func (g *GatewayKindVersion) GetListPath() string {
	return g.ListPath
}

func (g *GatewayKindVersion) GetName() string {
	return g.Name
}

func (g *GatewayKindVersion) GetParentPathParam() []string {
	return g.ParentPathParam
}

func (g *GatewayKindVersion) GetApplyExample() string {
	return g.ApplyExample
}

func (g *GatewayKindVersion) GetOrder() int {
	return g.Order
}

func (g *GatewayKindVersion) GetListQueryParamter() map[string]QueryParameterOption {
	return g.ListQueryParameter
}

const DefaultPriority = 1000 //update  json annotation for Order when changing this value

type Kind struct {
	Versions map[int]KindVersion
}

type KindCatalog = map[string]Kind

//go:embed console-default-schema.json
var consoleDefaultByteSchema []byte

//go:embed gateway-default-schema.json
var gatewayDefaultByteSchema []byte

type KindGeneric[T KindVersion] struct {
	Versions map[int]T
}

func buildKindCatalogFromByteSchema[T KindVersion](byteSchema []byte) KindCatalog {
	var jsonResult map[string]KindGeneric[T]
	err := json.Unmarshal(byteSchema, &jsonResult)
	if err != nil {
		panic(err)
	}
	var result KindCatalog = make(map[string]Kind)
	for kindName, kindGeneric := range jsonResult {
		kind := Kind{
			Versions: make(map[int]KindVersion),
		}
		for version, kindVersion := range kindGeneric.Versions {
			kind.Versions[version] = kindVersion
		}
		result[kindName] = kind
	}
	return result
}

func ConsoleDefaultKind() KindCatalog {
	return buildKindCatalogFromByteSchema[*ConsoleKindVersion](consoleDefaultByteSchema)
}

func GatewayDefaultKind() KindCatalog {
	return buildKindCatalogFromByteSchema[*GatewayKindVersion](gatewayDefaultByteSchema)
}

func NewKind(version int, kindVersion KindVersion) Kind {
	return Kind{
		Versions: map[int]KindVersion{version: kindVersion},
	}
}

func extractVersionFromApiVersion(apiVersion string) int {
	// we extract the number after v in a apiVersion
	// e.g. v1 1
	// e.g. v42-> 42

	re := regexp.MustCompile(`v(\d+)`)
	matches := re.FindStringSubmatch(apiVersion)

	if len(matches) < 2 {
		fmt.Fprintf(os.Stderr, "Invalid api version format \"%s\", could not extract version\n", apiVersion)
		os.Exit(1)
	}

	version, err := strconv.Atoi(matches[1])
	if err != nil {
		panic(fmt.Sprintf("Invalid version number in apiVersion: %s", matches[1]))
	}

	return version
}

func (kind *Kind) AddVersion(version int, kindVersion KindVersion) error {
	name := kind.GetName()
	if name != kindVersion.GetName() {
		return fmt.Errorf("Adding kind version of kind %s to different kind %s", kindVersion.GetName(), name)
	}
	kind.Versions[version] = kindVersion
	return nil
}

func (kind *Kind) GetParentFlag() []string {
	kindVersion := kind.GetLatestKindVersion()
	return kindVersion.GetParentPathParam()
}

func (kind *Kind) GetListFlag() map[string]QueryParameterOption {
	kindVersion := kind.GetLatestKindVersion()
	return kindVersion.GetListQueryParamter()
}

func (kind *Kind) MaxVersion() int {
	maxVersion := -1
	for version := range kind.Versions {
		if version > maxVersion {
			maxVersion = version
		}
	}
	return maxVersion
}

func (kind *Kind) GetLatestKindVersion() KindVersion {
	kindVersion, ok := kind.Versions[kind.MaxVersion()]
	if !ok {
		panic("Max numVersion on kind return a numVersion that does not exist")
	}
	return kindVersion
}

func (Kind *Kind) GetName() string {
	for _, kindVersion := range Kind.Versions {
		return kindVersion.GetName()
	}
	panic("No kindVersion in kind") //should never happen
}

func (kind *Kind) ListPath(parentPathValues []string) string {
	kindVersion := kind.GetLatestKindVersion()
	if len(parentPathValues) != len(kindVersion.GetParentPathParam()) {
		panic(fmt.Sprintf("For kind %s expected %d parent apiVersion values, got %d", kindVersion.GetName(), len(kindVersion.GetParentPathParam()), len(parentPathValues)))
	}
	path := kindVersion.GetListPath()
	for i, pathValue := range parentPathValues {
		path = strings.Replace(path, fmt.Sprintf("{%s}", kindVersion.GetParentPathParam()[i]), pathValue, 1)
	}
	return path
}

func (kind *Kind) DescribePath(parentPathValues []string, name string) string {
	return kind.ListPath(parentPathValues) + "/" + name
}

func (kind *Kind) ApplyPath(resource *resource.Resource) (string, error) {
	kindVersion, ok := kind.Versions[extractVersionFromApiVersion(resource.Version)]
	if !ok {
		return "", fmt.Errorf("Could not find version %s for kind %s", resource.Version, resource.Kind)
	}
	parentPathValues := make([]string, len(kindVersion.GetParentPathParam()))
	var err error
	for i, param := range kindVersion.GetParentPathParam() {
		parentPathValues[i], err = resource.StringFromMetadata(param)
		if err != nil {
			return "", err
		}
	}
	return kind.ListPath(parentPathValues), nil
}

func (kind *Kind) DeletePath(resource *resource.Resource) (string, error) {
	applyPath, err := kind.ApplyPath(resource)
	if err != nil {
		return "", err
	}

	return applyPath + "/" + resource.Name, nil
}
