package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/conduktor/ctl/internal/state"
	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/internal/state/storage"
	"github.com/spf13/cobra"
)

func initApply(rootContext cli.RootContext) {
	// applyCmd represents the apply command
	var recursiveFolder *bool
	var filePath *[]string
	var dryRun *bool
	var printDiff *bool
	var maxParallel *int
	var stateEnabled *bool
	var stateFile *string

	var applyCmd = &cobra.Command{
		Use:          "apply",
		Short:        "Upsert a resource on Conduktor",
		Long:         ``,
		SilenceUsage: true, // do not print usage on run error
		RunE: func(cmd *cobra.Command, args []string) error {
			stateCfg := storage.NewStorageConfig(stateEnabled, stateFile)
			return state.RunWithState(stateCfg, *dryRun, *rootContext.Debug, func(stateRef *model.State) error {

				cmdCtx := cli.ApplyHandlerContext{
					FilePaths:       *filePath,
					RecursiveFolder: *recursiveFolder,
					DryRun:          *dryRun,
					PrintDiff:       *printDiff,
					MaxParallel:     *maxParallel,
					StateEnabled:    stateCfg.Enabled,
					StateRef:        stateRef,
				}

				return runApply(rootContext, cmdCtx)
			})
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

	stateEnabled = applyCmd.
		PersistentFlags().Bool("enable-state", false, "Enable state management for the resource.")

	stateFile = applyCmd.
		PersistentFlags().String("state-file", "", "Path to the state file to use for state management. By default, use $HOME/.conduktor/ctl/state.yaml")

	_ = applyCmd.MarkPersistentFlagRequired("file")

	applyCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *maxParallel > 100 || *maxParallel < 1 {
			return fmt.Errorf("argument --parallelism must be between 1 and 100 (got %d)\n", *maxParallel)
		}
		return nil
	}
}

func runApply(rootContext cli.RootContext, cmdCtx cli.ApplyHandlerContext) error {
	applyHandler := cli.NewApplyHandler(rootContext)

	results, err := applyHandler.Handle(cmdCtx)
	if err != nil {
		return fmt.Errorf("failed to run apply: %s\n", err)
	}

	allSuccess := true
	for _, result := range results {
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "Could not apply resource %s/%s: %s\n", result.Resource.Kind, result.Resource.Name, result.Err)
			allSuccess = false
		} else if result.UpsertResult.UpsertResult != "" {
			fmt.Printf("%s", result.UpsertResult.Diff)
			fmt.Printf("%s/%s: %s\n", result.Resource.Kind, result.Resource.Name, result.UpsertResult.UpsertResult)
		}
	}

	if !allSuccess {
		return fmt.Errorf("one or more resources could not be applied")
	}
	return nil
}
