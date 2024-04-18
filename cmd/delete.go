package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// applyCmd represents the apply command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resource of a given kind and name",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command does nothing
		cmd.Help()
		os.Exit(1)
	},
}

func initDelete(kinds schema.KindCatalog) {
	rootCmd.AddCommand(deleteCmd)

	for name, kind := range kinds {
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
				err := apiClient().Delete(&kind, parentValue, args[0])
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
