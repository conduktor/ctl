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
var runInParallel *bool

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

// Globally accessible for testing purposes
func ApplyResources(resources []resource.Resource,
	applyFunc func(*resource.Resource, bool) (string, error),
	dryRun bool,
	runInParallel bool,
	logFunc func(res resource.Resource, upsertResult string, err error)) []struct {
	Resource     resource.Resource
	UpsertResult string
	Err          error
} {
	results := make([]struct {
		Resource     resource.Resource
		UpsertResult string
		Err          error
	}, len(resources))

	if runInParallel {
		var wg sync.WaitGroup
		var mu sync.Mutex // for logging in parallel; prevents interleaving of log messages
		// by multiple goroutines writing to stdout
		for i, resrc := range resources {
			wg.Add(1)
			go func(i int, resrc resource.Resource) {
				defer wg.Done()
				upsertResult, err := applyFunc(&resrc, dryRun)
				results[i] = struct {
					Resource     resource.Resource
					UpsertResult string
					Err          error
				}{resrc, upsertResult, err}
				mu.Lock()
				logFunc(resrc, upsertResult, err)
				mu.Unlock()
			}(i, resrc)
		}
		wg.Wait()
	} else {
		for i, resrc := range resources {
			upsertResult, err := applyFunc(&resrc, dryRun)
			results[i] = struct {
				Resource     resource.Resource
				UpsertResult string
				Err          error
			}{resrc, upsertResult, err}
			logFunc(resrc, upsertResult, err)
		}
	}
	return results
}

func runApply(kinds schema.KindCatalog, filePath []string, strict bool) {
	resources := loadResourceFromFileFlag(filePath, strict)
	schema.SortResourcesForApply(kinds, resources, *debug)
	// Group resources by kind
	kindGroups := make(map[string][]resource.Resource)
	var kindOrder []string
	for _, resrc := range resources {
		if _, exists := kindGroups[resrc.Kind]; !exists {
			kindOrder = append(kindOrder, resrc.Kind)
		}
		kindGroups[resrc.Kind] = append(kindGroups[resrc.Kind], resrc)
	}

	allSuccess := true
	for _, group := range kindGroups {
		if len(group) == 0 {
			continue
		}
		logFunc := func(res resource.Resource, upsertResult string, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not apply resource %s/%s: %s\n", res.Kind, res.Name, err)
			} else if upsertResult != "" {
				fmt.Printf("%s/%s: %s\n", res.Kind, res.Name, upsertResult)
			}
		}
		if isGatewayResource(group[0], kinds) {
			ApplyResources(group, gatewayApiClient().Apply, *dryRun, *runInParallel, logFunc)
		} else {
			ApplyResources(group, consoleApiClient().Apply, *dryRun, *runInParallel, logFunc)
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
		PersistentFlags().Bool("dry-run", false, "Test potential changes without the effects being applied")

	runInParallel = applyCmd.
		PersistentFlags().Bool("parallel-run", false, "Run each apply in parallel, useful when applying a large number of resources")

	applyCmd.MarkPersistentFlagRequired("file")
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
