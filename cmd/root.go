/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var debug *bool
var key *string
var cert *string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "conduktor",
	Short: "command line tools for conduktor",
	Long: `You need to define the CDK_TOKEN and CDK_BASE_URL environment variables to use this tool.
You can also use the CDK_KEY,CDK_CERT instead of --key and --cert flags to use a certificate for tls authentication.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	debug = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show more information for debugging")
	key = rootCmd.PersistentFlags().String("key", "", "Set pem key for certificate authentication (useful for teleport)")
	cert = rootCmd.PersistentFlags().String("cert", "", "Set pem cert for certificate authentication (useful for teleport)")
}
