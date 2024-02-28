package resource

import (
	"testing"
)

func checkResource(t *testing.T, result, expected Resource) {
	if result.Name != expected.Name {
		t.Errorf("Expected name %s got %s", expected.Name, result.Name)
	}

	if result.Kind != expected.Kind {
		t.Errorf("Expected kind %s got %s", expected.Kind, result.Kind)
	}

	if result.Version != expected.Version {
		t.Errorf("Expected version %s got %s", expected.Version, result.Version)
	}

	if string(result.Json) != string(expected.Json) {
		t.Errorf("Expected json:\n%s got\n%s", string(expected.Json), string(result.Json))
	}
}

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

	checkResource(t, results[0], Resource{
		Version: "v1",
		Kind:    "Topic",
		Name:    "abc.myTopic",
		Json:    []byte(`{"kind":"Topic","metadata":{"name":"abc.myTopic"},"spec":{"replicationFactor":1},"version":"v1"}`),
	})

	checkResource(t, results[1], Resource{
		Version: "v2",
		Kind:    "ConsumerGroup",
		Name:    "cg1",
		Json:    []byte(`{"kind":"ConsumerGroup","metadata":{"name":"cg1"},"version":"v2"}`),
	})
}

func TestFromFolder(t *testing.T) {
	resources, err := FromFolder("yamls")
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 4 {
		t.Fatalf("Expected to read 4 resources, readed %d", len(resources))
	}

	checkResource(t, resources[0], Resource{
		Version: "v1",
		Kind:    "a",
		Name:    "a",
		Json:    []byte(`{"kind":"a","metadata":{"name":"a"},"spec":{"data":"data"},"version":"v1"}`),
	})

	checkResource(t, resources[1], Resource{
		Version: "v1",
		Kind:    "a",
		Name:    "b",
		Json:    []byte(`{"kind":"a","metadata":{"name":"b"},"spec":{"data":"data2"},"version":"v1"}`),
	})

	checkResource(t, resources[2], Resource{
		Version: "v1",
		Kind:    "b",
		Name:    "a",
		Json:    []byte(`{"kind":"b","metadata":{"name":"a"},"spec":{"data":"yo"},"version":"v1"}`),
	})

	checkResource(t, resources[3], Resource{
		Version: "v1",
		Kind:    "b",
		Name:    "b",
		Json:    []byte(`{"kind":"b","metadata":{"name":"b"},"spec":{"data":"lo"},"version":"v1"}`),
	})
}
