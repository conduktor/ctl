package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

var dryRun *bool

func resourceForPath(path string) ([]resource.Resource, error) {
	directory, err := isDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if directory {
		return resource.FromFolder(path)
	} else {
		return resource.FromFile(path)
	}
}

func initApply(kinds schema.KindCatalog) {
	// applyCmd represents the apply command
	var filePath *[]string
	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Upsert a resource on Conduktor",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			resources := loadResourceFromFileFlag(*filePath)
			schema.SortResourcesForApply(kinds, resources, *debug)
			allSuccess := true
			for _, resource := range resources {
				var upsertResult string
				var err error
				if strings.Contains(resource.Version, "gateway") {
					upsertResult, err = gatewayApiClient().Apply(&resource, *dryRun)
				} else {
					upsertResult, err = apiClient().Apply(&resource, *dryRun)
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not apply resource %s/%s: %s\n", resource.Kind, resource.Name, err)
					allSuccess = false
				} else {
					fmt.Printf("%s/%s: %s\n", resource.Kind, resource.Name, upsertResult)
				}
			}
			if !allSuccess {
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(applyCmd)

	filePath = applyCmd.
		PersistentFlags().StringArrayP("file", "f", make([]string, 0, 0), "Specify the files to apply")

	dryRun = applyCmd.
		PersistentFlags().Bool("dry-run", false, "Don't really apply change but check on backend the effect if applied")

	applyCmd.MarkPersistentFlagRequired("file")
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
