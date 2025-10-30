package schema

import (
	"reflect"
	"testing"

	"github.com/conduktor/ctl/internal/resource"
)

func TestSortResources(t *testing.T) {
	catalog := KindCatalog{
		"kind1": Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{Order: 1},
			},
		},
		"kind2": Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{Order: 2},
			},
		},
		"kind3": Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{Order: 3},
				2: &ConsoleKindVersion{Order: 4},
			},
		},
	}

	resources := []resource.Resource{
		{Kind: "kind3", Version: "v1"},
		{Kind: "kind3", Version: "v2"},
		{Kind: "kind3", Version: "v4"},
		{Kind: "kind1", Version: "v1"},
		{Kind: "kind2", Version: "v1"},
	}

	SortResourcesForApply(catalog, resources, false)

	expected := []resource.Resource{
		{Kind: "kind1", Version: "v1"},
		{Kind: "kind2", Version: "v1"},
		{Kind: "kind3", Version: "v1"},
		{Kind: "kind3", Version: "v2"},
		{Kind: "kind3", Version: "v4"},
	}

	// Check if the resources are sorted in the expected order
	if !reflect.DeepEqual(resources, expected) {
		t.Errorf("Resources are not sorted in the expected order. Got: %v, want: %v", resources, expected)
	}
}
