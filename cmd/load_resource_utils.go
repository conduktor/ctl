package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/resource"
)

func loadResourceFromFileFlag(filePath []string, strict bool) []resource.Resource {
	var resources = make([]resource.Resource, 0)
	for _, path := range filePath {
		r, err := resourceForPath(path, strict)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		resources = append(resources, r...)
	}
	return resources
}
