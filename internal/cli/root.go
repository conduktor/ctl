package cli

import (
	"fmt"
	"os"

	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/schema"
)

type RootContext struct {
	consoleAPIClient      *client.Client
	consoleAPIClientError error
	gatewayAPIClient      *client.GatewayClient
	gatewayAPIClientError error
	kinds                 schema.KindCatalog
	strict                bool
	debug                 *bool
}

func NewRootContext(
	consoleAPIClient *client.Client,
	consoleAPIClientError error,
	gatewayAPIClient *client.GatewayClient,
	gatewayAPIClientError error,
	kinds schema.KindCatalog,
	strict bool,
	debug *bool,
) RootContext {
	return RootContext{
		consoleAPIClient:      consoleAPIClient,
		consoleAPIClientError: consoleAPIClientError,
		gatewayAPIClient:      gatewayAPIClient,
		gatewayAPIClientError: gatewayAPIClientError,
		kinds:                 kinds,
		strict:                strict,
		debug:                 debug,
	}
}

func (c *RootContext) ConsoleAPIClient() *client.Client {
	if c.consoleAPIClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create client: %s", c.consoleAPIClientError)
		os.Exit(1)
	}
	return c.consoleAPIClient
}

func (c *RootContext) GatewayAPIClient() *client.GatewayClient {
	if c.gatewayAPIClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create gateway client: %s", c.gatewayAPIClientError)
		os.Exit(1)
	}
	return c.gatewayAPIClient
}
