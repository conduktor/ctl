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
var maxParallel *int

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
	logFunc func(resource.Resource, string, error),
	dryRun bool,
	maxParallel int) []struct {
	Resource     resource.Resource
	UpsertResult string
	Err          error
} {
	results := make([]struct {
		Resource     resource.Resource
		UpsertResult string
		Err          error
	}, len(resources))

	if maxParallel > 1 {
		var wg sync.WaitGroup
		var mu sync.Mutex // for logging in parallel; prevents interleaving of log messages by multiple goroutines trying to write to stdout
		sem := make(chan struct{}, maxParallel)
		for i, resrc := range resources {
			wg.Add(1)
			sem <- struct{}{} // acquire a slot
			go func(i int, resrc resource.Resource) {
				defer func() {
					wg.Done()
					<-sem // release the slot
				}()
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
	for _, kind := range kindOrder {
		for _, res := range kindGroups[kind] {
			fmt.Printf("- Kind: %s, Name: %s\n", res.Kind, res.Name)
		}
	}

	allSuccess := true
	for _, kind := range kindOrder {
		group := kindGroups[kind]
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
		var results []struct {
			Resource     resource.Resource
			UpsertResult string
			Err          error
		}
		if isGatewayResource(group[0], kinds) {
			results = ApplyResources(group, gatewayApiClient().Apply, logFunc, *dryRun, *maxParallel)
		} else {
			results = ApplyResources(group, consoleApiClient().Apply, logFunc, *dryRun, *maxParallel)
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

	maxParallel = applyCmd.
		PersistentFlags().Int("parallelism", 1, "Run each apply in parallel, useful when applying a large number of resources. Must be less than 100.")

	applyCmd.MarkPersistentFlagRequired("file")

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
