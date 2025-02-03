package schema

import (
	"reflect"
	"testing"
)

// TODO: test for colision
func TestMerge(t *testing.T) {
	catalog1 := &Catalog{
		Kind: KindCatalog{
			"kind1": Kind{Versions: map[int]KindVersion{1: &ConsoleKindVersion{}}},
		},
		Run: RunCatalog{
			"run1": Run{BackendType: CONSOLE},
		},
	}

	catalog2 := &Catalog{
		Kind: KindCatalog{
			"kind2": Kind{Versions: map[int]KindVersion{1: &GatewayKindVersion{}}},
		},
		Run: RunCatalog{
			"run2": Run{BackendType: GATEWAY},
		},
	}

	expected := Catalog{
		Kind: KindCatalog{
			"kind1": Kind{Versions: map[int]KindVersion{1: &ConsoleKindVersion{}}},
			"kind2": Kind{Versions: map[int]KindVersion{1: &GatewayKindVersion{}}},
		},
		Run: RunCatalog{
			"run1": Run{BackendType: CONSOLE},
			"run2": Run{BackendType: GATEWAY},
		},
	}

	mergedCatalog := catalog1.Merge(catalog2)

	if !reflect.DeepEqual(mergedCatalog, expected) {
		t.Errorf("expected %v, got %v", expected, mergedCatalog)
	}
}
