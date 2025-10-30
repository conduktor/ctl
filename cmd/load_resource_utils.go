package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/resource"
)

func loadResourceFromFileFlag(filePath []string, strict, recursiveFolder bool) []resource.Resource {
	var resources = make([]resource.Resource, 0)
	for _, path := range filePath {
		r, err := resourceForPath(path, strict, recursiveFolder)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		resources = append(resources, r...)
	}
	return resources
}

//nolint:staticcheck
const FILE_ARGS_DOC = "Specify the files or folders to apply. For folders, all .yaml or .yml files within the folder will be applied, while files in subfolders will be ignored."
