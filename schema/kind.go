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
)

type KindVersion struct {
	ListPath        string
	Name            string
	ParentPathParam []string
	Order           int `json:1000` //same value DefaultPriority
}

const DefaultPriority = 1000 //update  json annotation for Order when changing this value

type Kind struct {
	Versions map[int]KindVersion
}

type KindCatalog = map[string]Kind

//go:embed default-schema.json
var defaultByteSchema []byte

func DefaultKind() KindCatalog {
	var result KindCatalog
	err := json.Unmarshal(defaultByteSchema, &result)
	if err != nil {
		panic(err)
	}
	return result
}

func NewKind(version int, kindVersion *KindVersion) Kind {
	return Kind{
		Versions: map[int]KindVersion{version: *kindVersion},
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

func (kind *Kind) AddVersion(version int, kindVersion *KindVersion) error {
	name := kind.GetName()
	if name != kindVersion.Name {
		return fmt.Errorf("Adding kind version of kind %s to different kind %s", kindVersion.Name, name)
	}
	kind.Versions[version] = *kindVersion
	return nil
}

func (kind *Kind) GetFlag() []string {
	kindVersion := kind.GetLatestKindVersion()
	result := make([]string, len(kindVersion.ParentPathParam))
	copy(result, kindVersion.ParentPathParam)
	return result
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

func (kind *Kind) GetLatestKindVersion() *KindVersion {
	kindVersion, ok := kind.Versions[kind.MaxVersion()]
	if !ok {
		panic("Max numVersion on kind return a numVersion that does not exist")
	}
	return &kindVersion
}

func (Kind *Kind) GetName() string {
	for _, kindVersion := range Kind.Versions {
		return kindVersion.Name
	}
	panic("No kindVersion in kind") //should never happen
}

func (kind *Kind) ListPath(parentPathValues []string) string {
	kindVersion := kind.GetLatestKindVersion()
	if len(parentPathValues) != len(kindVersion.ParentPathParam) {
		panic(fmt.Sprintf("For kind %s expected %d parent apiVersion values, got %d", kindVersion.Name, len(kindVersion.ParentPathParam), len(parentPathValues)))
	}
	path := kindVersion.ListPath
	for i, pathValue := range parentPathValues {
		path = strings.Replace(path, fmt.Sprintf("{%s}", kindVersion.ParentPathParam[i]), pathValue, 1)
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
	parentPathValues := make([]string, len(kindVersion.ParentPathParam))
	var err error
	for i, param := range kindVersion.ParentPathParam {
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
