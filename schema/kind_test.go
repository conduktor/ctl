package schema

import (
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

		got := kind.GetFlag()
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

		got := kind.GetFlag()

		if len(got) != 0 {
			t.Fatalf("got %d flags, want %d", len(got), 0)
		}
	})
}

func TestKindListPath(t *testing.T) {
	t.Run("replaces parent parameters in ListPath", func(t *testing.T) {
		kind := Kind{
			Versions: map[int]KindVersion{
				1: &ConsoleKindVersion{
					ListPath:        "/ListPath/{param-1}/{param-2}",
					ParentPathParam: []string{"param-1", "param-2"},
				},
			},
		}

		got := kind.ListPath([]string{"value1", "value2"})
		want := "/ListPath/value1/value2"

		if got != want {
			t.Errorf("got ListPath %q, want %q", got, want)
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

		kind.ListPath([]string{"value1"})
	})
}
