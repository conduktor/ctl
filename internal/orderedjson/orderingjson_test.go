package orderedjson

import (
	"encoding/json"
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestOrderedRecursiveMap(t *testing.T) {
	testForJSON(t, `{"name":"John","age":30,"city":"New York","children":[{"name":"Alice","age":5},{"name":"Bob","age":7}],"parent":{"name":"Jane","age":60,"city":"New York"}}`)
	testForJSON(t, `"yo"`)
	testForJSON(t, `true`)
	testForJSON(t, `false`)
	testForJSON(t, `42`)
	testForJSON(t, `42.2`)
	testForJSON(t, `[]`)
	testForJSON(t, `{}`)
	testForJSON(t, `{"z":{"x":{"v":{}}},"y":{"u":{"t":"p"}}}`)
	testForJSON(t, `[[[[]]]]`)
	testForJSON(t, `[{"z":42},{"b":{},"y":41,"a":[[{"z":42},{"b":{},"y":41,"a":[[{"z":42},{"b":{},"y":41,"a":[[{"z":42},{"b":{},"y":41,"a":[]}]]}]]}]]}]`)
}

func testForJSON(t *testing.T, originalJSON string) {
	// Unmarshal the JSON into an OrderedRecursiveMap
	var omap OrderedData
	err := json.Unmarshal([]byte(originalJSON), &omap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %+v", err)
	}

	fmt.Printf("%v\n", omap)
	// Marshal the OrderedRecursiveMap back into JSON
	marshaledJSON, err := json.Marshal(&omap)
	if err != nil {
		t.Fatalf("Failed to marshal OrderedRecursiveMap: %v", err)
	}

	// Check if the original JSON and the marshaled JSON are the same
	if originalJSON != string(marshaledJSON) {
		t.Errorf("Original JSON and marshaled JSON do not match. Original: %s, Marshaled: %s", originalJSON, string(marshaledJSON))
	}
}

func TestYamlMarshallingKeepOrderTo(t *testing.T) {
	// Unmarshal the JSON into an OrderedRecursiveMap
	var omap OrderedData
	err := json.Unmarshal([]byte(`{"name":"John","age":30,"city":"New York","children":[{"name":"Alice","age":5},{"name":"Bob","age":7}],"parent":{"name":"Jane","age":60,"city":"New York"}}`), &omap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %+v", err)
	}

	fmt.Printf("%v\n", omap)
	// Marshal the OrderedRecursiveMap back into JSON
	marshaledYaml, err := yaml.Marshal(&omap)
	if err != nil {
		t.Fatalf("Failed to marshal OrderedRecursiveMap: %v", err)
	}

	expected := `name: John
age: 30
city: New York
children:
    - name: Alice
      age: 5
    - name: Bob
      age: 7
parent:
    name: Jane
    age: 60
    city: New York
`

	// Check if the original JSON and the marshaled JSON are the same
	if expected != string(marshaledYaml) {
		t.Errorf("Marshalled yaml is not valid. Got:\n##\n%s\n##\n,\nMarshaled:\n##\n%s\n##", string(marshaledYaml), expected)
	}
}
