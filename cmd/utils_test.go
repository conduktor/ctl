package cmd

import (
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"strconv"
	"testing"
)

func TestExtractFlagValueForQueryParam(t *testing.T) {
	params := map[string]interface{}{
		"stringParam":   func() *string { s := "test"; return &s }(),
		"emptyString":   func() *string { s := ""; return &s }(),
		"boolParam":     func() *bool { b := true; return &b }(),
		"intParam":      func() *int { i := 123; return &i }(),
		"unprobableInt": func() *int { i := unprobableInt; return &i }(),
	}

	expected := map[string]string{
		"stringParam": "test",
		"boolParam":   strconv.FormatBool(true),
		"intParam":    strconv.Itoa(123),
	}

	result := extractFlagValueForQueryParam(params)

	if !reflect.DeepEqual(result, expected) {
		t.Error(spew.Printf("got %v, want %v", result, expected))
	}
}

func TestExtractFlagValueForBodyParam(t *testing.T) {
	params := map[string]interface{}{
		"stringParam":   func() *string { s := "test"; return &s }(),
		"emptyString":   func() *string { s := ""; return &s }(),
		"boolParam":     func() *bool { b := true; return &b }(),
		"intParam":      func() *int { i := 123; return &i }(),
		"unprobableInt": func() *int { i := unprobableInt; return &i }(),
		"otherParam":    "otherValue",
	}

	expected := map[string]interface{}{
		"stringParam": func() *string { s := "test"; return &s }(),
		"boolParam":   func() *bool { b := true; return &b }(),
		"intParam":    123,
		"otherParam":  "otherValue",
	}

	result := extractFlagValueForBodyParam(params)
	if !reflect.DeepEqual(result, expected) {
		t.Error(spew.Printf("got %v, want %v", result, expected))
	}
}
