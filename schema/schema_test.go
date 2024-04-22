package schema

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestGetKindWithYamlFromConsolePlus(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("docs.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := New(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		kinds, err := schema.GetKinds(true)
		if err != nil {
			t.Fatalf("failed getting kinds: %s", err)
		}

		expected := KindCatalog{
			"Application": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "Application",
						ListPath:        "/public/self-serve/v1/application",
						ParentPathParam: make([]string, 0),
					},
				},
			},
			"ApplicationInstance": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "ApplicationInstance",
						ListPath:        "/public/self-serve/v1/application-instance",
						ParentPathParam: make([]string, 0),
					},
				},
			},
			"ApplicationInstancePermission": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "ApplicationInstancePermission",
						ListPath:        "/public/self-serve/v1/application-instance-permission",
						ParentPathParam: make([]string, 0),
					},
				},
			},
			"TopicPolicy": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "TopicPolicy",
						ListPath:        "/public/self-serve/v1/topic-policy",
						ParentPathParam: make([]string, 0),
					},
				},
			},
			"Topic": {
				Versions: map[int]KindVersion{
					2: {
						Name:            "Topic",
						ListPath:        "/public/kafka/v2/cluster/{cluster}/topic",
						ParentPathParam: []string{"cluster"},
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
		schemaContent, err := os.ReadFile("multiple_version.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := New(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		kinds, err := schema.GetKinds(true)
		if err != nil {
			t.Fatalf("failed getting kinds: %s", err)
		}

		expected := KindCatalog{
			"Topic": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "Topic",
						ListPath:        "/public/v1/cluster/{cluster}/topic",
						ParentPathParam: []string{"cluster"},
					},
					2: {
						Name:            "Topic",
						ListPath:        "/public/v2/cluster/{cluster}/sa/{sa}/topic",
						ParentPathParam: []string{"cluster", "sa"},
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
		schemaContent, err := os.ReadFile("missing_field_in_metadata.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := New(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		_, err = schema.GetKinds(true)
		if !strings.Contains(err.Error(), "Parent path param sa not found in metadata for kind Topic") {
			t.Fatalf("Not expected error: %s", err)
		}
	})
}
func TestKindNotRequiredMetadataField(t *testing.T) {
	t.Run("gets kinds from schema", func(t *testing.T) {
		schemaContent, err := os.ReadFile("not_required_field_in_metadata.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := New(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		_, err = schema.GetKinds(true)
		if !strings.Contains(err.Error(), "Parent path param sa in metadata for kind Topic not required") {
			t.Fatalf("Not expected error: %s", err)
		}
	})
}
