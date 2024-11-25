package schema

import (
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetKindWithYamlFromGateway(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("gateway.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := New(schemaContent)
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
						ListQueryParameter: map[string]QueryParameterOption{},
						GetAvailable:       true,
						Order:              7,
					},
				},
			},
			"AliasTopic": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "AliasTopic",
						ListPath:        "/gateway/v2/alias-topic",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]QueryParameterOption{
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
					},
				},
			},
			"ConcentrationRule": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "ConcentrationRule",
						ListPath:        "/gateway/v2/concentration-rule",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]QueryParameterOption{
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
					},
				},
			},
			"GatewayGroup": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "GatewayGroup",
						ListPath:        "/gateway/v2/group",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]QueryParameterOption{
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
						ListQueryParameter: map[string]QueryParameterOption{
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
					},
				},
			},
			"Interceptor": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "Interceptor",
						ListPath:        "/gateway/v2/interceptor",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]QueryParameterOption{
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
