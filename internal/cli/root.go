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
	Catalog               schema.Catalog
	Strict                bool
	Debug                 *bool
}

func NewRootContext(
	consoleAPIClient *client.Client,
	consoleAPIClientError error,
	gatewayAPIClient *client.GatewayClient,
	gatewayAPIClientError error,
	catalog schema.Catalog,
	strict bool,
	debug *bool,
) RootContext {
	return RootContext{
		consoleAPIClient:      consoleAPIClient,
		consoleAPIClientError: consoleAPIClientError,
		gatewayAPIClient:      gatewayAPIClient,
		gatewayAPIClientError: gatewayAPIClientError,
		Catalog:               catalog,
		Strict:                strict,
		Debug:                 debug,
	}
}

func (c *RootContext) ConsoleAPIClient() *client.Client {
	if c.consoleAPIClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create client: %s", c.consoleAPIClientError)
		// Fail fast if client cannot be created
		os.Exit(1)
	}
	return c.consoleAPIClient
}

func (c *RootContext) GatewayAPIClient() *client.GatewayClient {
	if c.gatewayAPIClientError != nil {
		fmt.Fprintf(os.Stderr, "Cannot create gateway client: %s", c.gatewayAPIClientError)
		// Fail fast if client cannot be created
		os.Exit(1)
	}
	return c.gatewayAPIClient
}
