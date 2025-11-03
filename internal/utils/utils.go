package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func CamelToKebab(input string) string {
	// Split the input string into words
	words := make([]string, 0)
	currentWord := ""
	for _, char := range input {
		if char >= 'A' && char <= 'Z' {
			if currentWord != "" {
				words = append(words, currentWord)
			}
			currentWord = string(char)
		} else {
			currentWord += string(char)
		}
	}
	if currentWord != "" {
		words = append(words, currentWord)
	}

	// Join the words with hyphens
	kebabCase := strings.ToLower(strings.Join(words, "-"))

	return kebabCase
}

func KebabToUpperCamel(input string) string {
	// Split the input string into words
	words := strings.Split(input, "-")

	// Capitalize the first letter of each word
	upperCamelCase := ""
	for _, word := range words {
		upperCamelCase += cases.Title(language.English).String(word)
	}

	return upperCamelCase
}
