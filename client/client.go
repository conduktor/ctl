package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/printutils"
	"github.com/conduktor/ctl/resource"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	token   string
	baseUrl string
	client  *resty.Client
}

func Make(token string, baseUrl string, debug bool, key, cert string) Client {
	certificate, _ := tls.LoadX509KeyPair(cert, key)
	return Client{
		token:   token,
		baseUrl: baseUrl,
		client:  resty.New().SetDebug(debug).SetHeader("Authorization", "Bearer "+token).SetCertificates(certificate),
	}
}

func MakeFromEnv(debug bool, key, cert string) Client {
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
	finalKey := key
	finalCert := cert
	if finalKey == "" {
		finalKey = os.Getenv("CDK_KEY")
	}
	if finalCert == "" {
		finalCert = os.Getenv("CDK_CERT")
	}

	return Make(token, baseUrl, debug, finalKey, finalCert)
}

type UpsertResponse struct {
	UpsertResult string
}

func (client *Client) Apply(resource *resource.Resource, dryMode bool) (string, error) {
	url := client.baseUrl + "/" + UpperCamelToKebab(resource.Kind)
	builder := client.client.R().SetBody(resource.Json)
	if dryMode {
		builder = builder.SetQueryParam("dryMode", "true")
	}
	resp, err := builder.Put(url)
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("error applying resource %s/%s, got status code: %d:\n %s", resource.Kind, resource.Name, resp.StatusCode(), string(resp.Body()))
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
	url := client.baseUrl + "/" + UpperCamelToKebab(kind)
	resp, err := client.client.R().Get(url)
	if resp.IsError() {
		return fmt.Errorf("error listing resources of kind %s, got status code: %d:\n %s", kind, resp.StatusCode(), string(resp.Body()))
	}
	if err != nil {
		return err
	}
	return printResponseAsYaml(resp.Body())
}
func (client *Client) Describe(kind, name string) error {
	url := client.baseUrl + "/" + UpperCamelToKebab(kind) + "/" + name
	resp, err := client.client.R().Get(url)
	if resp.IsError() {
		return fmt.Errorf("error describing resources %s/%s, got status code: %d:\n %s", kind, name, resp.StatusCode(), string(resp.Body()))
	}
	if err != nil {
		return err
	}
	return printResponseAsYaml(resp.Body())
}

func (client *Client) Delete(kind, name string) error {
	url := client.baseUrl + "/" + UpperCamelToKebab(kind) + "/" + name
	resp, err := client.client.R().Delete(url)
	if resp.IsError() {
		return fmt.Errorf("error deleting resources %s/%s, got status code: %d:\n %s", kind, name, resp.StatusCode(), string(resp.Body()))
	} else {
		fmt.Printf("%s/%s deleted\n", kind, name)
	}

	return err
}

func UpperCamelToKebab(input string) string {
	// Split the input string into words
	words := make([]string, 0)
	currentWord := ""
	for _, char := range input {
		if char >= 'A' && char <= 'Z' {
			if currentWord != "" {
				words = append(words, currentWord)
			}
			currentWord = string(char)
		} else {
			currentWord += string(char)
		}
	}
	if currentWord != "" {
		words = append(words, currentWord)
	}

	// Join the words with hyphens
	kebabCase := strings.ToLower(strings.Join(words, "-"))

	return kebabCase
}
