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
			"VClusters": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:               "VClusters",
						ListPath:           "/gateway/v2/vclusters",
						ParentPathParam:    []string{},
						ListQueryParameter: map[string]QueryParameterOption{},
						GetAvailable:       true,
						Order:              7,
					},
				},
			},
			"AliasTopics": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "AliasTopics",
						ListPath:        "/gateway/v2/alias-topics",
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
			"ConcentrationRules": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "ConcentrationRules",
						ListPath:        "/gateway/v2/concentration-rules",
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
			"GatewayGroups": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "GatewayGroups",
						ListPath:        "/gateway/v2/gateway-groups",
						ParentPathParam: []string{},
						ListQueryParameter: map[string]QueryParameterOption{
							"showDefaults": {
								FlagName: "show-defaults",
								Required: false,
								Type:     "boolean",
							},
						},
						GetAvailable: true,
						Order:        10,
					},
				},
			},
			"ServiceAccounts": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "ServiceAccounts",
						ListPath:        "/gateway/v2/service-accounts",
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
						Order:        11,
					},
				},
			},
			"Interceptors": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "Interceptors",
						ListPath:        "/gateway/v2/interceptors",
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
