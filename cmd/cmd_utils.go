package cmd

import (
	"strings"
)

//nolint:staticcheck
const FILE_ARGS_DOC = "Specify the files or folders to apply. For folders, all .yaml or .yml files within the folder will be applied, while files in subfolders will be ignored."

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
