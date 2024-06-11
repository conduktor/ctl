package resource

import (
	"github.com/davecgh/go-spew/spew"
	"log"
	"os"
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
apiVersion: v1
kind: Topic
metadata:
  name: abc.myTopic
spec:
  replicationFactor: 1
---
apiVersion: v2
kind: ConsumerGroup
metadata:
  name: cg1
`)

	results, err := FromByte(yamlByte)
	spew.Dump(results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 2 {
		t.Errorf("results expected1 of length 2, got length %d", len(results))
	}

	checkResource(t, results[0], Resource{
		Version:  "v1",
		Kind:     "Topic",
		Name:     "abc.myTopic",
		Metadata: map[string]interface{}{"name": "abc.myTopic"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Topic","metadata":{"name":"abc.myTopic"},"spec":{"replicationFactor":1}}`),
	})

	checkResource(t, results[1], Resource{
		Version:  "v2",
		Kind:     "ConsumerGroup",
		Name:     "cg1",
		Metadata: map[string]interface{}{"name": "cg1"},
		Json:     []byte(`{"apiVersion":"v2","kind":"ConsumerGroup","metadata":{"name":"cg1"}}`),
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
		Version:  "v1",
		Kind:     "a",
		Name:     "a",
		Metadata: map[string]interface{}{"name": "a"},
		Json:     []byte(`{"apiVersion":"v1","kind":"a","metadata":{"name":"a"},"spec":{"data":"data"}}`),
	})

	checkResource(t, resources[1], Resource{
		Version:  "v1",
		Kind:     "a",
		Name:     "b",
		Metadata: map[string]interface{}{"name": "b"},
		Json:     []byte(`{"apiVersion":"v1","kind":"a","metadata":{"name":"b"},"spec":{"data":"data2"}}`),
	})

	checkResource(t, resources[2], Resource{
		Version:  "v1",
		Kind:     "b",
		Name:     "a",
		Metadata: map[string]interface{}{"name": "a"},
		Json:     []byte(`{"apiVersion":"v1","kind":"b","metadata":{"name":"a"},"spec":{"data":"yo"}}`),
	})

	checkResource(t, resources[3], Resource{
		Version:  "v1",
		Kind:     "b",
		Name:     "b",
		Metadata: map[string]interface{}{"name": "b"},
		Json:     []byte(`{"apiVersion":"v1","kind":"b","metadata":{"name":"b"},"spec":{"data":"lo"}}`),
	})
}

func TestResourceExpansion(t *testing.T) {

	avroSchema, err := os.CreateTemp("/tmp", "schema.avsc")
	if err != nil {
		t.Fatal(err)
	}
	defer avroSchema.Close()
	defer os.Remove(avroSchema.Name())
	if _, err := avroSchema.Write([]byte(`{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}`)); err != nil {
		log.Fatal(err)
	}

	jsonSchemaContent := []byte(`{
	"$id": "https://mycompany.com/myrecord",
	"$schema": "https://json-schema.org/draft/2019-09/schema",
	"type": "object",
	"title": "MyRecord",
	"description": "Json schema for MyRecord",
	"properties": {
		"id": { "type": "string" },
		"name": { "type": [ "string", "null" ] }
	},
	"required": [ "id" ],
	"additionalProperties": false
}`)
	jsonSchema, err := os.CreateTemp("/tmp", "schema.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonSchema.Close()
	defer os.Remove(jsonSchema.Name())
	if _, err := jsonSchema.Write(jsonSchemaContent); err != nil {
		log.Fatal(err)
	}

	yamlByte := []byte(`
# comment
---
apiVersion: v1
kind: Subject
metadata:
  cluster: cluster-a
  name: abc.mySchema
spec:
  format: avro
  schema: |
    {
      "type":"record",
      "name":"myrecord",
      "fields": [{ "name":"f1", "type":"string" }]
    }
---
apiVersion: v1
kind: Subject
metadata:
  cluster: cluster-a
  name: abc.mySchemaExtAvro
spec:
  format: json
  schemaFile: ` + avroSchema.Name() + `
---
apiVersion: v1
kind: Subject
metadata:
  cluster: cluster-a
  name: abc.mySchemaExtJson
spec:
  format: avro
  schemaFile: ` + jsonSchema.Name() + `
`)

	results, err := FromByte(yamlByte)
	spew.Dump(results)
	if err != nil {
		t.Error(err)
	}

	if len(results) != 3 {
		t.Errorf("results expected of length 3, got length %d", len(results))
	}

	checkResource(t, results[0], Resource{
		Version:  "v1",
		Kind:     "Subject",
		Name:     "abc.mySchema",
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "abc.mySchema"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchema"},"spec":{"format":"avro","schema":"{\n  \"type\":\"record\",\n  \"name\":\"myrecord\",\n  \"fields\": [{ \"name\":\"f1\", \"type\":\"string\" }]\n}\n"}}`),
	})

	checkResource(t, results[1], Resource{
		Version:  "v1",
		Kind:     "Subject",
		Name:     "abc.mySchemaExtAvro",
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "abc.mySchemaExtAvro"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchemaExtAvro"},"spec":{"format":"json","schema":"{\"type\":\"record\",\"name\":\"myrecord\",\"fields\":[{\"name\":\"f1\",\"type\":\"string\"}]}"}}`), // schemaFile is expanded
	})

	checkResource(t, results[2], Resource{
		Version:  "v1",
		Kind:     "Subject",
		Name:     "abc.mySchemaExtJson",
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "abc.mySchemaExtJson"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchemaExtJson"},"spec":{"format":"avro","schema":"{\n\t\"$id\": \"https://mycompany.com/myrecord\",\n\t\"$schema\": \"https://json-schema.org/draft/2019-09/schema\",\n\t\"type\": \"object\",\n\t\"title\": \"MyRecord\",\n\t\"description\": \"Json schema for MyRecord\",\n\t\"properties\": {\n\t\t\"id\": { \"type\": \"string\" },\n\t\t\"name\": { \"type\": [ \"string\", \"null\" ] }\n\t},\n\t\"required\": [ \"id\" ],\n\t\"additionalProperties\": false\n}"}}`),
	})

}
