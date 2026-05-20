package printutils

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// RenderTemplateAsKind takes a template's `spec` (with a `defaults` key holding
// metadata and spec) and turns it into a YAML resource of the given kind/version.
func RenderTemplateAsKind(templateSpec map[string]interface{}, kindName string, apiVersion int) (string, error) {
	defaults, ok := templateSpec["defaults"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Template response is missing spec.defaults")
	}

	out := map[string]interface{}{
		"apiVersion": fmt.Sprintf("v%d", apiVersion),
		"kind":       kindName,
	}
	if metadata, ok := defaults["metadata"]; ok {
		out["metadata"] = metadata
	}
	if spec, ok := defaults["spec"]; ok {
		out["spec"] = spec
	}

	data, err := yaml.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("Error marshaling template as YAML: %s", err)
	}
	return string(data), nil
}
