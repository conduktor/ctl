package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/pkg/schema"
	"github.com/spf13/cobra"
)

func buildDeleteByVClusterAndNameCmd(kind schema.Kind) *cobra.Command {
	const vClusterFlag = "vcluster"
	name := kind.GetName()
	var vClusterValue string
	var deleteCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [name]", name),
		Short:   "Delete resource of kind " + name,
		Args:    cobra.ExactArgs(1),
		Aliases: buildAlias(name),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			bodyParams := make(map[string]string)
			nameValue := args[0]
			if nameValue != "" {
				bodyParams["name"] = nameValue
			}
			bodyParams["vCluster"] = vClusterValue

			err = gatewayAPIClient().DeleteKindByNameAndVCluster(&kind, bodyParams)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	deleteCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "passthrough", "vCluster of the "+name)

	return deleteCmd
}

func buildDeleteInterceptorsCmd(kind schema.Kind) *cobra.Command {
	const vClusterFlag = "vcluster"
	const groupFlag = "group"
	const usernameFlag = "username"
	var vClusterValue string
	var groupValue string
	var usernameValue string
	name := kind.GetName()
	var interceptorDeleteCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [name]", name),
		Short:   "Delete resource of kind " + name,
		Args:    cobra.ExactArgs(1),
		Aliases: buildAlias(name),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			bodyParams := make(map[string]string)
			nameValue := args[0]
			if vClusterValue != "" {
				bodyParams["vCluster"] = vClusterValue
			}
			if groupValue != "" {
				bodyParams["group"] = groupValue
			}
			if usernameValue != "" {
				bodyParams["username"] = usernameValue
			}

			err = gatewayAPIClient().DeleteInterceptor(&kind, nameValue, bodyParams)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	interceptorDeleteCmd.Flags().StringVar(&vClusterValue, vClusterFlag, "", "vCluster of the "+name)
	interceptorDeleteCmd.Flags().StringVar(&groupValue, groupFlag, "", "Group of the "+name)
	interceptorDeleteCmd.Flags().StringVar(&usernameValue, usernameFlag, "", "Username of the "+name)

	return interceptorDeleteCmd
}
