package cmd

import (
	"fmt"
	"os"
	"sync"

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

func runApply(kinds schema.KindCatalog, filePath []string, strict bool) {
	resources := loadResourceFromFileFlag(filePath, strict)
	schema.SortResourcesForApply(kinds, resources, *debug)
	// Group resources by kind
	kindGroups := make(map[string][]resource.Resource)
	for _, resrc := range resources {
		kindGroups[resrc.Kind] = append(kindGroups[resrc.Kind], resrc)
	}

	results := make([]struct {
		Resource     resource.Resource
		UpsertResult string
		Err          error
	}, 0, len(resources))

	for _, group := range kindGroups {
		var wg sync.WaitGroup
		kindResults := make([]struct {
			Resource     resource.Resource
			UpsertResult string
			Err          error
		}, len(group))

		for i, resrc := range group {
			if isGatewayResource(resrc, kinds) {
				upsertResult, err := gatewayApiClient().Apply(&resrc, *dryRun)
				kindResults[i] = struct {
					Resource     resource.Resource
					UpsertResult string
					Err          error
				}{resrc, upsertResult, err}
			} else {
				wg.Add(1)
				go func(i int, resrc resource.Resource) {
					defer wg.Done()
					upsertResult, err := consoleApiClient().Apply(&resrc, *dryRun)
					kindResults[i] = struct {
						Resource     resource.Resource
						UpsertResult string
						Err          error
					}{resrc, upsertResult, err}
				}(i, resrc)
			}
		}
		wg.Wait()
		results = append(results, kindResults...)
	}

	allSuccess := true
	for _, res := range results {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "Could not apply resource %s/%s: %s\n", res.Resource.Kind, res.Resource.Name, res.Err)
			allSuccess = false
		} else if res.UpsertResult != "" {
			fmt.Printf("%s/%s: %s\n", res.Resource.Kind, res.Resource.Name, res.UpsertResult)
		}
	}
	if !allSuccess {
		os.Exit(1)
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
			runApply(kinds, *filePath, strict)
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
