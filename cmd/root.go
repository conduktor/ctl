package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/client"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
	"os"
)

var debug *bool
var apiClient_ *client.Client
var apiClientError error

func apiClient() *client.Client {
	if apiClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create client: %s", apiClientError)
		os.Exit(1)
	}
	return apiClient_
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "conduktor",
	Short: "Command line tools for conduktor",
	Long: `Make sure you've set the environment variables CDK_TOKEN (generated from Console) and CDK_BASE_URL.
Additionally, you can configure client TLS authentication by providing your certificate paths in CDK_KEY and CDK_CERT.
For the server TLS authentication, you can either not check the certificate by setting CDK_INSECURE=true, or provide your own certificate authority using CDK_CACERT.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if *debug {
			apiClient().ActivateDebug()
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
	apiClient_, apiClientError = client.MakeFromEnv()
	kinds := schema.KindCatalog{}
	if apiClientError == nil {
		kinds = apiClient_.GetKinds()
	} else {
		kinds = schema.DefaultKind()
	}
	debug = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more information for debugging")
	initGet(kinds)
	initDelete(kinds)
	initApply()
	initMkKind()
	initPrintKind(kinds)
}
