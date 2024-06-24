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
		schemaContent, err := os.ReadFile("docs_without_order.yaml")
		if err != nil {
			t.Fatalf("failed reading file: %s", err)
		}

		schema, err := New(schemaContent)
		if err != nil {
			t.Fatalf("failed creating new schema: %s", err)
		}

		kinds, err := schema.GetKinds(false)
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
						Order:           DefaultPriority,
					},
				},
			},
			"ApplicationInstance": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "ApplicationInstance",
						ListPath:        "/public/self-serve/v1/application-instance",
						ParentPathParam: make([]string, 0),
						Order:           DefaultPriority,
					},
				},
			},
			"ApplicationInstancePermission": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "ApplicationInstancePermission",
						ListPath:        "/public/self-serve/v1/application-instance-permission",
						ParentPathParam: make([]string, 0),
						Order:           DefaultPriority,
					},
				},
			},
			"TopicPolicy": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "TopicPolicy",
						ListPath:        "/public/self-serve/v1/topic-policy",
						ParentPathParam: make([]string, 0),
						Order:           DefaultPriority,
					},
				},
			},
			"Topic": {
				Versions: map[int]KindVersion{
					2: {
						Name:            "Topic",
						ListPath:        "/public/kafka/v2/cluster/{cluster}/topic",
						ParentPathParam: []string{"cluster"},
						Order:           DefaultPriority,
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
		schemaContent, err := os.ReadFile("docs_with_order.yaml")
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
						ParentPathParam: []string{},
						Order:           6,
					},
				},
			},
			"ApplicationInstance": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "ApplicationInstance",
						ListPath:        "/public/self-serve/v1/application-instance",
						ParentPathParam: []string{},
						Order:           7,
					},
				},
			},
			"ApplicationInstancePermission": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "ApplicationInstancePermission",
						ListPath:        "/public/self-serve/v1/application-instance-permission",
						ParentPathParam: []string{},
						Order:           8,
					},
				},
			},
			"ApplicationGroup": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "ApplicationGroup",
						ListPath:        "/public/self-serve/v1/application-group",
						ParentPathParam: []string{},
						Order:           9,
					},
				},
			},
			"TopicPolicy": {
				Versions: map[int]KindVersion{
					1: {
						Name:            "TopicPolicy",
						ListPath:        "/public/self-serve/v1/topic-policy",
						ParentPathParam: []string{},
						Order:           5,
					},
				},
			},
			"Topic": {
				Versions: map[int]KindVersion{
					2: {
						Name:            "Topic",
						ListPath:        "/public/kafka/v2/cluster/{cluster}/topic",
						ParentPathParam: []string{"cluster"},
						Order:           3,
					},
				},
			},
			"Subject": {
				Versions: map[int]KindVersion{
					2: {
						Name:            "Subject",
						ListPath:        "/public/kafka/v2/cluster/{cluster}/subject",
						ParentPathParam: []string{"cluster"},
						Order:           4,
					},
				},
			},
			"User": {
				Versions: map[int]KindVersion{
					2: {
						Name:            "User",
						ListPath:        "/public/iam/v2/user",
						ParentPathParam: []string{},
						Order:           0,
					},
				},
			},
			"Group": {
				Versions: map[int]KindVersion{
					2: {
						Name:            "Group",
						ListPath:        "/public/iam/v2/group",
						ParentPathParam: []string{},
						Order:           1,
					},
				},
			},
			"KafkaCluster": {
				Versions: map[int]KindVersion{
					2: {
						Name:            "KafkaCluster",
						ListPath:        "/public/console/v2/kafka-cluster",
						ParentPathParam: []string{},
						Order:           2,
					},
				},
			},
		}
		if !reflect.DeepEqual(kinds, expected) {
			t.Error(spew.Printf("got kinds %v, want %v", kinds["Subject"], expected["Subject"]))
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

		kinds, err := schema.GetKinds(false)
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
						Order:           DefaultPriority,
					},
					2: {
						Name:            "Topic",
						ListPath:        "/public/v2/cluster/{cluster}/sa/{sa}/topic",
						ParentPathParam: []string{"cluster", "sa"},
						Order:           42,
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
