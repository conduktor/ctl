package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

func initDelete(kinds schema.KindCatalog, strict bool) {
	var filePath *[]string
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete resource of a given kind and name",
		Long:  ``,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// Root command does nothing
			resources := loadResourceFromFileFlag(*filePath, strict)
			schema.SortResourcesForDelete(kinds, resources, *debug)
			allSuccess := true
			for _, resource := range resources {
				var err error
				kind := kinds[resource.Kind]
				if isGatewayKind(kind) {
					if isResourceIdentifiedByName(resource) {
						err = gatewayApiClient().DeleteResourceByName(&resource)
					} else if isResourceIdentifiedByNameAndVCluster(resource) {
						err = gatewayApiClient().DeleteResourceByNameAndVCluster(&resource)
					} else if isResourceInterceptor(resource) {
						err = gatewayApiClient().DeleteResourceInterceptors(&resource)
					}
				} else {
					err = consoleApiClient().DeleteResource(&resource)
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

	filePath = deleteCmd.Flags().StringArrayP("file", "f", make([]string, 0, 0), "the files to apply")

	deleteCmd.MarkFlagRequired("file")

	for name, kind := range kinds {
		if isKindIdentifiedByNameAndVCluster(kind) {
			byVClusterAndNameDeleteCmd := buildDeleteByVClusterAndNameCmd(kind)
			deleteCmd.AddCommand(byVClusterAndNameDeleteCmd)
		} else if isKindInterceptor(kind) {
			interceptorsDeleteCmd := buildDeleteInterceptorsCmd(kind)
			deleteCmd.AddCommand(interceptorsDeleteCmd)
		} else {
			flags := kind.GetParentFlag()
			parentFlagValue := make([]*string, len(flags))
			kindCmd := &cobra.Command{
				Use:     fmt.Sprintf("%s [name]", name),
				Short:   "Delete resource of kind " + name,
				Args:    cobra.MatchAll(cobra.ExactArgs(1)),
				Aliases: buildAlias(name),
				Run: func(cmd *cobra.Command, args []string) {
					parentValue := make([]string, len(parentFlagValue))
					for i, v := range parentFlagValue {
						parentValue[i] = *v
					}
					var err error
					if isGatewayKind(kind) {
						err = gatewayApiClient().Delete(&kind, parentValue, args[0])
					} else {
						err = consoleApiClient().Delete(&kind, parentValue, args[0])
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s\n", err)
						os.Exit(1)
					}
				},
			}
			for i, flag := range kind.GetParentFlag() {
				parentFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
				kindCmd.MarkFlagRequired(flag)
			}
			deleteCmd.AddCommand(kindCmd)
		}
	}
}
