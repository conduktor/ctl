package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

var filePath *[]string
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
	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Upsert a resource on Conduktor",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			var resources = make([]resource.Resource, 0)
			for _, path := range *filePath {
				r, err := resourceForPath(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
				resources = append(resources, r...)
			}
			var allSuccess = true
			schema.SortResources(kinds, resources, *debug)
			for _, resource := range resources {
				upsertResult, err := apiClient().Apply(&resource, *dryRun)
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

	// Here you will define your flags and configuration settings.
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
