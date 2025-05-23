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

		schema, err := NewOpenApiParser(schemaContent)
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
						ApplyExample: `kind: VirtualCluster
apiVersion: gateway/v2
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
						ApplyExample: `kind: AliasTopic
apiVersion: gateway/v2
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
						ApplyExample: `kind: ConcentrationRule
apiVersion: gateway/v2
metadata:
    name: concentrationRule1
    vCluster: vCluster1
spec:
    pattern: topic.*
    physicalTopics:
        delete: topic
        compact: compact_topic
        deleteCompact: compact_delete_topic
    autoManaged: false
    offsetCorrectness: false
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
						ApplyExample: `kind: GatewayGroup
apiVersion: gateway/v2
metadata:
    name: group1
spec:
    members:
        - vCluster: vCluster1
          name: serviceAccount1
        - vCluster: vCluster2
          name: serviceAccount2
        - vCluster: vCluster3
          name: serviceAccount3
    externalGroups:
        - GROUP_READER
        - GROUP_WRITER
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
						ApplyExample: `kind: GatewayServiceAccount
apiVersion: gateway/v2
metadata:
    name: user1
    vCluster: vcluster1
spec:
    type: EXTERNAL
    externalNames:
        - externalName
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
						ApplyExample: `kind: Interceptor
apiVersion: gateway/v2
metadata:
    name: yellow_cars_filter
    scope:
        vCluster: vCluster1
spec:
    comment: Filter yellow cars
    pluginClass: io.conduktor.gateway.interceptor.VirtualSqlTopicPlugin
    priority: 1
    config:
        virtualTopic: yellow_cars
        statement: SELECT '$.type' as type, '$.price' as price FROM cars WHERE '$.color' = 'yellow'
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

		schema, err := NewOpenApiParser(schemaContent)
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
