package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/client"
	"github.com/spf13/cobra"
	"os"
)

// applyCmd represents the apply command
var getCmd = &cobra.Command{
	Use:   "get kind [name]",
	Short: "get resource of a given kind",
	Long: `If name not provided it will list all resource. For example:
conduktor get application
will list all applications. Whereas:
conduktor get application myapp
will describe the application myapp`,
	Args: cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		client := client.MakeFromEnv(*debug, *key, *cert)
		var err error
		if len(args) == 1 {
			err = client.Get(args[0])
		} else if len(args) == 2 {
			err = client.Describe(args[0], args[1])
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
