package resource

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func checkResource(t *testing.T, result, expected Resource) {
	checkResourceGeneric(t, result, expected, true)
}

func checkResourceWithoutJSONOrder(t *testing.T, result, expected Resource) {
	checkResourceGeneric(t, result, expected, false)
}

func checkResourceGeneric(t *testing.T, result, expected Resource, jsonOrderMatter bool) {
	if result.Name != expected.Name {
		t.Errorf("Expected name %s got %s", expected.Name, result.Name)
	}

	if result.Kind != expected.Kind {
		t.Errorf("Expected kind %s got %s", expected.Kind, result.Kind)
	}

	if result.Version != expected.Version {
		t.Errorf("Expected version %s got %s", expected.Version, result.Version)
	}

	if !reflect.DeepEqual(result.Metadata, expected.Metadata) {
		t.Errorf("Expected Metadata %s got %s", expected.Metadata, result.Metadata)
	}
	if !reflect.DeepEqual(result.Spec, expected.Spec) {
		t.Errorf("Expected Spec %s got %s", expected.Spec, result.Spec)
	}

	if jsonOrderMatter {
		if string(result.Json) != string(expected.Json) {
			t.Errorf("Expected json:\n%s got\n%s", string(expected.Json), string(result.Json))
		}
	} else {
		var gotJSON map[string]interface{}
		var expectedJSON map[string]interface{}
		err := json.Unmarshal(result.Json, &gotJSON)
		if err != nil {
			t.Error(err)
		}
		err = json.Unmarshal(expected.Json, &expectedJSON)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(gotJSON, expectedJSON) {
			t.Errorf("Expected json:\n%s got\n%s", string(expected.Json), string(result.Json))
		}
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

	results, err := FromYamlByte(yamlByte, true)
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
		Spec:     map[string]interface{}{"replicationFactor": 1.0},
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
	resources, err := FromFolder("testdata/yamls", true, false)
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
		Spec:     map[string]interface{}{"data": "data"},
		Json:     []byte(`{"apiVersion":"v1","kind":"a","metadata":{"name":"a"},"spec":{"data":"data"}}`),
	})

	checkResource(t, resources[1], Resource{
		Version:  "v1",
		Kind:     "a",
		Name:     "b",
		Metadata: map[string]interface{}{"name": "b"},
		Spec:     map[string]interface{}{"data": "data2"},
		Json:     []byte(`{"apiVersion":"v1","kind":"a","metadata":{"name":"b"},"spec":{"data":"data2"}}`),
	})

	checkResource(t, resources[2], Resource{
		Version:  "v1",
		Kind:     "b",
		Name:     "a",
		Metadata: map[string]interface{}{"name": "a"},
		Spec:     map[string]interface{}{"data": "yo"},
		Json:     []byte(`{"apiVersion":"v1","kind":"b","metadata":{"name":"a"},"spec":{"data":"yo"}}`),
	})

	checkResource(t, resources[3], Resource{
		Version:  "v1",
		Kind:     "b",
		Name:     "b",
		Metadata: map[string]interface{}{"name": "b"},
		Spec:     map[string]interface{}{"data": "lo"},
		Json:     []byte(`{"apiVersion":"v1","kind":"b","metadata":{"name":"b"},"spec":{"data":"lo"}}`),
	})
}

func TestFromFolderRecursive(t *testing.T) {
	resources, err := FromFolder("testdata/yamls", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 6 {
		t.Fatalf("Expected to read 4 resources, readed %d", len(resources))
	}

	checkResource(t, resources[0], Resource{
		Version:  "v1",
		Kind:     "a",
		Name:     "a",
		Metadata: map[string]interface{}{"name": "a"},
		Spec:     map[string]interface{}{"data": "data"},
		Json:     []byte(`{"apiVersion":"v1","kind":"a","metadata":{"name":"a"},"spec":{"data":"data"}}`),
	})

	checkResource(t, resources[1], Resource{
		Version:  "v1",
		Kind:     "a",
		Name:     "b",
		Metadata: map[string]interface{}{"name": "b"},
		Spec:     map[string]interface{}{"data": "data2"},
		Json:     []byte(`{"apiVersion":"v1","kind":"a","metadata":{"name":"b"},"spec":{"data":"data2"}}`),
	})

	checkResource(t, resources[2], Resource{
		Version:  "v1",
		Kind:     "b",
		Name:     "a",
		Metadata: map[string]interface{}{"name": "a"},
		Spec:     map[string]interface{}{"data": "yo"},
		Json:     []byte(`{"apiVersion":"v1","kind":"b","metadata":{"name":"a"},"spec":{"data":"yo"}}`),
	})

	checkResource(t, resources[3], Resource{
		Version:  "v1",
		Kind:     "b",
		Name:     "b",
		Metadata: map[string]interface{}{"name": "b"},
		Spec:     map[string]interface{}{"data": "lo"},
		Json:     []byte(`{"apiVersion":"v1","kind":"b","metadata":{"name":"b"},"spec":{"data":"lo"}}`),
	})

	checkResource(t, resources[4], Resource{
		Version:  "v1",
		Kind:     "c",
		Name:     "c",
		Metadata: map[string]interface{}{"name": "c"},
		Spec:     map[string]interface{}{"data": "ccc"},
		Json:     []byte(`{"apiVersion":"v1","kind":"c","metadata":{"name":"c"},"spec":{"data":"ccc"}}`),
	})

	checkResource(t, resources[5], Resource{
		Version:  "v2",
		Kind:     "d",
		Name:     "d",
		Metadata: map[string]interface{}{"name": "d"},
		Spec:     map[string]interface{}{"data": "ddd"},
		Json:     []byte(`{"apiVersion":"v2","kind":"d","metadata":{"name":"d"},"spec":{"data":"ddd"}}`),
	})
}

func TestResourceExpansionVariableEnv(t *testing.T) {
	topicDesc, err := os.CreateTemp("/tmp", "topic.md")
	if err != nil {
		t.Fatal(err)
	}
	defer topicDesc.Close()
	defer os.Remove(topicDesc.Name())
	if _, err := topicDesc.Write([]byte(`This topic is awesome`)); err != nil {
		log.Fatal(err)
	}

	yamlByte := []byte(`
# comment
---
apiVersion: v1
kind: Topic
metadata:
  cluster: ${CLUSTER_NAME}
  name: ${TOPIC_NAME:-toto}
  labels:
    conduktor.io/descriptionFile: ` + topicDesc.Name() + `
spec:
  replicationFactor: 2
  partition: 3
`)
	os.Setenv("CLUSTER_NAME", "cluster-a")

	results, err := FromYamlByte(yamlByte, true)
	spew.Dump(results)
	if err != nil {
		t.Error(err)
	}

	if len(results) != 1 {
		t.Errorf("results expected of length 1, got length %d", len(results))
	}

	checkResourceWithoutJSONOrder(t, results[0], Resource{
		Version:  "v1",
		Kind:     "Topic",
		Name:     "toto",
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "toto", "labels": map[string]interface{}{"conduktor.io/description": "This topic is awesome"}},
		Spec:     map[string]interface{}{"replicationFactor": 2.0, "partition": 3.0},
		Json:     []byte(`{"apiVersion":"v1","kind":"Topic","metadata":{"cluster":"cluster-a","name":"toto","labels":{"conduktor.io/description":"This topic is awesome"}},"spec":{"replicationFactor":2,"partition":3}}`),
	})

	yamlByte2 := []byte(`
# comment
---
apiVersion: gateway/v2
kind: Interceptor
metadata:
  name: decryption
spec:
  priority: 100
  pluginClass: io.conduktor.gateway.interceptor.DecryptPlugin
  config:
    topic: .*
    kmsConfig:
      vault:
        uri: http://$${VAULT_URI}
        username: $${VAULT_USERNAME}
        password: $${VAULT_PASSWORD}
`)

	results2, err := FromYamlByte(yamlByte2, true)
	spew.Dump(results2)
	if err != nil {
		t.Error(err)
	}

	if len(results2) != 1 {
		t.Errorf("results expected of length 1, got length %d", len(results2))
	}

	checkResourceWithoutJSONOrder(t, results2[0], Resource{
		Version:  "gateway/v2",
		Kind:     "Interceptor",
		Name:     "decryption",
		Metadata: map[string]interface{}{"name": "decryption"},
		Spec:     map[string]interface{}{"priority": 100.0, "pluginClass": "io.conduktor.gateway.interceptor.DecryptPlugin", "config": map[string]interface{}{"topic": ".*", "kmsConfig": map[string]interface{}{"vault": map[string]interface{}{"uri": "http://${VAULT_URI}", "username": "${VAULT_USERNAME}", "password": "${VAULT_PASSWORD}"}}}},
		Json:     []byte(`{"apiVersion":"gateway/v2","kind":"Interceptor","metadata":{"name":"decryption"},"spec":{"priority":100,"pluginClass":"io.conduktor.gateway.interceptor.DecryptPlugin","config":{"topic":".*","kmsConfig":{"vault":{"uri":"http://${VAULT_URI}","username":"${VAULT_USERNAME}", "password": "${VAULT_PASSWORD}"}}}}}`),
	})
}

func TestResourceExpansionForTopic(t *testing.T) {
	topicDesc, err := os.CreateTemp("/tmp", "topic.md")
	if err != nil {
		t.Fatal(err)
	}
	defer topicDesc.Close()
	defer os.Remove(topicDesc.Name())
	if _, err := topicDesc.Write([]byte(`This topic is awesome`)); err != nil {
		log.Fatal(err)
	}

	yamlByte := []byte(`
# comment
---
apiVersion: v1
kind: Topic
metadata:
  cluster: cluster-a
  name: toto
  labels:
    conduktor.io/descriptionFile: ` + topicDesc.Name() + `
spec:
  replicationFactor: 2
  partition: 3
`)

	results, err := FromYamlByte(yamlByte, true)
	spew.Dump(results)
	if err != nil {
		t.Error(err)
	}

	if len(results) != 1 {
		t.Errorf("results expected of length 1, got length %d", len(results))
	}

	checkResourceWithoutJSONOrder(t, results[0], Resource{
		Version:  "v1",
		Kind:     "Topic",
		Name:     "toto",
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "toto", "labels": map[string]interface{}{"conduktor.io/description": "This topic is awesome"}},
		Spec:     map[string]interface{}{"replicationFactor": 2.0, "partition": 3.0},
		Json:     []byte(`{"apiVersion":"v1","kind":"Topic","metadata":{"cluster":"cluster-a","name":"toto","labels":{"conduktor.io/description":"This topic is awesome"}},"spec":{"replicationFactor":2,"partition":3}}`),
	})
}

func TestResourceExpansionForSchema(t *testing.T) {

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

	results, err := FromYamlByte(yamlByte, true)
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
		Spec:     map[string]interface{}{"format": "avro", "schema": "{\n  \"type\":\"record\",\n  \"name\":\"myrecord\",\n  \"fields\": [{ \"name\":\"f1\", \"type\":\"string\" }]\n}\n"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchema"},"spec":{"format":"avro","schema":"{\n  \"type\":\"record\",\n  \"name\":\"myrecord\",\n  \"fields\": [{ \"name\":\"f1\", \"type\":\"string\" }]\n}\n"}}`),
	})

	checkResource(t, results[1], Resource{
		Version:  "v1",
		Kind:     "Subject",
		Name:     "abc.mySchemaExtAvro",
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "abc.mySchemaExtAvro"},
		Spec:     map[string]interface{}{"format": "json", "schema": "{\"type\":\"record\",\"name\":\"myrecord\",\"fields\":[{\"name\":\"f1\",\"type\":\"string\"}]}"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchemaExtAvro"},"spec":{"format":"json","schema":"{\"type\":\"record\",\"name\":\"myrecord\",\"fields\":[{\"name\":\"f1\",\"type\":\"string\"}]}"}}`), // schemaFile is expanded
	})

	checkResource(t, results[2], Resource{
		Version:  "v1",
		Kind:     "Subject",
		Name:     "abc.mySchemaExtJson",
		Spec:     map[string]interface{}{"format": "avro", "schema": "{\n\t\"$id\": \"https://mycompany.com/myrecord\",\n\t\"$schema\": \"https://json-schema.org/draft/2019-09/schema\",\n\t\"type\": \"object\",\n\t\"title\": \"MyRecord\",\n\t\"description\": \"Json schema for MyRecord\",\n\t\"properties\": {\n\t\t\"id\": { \"type\": \"string\" },\n\t\t\"name\": { \"type\": [ \"string\", \"null\" ] }\n\t},\n\t\"required\": [ \"id\" ],\n\t\"additionalProperties\": false\n}"},
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "abc.mySchemaExtJson"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchemaExtJson"},"spec":{"format":"avro","schema":"{\n\t\"$id\": \"https://mycompany.com/myrecord\",\n\t\"$schema\": \"https://json-schema.org/draft/2019-09/schema\",\n\t\"type\": \"object\",\n\t\"title\": \"MyRecord\",\n\t\"description\": \"Json schema for MyRecord\",\n\t\"properties\": {\n\t\t\"id\": { \"type\": \"string\" },\n\t\t\"name\": { \"type\": [ \"string\", \"null\" ] }\n\t},\n\t\"required\": [ \"id\" ],\n\t\"additionalProperties\": false\n}"}}`),
	})

}

func TestJsonUnmarshal(t *testing.T) {
	aResource := Resource{
		Version:  "v1",
		Kind:     "Subject",
		Name:     "abc.mySchemaExtJson",
		Spec:     map[string]interface{}{"format": "avro", "schema": "{\n\t\"$id\": \"https://mycompany.com/myrecord\",\n\t\"$schema\": \"https://json-schema.org/draft/2019-09/schema\",\n\t\"type\": \"object\",\n\t\"title\": \"MyRecord\",\n\t\"description\": \"Json schema for MyRecord\",\n\t\"properties\": {\n\t\t\"id\": { \"type\": \"string\" },\n\t\t\"name\": { \"type\": [ \"string\", \"null\" ] }\n\t},\n\t\"required\": [ \"id\" ],\n\t\"additionalProperties\": false\n}"},
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "abc.mySchemaExtJson"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchemaExtJson"},"spec":{"format":"avro","schema":"{\n\t\"$id\": \"https://mycompany.com/myrecord\",\n\t\"$schema\": \"https://json-schema.org/draft/2019-09/schema\",\n\t\"type\": \"object\",\n\t\"title\": \"MyRecord\",\n\t\"description\": \"Json schema for MyRecord\",\n\t\"properties\": {\n\t\t\"id\": { \"type\": \"string\" },\n\t\t\"name\": { \"type\": [ \"string\", \"null\" ] }\n\t},\n\t\"required\": [ \"id\" ],\n\t\"additionalProperties\": false\n}"}}`),
	}

	var decodedResource Resource
	err := json.Unmarshal(aResource.Json, &decodedResource)
	if err != nil {
		t.Error(err)
	}
	checkResource(t, decodedResource, aResource)
}

func TestJsonMarshall(t *testing.T) {
	aResource := Resource{
		Version:  "v1",
		Kind:     "Subject",
		Name:     "abc.mySchemaExtJson",
		Metadata: map[string]interface{}{"cluster": "cluster-a", "name": "abc.mySchemaExtJson"},
		Json:     []byte(`{"apiVersion":"v1","kind":"Subject","metadata":{"cluster":"cluster-a","name":"abc.mySchemaExtJson"},"spec":{"format":"avro","schema":"{\n\t\"$id\": \"https://mycompany.com/myrecord\",\n\t\"$schema\": \"https://json-schema.org/draft/2019-09/schema\",\n\t\"type\": \"object\",\n\t\"title\": \"MyRecord\",\n\t\"description\": \"Json schema for MyRecord\",\n\t\"properties\": {\n\t\t\"id\": { \"type\": \"string\" },\n\t\t\"name\": { \"type\": [ \"string\", \"null\" ] }\n\t},\n\t\"required\": [ \"id\" ],\n\t\"additionalProperties\": false\n}"}}`),
	}

	bytes, err := json.Marshal(aResource)
	if err != nil {
		t.Error(err)
	}
	if string(bytes) != string(aResource.Json) {
		t.Errorf("Expected %s got %s", string(aResource.Json), string(bytes))
	}
}
