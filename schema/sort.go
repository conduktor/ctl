package schema

import (
	_ "embed"
	"fmt"
	"os"
	"sort"

	"github.com/conduktor/ctl/resource"
)

func resourcePriority(catalog KindCatalog, resource resource.Resource, debug, fallbackToDefaultCatalog bool) int {
	kind, ok := catalog[resource.Kind]
	if !ok {
		if debug {
			fmt.Fprintf(os.Stderr, "Could not find kind: %s in catalog, default to DefaultPriority for resource ordering\n", resource.Kind)
		}
		return DefaultPriority
	}
	version := extractVersionFromApiVersion(resource.Version)
	kindVersion, ok := kind.Versions[version]
	if !ok {
		if debug {
			fmt.Fprintf(os.Stderr, "Could not find version: %d of kind %s in catalog, default to DefaultPriority for resource ordering\n", version, resource.Kind)
		}
		return DefaultPriority
	} else {
		order := kindVersion.Order
		if order == DefaultPriority && fallbackToDefaultCatalog {
			defaultCatalog := DefaultKind()
			orderFromDefaultCatalog := resourcePriority(defaultCatalog, resource, false, false)
			if orderFromDefaultCatalog != DefaultPriority && debug {
				fmt.Fprintf(os.Stderr, "Could not find version: %d of kind %s in catalog, but find it in default catalog with priority %d\n", version, resource.Kind, orderFromDefaultCatalog)
			}
			return orderFromDefaultCatalog
		} else {
			return kindVersion.Order
		}
	}
}

func SortResources(catalog KindCatalog, resources []resource.Resource, debug bool) {
	sort.SliceStable(resources, func(i, j int) bool {
		return resourcePriority(catalog, resources[i], debug, true) < resourcePriority(catalog, resources[j], debug, true)
	})

}
