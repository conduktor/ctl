package schema

import (
	"fmt"
	"strings"
)

type Run struct {
	Path           string
	Name           string
	Doc            string
	QueryParameter map[string]FlagParameterOption
	PathParameter  []string
	BodyFields     map[string]FlagParameterOption
	Method         string

	BackendType BackendType `json:"-"`
}

func (c *Run) BuildPath(pathValue []string) string {
	path := c.Path
	if len(pathValue) != len(c.PathParameter) {
		panic(fmt.Sprintf("BuildPath: pathValue lentgth (%d) does not match execution path parameter length (%d) for %s", len(pathValue), len(c.PathParameter), c.Name))
	}
	for i, pathValue := range pathValue {
		path = strings.Replace(path, fmt.Sprintf("{%s}", c.PathParameter[i]), pathValue, 1)
	}
	return path
}
