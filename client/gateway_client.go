package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/go-resty/resty/v2"
)

type GatewayClient struct {
	cdkGatewayUser     string
	cdkGatewayPassword string
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

func MakeGateway(apiParameter GatewayApiParameter) (*GatewayClient, error) {
	restyClient := resty.New().SetDebug(apiParameter.Debug).SetHeader("X-CDK-CLIENT", "CLI/"+utils.GetConduktorVersion())

	if apiParameter.BaseUrl == "" {
		return nil, fmt.Errorf("Please set CDK_GATEWAY_BASE_URL")
	}

	if apiParameter.CdkGatewayUser == "" || apiParameter.CdkGatewayPassword == "" {
		return nil, fmt.Errorf("CDK_GATEWAY_USER and CDK_GATEWAY_PASSWORD must be provided")
	}

	result := &GatewayClient{
		cdkGatewayUser:     apiParameter.CdkGatewayUser,
		cdkGatewayPassword: apiParameter.CdkGatewayPassword,
		baseUrl:            apiParameter.BaseUrl,
		client:             restyClient,
		kinds:              nil,
	}

	result.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	result.client.SetDisableWarn(true)
	result.client.SetBasicAuth(apiParameter.CdkGatewayUser, apiParameter.CdkGatewayPassword)

	err := result.initKindFromApi()
	if err != nil {
		if apiParameter.Debug {
			fmt.Fprintf(os.Stderr, "Cannot access the Gateway Conduktor API: %s\nUsing offline defaults.\n", err)
		}
		result.kinds = schema.GatewayDefaultKind()
	}

	return result, nil
}

func MakeGatewayClientFromEnv() (*GatewayClient, error) {
	apiParameter := GatewayApiParameter{
		BaseUrl:            os.Getenv("CDK_GATEWAY_BASE_URL"),
		Debug:              utils.CdkDebug(),
		CdkGatewayUser:     os.Getenv("CDK_GATEWAY_USER"),
		CdkGatewayPassword: os.Getenv("CDK_GATEWAY_PASSWORD"),
	}

	client, err := MakeGateway(apiParameter)
	if err != nil {
		return nil, fmt.Errorf("Cannot create client: %s", err)
	}
	return client, nil
}

func (client *GatewayClient) Get(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, queryParams map[string]string) ([]resource.Resource, error) {
	var result []resource.Resource
	url := client.baseUrl + kind.ListPath(parentPathValue, parentQueryValue).Path
	requestBuilder := client.client.R()
	if queryParams != nil {
		requestBuilder = requestBuilder.SetQueryParams(queryParams)
	}
	resp, err := requestBuilder.Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractApiError(resp))
	}
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

func (client *GatewayClient) Describe(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) (resource.Resource, error) {
	var result resource.Resource
	url := client.baseUrl + kind.DescribePath(parentPathValue, parentQueryValue, name).Path
	resp, err := client.client.R().Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf("error describing resources %s/%s, got status code: %d:\n %s", kind.GetName(), name, resp.StatusCode(), string(resp.Body()))
	}
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

func (client *GatewayClient) Delete(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) error {
	url := client.baseUrl + kind.DescribePath(parentPathValue, parentQueryValue, name).Path
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

func (client *GatewayClient) DeleteResourceByName(resource *resource.Resource) error {
	kinds := client.GetKinds()
	kind, ok := kinds[resource.Kind]
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	deletePath, err := kind.DeletePath(resource)
	if err != nil {
		return err
	}
	url := client.baseUrl + deletePath
	resp, err := client.client.R().Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	} else {
		fmt.Printf("%s/%s deleted\n", kind.GetName(), resource.Name)
	}

	return err
}

func (client *GatewayClient) DeleteResourceByNameAndVCluster(resource *resource.Resource) error {
	kinds := client.GetKinds()
	kind, ok := kinds[resource.Kind]
	name := resource.Name
	vCluster := resource.Metadata["vCluster"]
	if vCluster == nil {
		vCluster = "passthrough"
	}
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	deletePath := kind.ListPath(nil, nil)
	url := client.baseUrl + deletePath.Path
	resp, err := client.client.R().SetBody(map[string]string{"name": name, "vCluster": vCluster.(string)}).Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	} else {
		fmt.Printf("%s/%s deleted\n", kind.GetName(), resource.Name)
	}

	return err
}

type DeleteInterceptorPayload struct {
	VCluster *string `json:"vCluster"`
	Group    *string `json:"group"`
	Username *string `json:"username"`
}

func (client *GatewayClient) DeleteResourceInterceptors(resource *resource.Resource) error {
	kinds := client.GetKinds()
	kind, ok := kinds[resource.Kind]
	scope := resource.Metadata["scope"]
	var deleteInterceptorPayload DeleteInterceptorPayload
	if scope == nil {
		deleteInterceptorPayload = DeleteInterceptorPayload{
			VCluster: nil,
			Group:    nil,
			Username: nil,
		}
	} else {
		vCluster := scope.(map[string]interface{})["vCluster"]
		var vClusterValue string
		if vCluster != nil && vCluster.(string) != "" {
			vClusterValue = vCluster.(string)
			deleteInterceptorPayload.VCluster = &vClusterValue
		}
		group := scope.(map[string]interface{})["group"]
		var groupValue string
		if group != nil && group.(string) != "" {
			groupValue = group.(string)
			deleteInterceptorPayload.Group = &groupValue
		}
		username := scope.(map[string]interface{})["username"]
		var usernameValue string
		if username != nil && username.(string) != "" {
			usernameValue = username.(string)
			deleteInterceptorPayload.Username = &usernameValue
		}
	}
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	deletePath, err := kind.DeletePath(resource)
	if err != nil {
		return err
	}
	url := client.baseUrl + deletePath
	resp, err := client.client.R().SetBody(deleteInterceptorPayload).Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	} else {
		fmt.Printf("%s/%s deleted\n", kind.GetName(), resource.Name)
	}

	return err
}

func (client *GatewayClient) DeleteKindByNameAndVCluster(kind *schema.Kind, param map[string]string) error {
	url := client.baseUrl + kind.ListPath(nil, nil).Path
	req := client.client.R()
	req.SetBody(param)
	resp, err := req.Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	} else {
		fmt.Printf("%s/%s deleted\n", kind.GetName(), param)
	}

	return err
}

func (client *GatewayClient) DeleteInterceptor(kind *schema.Kind, name string, param map[string]string) error {
	url := client.baseUrl + kind.ListPath(nil, nil).Path + "/" + name
	req := client.client.R()
	var bodyParams = make(map[string]interface{})
	for k, v := range param {
		if v == "" {
			bodyParams[k] = nil
		} else {
			bodyParams[k] = v
		}
	}
	req.SetBody(bodyParams)
	resp, err := req.Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	} else {
		fmt.Printf("%s/%s deleted\n", kind.GetName(), param)
	}

	return err
}

func (client *GatewayClient) ActivateDebug() {
	client.client.SetDebug(true)
}

func (client *GatewayClient) Apply(resource *resource.Resource, dryMode bool) (string, error) {
	kinds := client.GetKinds()
	kind, ok := kinds[resource.Kind]
	if !ok {
		return "", fmt.Errorf("kind %s not found", resource.Kind)
	}
	applyQueryInfo, err := kind.ApplyPath(resource)
	if err != nil {
		return "", err
	}
	url := client.baseUrl + applyQueryInfo.Path
	builder := client.client.R().SetBody(resource.Json)
	for _, param := range applyQueryInfo.QueryParams {
		builder.SetQueryParam(param.Name, param.Value)
	}
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
	client.kinds, err = schema.GetGatewayKinds(strict)
	if err != nil {
		fmt.Errorf("Cannot extract kinds from openapi: %s", err)
	}
	return nil
}

func (client *GatewayClient) GetKinds() schema.KindCatalog {
	return client.kinds
}
