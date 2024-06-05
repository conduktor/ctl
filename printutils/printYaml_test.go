package printutils

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintResourceLikeYamlOnSingleResource(t *testing.T) {
	resourceFromBe := `{"spec": "someSpec", "apiVersion": "v4", "kind": "Gelato", "metadata": "arancia"}`
	var data interface{}
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	PrintResourceLikeYamlFile(&output, data)
	expected := strings.TrimSpace(`
apiVersion: v4
kind: Gelato
metadata: arancia
spec: someSpec`)
	got := strings.TrimSpace(output.String())
	if got != expected {
		t.Errorf("got:\n%s \nexpected:\n%s", got, expected)
	}
}

func TestPrintResourceLikeYamlInCaseOfScalarValue(t *testing.T) {
	resourceFromBe := `[[1], 3, true, "cat"]`
	var data interface{}
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	PrintResourceLikeYamlFile(&output, data)
	expected := strings.TrimSpace(`
---
- 1
---
3
---
true
---
cat`)
	got := strings.TrimSpace(output.String())
	if got != expected {
		t.Errorf("got:\n%s \nexpected:\n%s", got, expected)
	}
}

func TestPrintResourceLikeYamlOnMultileResources(t *testing.T) {
	resourceFromBe := `{"spec": "someSpec", "apiVersion": "v4", "newKind": "Gelato", "metadata": "arancia"}`
	var data interface{}
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	PrintResourceLikeYamlFile(&output, data)
	expected := strings.TrimSpace(`
apiVersion: v4
metadata: arancia
spec: someSpec
newKind: Gelato
`)
	got := strings.TrimSpace(output.String())
	if got != expected {
		t.Errorf("got:\n%s \nexpected:\n%s", got, expected)
	}
}
func TestPrintResourceWithMissingFieldAndUnexpectedField(t *testing.T) {
	resourceFromBe := `[[1], 3, true, "cat"]`
	var data interface{}
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	PrintResourceLikeYamlFile(&output, data)
	expected := strings.TrimSpace(`
---
- 1
---
3
---
true
---
cat`)
	got := strings.TrimSpace(output.String())
	if got != expected {
		t.Errorf("got:\n%s \nexpected:\n%s", got, expected)
	}
}
