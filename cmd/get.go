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
		if name == "AliasTopics" || name == "ConcentrationRules" {
			aliasTopicGetCmd := buildListFilteredByVClusterOrNameCmd(name, kind)
			getCmd.AddCommand(aliasTopicGetCmd)
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
			getCmd.AddCommand(kindCmd)
		}
	}
}

func buildListFilteredByVClusterOrNameCmd(name string, kind schema.Kind) *cobra.Command {
	const nameFlag = "name"
	const vClusterFlag = "vcluster"
	const showDefaultsFlag = "showDefaults"
	var nameValue string
	var vClusterValue string
	var showDefaultsValue string
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

	aliasTopicGetCmd.Flags().StringVar(&nameValue, nameFlag, "", "filter the "+kind.GetName()+" result list by name")
	aliasTopicGetCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "filter the "+kind.GetName()+" result list by vcluster")
	aliasTopicGetCmd.Flags().StringVar(&showDefaultsValue, "showDefaults", "", "Toggle show defaults values (true|false, default false)")

	return aliasTopicGetCmd
}
