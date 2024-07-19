package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

func initDelete(kinds schema.KindCatalog) {
	var filePath *[]string
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete resource of a given kind and name",
		Long:  ``,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// Root command does nothing
			resources := loadResourceFromFileFlag(*filePath)
			schema.SortResourcesForDelete(kinds, resources, *debug)
			allSuccess := true
			for _, resource := range resources {
				err := apiClient().DeleteResource(&resource)
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
		if name == "AliasTopics" || name == "ConcentrationRules" || name == "ServiceAccounts" {
			byVClusterAndNameDeleteCmd := buildDeleteByVClusterAndNameCmd(kind)
			deleteCmd.AddCommand(byVClusterAndNameDeleteCmd)
		} else if name == "Interceptors" {
			interceptorsDeleteCmd := buildDeleteInterceptorsCmd(kind)
			deleteCmd.AddCommand(interceptorsDeleteCmd)
		} else {
			flags := kind.GetFlag()
			parentFlagValue := make([]*string, len(flags))
			kindCmd := &cobra.Command{
				Use:     fmt.Sprintf("%s [name]", name),
				Short:   "Delete resource of kind " + name,
				Args:    cobra.MatchAll(cobra.ExactArgs(1)),
				Aliases: []string{strings.ToLower(name), strings.ToLower(name) + "s", name + "s"},
				Run: func(cmd *cobra.Command, args []string) {
					parentValue := make([]string, len(parentFlagValue))
					for i, v := range parentFlagValue {
						parentValue[i] = *v
					}
					var err error
					if strings.Contains(kind.GetLatestKindVersion().ListPath, "gateway") {
						err = gatewayApiClient().Delete(&kind, parentValue, args[0])
					} else {
						err = apiClient().Delete(&kind, parentValue, args[0])
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s\n", err)
						os.Exit(1)
					}
				},
			}
			for i, flag := range kind.GetFlag() {
				parentFlagValue[i] = kindCmd.Flags().String(flag, "", "Parent "+flag)
				kindCmd.MarkFlagRequired(flag)
			}
			deleteCmd.AddCommand(kindCmd)
		}
	}
}

func buildDeleteByVClusterAndNameCmd(kind schema.Kind) *cobra.Command {
	const nameFlag = "name"
	const vClusterFlag = "vcluster"
	name := kind.GetName()
	var nameValue string
	var vClusterValue string
	var aliasTopicDeleteCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [name]", name),
		Short:   "Delete resource of kind " + name,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{strings.ToLower(name), strings.ToLower(name) + "s", name + "s"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			queryParams := make(map[string]string)
			if nameValue != "" {
				queryParams[nameFlag] = nameValue
			}
			if vClusterValue != "" {
				queryParams[vClusterFlag] = vClusterValue
			} else {
				queryParams[vClusterFlag] = "passthrough"
			}

			err = gatewayApiClient().DeleteKindByNameAndVCluster(&kind, queryParams)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	aliasTopicDeleteCmd.Flags().StringVar(&nameValue, nameFlag, "", "name of the "+name)
	aliasTopicDeleteCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "vCluster of the "+name)

	aliasTopicDeleteCmd.MarkFlagRequired(nameFlag)

	return aliasTopicDeleteCmd
}

func buildDeleteInterceptorsCmd(kind schema.Kind) *cobra.Command {
	const nameFlag = "name"
	const vClusterFlag = "vcluster"
	const groupFlag = "group"
	const usernameFlag = "username"
	var nameValue string
	var vClusterValue string
	var groupValue string
	var usernameValue string
	name := kind.GetName()
	var interceptorDeleteCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [name]", name),
		Short:   "Delete resource of kind " + name,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{strings.ToLower(name), strings.ToLower(name) + "s", name + "s"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			queryParams := make(map[string]string)
			if nameValue != "" {
				queryParams[nameFlag] = nameValue
			}
			if vClusterValue != "" {
				queryParams[vClusterFlag] = vClusterValue
			}
			if groupValue != "" {
				queryParams[groupFlag] = groupValue
			}
			if usernameValue != "" {
				queryParams[usernameFlag] = usernameValue
			}

			err = gatewayApiClient().DeleteInterceptor(&kind, nameValue, queryParams)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	interceptorDeleteCmd.Flags().StringVar(&nameValue, nameFlag, "", "name of the "+name)
	interceptorDeleteCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "vCluster of the "+name)
	interceptorDeleteCmd.Flags().StringVar(&groupValue, groupFlag, "", "group of the "+name)
	interceptorDeleteCmd.Flags().StringVar(&usernameValue, usernameFlag, "", "username of the "+name)

	interceptorDeleteCmd.MarkFlagRequired(nameFlag)

	return interceptorDeleteCmd
}
