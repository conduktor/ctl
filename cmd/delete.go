package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/conduktor/ctl/internal/state"
	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/internal/state/storage"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/spf13/cobra"
)

func initDelete(rootContext cli.RootContext) {
	var recursiveFolder *bool
	var filePath *[]string
	var dryRun *bool
	var stateEnabled *bool
	var stateFile *string

	var deleteCmd = &cobra.Command{
		Use:          "delete",
		Short:        "Delete resource of a given kind and name",
		Long:         ``,
		Args:         cobra.NoArgs,
		SilenceUsage: true, // do not print usage on run error
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteFromFiles(rootContext, *filePath, *recursiveFolder, dryRun, stateEnabled, stateFile)
		},
	}

	rootCmd.AddCommand(deleteCmd)

	filePath = deleteCmd.Flags().StringArrayP("file", "f", make([]string, 0), FILE_ARGS_DOC)

	recursiveFolder = deleteCmd.
		Flags().BoolP("recursive", "r", false, "Delete all .yaml or .yml files in the specified folder and its subfolders. If not set, only files in the specified folder will be applied.")

	dryRun = deleteCmd.
		PersistentFlags().Bool("dry-run", false, "Test potential changes without the effects being deleted")

	stateEnabled = deleteCmd.
		PersistentFlags().Bool("enable-state", false, "Enable state management for the resource.")

	stateFile = deleteCmd.
		PersistentFlags().String("state-file", "", "Path to the state file to use for state management. By default, use $HOME/.conduktor/ctl/state.yaml")

	_ = deleteCmd.MarkFlagRequired("file")

	for name, kind := range rootContext.Catalog.Kind {
		if cli.IsKindIdentifiedByNameAndVCluster(kind) {
			byVClusterAndNameDeleteCmd := buildDeleteByVClusterAndNameCmd(rootContext, kind)
			deleteCmd.AddCommand(byVClusterAndNameDeleteCmd)
		} else if cli.IsKindInterceptor(kind) {
			interceptorsDeleteCmd := buildDeleteInterceptorsCmd(rootContext, kind)
			deleteCmd.AddCommand(interceptorsDeleteCmd)
		} else {
			flags := kind.GetParentFlag()
			parentQueryFlags := kind.GetParentQueryFlag()
			parentFlagValue := make([]*string, len(flags))
			parentQueryFlagValue := make([]*string, len(parentQueryFlags))
			kindCmd := &cobra.Command{
				Use:          fmt.Sprintf("%s [name]", name),
				Short:        "Delete resource of kind " + name,
				Args:         cobra.MatchAll(cobra.ExactArgs(1)),
				Aliases:      buildAlias(name),
				SilenceUsage: true, // do not print usage on run error
				RunE: func(cmd *cobra.Command, args []string) error {
					return runDeleteKind(rootContext, kind, args, parentFlagValue, parentQueryFlagValue)
				},
			}
			for i, flag := range kind.GetParentFlag() {
				parentFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
				_ = kindCmd.MarkFlagRequired(flag)
			}
			for i, flag := range parentQueryFlags {
				parentQueryFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
			}
			deleteCmd.AddCommand(kindCmd)
		}
	}
}

func runDeleteFromFiles(rootContext cli.RootContext, filePaths []string, recursiveFolder bool, dryRun *bool, stateEnabled *bool, stateFile *string) error {

	stateCfg := storage.NewStorageConfig(stateEnabled, stateFile)
	return state.RunWithState(stateCfg, *dryRun, *rootContext.Debug, func(stateRef *model.State) error {
		deleteHandler := cli.NewDeleteHandler(rootContext)

		cmdCtx := cli.DeleteFileHandlerContext{
			FilePaths:       filePaths,
			RecursiveFolder: recursiveFolder,
			IgnoreMissing:   false, // fail even if resource is missing (keep current behavior)
			DryRun:          *dryRun,
			StateEnabled:    *stateEnabled,
			StateRef:        stateRef,
		}

		results, err := deleteHandler.HandleFromFiles(cmdCtx)
		if err != nil {
			return fmt.Errorf("fail to delete: %s\n", err)
		}

		allSuccess := true
		for _, result := range results {
			if result.Err != nil {
				fmt.Fprintf(os.Stderr, "Could not delete resource %s/%s: %s\n", result.Resource.Kind, result.Resource.Name, result.Err)
				allSuccess = false
			}
		}

		if !allSuccess {
			return fmt.Errorf("some resources could not be deleted")
		}
		return nil
	})
}

func buildDeleteByVClusterAndNameCmd(rootContext cli.RootContext, kind schema.Kind) *cobra.Command {
	const vClusterFlag = "vcluster"
	name := kind.GetName()
	var vClusterValue string
	var deleteCmd = &cobra.Command{
		Use:          fmt.Sprintf("%s [name]", name),
		Short:        "Delete resource of kind " + name,
		Args:         cobra.ExactArgs(1),
		Aliases:      buildAlias(name),
		SilenceUsage: true, // do not print usage on run error
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteByVClusterAndName(rootContext, kind, args[0], vClusterValue)
		},
	}

	deleteCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "passthrough", "vCluster of the "+name)

	return deleteCmd
}

func buildDeleteInterceptorsCmd(rootContext cli.RootContext, kind schema.Kind) *cobra.Command {
	const vClusterFlag = "vcluster"
	const groupFlag = "group"
	const usernameFlag = "username"
	var vClusterValue string
	var groupValue string
	var usernameValue string
	name := kind.GetName()
	var interceptorDeleteCmd = &cobra.Command{
		Use:          fmt.Sprintf("%s [name]", name),
		Short:        "Delete resource of kind " + name,
		Args:         cobra.ExactArgs(1),
		Aliases:      buildAlias(name),
		SilenceUsage: true, // do not print usage on run error
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteInterceptor(rootContext, kind, args[0], vClusterValue, groupValue, usernameValue)
		},
	}

	interceptorDeleteCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "vCluster of the "+name)
	interceptorDeleteCmd.Flags().StringVar(&groupValue, groupFlag, "", "Group of the "+name)
	interceptorDeleteCmd.Flags().StringVar(&usernameValue, usernameFlag, "", "Username of the "+name)

	return interceptorDeleteCmd
}

func runDeleteByVClusterAndName(rootContext cli.RootContext, kind schema.Kind, name string, vCluster string) error {
	deleteHandler := cli.NewDeleteHandler(rootContext)

	cmdCtx := cli.DeleteByVClusterAndNameHandlerContext{
		Name:          name,
		VCluster:      vCluster,
		IgnoreMissing: false, // fail even if resource is missing (keep current behavior)
	}

	err := deleteHandler.HandleByVClusterAndName(kind, cmdCtx)
	if err != nil {
		return fmt.Errorf("%s\n", err)
	}
	return nil
}

func runDeleteInterceptor(rootContext cli.RootContext, kind schema.Kind, name string, vCluster string, group string, username string) error {
	deleteHandler := cli.NewDeleteHandler(rootContext)

	cmdCtx := cli.DeleteInterceptorHandlerContext{
		Name:          name,
		VCluster:      vCluster,
		Group:         group,
		Username:      username,
		IgnoreMissing: false, // fail even if resource is missing (keep current behavior)
	}

	err := deleteHandler.HandleInterceptor(kind, cmdCtx)
	if err != nil {
		return fmt.Errorf("%s\n", err)
	}
	return nil
}

func runDeleteKind(
	rootContext cli.RootContext,
	kind schema.Kind,
	args []string,
	parentFlagValue []*string,
	parentQueryFlagValue []*string) error {

	deleteHandler := cli.NewDeleteHandler(rootContext)

	cmdCtx := cli.DeleteKindHandlerContext{
		Args:                 args,
		ParentFlagValue:      parentFlagValue,
		ParentQueryFlagValue: parentQueryFlagValue,
		IgnoreMissing:        false, // fail even if resource is missing (keep current behavior)
	}

	err := deleteHandler.HandleKind(kind, cmdCtx)
	if err != nil {
		return fmt.Errorf("%s\n", err)
	}
	return nil
}
