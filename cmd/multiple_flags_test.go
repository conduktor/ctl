package cmd

import (
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"reflect"
	"strconv"
	"testing"
)

func TestExtractFlagValueForQueryParam(t *testing.T) {
	command := &cobra.Command{}
	multipleFlags := NewMultipleFlags(command, map[string]schema.FlagParameterOption{
		"stringParam":    {FlagName: "stringParam", Type: "string"},
		"notSetString":   {FlagName: "notSetStrign", Type: "string"},
		"emptyString":    {FlagName: "emptyString", Type: "string"},
		"boolParam":      {FlagName: "boolParam", Type: "boolean"},
		"boolParamFalse": {FlagName: "boolParamFalse", Type: "boolean"},
		"notSetBoolean":  {FlagName: "notSetBoolean", Type: "boolean"},
		"intParam":       {FlagName: "intParam", Type: "integer"},
		"notSetInt":      {FlagName: "notSetInt", Type: "integer"},
		"zeroParam":      {FlagName: "zeroParam", Type: "integer"},
	})
	multipleFlags.result = map[string]interface{}{
		"stringParam":    func() *string { s := "test"; return &s }(),
		"notSetString":   func() *string { s := ""; return &s }(),
		"emptyString":    func() *string { s := ""; return &s }(),
		"boolParam":      func() *bool { b := true; return &b }(),
		"boolParamFalse": func() *bool { b := false; return &b }(),
		"notSetBoolean":  func() *bool { b := false; return &b }(),
		"intParam":       func() *int { i := 123; return &i }(),
		"notSetInt":      func() *int { i := 0; return &i }(),
		"zeroParam":      func() *int { i := 0; return &i }(),
	}

	expected := map[string]string{
		"stringParam":    "test",
		"boolParam":      strconv.FormatBool(true),
		"intParam":       strconv.Itoa(123),
		"zeroParam":      strconv.Itoa(0),
		"emptyString":    "",
		"boolParamFalse": strconv.FormatBool(false),
	}

	command.Flags().Lookup("stringParam").Changed = true
	command.Flags().Lookup("boolParam").Changed = true
	command.Flags().Lookup("boolParamFalse").Changed = true
	command.Flags().Lookup("intParam").Changed = true
	command.Flags().Lookup("zeroParam").Changed = true
	command.Flags().Lookup("emptyString").Changed = true
	result := multipleFlags.ExtractFlagValueForQueryParam()

	if !reflect.DeepEqual(result, expected) {
		t.Error(spew.Printf("got %v, want %v", result, expected))
	}
}

func TestExtractFlagValueForBodyParam(t *testing.T) {
	command := &cobra.Command{}
	multipleFlags := NewMultipleFlags(command, map[string]schema.FlagParameterOption{
		"stringParam":    {FlagName: "stringParam", Type: "string"},
		"notSetString":   {FlagName: "notSetStrign", Type: "string"},
		"emptyString":    {FlagName: "emptyString", Type: "string"},
		"boolParam":      {FlagName: "boolParam", Type: "boolean"},
		"boolParamFalse": {FlagName: "boolParamFalse", Type: "boolean"},
		"notSetBoolean":  {FlagName: "notSetBoolean", Type: "boolean"},
		"intParam":       {FlagName: "intParam", Type: "integer"},
		"notSetInt":      {FlagName: "notSetInt", Type: "integer"},
		"zeroParam":      {FlagName: "zeroParam", Type: "integer"},
		"jsonArray":      {FlagName: "jsonArray", Type: "json"},
		"jsonObject":     {FlagName: "jsonObject", Type: "json"},
		"notSetJson":     {FlagName: "notSetJson", Type: "json"},
	})
	multipleFlags.result = map[string]interface{}{
		"stringParam":    func() *string { s := "test"; return &s }(),
		"notSetString":   func() *string { s := ""; return &s }(),
		"emptyString":    func() *string { s := ""; return &s }(),
		"boolParam":      func() *bool { b := true; return &b }(),
		"boolParamFalse": func() *bool { b := false; return &b }(),
		"notSetBoolean":  func() *bool { b := false; return &b }(),
		"intParam":       func() *int { i := 123; return &i }(),
		"notSetInt":      func() *int { i := 0; return &i }(),
		"zeroParam":      func() *int { i := 0; return &i }(),
		"jsonArray":      func() *string { s := `[{"partition":1,"offset":42}]`; return &s }(),
		"jsonObject":     func() *string { s := `{"key":"value"}`; return &s }(),
		"notSetJson":     func() *string { s := ""; return &s }(),
	}

	expected := map[string]interface{}{
		"stringParam":    func() *string { s := "test"; return &s }(),
		"emptyString":    func() *string { s := ""; return &s }(),
		"boolParam":      func() *bool { b := true; return &b }(),
		"boolParamFalse": func() *bool { b := false; return &b }(),
		"intParam":       func() *int { i := 123; return &i }(),
		"zeroParam":      func() *int { i := 0; return &i }(),
		"jsonArray":      []interface{}{map[string]interface{}{"partition": float64(1), "offset": float64(42)}},
		"jsonObject":     map[string]interface{}{"key": "value"},
	}

	command.Flags().Lookup("stringParam").Changed = true
	command.Flags().Lookup("boolParam").Changed = true
	command.Flags().Lookup("boolParamFalse").Changed = true
	command.Flags().Lookup("intParam").Changed = true
	command.Flags().Lookup("zeroParam").Changed = true
	command.Flags().Lookup("emptyString").Changed = true
	command.Flags().Lookup("jsonArray").Changed = true
	command.Flags().Lookup("jsonObject").Changed = true
	result, err := multipleFlags.ExtractFlagValueForBodyParam()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Error(spew.Printf("got %v, want %v", result, expected))
	}
}

func TestExtractFlagValueForBodyParamInvalidJSON(t *testing.T) {
	command := &cobra.Command{}
	multipleFlags := NewMultipleFlags(command, map[string]schema.FlagParameterOption{
		"jsonParam": {FlagName: "jsonParam", Type: "json"},
	})
	multipleFlags.result = map[string]interface{}{
		"jsonParam": func() *string { s := "not-json"; return &s }(),
	}
	command.Flags().Lookup("jsonParam").Changed = true

	_, err := multipleFlags.ExtractFlagValueForBodyParam()
	if err == nil {
		t.Error("expected an error for invalid JSON, got nil")
	}
}
