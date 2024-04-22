package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
	"os"
	"strings"
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
					err = apiClient().Get(&kind, parentValue)
				} else if len(args) == 1 {
					err = apiClient().Describe(&kind, parentValue, args[0])
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
