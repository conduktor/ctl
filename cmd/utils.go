package cmd

import "strings"

func removeTrailingSIfAny(name string) string {
	return strings.TrimSuffix(name, "s")
}

func buildAlias(name string) []string {
	aliases := []string{strings.ToLower(name)}
	// This doesn't seem to be needed since none of the kinds ends with an S
	// However I'm leaving it here as a conditional so it won't affect the usage
	if strings.HasSuffix(name, "s") {
		aliases = append(aliases, removeTrailingSIfAny(name), removeTrailingSIfAny(strings.ToLower(name)))
	}
	return aliases
}
