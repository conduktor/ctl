package utils

import (
	"testing"
)

func TestUpperCamelToKebab(t *testing.T) {
	t.Run("converts upper camel case to kebab case", func(t *testing.T) {
		got := UpperCamelToKebab("UpperCamelCase")
		want := "upper-camel-case"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("handles single word", func(t *testing.T) {
		got := UpperCamelToKebab("Word")
		want := "word"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		got := UpperCamelToKebab("")
		want := ""

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestKebabToUpperCamel(t *testing.T) {
	t.Run("converts kebab case to upper camel case", func(t *testing.T) {
		got := KebabToUpperCamel("kebab-case-to-upper")
		want := "KebabCaseToUpper"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("handles single word", func(t *testing.T) {
		got := KebabToUpperCamel("word")
		want := "Word"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		got := KebabToUpperCamel("")
		want := ""

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
