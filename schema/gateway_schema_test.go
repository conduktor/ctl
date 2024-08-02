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
			"Vclusters": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "Vclusters",
						ListPath:        "/gateway/v2/vclusters",
						ParentPathParam: []string{},
						Order:           7,
					},
				},
			},
			"AliasTopics": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "AliasTopics",
						ListPath:        "/gateway/v2/alias-topics",
						ParentPathParam: []string{},
						Order:           8,
					},
				},
			},
			"ConcentrationRules": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "ConcentrationRules",
						ListPath:        "/gateway/v2/concentration-rules",
						ParentPathParam: []string{},
						Order:           9,
					},
				},
			},
			"GatewayGroups": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "GatewayGroups",
						ListPath:        "/gateway/v2/gateway-groups",
						ParentPathParam: []string{},
						Order:           10,
					},
				},
			},
			"ServiceAccounts": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "ServiceAccounts",
						ListPath:        "/gateway/v2/service-accounts",
						ParentPathParam: []string{},
						Order:           11,
					},
				},
			},
			"Interceptors": {
				Versions: map[int]KindVersion{
					2: &GatewayKindVersion{
						Name:            "Interceptors",
						ListPath:        "/gateway/v2/interceptors",
						ParentPathParam: []string{},
						Order:           12,
					},
				},
			},
		}
		if !reflect.DeepEqual(kinds, expected) {
			t.Error(spew.Printf("got kinds %v, want %v", kinds, expected))
		}
	})
}
