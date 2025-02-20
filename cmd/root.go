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
			// ActivateDebug() will enable debug mode for the resty client.
			// Doesn't need to be set if the client was not initialised correctly.
			if apiClientError == nil {
				consoleApiClient().ActivateDebug()
			}
			if gatewayApiClientError == nil {
				gatewayApiClient().ActivateDebug()
			}
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
	var consoleKinds *schema.Catalog
	if apiClientError == nil {
		consoleKinds = apiClient_.GetCatalog()
	} else {
		consoleKinds = schema.ConsoleDefaultCatalog()
	}
	gatewayApiClient_, gatewayApiClientError = client.MakeGatewayClientFromEnv()
	var gatewayKinds *schema.Catalog
	if gatewayApiClientError == nil {
		gatewayKinds = gatewayApiClient().GetCatalog()
	} else {
		gatewayKinds = schema.GatewayDefaultCatalog()
	}
	catalog := consoleKinds.Merge(gatewayKinds)
	debug = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show more information for debugging")
	var permissive = rootCmd.PersistentFlags().Bool("permissive", false, "Permissive mode, allow undefined environment variables")
	initGet(catalog.Kind)
	initTemplate(catalog.Kind)
	initDelete(catalog.Kind, !*permissive)
	initApply(catalog.Kind, !*permissive)
	intConsoleMakeCatalog()
	initGatewayMakeCatalog()
	initPrintCatalog(catalog)
	initSql(catalog.Kind)
	initRun(catalog.Run)
}
