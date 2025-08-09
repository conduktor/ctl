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
	schemaCatalog      *schema.Catalog
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
		schemaCatalog:      nil,
	}

	result.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	result.client.SetDisableWarn(true)
	result.client.SetBasicAuth(apiParameter.CdkGatewayUser, apiParameter.CdkGatewayPassword)

	err := result.initCatalogFromApi()
	if err != nil {
		if apiParameter.Debug {
			fmt.Fprintf(os.Stderr, "Cannot access the Gateway Conduktor API: %s\nUsing offline defaults.\n", err)
		}
		result.schemaCatalog = schema.GatewayDefaultCatalog()
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
	queryInfo := kind.ListPath(parentPathValue, parentQueryValue)
	url := client.baseUrl + queryInfo.Path
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
		return result, fmt.Errorf(extractApiError(resp))
	}
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

func (client *GatewayClient) Describe(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) (resource.Resource, error) {
	var result resource.Resource
	queryInfo := kind.DescribePath(parentPathValue, parentQueryValue, name)
	url := client.baseUrl + queryInfo.Path
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
	url := client.baseUrl + queryInfo.Path
	requestBuilder := client.client.R()
	for _, p := range queryInfo.QueryParams {
		requestBuilder = requestBuilder.SetQueryParam(p.Name, p.Value)
	}
	resp, err := requestBuilder.Delete(url)
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
	requestBuilder := client.client.R()
	kind, ok := kinds[resource.Kind]
	if !ok {
		return fmt.Errorf("kind %s not found", resource.Kind)
	}
	deletePath, queryParams, err := kind.DeletePath(resource)
	if err != nil {
		return err
	}
	url := client.baseUrl + deletePath
	if queryParams != nil {
		requestBuilder = requestBuilder.SetQueryParams(queryParams)
	}
	resp, err := requestBuilder.Delete(url)
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
	var deleteInterceptorPayload *DeleteInterceptorPayload
	if scope == nil {
		deleteInterceptorPayload = nil
	} else {
		deleteInterceptorPayload = &DeleteInterceptorPayload{}
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
	deletePath, _, err := kind.DeletePath(resource)
	if err != nil {
		return err
	}
	url := client.baseUrl + deletePath
	req := client.client.R()
	if deleteInterceptorPayload != nil {
		req = req.SetBody(deleteInterceptorPayload)
	}
	resp, err := req.Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		msg := extractApiError(resp)
		if deleteInterceptorPayload == nil {
			msg += "\nThis error may be caused by a bug in Conduktor Gateway REST api defaults fixed in version 3.11.0.\nAs a quick fix, you can fetch your interceptor to see the exact scope and use this when deleting."
		}
		return fmt.Errorf("%s", msg)
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

func (client *GatewayClient) Run(run schema.Run, pathValue []string, queryParams map[string]string, body interface{}) ([]byte, error) {
	if run.BackendType != schema.GATEWAY {
		return nil, fmt.Errorf("Only console backend type is supported by console client")
	}
	path := run.BuildPath(pathValue)
	url := client.baseUrl + path
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
		return nil, fmt.Errorf(extractApiError(resp))
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
	url := client.baseUrl + applyQueryInfo.Path
	builder := client.client.R().SetBody(resource.Json)
	for _, param := range applyQueryInfo.QueryParams {
		builder = builder.SetQueryParam(param.Name, param.Value)
	}
	if dryMode {
		builder = builder.SetQueryParam("dryMode", "true")
	}
	if diffMode {
		currentRes, err := client.GetFromResource(resource)
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
		return result, fmt.Errorf(extractApiError(resp))
	}
	bodyBytes := resp.Body()
	err = json.Unmarshal(bodyBytes, &result)
	//in case backend format change (not json string anymore). Let not fail the client for that
	if err != nil {
		result.UpsertResult = resp.String()
		return result, nil
	}
	return result, nil
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

func (client *GatewayClient) initCatalogFromApi() error {
	data, err := client.GetOpenApi()
	if err != nil {
		return fmt.Errorf("Cannot get openapi: %s", err)
	}
	schema, err := schema.NewOpenApiParser(data)
	if err != nil {
		return fmt.Errorf("Cannot parse openapi: %s", err)
	}
	strict := false
	client.schemaCatalog, err = schema.GetGatewayCatalog(strict)
	if err != nil {
		fmt.Errorf("Cannot extract schemaCatalog from openapi: %s", err)
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
	url := client.baseUrl + applyQueryInfo.Path
	builder := client.client.R().SetBody(res.Json)

	for _, param := range applyQueryInfo.QueryParams {
		builder = builder.SetQueryParam(param.Name, param.Value)
	}

	resp, err := builder.Get(url)
	if err != nil {
		return resource.Resource{}, err
	}

	if resp.IsError() {
		return resource.Resource{}, fmt.Errorf(extractApiError(resp))
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
