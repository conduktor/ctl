package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/spf13/cobra"
)

var dryRun *bool
var printDiff *bool
var maxParallel *int

func resourceForPath(path string, strict, recursiveFolder bool) ([]resource.Resource, error) {
	directory, err := isDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if directory {
		return resource.FromFolder(path, strict, recursiveFolder)
	} else {
		return resource.FromFile(path, strict)
	}
}

// Globally accessible for testing purposes.
func ApplyResources(resources []resource.Resource,
	applyFunc func(*resource.Resource, bool, bool) (client.Result, error),
	logFunc func(resource.Resource, client.Result, error),
	dryRun bool,
	diffMode bool,
	maxParallel int) []struct {
	Resource     resource.Resource
	UpsertResult client.Result
	Err          error
} {
	results := make([]struct {
		Resource     resource.Resource
		UpsertResult client.Result
		Err          error
	}, len(resources))

	if maxParallel > 1 {
		var wg sync.WaitGroup
		var mu sync.Mutex // for logging in parallel; prevents interleaving of log messages by multiple goroutines trying to write to stdout
		sem := make(chan struct{}, maxParallel)
		for i, resrc := range resources {
			wg.Add(1)
			sem <- struct{}{} // acquire a slot
			go func(i int, res resource.Resource) {
				defer func() {
					wg.Done()
					<-sem // release the slot
				}()
				upsertResult, err := applyFunc(&res, dryRun, diffMode)
				results[i] = struct {
					Resource     resource.Resource
					UpsertResult client.Result
					Err          error
				}{res, upsertResult, err}
				mu.Lock()
				logFunc(res, upsertResult, err)
				mu.Unlock()
			}(i, resrc)
		}
		wg.Wait()
	} else {
		for i, res := range resources {
			var err error
			upsertResult, err := applyFunc(&res, dryRun, diffMode)
			results[i] = struct {
				Resource     resource.Resource
				UpsertResult client.Result
				Err          error
			}{res, upsertResult, err}
			logFunc(res, upsertResult, err)
		}
	}
	return results
}

func runApply(kinds schema.KindCatalog, filePath []string, strict bool, recursiveFolder bool) {
	resources := loadResourceFromFileFlag(filePath, strict, recursiveFolder)
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
	for _, kind := range kindOrder {
		group := kindGroups[kind]
		if len(group) == 0 {
			continue
		}
		logFunc := func(res resource.Resource, upsertResult client.Result, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not apply resource %s/%s: %s\n", res.Kind, res.Name, err)
			} else if upsertResult.UpsertResult != "" {
				fmt.Printf("%s", upsertResult.Diff)
				fmt.Printf("%s/%s: %s\n", res.Kind, res.Name, upsertResult.UpsertResult)
			}
		}
		var results []struct {
			Resource     resource.Resource
			UpsertResult client.Result
			Err          error
		}

		if isGatewayResource(group[0], kinds) {
			results = ApplyResources(group, gatewayAPIClient().Apply, logFunc, *dryRun, *printDiff, *maxParallel)
		} else {
			results = ApplyResources(group, consoleAPIClient().Apply, logFunc, *dryRun, *printDiff, *maxParallel)
		}
		for _, r := range results {
			if r.Err != nil {
				allSuccess = false
				break
			}
		}
	}
	if !allSuccess {
		os.Exit(1)
	}
}

func initApply(kinds schema.KindCatalog, strict bool) {
	// applyCmd represents the apply command
	var recursiveFolder *bool
	var filePath *[]string
	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Upsert a resource on Conduktor",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			runApply(kinds, *filePath, strict, *recursiveFolder)
		},
	}

	rootCmd.AddCommand(applyCmd)

	filePath = applyCmd.
		PersistentFlags().StringArrayP("file", "f", make([]string, 0), FILE_ARGS_DOC)

	dryRun = applyCmd.
		PersistentFlags().Bool("dry-run", false, "Test potential changes without the effects being applied")

	printDiff = applyCmd.
		PersistentFlags().Bool("print-diff", false, "Print the diff between the current resource and the one to be applied")

	recursiveFolder = applyCmd.
		PersistentFlags().BoolP("recursive", "r", false, "Apply all .yaml or .yml files in the specified folder and its subfolders. If not set, only files in the specified folder will be applied.")

	maxParallel = applyCmd.
		PersistentFlags().Int("parallelism", 1, "Run each apply in parallel, useful when applying a large number of resources. Must be less than 100.")

	_ = applyCmd.MarkPersistentFlagRequired("file")

	applyCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if *maxParallel > 100 || *maxParallel < 1 {
			fmt.Fprintf(os.Stderr, "Error: --parallelism must be between 1 and 100 (got %d)\n", *maxParallel)
			os.Exit(1)
		}
	}
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
