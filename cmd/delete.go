package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/pkg/schema"
	"github.com/spf13/cobra"
)

func initDelete(kinds schema.KindCatalog, strict bool) {
	var recursiveFolder *bool
	var filePath *[]string
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete resource of a given kind and name",
		Long:  ``,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// Root command does nothing
			resources := loadResourceFromFileFlag(*filePath, strict, *recursiveFolder)
			schema.SortResourcesForDelete(kinds, resources, *debug)
			allSuccess := true
			for _, resource := range resources {
				var err error
				kind := kinds[resource.Kind]
				if isGatewayKind(kind) {
					if isResourceIdentifiedByName(resource) {
						err = gatewayAPIClient().DeleteResourceByName(&resource)
					} else if isResourceIdentifiedByNameAndVCluster(resource) {
						err = gatewayAPIClient().DeleteResourceByNameAndVCluster(&resource)
					} else if isResourceInterceptor(resource) {
						err = gatewayAPIClient().DeleteResourceInterceptors(&resource)
					}
				} else {
					err = consoleAPIClient().DeleteResource(&resource)
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not delete resource %s/%s: %s\n", resource.Kind, resource.Name, err)
					allSuccess = false
				}
			}
			if !allSuccess {
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(deleteCmd)

	filePath = deleteCmd.Flags().StringArrayP("file", "f", make([]string, 0), FILE_ARGS_DOC)

	recursiveFolder = deleteCmd.
		Flags().BoolP("recursive", "r", false, "Delete all .yaml or .yml files in the specified folder and its subfolders. If not set, only files in the specified folder will be applied.")

	_ = deleteCmd.MarkFlagRequired("file")

	for name, kind := range kinds {
		if isKindIdentifiedByNameAndVCluster(kind) {
			byVClusterAndNameDeleteCmd := buildDeleteByVClusterAndNameCmd(kind)
			deleteCmd.AddCommand(byVClusterAndNameDeleteCmd)
		} else if isKindInterceptor(kind) {
			interceptorsDeleteCmd := buildDeleteInterceptorsCmd(kind)
			deleteCmd.AddCommand(interceptorsDeleteCmd)
		} else {
			flags := kind.GetParentFlag()
			parentQueryFlags := kind.GetParentQueryFlag()
			parentFlagValue := make([]*string, len(flags))
			parentQueryFlagValue := make([]*string, len(parentQueryFlags))
			kindCmd := &cobra.Command{
				Use:     fmt.Sprintf("%s [name]", name),
				Short:   "Delete resource of kind " + name,
				Args:    cobra.MatchAll(cobra.ExactArgs(1)),
				Aliases: buildAlias(name),
				Run: func(cmd *cobra.Command, args []string) {
					parentValue := make([]string, len(parentFlagValue))
					parentQueryValue := make([]string, len(parentQueryFlagValue))
					for i, v := range parentFlagValue {
						parentValue[i] = *v
					}
					for i, v := range parentQueryFlagValue {
						parentQueryValue[i] = *v
					}

					var err error
					if isGatewayKind(kind) {
						err = gatewayAPIClient().Delete(&kind, parentValue, parentQueryValue, args[0])
					} else {
						err = consoleAPIClient().Delete(&kind, parentValue, parentQueryValue, args[0])
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s\n", err)
						os.Exit(1)
					}
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
