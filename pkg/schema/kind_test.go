package schema

import (
	"reflect"
	"testing"
)

func TestKindGetFlag(t *testing.T) {
	t.Run("converts parent parameters to flags", func(t *testing.T) {

		kind := Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{
					ParentPathParam: []string{"param-1", "param-2", "param-3"},
				},
			},
		}

		got := kind.GetParentFlag()
		want := []string{"param-1", "param-2", "param-3"}

		if len(got) != len(want) {
			t.Fatalf("got %d flags, want %d", len(got), len(want))
		}

		for i, flag := range got {
			if flag != want[i] {
				t.Errorf("got flag %q at index %d, want %q", flag, i, want[i])
			}
		}
	})
}

func TestKindGetFlagWhenNoFlag(t *testing.T) {
	t.Run("converts parent parameters to flags", func(t *testing.T) {
		kind := Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{
					ParentPathParam: []string{},
				},
			},
		}

		got := kind.GetParentFlag()

		if len(got) != 0 {
			t.Fatalf("got %d flags, want %d", len(got), 0)
		}
	})
}

func TestKindListPath(t *testing.T) {
	t.Run("replaces parent parameters in ListPath and add parameters", func(t *testing.T) {
		kind := Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{
					ListPath:         "/ListPath/{param-1}/{param-2}",
					ParentPathParam:  []string{"param-1", "param-2"},
					ParentQueryParam: []string{"param-3"},
				},
			},
		}

		got := kind.ListPath([]string{"value1", "value2"}, []string{"value3"})
		wantPath := "/ListPath/value1/value2"
		wantParams := []QueryParam{{Name: "param-3", Value: "value3"}}

		if got.Path != wantPath {
			t.Errorf("got ListPath %q, want %q", got.Path, wantPath)
		}

		if !reflect.DeepEqual(got.QueryParams, wantParams) {
			t.Errorf("got ListPath params %q, want %q", got.QueryParams, wantParams)
		}
	})

	t.Run("panics when parent paths and parameters length mismatch", func(t *testing.T) {
		kind := Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{
					ListPath:        "/ListPath/{param1}/{param2}",
					ParentPathParam: []string{"param1", "param2"},
				},
			},
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()

		kind.ListPath([]string{"value1"}, []string{})
	})

	t.Run("panics when parent paths and parameters length mismatch", func(t *testing.T) {
		kind := Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{
					ListPath:         "/Test",
					ParentPathParam:  []string{},
					ParentQueryParam: []string{"param1", "param2"},
				},
			},
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()

		kind.ListPath([]string{}, []string{"value1"})
	})
}
