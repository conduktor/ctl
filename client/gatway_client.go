package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/go-resty/resty/v2"
)

type GatewayClient struct {
	cdkGatewayUser     string
	CdkGatewayPassword string
	baseUrl            string
	client             *resty.Client
	kinds              schema.KindCatalog
}

type GatewayApiParameter struct {
	BaseUrl            string
	Debug              bool
	CdkGatewayUser     string
	CdkGatewayPassword string
}

func MakeGateaway(apiParameter GatewayApiParameter) (*GatewayClient, error) {
	restyClient := resty.New().SetDebug(apiParameter.Debug).SetHeader("X-CDK-CLIENT", "CLI/"+utils.GetConduktorVersion())

	if apiParameter.BaseUrl == "" {
		return nil, fmt.Errorf("Please set CDK_GATEWAY_BASE_URL")
	}

	if apiParameter.CdkGatewayUser == "" || apiParameter.CdkGatewayPassword == "" {
		return nil, fmt.Errorf("CDK_GATEWAY_USER and CDK_GATEWAY_PASSWORD must be provided")
	}

	result := &GatewayClient{
		cdkGatewayUser:     apiParameter.CdkGatewayUser,
		CdkGatewayPassword: apiParameter.CdkGatewayPassword,
		baseUrl:            apiParameter.BaseUrl,
		client:             restyClient,
		kinds:              nil,
	}

	result.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	result.client.SetDisableWarn(true)
	result.client.SetBasicAuth(apiParameter.CdkGatewayUser, apiParameter.CdkGatewayPassword)

	err := result.initKindFromApi()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot access the Conduktor API: %s\nUsing offline defaults.\n", err)
		result.kinds = schema.DefaultKind()
	}

	return result, nil
}

func MakeGatewayClientFromEnv() (*GatewayClient, error) {
	apiParameter := GatewayApiParameter{
		BaseUrl:            os.Getenv("CDK_GATEWAY_BASE_URL"),
		Debug:              strings.ToLower(os.Getenv("CDK_DEBUG")) == "true",
		CdkGatewayUser:     os.Getenv("CDK_GATEWAY_USER"),
		CdkGatewayPassword: os.Getenv("CDK_GATEWAY_PASSWORD"),
	}

	client, err := MakeGateaway(apiParameter)
	if err != nil {
		return nil, fmt.Errorf("Cannot create client: %s", err)
	}
	return client, nil
}

func (client *GatewayClient) Get(kind *schema.Kind, parentPathValue []string) error {
	url := client.baseUrl + kind.ListPath(parentPathValue)
	resp, err := client.client.R().Get(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	}
	return printResponseAsYaml(resp.Body())
}

func (client *GatewayClient) Describe(kind *schema.Kind, parentPathValue []string, name string) error {
	url := client.baseUrl + kind.DescribePath(parentPathValue, name)
	resp, err := client.client.R().Get(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf("error describing resources %s/%s, got status code: %d:\n %s", kind.GetName(), name, resp.StatusCode(), string(resp.Body()))
	}
	return printResponseAsYaml(resp.Body())
}

func (client *GatewayClient) Delete(kind *schema.Kind, parentPathValue []string, name string) error {
	url := client.baseUrl + kind.DescribePath(parentPathValue, name)
	resp, err := client.client.R().Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	} else {
		fmt.Printf("%s/%s deleted\n", kind.GetName(), name)
	}

	return err
}

func (client *GatewayClient) Apply(resource *resource.Resource, dryMode bool) (string, error) {
	kinds := client.GetKinds()
	kind, ok := kinds[resource.Kind]
	if !ok {
		return "", fmt.Errorf("kind %s not found", resource.Kind)
	}
	applyPath, err := kind.ApplyPath(resource)
	if err != nil {
		return "", err
	}
	url := client.baseUrl + applyPath
	builder := client.client.R().SetBody(resource.Json)
	if dryMode {
		builder = builder.SetQueryParam("dryMode", "true")
	}
	resp, err := builder.Put(url)
	if err != nil {
		return "", err
	} else if resp.IsError() {
		return "", fmt.Errorf(extractApiError(resp))
	}
	bodyBytes := resp.Body()
	var upsertResponse string
	err = json.Unmarshal(bodyBytes, &upsertResponse)
	//in case backend format change (not json string anymore). Let not fail the client for that
	if err != nil {
		return resp.String(), nil
	}
	return upsertResponse, nil
}

func (client *GatewayClient) GetOpenApi() ([]byte, error) {
	url := client.baseUrl + "/gateway/v2/docs"
	resp, err := client.client.R().Get(url)
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	return resp.Body(), nil
}

func (client *GatewayClient) initKindFromApi() error {
	data, err := client.GetOpenApi()
	if err != nil {
		return fmt.Errorf("Cannot get openapi: %s", err)
	}
	schema, err := schema.New(data)
	if err != nil {
		return fmt.Errorf("Cannot parse openapi: %s", err)
	}
	strict := false
	client.kinds, err = schema.GetKinds(strict)
	if err != nil {
		fmt.Errorf("Cannot extract kinds from openapi: %s", err)
	}
	return nil
}

func (client *GatewayClient) GetKinds() schema.KindCatalog {
	return client.kinds
}
