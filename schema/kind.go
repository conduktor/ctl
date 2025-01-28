package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/conduktor/ctl/resource"
)

type Kind struct {
	Versions map[int]KindVersion
}

type KindCatalog = map[string]Kind

//go:embed default_schema/console.json
var consoleDefaultByteSchema []byte

//go:embed default_schema/gateway.json
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

func (kind *Kind) GetParentQueryFlag() []string {
	kindVersion := kind.GetLatestKindVersion()
	return kindVersion.GetParentQueryParam()
}

func (kind *Kind) GetListFlag() map[string]QueryParameterOption {
	kindVersion := kind.GetLatestKindVersion()
	kindVersion.GetParentQueryParam()
	flags := make(map[string]QueryParameterOption)
	// Filter out query params from parent to avoid duplicates
	for k, v := range kindVersion.GetListQueryParameter() {
		if !slices.Contains(kindVersion.GetParentQueryParam(), k) {
			flags[k] = v
		}
	}
	return flags
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

type QueryInfo struct {
	Path        string
	QueryParams []QueryParam
}

type QueryParam struct {
	Name  string
	Value string
}

func (kind *Kind) ListPath(parentValues []string, parentQueryValues []string) QueryInfo {
	kindVersion := kind.GetLatestKindVersion()
	if len(parentValues) != len(kindVersion.GetParentPathParam()) {
		panic(fmt.Sprintf("For kind %s expected %d parent apiVersion values, got %d", kindVersion.GetName(), len(kindVersion.GetParentPathParam()), len(parentValues)))
	}
	path := kindVersion.GetListPath()
	for i, pathValue := range parentValues {
		path = strings.Replace(path, fmt.Sprintf("{%s}", kindVersion.GetParentPathParam()[i]), pathValue, 1)
	}

	if len(parentQueryValues) != len(kindVersion.GetParentQueryParam()) {
		panic(fmt.Sprintf("For kind %s expected %d parent query parameter values, got %d", kindVersion.GetName(), len(kindVersion.GetParentPathParam()), len(parentValues)))
	}
	var params []QueryParam
	for i, value := range parentQueryValues {
		if value != "" {
			params = append(params, QueryParam{
				Name:  kindVersion.GetParentQueryParam()[i],
				Value: value,
			})
		}
	}

	return QueryInfo{
		Path:        path,
		QueryParams: params,
	}
}

func (kind *Kind) DescribePath(parentPathValues []string, parentQueryValues []string, name string) QueryInfo {
	queryInfo := kind.ListPath(parentPathValues, parentQueryValues)
	return QueryInfo{
		Path:        queryInfo.Path + "/" + name,
		QueryParams: queryInfo.QueryParams,
	}
}

func (kind *Kind) ApplyPath(resource *resource.Resource) (QueryInfo, error) {
	kindVersion, ok := kind.Versions[extractVersionFromApiVersion(resource.Version)]
	if !ok {
		return QueryInfo{}, fmt.Errorf("Could not find version %s for kind %s", resource.Version, resource.Kind)
	}
	parentPathValues := make([]string, len(kindVersion.GetParentPathParam()))
	var parentQueryValues []string
	var err error
	for i, param := range kindVersion.GetParentPathParam() {
		parentPathValues[i], err = resource.StringFromMetadata(param)
		if err != nil {
			return QueryInfo{}, err
		}
	}
	for _, param := range kindVersion.GetParentQueryParam() {
		var value string
		value, err = resource.StringFromMetadata(param)
		if err == nil {
			parentQueryValues = append(parentQueryValues, value)
		} else {
			parentQueryValues = append(parentQueryValues, "")
		}
	}
	return kind.ListPath(parentPathValues, parentQueryValues), nil
}

func (kind *Kind) DeletePath(resource *resource.Resource) (string, error) {
	applyPath, err := kind.ApplyPath(resource)
	if err != nil {
		return "", err
	}

	return applyPath.Path + "/" + resource.Name, nil
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
