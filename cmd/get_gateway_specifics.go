package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

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
