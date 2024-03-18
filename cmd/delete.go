package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/client"
	"github.com/spf13/cobra"
	"os"
)

// applyCmd represents the apply command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete resource of a given kind and name",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := client.MakeFromEnv(*debug, *key, *cert)
		err := client.Delete(args[0], args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
