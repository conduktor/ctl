package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/spf13/cobra"
)

var dryRun *bool
var printDiff *bool
var maxParallel *int
var serverSide *bool
var strategyFlag *string
var noProgress *bool
var assumeYes *bool

func initApply(rootContext cli.RootContext) {
	// applyCmd represents the apply command
	var recursiveFolder *bool
	var filePath *[]string
	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Upsert a resource on Conduktor",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if *serverSide {
				runServerApply(rootContext, *filePath, *recursiveFolder)
			} else {
				runApply(rootContext, *filePath, *recursiveFolder)
			}
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
		PersistentFlags().Int("parallelism", 1, "Run each apply in parallel, useful when applying a large number of resources. Must be less than 100. Only used without --server-side.")

	serverSide = applyCmd.
		PersistentFlags().Bool("server-side", false, "Use server-side processing for better performance with many resources")

	strategyFlag = applyCmd.
		PersistentFlags().String("strategy", "fail-fast", "Apply strategy: fail-fast or continue-on-error. Only used with --server-side.")

	noProgress = applyCmd.
		PersistentFlags().Bool("no-progress", false, "Do not display live progress (useful for CI logs). Only used with --server-side.")

	assumeYes = applyCmd.
		PersistentFlags().Bool("yes", false, "Skip confirmation when applying a large number of resources. Only used with --server-side.")

	_ = applyCmd.MarkPersistentFlagRequired("file")

	applyCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if !*serverSide && (*maxParallel > 100 || *maxParallel < 1) {
			fmt.Fprintf(os.Stderr, "Error: --parallelism must be between 1 and 100 (got %d)\n", *maxParallel)
			os.Exit(1)
		}
		if *serverSide && *strategyFlag != "fail-fast" && *strategyFlag != "continue-on-error" {
			fmt.Fprintf(os.Stderr, "Error: --strategy must be one of [fail-fast, continue-on-error]\n")
			os.Exit(1)
		}
	}
}

func runApply(rootContext cli.RootContext, filePath []string, recursiveFolder bool) {

	applyHandler := cli.NewApplyHandler(rootContext)

	cmdCtx := cli.ApplyHandlerContext{
		FilePaths:       filePath,
		DryRun:          *dryRun,
		PrintDiff:       *printDiff,
		RecursiveFolder: recursiveFolder,
		MaxParallel:     *maxParallel,
	}

	results, err := applyHandler.Handle(cmdCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during apply: %s\n", err)
		os.Exit(1)
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
		os.Exit(1)
	}
}

func runServerApply(rootContext cli.RootContext, filePath []string, recursiveFolder bool) {
	handler := cli.NewServerApplyHandler(rootContext)

	cmdCtx := cli.ServerApplyHandlerContext{
		FilePaths:       filePath,
		DryRun:          *dryRun,
		PrintDiff:       *printDiff,
		RecursiveFolder: recursiveFolder,
		Strategy:        *strategyFlag,
		NoProgress:      *noProgress,
		AssumeYes:       *assumeYes,
	}

	results, err := handler.Handle(cmdCtx)
	if err != nil {
		if err == cli.ErrCancelled {
			fmt.Fprintln(os.Stderr, "Operation cancelled")
			os.Exit(130) // Standard exit code for SIGINT
		}
		fmt.Fprintf(os.Stderr, "Error during server apply: %s\n", err)
		os.Exit(1)
	}

	successes := 0
	failures := 0
	for _, result := range results {
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "Could not apply resource %s/%s: %s\n", result.Resource.Kind, result.Resource.Name, result.Err)
			failures++
		} else if result.UpsertResult.UpsertResult != "" {
			fmt.Printf("%s", result.UpsertResult.Diff)
			fmt.Printf("%s/%s: %s\n", result.Resource.Kind, result.Resource.Name, result.UpsertResult.UpsertResult)
			successes++
		}
	}

	if failures > 0 {
		if successes > 0 {
			os.Exit(2) // partial success
		}
		os.Exit(1)
	}
}
