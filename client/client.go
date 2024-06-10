package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/conduktor/ctl/printutils"
	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/go-resty/resty/v2"
	"os"
	"strings"
)

type Client struct {
	apiKey  string
	baseUrl string
	client  *resty.Client
	kinds   schema.KindCatalog
}

func Make(apiKey string, baseUrl string, debug bool, key, cert, cacert string, insecure bool) (*Client, error) {
	//apiKey is set later because it's not mandatory for getting the openapi and parsing different kind
	restyClient := resty.New().SetDebug(debug).SetHeader("X-CDK-CLIENT", "CLI/"+utils.GetConduktorVersion())

	if (key == "" && cert != "") || (key != "" && cert == "") {
		return nil, fmt.Errorf("key and cert must be provided together")
	} else if key != "" && cert != "" {
		certificate, err := tls.LoadX509KeyPair(cert, key)
		restyClient.SetCertificates(certificate)
		if err != nil {
			return nil, err
		}
	}

	if cacert != "" {
		restyClient.SetRootCertificate(cacert)
	}

	result := &Client{
		apiKey:  apiKey,
		baseUrl: baseUrl,
		client:  restyClient,
		kinds:   nil,
	}

	if apiKey != "" {
		result.setApiKeyInRestClient()
	} else {
		//it will be set later only when really needed
		//so aim is not fail when CDK_API_KEY is not set before printing the cmd help
	}

	if insecure {
		result.IgnoreUntrustedCertificate()
	}

	err := result.initKindFromApi()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot access the Conduktor API: %s\nUsing offline defaults.\n", err)
		result.kinds = schema.DefaultKind()
	}

	return result, nil
}

func MakeFromEnv() (*Client, error) {
	baseUrl := os.Getenv("CDK_BASE_URL")
	if baseUrl == "" {
		return nil, fmt.Errorf("Please set CDK_BASE_URL")
	}
	debug := strings.ToLower(os.Getenv("CDK_DEBUG")) == "true"
	key := os.Getenv("CDK_KEY")
	cert := os.Getenv("CDK_CERT")
	cacert := os.Getenv("CDK_CACERT")
	insecure := strings.ToLower(os.Getenv("CDK_INSECURE")) == "true"

	client, err := Make("", baseUrl, debug, key, cert, cacert, insecure)
	if err != nil {
		return nil, fmt.Errorf("Cannot create client: %s", err)
	}
	return client, nil
}

type UpsertResponse struct {
	UpsertResult string
}

func (c *Client) IgnoreUntrustedCertificate() {
	c.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}

func (c *Client) setApiKeyFromEnvIfNeeded() {
	if c.apiKey == "" {
		apiKey := os.Getenv("CDK_API_KEY")
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Please set CDK_API_KEY")
			os.Exit(1)
		}
		c.apiKey = apiKey
		c.setApiKeyInRestClient()
	}
}

func (c *Client) setApiKeyInRestClient() {
	c.client = c.client.SetHeader("Authorization", "Bearer "+c.apiKey)
}
func extractApiError(resp *resty.Response) string {
	var apiError ApiError
	jsonError := json.Unmarshal(resp.Body(), &apiError)
	if jsonError != nil {
		return resp.String()
	} else {
		return apiError.String()
	}
}

func (client *Client) publicV1Url() string {
	return client.baseUrl + "/public/v1"
}

func (client *Client) ActivateDebug() {
	client.client.SetDebug(true)
}

func (client *Client) Apply(resource *resource.Resource, dryMode bool) (string, error) {
	client.setApiKeyFromEnvIfNeeded()
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
	var upsertResponse UpsertResponse
	err = json.Unmarshal(bodyBytes, &upsertResponse)
	//in case backend format change (not json string anymore). Let not fail the client for that
	if err != nil {
		return resp.String(), nil
	}
	return upsertResponse.UpsertResult, nil
}

func printResponseAsYaml(bytes []byte) error {
	var data interface{}
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}
	return printutils.PrintResourceLikeYamlFile(os.Stdout, data)
}

func (client *Client) Get(kind *schema.Kind, parentPathValue []string) error {
	client.setApiKeyFromEnvIfNeeded()
	url := client.baseUrl + kind.ListPath(parentPathValue)
	resp, err := client.client.R().Get(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	}
	return printResponseAsYaml(resp.Body())
}

func (client *Client) Describe(kind *schema.Kind, parentPathValue []string, name string) error {
	client.setApiKeyFromEnvIfNeeded()
	url := client.baseUrl + kind.DescribePath(parentPathValue, name)
	resp, err := client.client.R().Get(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf("error describing resources %s/%s, got status code: %d:\n %s", kind.GetName(), name, resp.StatusCode(), string(resp.Body()))
	}
	return printResponseAsYaml(resp.Body())
}

func (client *Client) Delete(kind *schema.Kind, parentPathValue []string, name string) error {
	client.setApiKeyFromEnvIfNeeded()
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

func (client *Client) GetOpenApi() ([]byte, error) {
	url := client.baseUrl + "/public/docs/docs.yaml"
	resp, err := client.client.R().Get(url)
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	return resp.Body(), nil
}

func (client *Client) initKindFromApi() error {
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

func (client *Client) GetKinds() schema.KindCatalog {
	return client.kinds
}
