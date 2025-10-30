package printutils

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/conduktor/ctl/internal/orderedjson"
)

func TestPrintResourceLikeYamlOnSingleResourceFromNormalJson(t *testing.T) {
	resourceFromBe := `{"spec": "someSpec", "apiVersion": "v4", "kind": "Gelato", "metadata": "arancia"}`
	var data interface{}
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
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

func TestPrintResourceLikeYamlOnSingleResourceFromOrderedJson(t *testing.T) {
	resourceFromBe := `{"spec": "someSpec", "apiVersion": "v4", "kind": "Gelato", "metadata": {"z": 1, "t": 2, "x": 3}}`
	var data orderedjson.OrderedData
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.TrimSpace(`
apiVersion: v4
kind: Gelato
metadata:
    z: 1
    t: 2
    x: 3
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
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
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

func TestPrintResourceLikeYamlOnResourcesWithNewUnexpectedFieldFromOrderedJson(t *testing.T) {
	resourceFromBe := `{"spec": "someSpec", "apiVersion": "v4", "newKind": "Gelato", "metadata": {"z": 1, "t": 2, "x": 3}}`
	var data orderedjson.OrderedData
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.TrimSpace(`
apiVersion: v4
metadata:
    z: 1
    t: 2
    x: 3
spec: someSpec
newKind: Gelato
`)
	got := strings.TrimSpace(output.String())
	if got != expected {
		t.Errorf("got:\n%s \nexpected:\n%s", got, expected)
	}
}

func TestPrintResourceLikeYamlOnResourcesWithNewUnexpectedFieldFromNormalJson(t *testing.T) {
	resourceFromBe := `{"spec": "someSpec", "apiVersion": "v4", "newKind": "Gelato", "metadata": "arancia"}`
	var data interface{}
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
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

func TestPrintResourceLikeYamlOnMultipleResourceUsingNormalJson(t *testing.T) {
	resourceFromBe := `[{"spec": "someSpec", "apiVersion": "v4", "newKind": "Gelato", "metadata": "arancia"}, {"spec": "someSpec2", "apiVersion": "v5", "newKind": "banana", "metadata": "bananaa"}]`
	var data interface{}
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.TrimSpace(`
---
apiVersion: v4
metadata: arancia
spec: someSpec
newKind: Gelato
---
apiVersion: v5
metadata: bananaa
spec: someSpec2
newKind: banana 
`)
	got := strings.TrimSpace(output.String())
	if got != expected {
		t.Errorf("got:\n%s \nexpected:\n%s", got, expected)
	}
}

func TestPrintResourceLikeYamlOnMultipleResourceUsingOrderedJson(t *testing.T) {
	resourceFromBe := `[{"spec": "someSpec", "apiVersion": "v4", "newKind": "Gelato", "metadata": "arancia"}, {"spec": "someSpec2", "apiVersion": "v5", "newKind": "banana", "metadata": {"z": 1, "t": 2, "x": 3}}]`
	var data orderedjson.OrderedData
	err := json.Unmarshal([]byte(resourceFromBe), &data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.TrimSpace(`
---
apiVersion: v4
metadata: arancia
spec: someSpec
newKind: Gelato
---
apiVersion: v5
metadata:
    z: 1
    t: 2
    x: 3
spec: someSpec2
newKind: banana 
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
	err = PrintResourceLikeYamlFile(&output, data)
	if err != nil {
		t.Fatal(err)
	}
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
