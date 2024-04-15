package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/utils"
	"github.com/spf13/cobra"
)

// versionCmd represents the apply command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of conduktor",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nHash: %s\n", utils.GetConduktorVersion(), utils.GetConduktorHash())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
