package cli

import (
	"os"

	"github.com/conduktor/ctl/pkg/resource"
)

// LoadResourcesFromFiles loads resources from multiple file paths.
func LoadResourcesFromFiles(filePaths []string, strict, recursiveFolder bool) ([]resource.Resource, error) {
	var allResources []resource.Resource

	for _, path := range filePaths {
		resources, err := ResourceForPath(path, strict, recursiveFolder)
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, resources...)
	}

	return allResources, nil
}

// ResourceForPath loads resources from a single path (file or directory).
func ResourceForPath(path string, strict, recursiveFolder bool) ([]resource.Resource, error) {
	directory, err := IsDirectory(path)
	if err != nil {
		return nil, err
	}
	if directory {
		return resource.FromFolder(path, strict, recursiveFolder)
	} else {
		return resource.FromFile(path, strict)
	}
}

// IsDirectory checks if the given path is a directory.
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}
