package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

var dryRun *bool

func resourceForPath(path string, strict bool) ([]resource.Resource, error) {
	directory, err := isDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if directory {
		return resource.FromFolder(path, strict)
	} else {
		return resource.FromFile(path, strict)
	}
}

func initApply(kinds schema.KindCatalog, strict bool) {
	// applyCmd represents the apply command
	var filePath *[]string
	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Upsert a resource on Conduktor",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			resources := loadResourceFromFileFlag(*filePath, strict)
			schema.SortResourcesForApply(kinds, resources, *debug)
			allSuccess := true
			for _, resource := range resources {
				var upsertResult string
				var err error
				if isGatewayResource(resource, kinds) {
					upsertResult, err = gatewayApiClient().Apply(&resource, *dryRun)
				} else {
					upsertResult, err = consoleApiClient().Apply(&resource, *dryRun)
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
