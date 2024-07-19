package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resource of a given kind",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		cmd.Help()
		os.Exit(1)
	},
}

func initGet(kinds schema.KindCatalog) {
	rootCmd.AddCommand(getCmd)

	for name, kind := range kinds {
		if name == "AliasTopics" || name == "ConcentrationRules" || name == "ServiceAccounts" {
			aliasTopicGetCmd := buildListFilteredByVClusterOrNameCmd(kind)
			getCmd.AddCommand(aliasTopicGetCmd)
		} else if name == "Interceptors" {
			interceptorsGetCmd := buildListFilteredIntercpetorsCmd(kind)
			getCmd.AddCommand(interceptorsGetCmd)
		} else {
			flags := kind.GetFlag()
			parentFlagValue := make([]*string, len(flags))
			kindCmd := &cobra.Command{
				Use:     fmt.Sprintf("%s [name]", name),
				Short:   "Get resource of kind " + name,
				Args:    cobra.MatchAll(cobra.MaximumNArgs(1)),
				Long:    `If name not provided it will list all resource`,
				Aliases: []string{strings.ToLower(name), strings.ToLower(name) + "s", name + "s"},
				Run: func(cmd *cobra.Command, args []string) {
					parentValue := make([]string, len(parentFlagValue))
					for i, v := range parentFlagValue {
						parentValue[i] = *v
					}
					var err error
					if len(args) == 0 {
						var result []resource.Resource
						if strings.Contains(kind.GetLatestKindVersion().ListPath, "gateway") {
							result, err = gatewayApiClient().Get(&kind, parentValue)
						} else {
							result, err = apiClient().Get(&kind, parentValue)
						}
						for _, r := range result {
							r.PrintPreservingOriginalFieldOrder()
							fmt.Println("---")
						}
					} else if len(args) == 1 {
						var result resource.Resource
						if strings.Contains(kind.GetLatestKindVersion().ListPath, "gateway") {
							result, err = gatewayApiClient().Describe(&kind, parentValue, args[0])
						} else {
							result, err = apiClient().Describe(&kind, parentValue, args[0])
						}
						result.PrintPreservingOriginalFieldOrder()
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
			getCmd.AddCommand(kindCmd)
		}
	}
}

func buildListFilteredByVClusterOrNameCmd(kind schema.Kind) *cobra.Command {
	const nameFlag = "name"
	const vClusterFlag = "vcluster"
	const showDefaultsFlag = "showDefaults"
	var nameValue string
	var vClusterValue string
	var showDefaultsValue string
	name := kind.GetName()
	var aliasTopicGetCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [name]", name),
		Short:   "Get resource of kind " + kind.GetName(),
		Args:    cobra.ExactArgs(0),
		Aliases: []string{strings.ToLower(name), strings.ToLower(name) + "s", name + "s"},
		Run: func(cmd *cobra.Command, args []string) {
			var result []resource.Resource
			var err error
			queryParams := make(map[string]string)
			if nameValue != "" {
				queryParams[nameFlag] = nameValue
			}
			if vClusterValue != "" {
				queryParams[vClusterFlag] = vClusterValue
			}
			if showDefaultsValue != "" {
				queryParams[showDefaultsFlag] = showDefaultsValue
			}

			result, err = gatewayApiClient().ListKindWithFilters(&kind, queryParams)
			for _, r := range result {
				r.PrintPreservingOriginalFieldOrder()
				fmt.Println("---")
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	aliasTopicGetCmd.Flags().StringVar(&nameValue, nameFlag, "", "filter the "+name+" result list by name")
	aliasTopicGetCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "filter the "+name+" result list by vcluster")
	aliasTopicGetCmd.Flags().StringVar(&showDefaultsValue, "showDefaults", "", "Toggle show defaults values (true|false, default false)")

	return aliasTopicGetCmd
}

func buildListFilteredIntercpetorsCmd(kind schema.Kind) *cobra.Command {
	const nameFlag = "name"
	const globalFlag = "global"
	const vClusterFlag = "vcluster"
	const groupFlag = "group"
	const usernameFlag = "username"
	var nameValue string
	var vClusterValue string
	var groupValue string
	var usernameValue string
	var globalValue bool
	name := kind.GetName()
	var aliasTopicGetCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [name]", name),
		Short:   "Get resource of kind " + name,
		Args:    cobra.ExactArgs(0),
		Aliases: []string{strings.ToLower(name), strings.ToLower(name) + "s", name + "s"},
		Run: func(cmd *cobra.Command, args []string) {
			var result []resource.Resource
			var err error
			if g, _ := cmd.Flags().GetBool("global"); g {
				globalValue = true
			} else {
				globalValue = false
			}
			result, err = gatewayApiClient().ListInterceptorsFilters(&kind, nameValue, globalValue, vClusterValue, groupValue, usernameValue)
			for _, r := range result {
				r.PrintPreservingOriginalFieldOrder()
				fmt.Println("---")
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	aliasTopicGetCmd.Flags().StringVar(&nameValue, nameFlag, "", "filter the "+name+" result list by name")
	aliasTopicGetCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "filter the "+name+" result list by vcluster")
	aliasTopicGetCmd.Flags().StringVar(&groupValue, groupFlag, "", "filter the "+name+" result list by group")
	aliasTopicGetCmd.Flags().StringVar(&usernameValue, usernameFlag, "", "filter the "+name+" result list by username")
	aliasTopicGetCmd.Flags().Bool(globalFlag, false, "Keep only global interceptors")

	return aliasTopicGetCmd
}
