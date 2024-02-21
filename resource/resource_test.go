package resource

import (
	"testing"
)

func TestFromByteForOneResourceWithValidResource(t *testing.T) {
	yamlByte := []byte(`
# comment
---
version: v1
kind: Topic
metadata:
  name: abc.myTopic
spec:
  replicationFactor: 1
---
version: v2
kind: ConsumerGroup
metadata:
  name: cg1
`)

	results, err := FromByte(yamlByte)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 2 {
		t.Errorf("results expected1 of length 2, got length %d", len(results))
	}

	result1 := results[0]
	expected1 := Resource{
		Version: "v1",
		Kind:    "Topic",
		Name:    "abc.myTopic",
		Json:    []byte(`{"kind":"Topic","metadata":{"name":"abc.myTopic"},"spec":{"replicationFactor":1},"version":"v1"}`),
	}

	if result1.Name != expected1.Name {
		t.Errorf("Expected name %s got %s", expected1.Name, result1.Name)
	}

	if result1.Kind != expected1.Kind {
		t.Errorf("Expected name %s got %s", expected1.Kind, result1.Kind)
	}

	if result1.Version != expected1.Version {
		t.Errorf("Expected name %s got %s", expected1.Version, result1.Version)
	}

	expectedJsonString1 := string(expected1.Json)
	resultJsonString1 := string(result1.Json)
	if expectedJsonString1 != resultJsonString1 {
		t.Errorf("\nExpected json:\n%s got:\n%s", expectedJsonString1, resultJsonString1)
	}

	result2 := results[1]
	expected2 := Resource{
		Version: "v2",
		Kind:    "ConsumerGroup",
		Name:    "cg1",
		Json:    []byte(`{"kind":"ConsumerGroup","metadata":{"name":"cg1"},"version":"v2"}`),
	}

	if result2.Name != expected2.Name {
		t.Errorf("Expected name %s got %s", expected2.Name, result2.Name)
	}

	if result2.Kind != expected2.Kind {
		t.Errorf("Expected name %s got %s", expected2.Kind, result2.Kind)
	}

	if result2.Version != expected2.Version {
		t.Errorf("Expected name %s got %s", expected2.Version, result2.Version)
	}

	expectedJsonString2 := string(expected2.Json)
	resultJsonString2 := string(result2.Json)
	if expectedJsonString2 != resultJsonString2 {
		t.Errorf("\nExpected json:\n%s got:\n%s", expectedJsonString2, resultJsonString2)
	}
}

func TestFromByte(t *testing.T) {

}
