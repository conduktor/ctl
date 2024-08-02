package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

func buildDeleteByVClusterAndNameCmd(kind schema.Kind) *cobra.Command {
	const nameFlag = "name"
	const vClusterFlag = "vcluster"
	name := kind.GetName()
	var nameValue string
	var vClusterValue string
	var deleteCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [name]", name),
		Short:   "Delete resource of kind " + name,
		Args:    cobra.ExactArgs(0),
		Aliases: buildAlias(name),
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

	deleteCmd.Flags().StringVar(&nameValue, nameFlag, "", "name of the "+name)
	deleteCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "vCluster of the "+name)

	deleteCmd.MarkFlagRequired(nameFlag)

	return deleteCmd
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
		Aliases: buildAlias(name),
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
