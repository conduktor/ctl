package printutils

import "testing"

func TestRenderTemplateAsKindBuildsResourceFromDefaults(t *testing.T) {
	templateSpec := map[string]interface{}{
		"displayName": "High Partition Topic",
		"description": "Optimised for high throughput",
		"defaults": map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":   "my-topic-${dept}",
				"labels": map[string]interface{}{"throughput": "high"},
			},
			"spec": map[string]interface{}{
				"partitions":        24,
				"replicationFactor": 3,
				"configs":           map[string]interface{}{"retention.ms": "604800000"},
			},
		},
	}

	got, err := RenderTemplateAsKind(templateSpec, "Topic", 2)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := `apiVersion: v2
kind: Topic
metadata:
    labels:
        throughput: high
    name: my-topic-${dept}
spec:
    configs:
        retention.ms: "604800000"
    partitions: 24
    replicationFactor: 3
`
	if got != want {
		t.Errorf("rendered template mismatch.\nwant:\n%s\ngot:\n%s", want, got)
	}
}
