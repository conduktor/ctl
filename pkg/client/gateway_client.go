package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/utils"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/go-resty/resty/v2"
)

type GatewayClient struct {
	cdkGatewayUser     string
	cdkGatewayPassword string
	baseURL            string
	client             *resty.Client
	schemaCatalog      *schema.Catalog
}

type GatewayAPIParameter struct {
	BaseURL            string
	Debug              bool
	CdkGatewayUser     string
	CdkGatewayPassword string
}

func MakeGateway(apiParameter GatewayAPIParameter) (*GatewayClient, error) {
	restyClient := resty.New().SetDebug(apiParameter.Debug).SetHeader("X-CDK-CLIENT", "CLI/"+utils.GetConduktorVersion())

	if apiParameter.BaseURL == "" {
		return nil, fmt.Errorf("Please set CDK_GATEWAY_BASE_URL")
	}

	if apiParameter.CdkGatewayUser == "" || apiParameter.CdkGatewayPassword == "" {
		return nil, fmt.Errorf("CDK_GATEWAY_USER and CDK_GATEWAY_PASSWORD must be provided")
	}

	result := &GatewayClient{
		cdkGatewayUser:     apiParameter.CdkGatewayUser,
		cdkGatewayPassword: apiParameter.CdkGatewayPassword,
		baseURL:            apiParameter.BaseURL,
		client:             restyClient,
		schemaCatalog:      nil,
	}

	result.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	result.client.SetDisableWarn(true)
	result.client.SetBasicAuth(apiParameter.CdkGatewayUser, apiParameter.CdkGatewayPassword)

	err := result.initCatalogFromAPI()
	if err != nil {
		if apiParameter.Debug {
			fmt.Fprintf(os.Stderr, "Cannot access the Gateway Conduktor API: %s\nUsing offline defaults.\n", err)
		}
		result.schemaCatalog = schema.GatewayDefaultCatalog()
	}

	return result, nil
}

func MakeGatewayClientFromEnv() (*GatewayClient, error) {
	apiParameter := GatewayAPIParameter{
		BaseURL:            os.Getenv("CDK_GATEWAY_BASE_URL"),
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
	queryInfo := kind.ListPath(parentPathValue, parentQueryValue)
	url := client.baseURL + queryInfo.Path
	requestBuilder := client.client.R()
	for _, p := range queryInfo.QueryParams {
		requestBuilder = requestBuilder.SetQueryParam(p.Name, p.Value)
	}
	if queryParams != nil {
		requestBuilder = requestBuilder.SetQueryParams(queryParams)
	}
	resp, err := requestBuilder.Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractAPIError(resp))
	}
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

func (client *GatewayClient) Describe(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) (resource.Resource, error) {
	var result resource.Resource
	queryInfo := kind.DescribePath(parentPathValue, parentQueryValue, name)
	url := client.baseURL + queryInfo.Path
	requestBuilder := client.client.R()
	for _, p := range queryInfo.QueryParams {
		requestBuilder = requestBuilder.SetQueryParam(p.Name, p.Value)
	}
	resp, err := requestBuilder.Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf("error describing resources %s/%s, got status code: %d:\n %s", kind.GetName(), name, resp.StatusCode(), string(resp.Body()))
	}
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

func (client *GatewayClient) Delete(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) error {
	queryInfo := kind.DescribePath(parentPathValue, parentQueryValue, name)
	url := client.baseURL + queryInfo.Path
	requestBuilder := client.client.R()
	for _, p := range queryInfo.QueryParams {
		requestBuilder = requestBuilder.SetQueryParam(p.Name, p.Value)
	}
	resp, err := requestBuilder.Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractAPIError(resp))
	} else {
		fmt.Printf("%s/%s: Deleted\n", kind.GetName(), name)
	}

	return err
}

func (client *GatewayClient) DeleteResourceByName(resource *resource.Resource) error {
	kinds := client.GetKinds()
	requestBuilder := client.client.R()
	kind, ok := kinds[resource.Kind]
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	deletePath, queryParams, err := kind.DeletePath(resource)
	if err != nil {
		return err
	}
	url := client.baseURL + deletePath
	if queryParams != nil {
		requestBuilder = requestBuilder.SetQueryParams(queryParams)
	}
	resp, err := requestBuilder.Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractAPIError(resp))
	} else {
		fmt.Printf("%s/%s: Deleted\n", kind.GetName(), resource.Name)
	}

	return err
}

func (client *GatewayClient) DeleteResourceByNameAndVCluster(resource *resource.Resource) error {
	kinds := client.GetKinds()
	kind, ok := kinds[resource.Kind]
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	name := resource.Name
	vCluster, ok := resource.Metadata["vCluster"].(string)
	if !ok || vCluster == "" {
		vCluster = "passthrough"
	}
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	deletePath := kind.ListPath(nil, nil)
	url := client.baseURL + deletePath.Path
	if !ok {
		return fmt.Errorf("vCluster value is not a string for resource %s/%s", resource.Kind, resource.Name)
	}
	resp, err := client.client.R().SetBody(map[string]string{"name": name, "vCluster": vCluster}).Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractAPIError(resp))
	} else {
		fmt.Printf("%s/%s: Deleted\n", kind.GetName(), resource.Name)
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
	var deleteInterceptorPayload *DeleteInterceptorPayload
	if scope == nil {
		deleteInterceptorPayload = nil
	} else {
		deleteInterceptorPayload = &DeleteInterceptorPayload{}

		scopeMap, ok := scope.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid scope format for resource %s/%s", resource.Kind, resource.Name)
		}

		vClusterValue, ok := scopeMap["vCluster"].(string)
		if ok && vClusterValue != "" {
			deleteInterceptorPayload.VCluster = &vClusterValue
		}
		groupValue, ok := scopeMap["group"].(string)
		if ok && groupValue != "" {
			deleteInterceptorPayload.Group = &groupValue
		}
		usernameValue, ok := scopeMap["username"].(string)
		if ok && usernameValue != "" {
			deleteInterceptorPayload.Username = &usernameValue
		}
	}
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	deletePath, _, err := kind.DeletePath(resource)
	if err != nil {
		return err
	}
	url := client.baseURL + deletePath
	req := client.client.R()
	if deleteInterceptorPayload != nil {
		req = req.SetBody(deleteInterceptorPayload)
	}
	resp, err := req.Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		msg := extractAPIError(resp)
		if deleteInterceptorPayload == nil {
			msg += "\nThis error may be caused by a bug in Conduktor Gateway REST api defaults fixed in version 3.11.0.\nAs a quick fix, you can fetch your interceptor to see the exact scope and use this when deleting."
		}
		return fmt.Errorf("%s", msg)
	} else {
		fmt.Printf("%s/%s: Deleted\n", kind.GetName(), resource.Name)
	}

	return err
}

func (client *GatewayClient) DeleteKindByNameAndVCluster(kind *schema.Kind, param map[string]string) error {
	url := client.baseURL + kind.ListPath(nil, nil).Path
	req := client.client.R()
	req.SetBody(param)
	resp, err := req.Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractAPIError(resp))
	} else {
		fmt.Printf("%s/%s: Deleted\n", kind.GetName(), param)
	}

	return err
}

func (client *GatewayClient) DeleteInterceptor(kind *schema.Kind, name string, param map[string]string) error {
	url := client.baseURL + kind.ListPath(nil, nil).Path + "/" + name
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
		return fmt.Errorf(extractAPIError(resp))
	} else {
		fmt.Printf("%s/%s: Deleted\n", kind.GetName(), param)
	}

	return err
}

func (client *GatewayClient) ActivateDebug() {
	client.client.SetDebug(true)
}

func (client *GatewayClient) Run(run schema.Run, pathValue []string, queryParams map[string]string, body interface{}) ([]byte, error) {
	if run.BackendType != schema.GATEWAY {
		return nil, fmt.Errorf("Only console backend type is supported by console client")
	}
	path := run.BuildPath(pathValue)
	url := client.baseURL + path
	requestBuilder := client.client.R()
	for k, v := range queryParams {
		requestBuilder.SetQueryParam(k, v)
	}
	if body != nil {
		requestBuilder = requestBuilder.SetBody(body)
	}
	resp, err := requestBuilder.Execute(run.Method, url)
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, fmt.Errorf(extractAPIError(resp))
	}
	return resp.Body(), nil
}

func (client *GatewayClient) Apply(resource *resource.Resource, dryMode bool, diffMode bool) (Result, error) {
	var result = Result{}

	kinds := client.GetKinds()
	kind, ok := kinds[resource.Kind]
	if !ok {
		return result, fmt.Errorf("kind %s not found", resource.Kind)
	}
	applyQueryInfo, err := kind.ApplyPath(resource)
	if err != nil {
		return result, err
	}
	url := client.baseURL + applyQueryInfo.Path
	builder := client.client.R().SetBody(resource.Json)
	for _, param := range applyQueryInfo.QueryParams {
		builder = builder.SetQueryParam(param.Name, param.Value)
	}
	if dryMode {
		builder = builder.SetQueryParam("dryMode", "true")
	}
	if diffMode {
		currentRes, err := client.GetFromResource(resource)
		if err != nil {
			return result, err
		}
		diff, err := utils.DiffResources(&currentRes, resource)
		if err != nil {
			return result, err
		}
		result.Diff = diff
	}
	resp, err := builder.Put(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractAPIError(resp))
	}
	bodyBytes := resp.Body()
	err = json.Unmarshal(bodyBytes, &result)
	//in case backend format change (not json string anymore). Let not fail the client for that
	if err != nil {
		result.UpsertResult = resp.String()
		//nolint:nilerr
		return result, nil
	}
	return result, nil
}

func (client *GatewayClient) GetOpenAPI() ([]byte, error) {
	url := client.baseURL + "/gateway/v2/docs"
	resp, err := client.client.R().Get(url)
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	return resp.Body(), nil
}

func (client *GatewayClient) initCatalogFromAPI() error {
	data, err := client.GetOpenAPI()
	if err != nil {
		return fmt.Errorf("Cannot get openapi: %s", err)
	}
	schema, err := schema.NewOpenAPIParser(data)
	if err != nil {
		return fmt.Errorf("Cannot parse openapi: %s", err)
	}
	strict := false
	client.schemaCatalog, err = schema.GetGatewayCatalog(strict)
	if err != nil {
		return fmt.Errorf("annot extract schemaCatalog from openapi: %s", err)
	}
	return nil
}

func (client *GatewayClient) GetKinds() schema.KindCatalog {
	return client.schemaCatalog.Kind
}

func (client *GatewayClient) GetCatalog() *schema.Catalog {
	return client.schemaCatalog
}

func (client *GatewayClient) GetFromResource(res *resource.Resource) (resource.Resource, error) {
	var results []resource.Resource
	kinds := client.GetKinds()
	kind, ok := kinds[res.Kind]
	if !ok {
		return resource.Resource{}, fmt.Errorf("kind %s not found", res.Kind)
	}
	applyQueryInfo, err := kind.ApplyPath(res)
	if err != nil {
		return resource.Resource{}, err
	}
	url := client.baseURL + applyQueryInfo.Path
	builder := client.client.R().SetBody(res.Json)

	for _, param := range applyQueryInfo.QueryParams {
		builder = builder.SetQueryParam(param.Name, param.Value)
	}

	resp, err := builder.Get(url)
	if err != nil {
		return resource.Resource{}, err
	}

	if resp.IsError() {
		return resource.Resource{}, fmt.Errorf(extractAPIError(resp))
	}

	err = json.Unmarshal(resp.Body(), &results)
	if err != nil {
		return resource.Resource{}, err
	}

	// Find the resource by name from the response list
	for _, element := range results {
		if element.Name == res.Name {
			return element, nil
		}
	}
	return resource.Resource{}, fmt.Errorf("could not find any matching resource")
}
