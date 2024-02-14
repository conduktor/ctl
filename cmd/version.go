package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "unknown"
var hash = "unknown"

// versionCmd represents the apply command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display the version of conduktor",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nHash: %s\n", version, hash)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
