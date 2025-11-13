package cmd

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/cli"
	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/spf13/cobra"
)

var rootContext cli.RootContext
var debug *bool
var apiClient_ *client.Client
var consoleAPIClientError error
var gatewayAPIClient_ *client.GatewayClient
var gatewayAPIClientError error

func consoleAPIClient() *client.Client {
	if consoleAPIClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create client: %s", consoleAPIClientError)
		os.Exit(1)
	}
	return apiClient_
}

func gatewayAPIClient() *client.GatewayClient {
	if gatewayAPIClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create gateway client: %s", gatewayAPIClientError)
		os.Exit(1)
	}
	return gatewayAPIClient_
}

// rootCmd represents the base command when called without any subcommands.
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
			if consoleAPIClientError == nil {
				consoleAPIClient().ActivateDebug()
			}
			if gatewayAPIClientError == nil {
				gatewayAPIClient().ActivateDebug()
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
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
	apiClient_, consoleAPIClientError = client.MakeFromEnv()
	var consoleKinds *schema.Catalog
	if consoleAPIClientError == nil {
		consoleKinds = apiClient_.GetCatalog()
	} else {
		consoleKinds = schema.ConsoleDefaultCatalog()
	}
	gatewayAPIClient_, gatewayAPIClientError = client.MakeGatewayClientFromEnv()
	var gatewayKinds *schema.Catalog
	if gatewayAPIClientError == nil {
		gatewayKinds = gatewayAPIClient().GetCatalog()
	} else {
		gatewayKinds = schema.GatewayDefaultCatalog()
	}
	catalog := consoleKinds.Merge(gatewayKinds)
	debug = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show more information for debugging")
	var permissive = rootCmd.PersistentFlags().Bool("permissive", false, "Permissive mode, allow undefined environment variables")
	strict := !*permissive

	rootContext = cli.NewRootContext(
		apiClient_,
		consoleAPIClientError,
		gatewayAPIClient_,
		gatewayAPIClientError,
		catalog,
		strict,
		debug,
	)

	initGet(rootContext)
	initTemplate(rootContext)
	initEdit(rootContext)
	initDelete(rootContext)
	initApply(rootContext)
	intConsoleMakeCatalog()
	initGatewayMakeCatalog()
	initPrintCatalog(catalog)
	initSQL(catalog.Kind)
	initRun(catalog.Run)
}
