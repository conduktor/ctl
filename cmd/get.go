package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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

var getCmdWhenNoSchema = &cobra.Command{
	Use:   "get kind [name]",
	Short: "Get resource of a given kind",
	Long: `If name not provided it will list all resource. For example:
conduktor get application
will list all applications. Whereas:
conduktor get application myapp
will describe the application myapp`,
	Args: cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if len(args) == 1 {
			err = apiClient.Get(args[0])
		} else if len(args) == 2 {
			err = apiClient.Describe(args[0], args[1])
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	},
}

func initGet() {
	if schemaClient == nil {
		rootCmd.AddCommand(getCmdWhenNoSchema)
		return
	}
	tags, err := schemaClient.GetKind()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load kind from openapi: %s\n", err)
		rootCmd.AddCommand(getCmdWhenNoSchema)
		return
	}
	rootCmd.AddCommand(getCmd)

	for _, tag := range tags {
		tagCmd := &cobra.Command{
			Use:   fmt.Sprintf("%s [name]", tag),
			Short: "Get resource of kind " + tag,
			Args:  cobra.MatchAll(cobra.MaximumNArgs(1)),
			Long:  `If name not provided it will list all resource`,
			Run: func(cmd *cobra.Command, args []string) {
				var err error
				if len(args) == 0 {
					err = apiClient.Get(tag)
				} else if len(args) == 1 {
					err = apiClient.Describe(tag, args[0])
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			},
		}
		getCmd.AddCommand(tagCmd)
	}
}
