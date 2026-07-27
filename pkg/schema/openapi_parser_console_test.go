package schema

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetKindWithYamlFromOldConsolePlusWithoutOrder(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/docs_without_order.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		kinds, err := schema.GetConsoleKinds(false)
		if err != nil {
			t.Fatalf("failed getting kinds: %s", err)
		}

		expected := KindCatalog{
			"Application": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:               "Application",
						ListPath:           "/public/self-serve/v1/application",
						ParentPathParam:    make([]string, 0),
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              DefaultPriority,
					},
				},
			},
			"ApplicationInstance": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:            "ApplicationInstance",
						ListPath:        "/public/self-serve/v1/application-instance",
						ParentPathParam: make([]string, 0),
						ListQueryParameter: map[string]FlagParameterOption{
							"application": {
								FlagName: "application",
								Required: false,
								Type:     "string",
							},
						},
						Order: DefaultPriority,
					},
				},
			},
			"ApplicationInstancePermission": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:            "ApplicationInstancePermission",
						ListPath:        "/public/self-serve/v1/application-instance-permission",
						ParentPathParam: make([]string, 0),
						ListQueryParameter: map[string]FlagParameterOption{
							"filterByApplication": {
								FlagName: "application",
								Required: false,
								Type:     "string",
							},
							"filterByApplicationInstance": {
								FlagName: "application-instance",
								Required: false,
								Type:     "string",
							},
							"filterByGrantedTo": {
								FlagName: "granted-to",
								Required: false,
								Type:     "string",
							},
						},
						Order: DefaultPriority,
					},
				},
			},
			"TopicPolicy": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:            "TopicPolicy",
						ListPath:        "/public/self-serve/v1/topic-policy",
						ParentPathParam: make([]string, 0),
						ListQueryParameter: map[string]FlagParameterOption{
							"app-instance": {
								FlagName: "application-instance",
								Required: false,
								Type:     "string",
							},
						},
						Order: DefaultPriority,
					},
				},
			},
			"Topic": {
				Versions: map[int]KindVersion{
					2: &ConsoleKindVersion{
						Name:               "Topic",
						ListPath:           "/public/kafka/v2/cluster/{cluster}/topic",
						ParentPathParam:    []string{"cluster"},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              DefaultPriority,
					},
				},
			},
		}
		if !reflect.DeepEqual(kinds, expected) {
			t.Error(spew.Printf("got kinds %v, want %v", kinds, expected))
		}
	})
}

func TestGetKindWithYamlFromConsolePlus(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/docs_with_order.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		kinds, err := schema.GetConsoleKinds(true)
		if err != nil {
			t.Fatalf("failed getting kinds: %s", err)
		}

		expected := KindCatalog{
			"Application": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:               "Application",
						ListPath:           "/public/self-serve/v1/application",
						ParentPathParam:    []string{},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              6,
					},
				},
			},
			"ApplicationInstance": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:            "ApplicationInstance",
						ListPath:        "/public/self-serve/v1/application-instance",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]FlagParameterOption{
							"application": {
								FlagName: "application",
								Required: false,
								Type:     "string",
							},
						},
						Order: 7,
					},
				},
			},
			"ApplicationInstancePermission": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:            "ApplicationInstancePermission",
						ListPath:        "/public/self-serve/v1/application-instance-permission",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]FlagParameterOption{
							"filterByApplication": {
								FlagName: "application",
								Required: false,
								Type:     "string",
							},
							"filterByApplicationInstance": {
								FlagName: "application-instance",
								Required: false,
								Type:     "string",
							},
							"filterByGrantedTo": {
								FlagName: "granted-to",
								Required: false,
								Type:     "string",
							},
						},
						Order: 8,
					},
				},
			},
			"ApplicationGroup": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:               "ApplicationGroup",
						ListPath:           "/public/self-serve/v1/application-group",
						ParentPathParam:    []string{},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              9,
					},
				},
			},
			"TopicPolicy": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:            "TopicPolicy",
						ListPath:        "/public/self-serve/v1/topic-policy",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]FlagParameterOption{
							"app-instance": {
								FlagName: "application-instance",
								Required: false,
								Type:     "string",
							},
						},
						Order: 5,
					},
				},
			},
			"Topic": {
				Versions: map[int]KindVersion{
					2: &ConsoleKindVersion{
						Name:               "Topic",
						ListPath:           "/public/kafka/v2/cluster/{cluster}/topic",
						ParentPathParam:    []string{"cluster"},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              3,
					},
				},
			},
			"Subject": {
				Versions: map[int]KindVersion{
					2: &ConsoleKindVersion{
						Name:               "Subject",
						ListPath:           "/public/kafka/v2/cluster/{cluster}/subject",
						ParentPathParam:    []string{"cluster"},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              4,
					},
				},
			},
			"User": {
				Versions: map[int]KindVersion{
					2: &ConsoleKindVersion{
						Name:               "User",
						ListPath:           "/public/iam/v2/user",
						ParentPathParam:    []string{},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              0,
					},
				},
			},
			"Group": {
				Versions: map[int]KindVersion{
					2: &ConsoleKindVersion{
						Name:               "Group",
						ListPath:           "/public/iam/v2/group",
						ParentPathParam:    []string{},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              1,
					},
				},
			},
			"KafkaCluster": {
				Versions: map[int]KindVersion{
					2: &ConsoleKindVersion{
						Name:               "KafkaCluster",
						ListPath:           "/public/console/v2/kafka-cluster",
						ParentPathParam:    []string{},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              2,
					},
				},
			},
		}
		if !reflect.DeepEqual(kinds, expected) {
			t.Error(spew.Printf("got kinds %v, want %v", kinds, expected))
		}
	})
}

func TestGetKindWithMultipleVersion(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/multiple_version.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		kinds, err := schema.GetConsoleKinds(false)
		if err != nil {
			t.Fatalf("failed getting kinds: %s", err)
		}

		expected := KindCatalog{
			"Topic": {
				Versions: map[int]KindVersion{
					1: &ConsoleKindVersion{
						Name:               "Topic",
						ListPath:           "/public/v1/cluster/{cluster}/topic",
						ParentPathParam:    []string{"cluster"},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              DefaultPriority,
					},
					2: &ConsoleKindVersion{
						Name:               "Topic",
						ListPath:           "/public/v2/cluster/{cluster}/sa/{sa}/topic",
						ParentPathParam:    []string{"cluster", "sa"},
						ListQueryParameter: map[string]FlagParameterOption{},
						Order:              42,
					},
				},
			},
		}
		if !reflect.DeepEqual(kinds, expected) {
			t.Error(spew.Printf("got kinds %v, want %v", kinds, expected))
		}
	})
}
func TestKindWithMissingMetadataField(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/missing_field_in_metadata.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		_, err = schema.GetConsoleKinds(true)
		if !strings.Contains(err.Error(), "parent path param sa not found in metadata for kind Topic") {
			t.Fatalf("Not expected error: %s", err)
		}
	})
}
func TestKindNotRequiredMetadataField(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/not_required_field_in_metadata.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		_, err = schema.GetConsoleKinds(true)
		if !strings.Contains(err.Error(), "parent path param sa in metadata for kind Topic not required") {
			t.Fatalf("Not expected error: %s", err)
		}
	})
}

func TestGetExecutes(t *testing.T) {
	t.Run("gets execute endpoint from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/console_run.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		result, err := schema.getRuns(CONSOLE)
		if err != nil {
			t.Fatalf("failed getting execute: %s", err)
		}

		//all the token runs are not present in real life just present in the yaml test file used here
		expected := RunCatalog{
			"partnerZoneGenerateCredentials": Run{
				Path:           "/public/partner-zone/v2/{partner-zone-name}/generate-credentials",
				Name:           "partnerZoneGenerateCredentials",
				PathParameter:  []string{"partner-zone-name"},
				Doc:            "generate a token for a partner zone service account",
				QueryParameter: map[string]FlagParameterOption{},
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "POST",
				BackendType:    CONSOLE,
			},
			"createAdminToken": Run{
				Path:           "/token/v1/admin_tokens",
				Name:           "createAdminToken",
				Doc:            "Create an admin token",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  []string{},
				BodyFields: map[string]FlagParameterOption{
					"name": {
						FlagName: "name",
						Required: true,
						Type:     "string",
					},
				},
				Method:      "POST",
				BackendType: CONSOLE,
			},
			"createApplicationInstanceToken": Run{
				Path:           "/token/v1/application_instance_tokens/{application-instance-name}",
				Name:           "createApplicationInstanceToken",
				Doc:            "Create an application instance token",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter: []string{
					"application-instance-name",
				},
				BodyFields: map[string]FlagParameterOption{
					"name": {
						FlagName: "name",
						Required: true,
						Type:     "string",
					},
				},
				Method:      "POST",
				BackendType: CONSOLE,
			},
			"deleteToken": Run{
				Path:           "/token/v1/{token-id}",
				Name:           "deleteToken",
				Doc:            "Delete a token",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  []string{"token-id"},
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "DELETE",
				BackendType:    CONSOLE,
			},
			"listAdminToken": Run{
				Path:           "/token/v1/admin_tokens",
				Name:           "listAdminToken",
				Doc:            "List admin token",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  []string{},
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "GET",
				BackendType:    CONSOLE,
			},
			"listApplicationInstanceToken": Run{
				Path:           "/token/v1/application_instance_tokens/{application-instance-name}",
				Name:           "listApplicationInstanceToken",
				Doc:            "List application instance token",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter: []string{
					"application-instance-name",
				},
				BodyFields:  map[string]FlagParameterOption{},
				Method:      "GET",
				BackendType: CONSOLE,
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Error(spew.Printf("got %v, want %v", result, expected))
		}
	})
}

func TestGetConnectorRuns(t *testing.T) {
	t.Run("parses connector stop / offsets runs, including the PATCH with a JSON array body", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/connector_run.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		result, err := schema.getRuns(CONSOLE)
		if err != nil {
			t.Fatalf("failed getting runs: %s", err)
		}

		offsetsPath := "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/offsets"
		connectorPathParams := []string{"cluster", "connectCluster", "connector-name"}
		expected := RunCatalog{
			"connectorStop": Run{
				Path:           "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/stop",
				Name:           "connectorStop",
				Doc:            "Stop a connector",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  connectorPathParams,
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "PUT",
				BackendType:    CONSOLE,
			},
			"connectorPause": Run{
				Path:           "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/pause",
				Name:           "connectorPause",
				Doc:            "Pause a connector",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  connectorPathParams,
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "PUT",
				BackendType:    CONSOLE,
			},
			"connectorResume": Run{
				Path:           "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/resume",
				Name:           "connectorResume",
				Doc:            "Resume a paused connector",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  connectorPathParams,
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "PUT",
				BackendType:    CONSOLE,
			},
			"connectorRestart": Run{
				Path:           "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/restart",
				Name:           "connectorRestart",
				Doc:            "Restart a connector",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  connectorPathParams,
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "PUT",
				BackendType:    CONSOLE,
			},
			"connectorGetOffsets": Run{
				Path:           offsetsPath,
				Name:           "connectorGetOffsets",
				Doc:            "Get the offsets of a connector",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  connectorPathParams,
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "GET",
				BackendType:    CONSOLE,
			},
			"connectorResetOffsets": Run{
				Path:           offsetsPath,
				Name:           "connectorResetOffsets",
				Doc:            "Reset all offsets of a stopped connector",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  connectorPathParams,
				BodyFields:     map[string]FlagParameterOption{},
				Method:         "DELETE",
				BackendType:    CONSOLE,
			},
			"connectorAlterOffsets": Run{
				Path:           offsetsPath,
				Name:           "connectorAlterOffsets",
				Doc:            "Alter the offsets of a stopped connector",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  connectorPathParams,
				BodyFields: map[string]FlagParameterOption{
					"offsets": {
						FlagName: "offsets",
						Required: false,
						Type:     "json",
					},
				},
				Method:      "PATCH",
				BackendType: CONSOLE,
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Error(spew.Printf("got %v, want %v", result, expected))
		}
	})

	t.Run("reads x-cdk-run-name-v2 and prefers it over x-cdk-run-name", func(t *testing.T) {
		// New endpoints are annotated with x-cdk-run-name-v2 only, so old CLIs
		// (which look only for x-cdk-run-name) skip them. The current CLI must
		// expose both the v2-only endpoints and the legacy x-cdk-run-name ones,
		// preferring the v2 name when an endpoint carries both keys.
		spec := []byte(`openapi: 3.0.0
info:
  title: run-version
  version: "1.0"
paths:
  /v2only:
    get:
      x-cdk-run-name-v2: newOnlyRun
      x-cdk-run-doc: A run only the new CLI can see
      responses:
        "200":
          description: ok
  /both:
    get:
      x-cdk-run-name: legacyName
      x-cdk-run-name-v2: preferredName
      responses:
        "200":
          description: ok
  /legacyonly:
    get:
      x-cdk-run-name: legacyOnlyRun
      responses:
        "200":
          description: ok
`)

		schema, err := NewOpenAPIParser(spec)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		result, err := schema.getRuns(CONSOLE)
		if err != nil {
			t.Fatalf("failed getting runs: %s", err)
		}

		if _, ok := result["newOnlyRun"]; !ok {
			t.Errorf("expected v2-only endpoint to be exposed as %q, got %v", "newOnlyRun", result)
		}
		if _, ok := result["preferredName"]; !ok {
			t.Errorf("expected v2 name to take precedence (%q), got %v", "preferredName", result)
		}
		if _, ok := result["legacyName"]; ok {
			t.Errorf("legacy name should be shadowed by the v2 name when both are present, got %v", result)
		}
		if _, ok := result["legacyOnlyRun"]; !ok {
			t.Errorf("expected legacy-only endpoint to still be exposed, got %v", result)
		}
	})

	t.Run("parses the topic add-partitions / empty / low-watermark runs", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/topic_run.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		result, err := schema.getRuns(CONSOLE)
		if err != nil {
			t.Fatalf("failed getting runs: %s", err)
		}

		topicPathParams := []string{"cluster", "topic-name"}
		expected := RunCatalog{
			// The partition count is a scalar (int) property of an object body,
			// so it surfaces as a typed --partition-count flag.
			"topicAddPartitions": Run{
				Path:           "/public/kafka/v2/cluster/{cluster}/topic/{topic-name}/partitions",
				Name:           "topicAddPartitions",
				Doc:            "Increase the number of partitions of a topic",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  topicPathParams,
				BodyFields: map[string]FlagParameterOption{
					"partitionCount": {
						FlagName: "partition-count",
						Required: true,
						Type:     "integer",
					},
				},
				Method:      "PUT",
				BackendType: CONSOLE,
			},
			// The partition to empty is an optional query parameter, not a body.
			"topicEmpty": Run{
				Path: "/public/kafka/v2/cluster/{cluster}/topic/{topic-name}/empty",
				Name: "topicEmpty",
				Doc:  "Empty a topic, or a single partition of it",
				QueryParameter: map[string]FlagParameterOption{
					"partition": {
						FlagName: "partition",
						Required: false,
						Type:     "integer",
					},
				},
				PathParameter: topicPathParams,
				BodyFields:    map[string]FlagParameterOption{},
				Method:        "PUT",
				BackendType:   CONSOLE,
			},
			// offsets is a map (object) property, so it is passed as a
			// JSON-encoded string flag we decode back into the body.
			"topicSetLowWatermark": Run{
				Path:           "/public/kafka/v2/cluster/{cluster}/topic/{topic-name}/low-watermark",
				Name:           "topicSetLowWatermark",
				Doc:            "Set a topic's low watermark to an arbitrary offset per partition",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  topicPathParams,
				BodyFields: map[string]FlagParameterOption{
					"offsets": {
						FlagName: "offsets",
						Required: true,
						Type:     "json",
					},
				},
				Method:      "PUT",
				BackendType: CONSOLE,
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Error(spew.Printf("got %v, want %v", result, expected))
		}
	})

	t.Run("exposes polymorphic (oneOf) body properties as json flags", func(t *testing.T) {
		// A body property that is a oneOf/anyOf/allOf carries no scalar type;
		// it must still surface as a json-encoded string flag (like the
		// consumerGroupResetOffsets reset payload) rather than being dropped.
		spec := []byte(`openapi: 3.0.0
info:
  title: run-version
  version: "1.0"
paths:
  /reset:
    post:
      x-cdk-run-name: resetRun
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ResetRequest'
      responses:
        "200":
          description: ok
components:
  schemas:
    ResetRequest:
      type: object
      required:
      - selection
      properties:
        selection:
          $ref: '#/components/schemas/Selection'
        note:
          type: string
    Selection:
      oneOf:
      - $ref: '#/components/schemas/AllTopics'
      - $ref: '#/components/schemas/OneTopic'
    AllTopics:
      type: object
      properties:
        type:
          type: string
    OneTopic:
      type: object
      properties:
        type:
          type: string
        topic:
          type: string
`)

		schema, err := NewOpenAPIParser(spec)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		result, err := schema.getRuns(CONSOLE)
		if err != nil {
			t.Fatalf("failed getting runs: %s", err)
		}

		run, ok := result["resetRun"]
		if !ok {
			t.Fatalf("expected resetRun to be present, got %v", result)
		}
		selection, ok := run.BodyFields["selection"]
		if !ok {
			t.Fatalf("expected oneOf property 'selection' to be exposed as a body field, got %v", run.BodyFields)
		}
		if selection.Type != "json" {
			t.Errorf("expected oneOf property to be a json flag, got %q", selection.Type)
		}
		if !selection.Required {
			t.Errorf("expected 'selection' to be required")
		}
		if note, ok := run.BodyFields["note"]; !ok || note.Type != "string" {
			t.Errorf("expected scalar sibling 'note' to remain a string flag, got %v", run.BodyFields)
		}
	})
}
