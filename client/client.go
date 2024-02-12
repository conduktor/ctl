package client

import (
	"fmt"
	"github.com/conduktor/ctl/resource"
	"github.com/go-resty/resty/v2"
	"os"
)

type Client struct {
	token   string
	baseUrl string
	client  *resty.Client
}

func Make(token string, baseUrl string) Client {
	return Client{
		token:   token,
		baseUrl: baseUrl,
		client:  resty.New(),
	}
}

func MakeFromEnv() Client {
	token := os.Getenv("CDK_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Please seat CDK_TOKEN")
		os.Exit(1)
	}
	baseUrl := os.Getenv("CDK_BASE_URL")
	if baseUrl == "" {
		fmt.Fprintln(os.Stderr, "Please set CDK_BASE_URL")
		os.Exit(2)
	}

	return Make(token, baseUrl)
}

func (client *Client) Apply(resource *resource.Resource) error {
	url := client.baseUrl + "/" + resource.Kind
	resp, err := client.client.R().SetHeader("Authentication", "Bearer "+client.token).SetBody(resource.Json).Put(url)
	if resp.IsError() {
		return fmt.Errorf("Error applying resource %s/%s, got status code: %d:\n %s", resource.Kind, resource.Name, resp.StatusCode(), string(resp.Body()))
	}
	return err
}
