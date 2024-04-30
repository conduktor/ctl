/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/client"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
	"os"
)

var debug *bool
var key *string
var cert *string
var apiClient *client.Client
var schemaClient *schema.Schema = nil

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "conduktor",
	Short: "Command line tools for conduktor",
	Long: `You need to define the CDK_TOKEN and CDK_BASE_URL environment variables to use this tool.
You can also use the CDK_KEY,CDK_CERT instead of --key and --cert flags to use a certificate for tls authentication.
If you have an untrusted certificate you can use the CDK_INSECURE=true or CDK_CACERT variable to disable tls verification`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if *debug {
			apiClient.ActivateDebug()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
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
	apiClient = client.MakeFromEnv()
	openApi, err := apiClient.GetOpenApi()
	if err == nil {
		schemaClient, err = schema.New(openApi)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load server openapi: %s\n", err)
	}
	debug = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more information for debugging")
	key = rootCmd.PersistentFlags().String("key", "", "set pem key for certificate authentication (useful for teleport)")
	cert = rootCmd.PersistentFlags().String("cert", "", "set pem cert for certificate authentication (useful for teleport)")
	initGet()
	initDelete()
	initApply()
}
