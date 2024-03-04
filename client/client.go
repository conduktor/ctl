package client

import (
	"encoding/json"
	"fmt"
	"github.com/conduktor/ctl/printutils"
	"github.com/conduktor/ctl/resource"
	"github.com/go-resty/resty/v2"
	"os"
)

type Client struct {
	token   string
	baseUrl string
	client  *resty.Client
}

func Make(token string, baseUrl string, debug bool) Client {
	return Client{
		token:   token,
		baseUrl: baseUrl,
		client:  resty.New().SetDebug(debug).SetHeader("Authorization", "Bearer "+token),
	}
}

func MakeFromEnv(debug bool) Client {
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

	return Make(token, baseUrl, debug)
}

type UpsertResponse struct {
	UpsertResult string
}

func (client *Client) Apply(resource *resource.Resource, dryMode bool) (string, error) {
	url := client.baseUrl + "/" + resource.Kind
	builder := client.client.R().SetBody(resource.Json)
	if dryMode {
		builder = builder.SetQueryParam("dryMode", "true")
	}
	resp, err := builder.Put(url)
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("Error applying resource %s/%s, got status code: %d:\n %s", resource.Kind, resource.Name, resp.StatusCode(), string(resp.Body()))
	}
	bodyBytes := resp.Body()
	var upsertResponse UpsertResponse
	err = json.Unmarshal(bodyBytes, &upsertResponse)
	//in case backend format change (not json string anymore). Let not fail the client for that
	if err != nil {
		return resp.String(), nil
	}
	if dryMode && upsertResponse.UpsertResult == "Created" {
		return "To be created", nil
	}
	if dryMode && upsertResponse.UpsertResult == "Updated" {
		return "To be updated", nil
	}
	if dryMode && upsertResponse.UpsertResult == "NotChanged" {
		return "Nothing to do", nil
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
	url := client.baseUrl + "/" + kind
	resp, err := client.client.R().Get(url)
	if resp.IsError() {
		return fmt.Errorf("Error listing resources of kind %s, got status code: %d:\n %s", kind, resp.StatusCode(), string(resp.Body()))
	}
	if err != nil {
		return err
	}
	return printResponseAsYaml(resp.Body())
}
func (client *Client) Describe(kind, name string) error {
	url := client.baseUrl + "/" + kind + "/" + name
	resp, err := client.client.R().Get(url)
	if resp.IsError() {
		return fmt.Errorf("Error describing resources %s/%s, got status code: %d:\n %s", kind, name, resp.StatusCode(), string(resp.Body()))
	}
	if err != nil {
		return err
	}
	return printResponseAsYaml(resp.Body())
}

func (client *Client) Delete(kind, name string) error {
	url := client.baseUrl + "/" + kind + "/" + name
	resp, err := client.client.R().Delete(url)
	if resp.IsError() {
		return fmt.Errorf("Error deleting resources %s/%s, got status code: %d:\n %s", kind, name, resp.StatusCode(), string(resp.Body()))
	} else {
		fmt.Printf("%s/%s deleted\n", kind, name)
	}

	return err
}
