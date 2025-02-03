package schema

import (
	"testing"
)

func TestBuildPath(t *testing.T) {
	run := &Run{
		Path:          "/api/v1/resource/{p1}/subresource/{p2}",
		PathParameter: []string{"p1", "p2"},
		Name:          "TestRun",
	}

	pathValues := []string{"123", "456"}
	expectedPath := "/api/v1/resource/123/subresource/456"

	result := run.BuildPath(pathValues)

	if result != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, result)
	}
}

func TestBuildPathPanicOnWrongSizeTooSmall(t *testing.T) {
	run := &Run{
		Path:          "/api/v1/resource/{p1}/subresource/{p2}",
		PathParameter: []string{"p1", "p2"},
		Name:          "TestRun",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but did not occur")
		}
	}()

	run.BuildPath([]string{"123"})
}

func TestBuildPathPanicOnWrongSizeTooBig(t *testing.T) {
	run := &Run{
		Path:          "/api/v1/resource/{p1}/subresource/{p2}",
		PathParameter: []string{"p1", "p2"},
		Name:          "TestRun",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but did not occur")
		}
	}()

	run.BuildPath([]string{"123", "3", "4"})
}
