package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/client"
	"github.com/conduktor/ctl/schema"
	"github.com/spf13/cobra"
)

var debug *bool
var apiClient_ *client.Client
var apiClientError error
var gatewayApiClient_ *client.GatewayClient
var gatewayApiClientError error

func consoleApiClient() *client.Client {
	if apiClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create client: %s", apiClientError)
		os.Exit(1)
	}
	return apiClient_
}

func gatewayApiClient() *client.GatewayClient {
	if gatewayApiClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create gateway client: %s", gatewayApiClientError)
		os.Exit(1)
	}
	return gatewayApiClient_
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "conduktor",
	Short: "Command line tools for conduktor",
	Long: `Make sure you've set the environment variables CDK_USER/CDK_PASSWORD or CDK_API_KEY (generated from Console) and CDK_BASE_URL.
Additionally, you can configure client TLS authentication by providing your certificate paths in CDK_KEY and CDK_CERT.
For server TLS authentication, you can ignore the certificate by setting CDK_INSECURE=true, or provide a certificate authority using CDK_CACERT.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if *debug {
			consoleApiClient().ActivateDebug()
			gatewayApiClient().ActivateDebug()
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
	var kinds schema.KindCatalog
	if apiClientError == nil {
		kinds = apiClient_.GetKinds()
	} else {
		kinds = schema.ConsoleDefaultKind()
	}
	gatewayApiClient_, gatewayApiClientError = client.MakeGatewayClientFromEnv()
	var gatewayKinds schema.KindCatalog
	if gatewayApiClientError == nil {
		gatewayKinds = gatewayApiClient().GetKinds()
	} else {
		gatewayKinds = schema.GatewayDefaultKind()
	}
	for k, v := range gatewayKinds {
		kinds[k] = v
	}
	debug = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more information for debugging")
	var permissive = rootCmd.PersistentFlags().Bool("permissive", false, "permissive mode, allow undefined environment variables")
	initGet(kinds)
	initTemplate(kinds)
	initDelete(kinds, !*permissive)
	initApply(kinds, !*permissive)
	initConsoleMkKind()
	initGatewayMkKind()
	initPrintKind(kinds)
	initSql(kinds)
}
