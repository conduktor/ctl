package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/conduktor/ctl/printutils"
	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/utils"
	"github.com/go-resty/resty/v2"
	"os"
	"strings"
)

type Client struct {
	token   string
	baseUrl string
	client  *resty.Client
}

func Make(token string, baseUrl string, debug bool, key, cert string) *Client {
	certificate, _ := tls.LoadX509KeyPair(cert, key)
	return &Client{
		token:   token,
		baseUrl: baseUrl,
		client:  resty.New().SetDebug(debug).SetHeader("Authorization", "Bearer "+token).SetCertificates(certificate),
	}
}

func MakeFromEnv() *Client {
	token := os.Getenv("CDK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Please set CDK_TOKEN")
		os.Exit(1)
	}
	baseUrl := os.Getenv("CDK_BASE_URL")
	if baseUrl == "" {
		fmt.Fprintln(os.Stderr, "Please set CDK_BASE_URL")
		os.Exit(2)
	}
	debug := strings.ToLower(os.Getenv("CDK_DEBUG")) == "true"
	key := os.Getenv("CDK_KEY")
	cert := os.Getenv("CDK_CERT")

	return Make(token, baseUrl, debug, key, cert)
}

type UpsertResponse struct {
	UpsertResult string
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
	url := client.publicV1Url() + "/" + utils.UpperCamelToKebab(resource.Kind)
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

func (client *Client) Get(kind string) error {
	url := client.publicV1Url() + "/" + utils.UpperCamelToKebab(kind)
	resp, err := client.client.R().Get(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	}
	return printResponseAsYaml(resp.Body())
}

func (client *Client) Describe(kind, name string) error {
	url := client.publicV1Url() + "/" + utils.UpperCamelToKebab(kind) + "/" + name
	resp, err := client.client.R().Get(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf("error describing resources %s/%s, got status code: %d:\n %s", kind, name, resp.StatusCode(), string(resp.Body()))
	}
	return printResponseAsYaml(resp.Body())
}

func (client *Client) Delete(kind, name string) error {
	url := client.publicV1Url() + "/" + utils.UpperCamelToKebab(kind) + "/" + name
	resp, err := client.client.R().Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	} else {
		fmt.Printf("%s/%s deleted\n", kind, name)
	}

	return err
}

func (client *Client) GetOpenApi() ([]byte, error) {
	url := client.baseUrl + "public/docs/docs.yaml"
	resp, err := client.client.R().Get(url)
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	return resp.Body(), nil
}
