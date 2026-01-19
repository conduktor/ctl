package schema

import (
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetKindWithYamlFromGateway(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/gateway.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		kinds, err := schema.GetGatewayKinds(true)
		if err != nil {
			t.Fatalf("failed getting kinds: %s", err)
		}

		expected := KindCatalog{
			"VirtualCluster": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:               "VirtualCluster",
						ListPath:           "/gateway/v2/virtual-cluster",
						ParentPathParam:    []string{},
						ListQueryParameter: map[string]FlagParameterOption{},
						GetAvailable:       true,
						ApplyExample: `apiVersion: gateway/v2
kind: VirtualCluster
metadata:
    name: vcluster1
spec:
    aclEnabled: false
    superUsers:
        - username1
        - username2
    type: Standard
`,
						Order: 7,
					},
				},
			},
			"AliasTopic": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "AliasTopic",
						ListPath:        "/gateway/v2/alias-topic",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]FlagParameterOption{
							"name": {
								FlagName: "name",
								Required: false,
								Type:     "string",
							},
							"vcluster": {
								FlagName: "vcluster",
								Required: false,
								Type:     "string",
							},
							"showDefaults": {
								FlagName: "show-defaults",
								Required: false,
								Type:     "boolean",
							},
						},
						GetAvailable: false,
						Order:        8,
						ApplyExample: `apiVersion: gateway/v2
kind: AliasTopic
metadata:
    name: name1
    vCluster: vCluster1
spec:
    physicalName: physicalName1
`,
					},
				},
			},
			"ConcentrationRule": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "ConcentrationRule",
						ListPath:        "/gateway/v2/concentration-rule",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]FlagParameterOption{
							"vcluster": {
								FlagName: "vcluster",
								Required: false,
								Type:     "string",
							},
							"name": {
								FlagName: "name",
								Required: false,
								Type:     "string",
							},
							"showDefaults": {
								FlagName: "show-defaults",
								Required: false,
								Type:     "boolean",
							},
						},
						GetAvailable: false,
						Order:        9,
						ApplyExample: `apiVersion: gateway/v2
kind: ConcentrationRule
metadata:
    name: concentrationRule1
    vCluster: vCluster1
spec:
    autoManaged: false
    offsetCorrectness: false
    pattern: topic.*
    physicalTopics:
        compact: compact_topic
        delete: topic
        deleteCompact: compact_delete_topic
`,
					},
				},
			},
			"GatewayGroup": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "GatewayGroup",
						ListPath:        "/gateway/v2/group",
						ParentPathParam: []string{},
						ApplyExample: `apiVersion: gateway/v2
kind: GatewayGroup
metadata:
    name: group1
spec:
    externalGroups:
        - GROUP_READER
        - GROUP_WRITER
    members:
        - name: serviceAccount1
          vCluster: vCluster1
        - name: serviceAccount2
          vCluster: vCluster2
        - name: serviceAccount3
          vCluster: vCluster3
`,
						ListQueryParameter: map[string]FlagParameterOption{
							"showDefaults": {
								FlagName: "show-defaults",
								Required: false,
								Type:     "boolean",
							},
						},
						GetAvailable: true,
						Order:        11,
					},
				},
			},
			"GatewayServiceAccount": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "GatewayServiceAccount",
						ListPath:        "/gateway/v2/service-account",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]FlagParameterOption{
							"name": {
								FlagName: "name",
								Required: false,
								Type:     "string",
							},
							"type": {
								FlagName: "type",
								Required: false,
								Type:     "string",
							},
							"vcluster": {
								FlagName: "vcluster",
								Required: false,
								Type:     "string",
							},
							"showDefaults": {
								FlagName: "show-defaults",
								Required: false,
								Type:     "boolean",
							},
						},
						GetAvailable: false,
						Order:        10,
						ApplyExample: `apiVersion: gateway/v2
kind: GatewayServiceAccount
metadata:
    name: user1
    vCluster: vcluster1
spec:
    externalNames:
        - externalName
    type: EXTERNAL
`,
					},
				},
			},
			"Interceptor": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "Interceptor",
						ListPath:        "/gateway/v2/interceptor",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]FlagParameterOption{
							"username": {
								FlagName: "username",
								Required: false,
								Type:     "string",
							},
							"name": {
								FlagName: "name",
								Required: false,
								Type:     "string",
							},
							"global": {
								FlagName: "global",
								Required: false,
								Type:     "boolean",
							},
							"vcluster": {
								FlagName: "vcluster",
								Required: false,
								Type:     "string",
							},
							"group": {
								FlagName: "group",
								Required: false,
								Type:     "string",
							},
						},
						ApplyExample: `apiVersion: gateway/v2
kind: Interceptor
metadata:
    name: yellow_cars_filter
    scope:
        vCluster: vCluster1
spec:
    comment: Filter yellow cars
    config:
        statement: SELECT '$.type' as type, '$.price' as price FROM cars WHERE '$.color' = 'yellow'
        virtualTopic: yellow_cars
    pluginClass: io.conduktor.gateway.interceptor.VirtualSqlTopicPlugin
    priority: 1
`,
						GetAvailable: false,
						Order:        12,
					},
				},
			},
		}
		if !reflect.DeepEqual(kinds, expected) {
			t.Error(spew.Printf("got kinds %v, want %v", kinds, expected))
		}
	})
}
func TestGetExecutesForGateway(t *testing.T) {
	t.Run("gets execute endpoint from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("testdata/gateway_run.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := NewOpenAPIParser(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		result, err := schema.getRuns(GATEWAY)
		if err != nil {
			t.Fatalf("failed getting execute: %s", err)
		}

		//all the token runs are not present in real life just present in the yaml test file used here
		expected := RunCatalog{
			"generateServiceAccountToken": Run{
				Path:           "/gateway/v2/token",
				Name:           "generateServiceAccountToken",
				Doc:            "Generate a token for a service account on a virtual cluster",
				QueryParameter: map[string]FlagParameterOption{},
				PathParameter:  []string{},
				BodyFields: map[string]FlagParameterOption{
					"vCluster": {
						FlagName: "v-cluster",
						Required: false,
						Type:     "string",
					},
					"lifeTimeSeconds": {
						Type:     "integer",
						Required: true,
						FlagName: "life-time-seconds",
					},
					"username": {
						FlagName: "username",
						Required: true,
						Type:     "string",
					},
				},
				Method:      "POST",
				BackendType: GATEWAY,
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Error(spew.Printf("got %v, want %v", result, expected))
		}
	})
}
