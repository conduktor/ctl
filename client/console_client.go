package client

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/conduktor/ctl/resource"
	"github.com/conduktor/ctl/schema"
	"github.com/conduktor/ctl/utils"
	"github.com/go-resty/resty/v2"
)

type AuthMethod interface {
	AuthorizationHeader() string
}

type BearerToken struct {
	Token string
}

func (t BearerToken) AuthorizationHeader() string {
	return fmt.Sprintf("Bearer %s", t.Token)
}

type BasicAuth struct {
	Username string
	Password string
}

func (t BasicAuth) AuthorizationHeader() string {
	credentials := fmt.Sprintf("%s:%s", t.Username, t.Password)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encodedCredentials)
}

type Client struct {
	authMethod AuthMethod
	baseUrl    string
	client     *resty.Client
	kinds      schema.KindCatalog
}

type ApiParameter struct {
	ApiKey      string
	BaseUrl     string
	Debug       bool
	Key         string
	Cert        string
	Cacert      string
	CdkUser     string
	CdkPassword string
	AuthMode    string
	Insecure    bool
}

func uniformizeBaseUrl(baseUrl string) string {
	regex := regexp.MustCompile(`(/api)?/?$`)
	return regex.ReplaceAllString(baseUrl, "/api")
}

func Make(apiParameter ApiParameter) (*Client, error) {
	//apiKey is set later because it's not mandatory for getting the openapi and parsing different kind
	//or to get jwt token
	restyClient := resty.New().SetDebug(apiParameter.Debug).SetHeader("X-CDK-CLIENT", "CLI/"+utils.GetConduktorVersion())

	if apiParameter.BaseUrl == "" {
		return nil, fmt.Errorf("Please set CDK_BASE_URL")
	}

	if (apiParameter.Key == "" && apiParameter.Cert != "") || (apiParameter.Key != "" && apiParameter.Cert == "") {
		return nil, fmt.Errorf("CDK_KEY and CDK_CERT must be provided together")
	} else if apiParameter.Key != "" && apiParameter.Cert != "" {
		certificate, err := tls.LoadX509KeyPair(apiParameter.Cert, apiParameter.Key)
		restyClient.SetCertificates(certificate)
		if err != nil {
			return nil, err
		}
	}

	if (apiParameter.CdkUser != "" && apiParameter.CdkPassword == "") || (apiParameter.CdkUser == "" && apiParameter.CdkPassword != "") {
		return nil, fmt.Errorf("CDK_USER and CDK_PASSWORD must be provided together")
	}

	if apiParameter.CdkUser != "" && apiParameter.ApiKey != "" {
		return nil, fmt.Errorf("Can't set both CDK_USER and CDK_API_KEY")
	}

	if apiParameter.Cacert != "" {
		restyClient.SetRootCertificate(apiParameter.Cacert)
	}

	result := &Client{
		authMethod: nil,
		baseUrl:    uniformizeBaseUrl(apiParameter.BaseUrl),
		client:     restyClient,
		kinds:      nil,
	}

	if apiParameter.Insecure {
		result.IgnoreUntrustedCertificate()
	}

	if apiParameter.ApiKey != "" {
		result.authMethod = BearerToken{apiParameter.ApiKey}
	}

	if apiParameter.CdkUser != "" {
		if strings.ToLower(apiParameter.AuthMode) == "external" {
			result.authMethod = BasicAuth{apiParameter.CdkUser, apiParameter.CdkPassword}
		} else if apiParameter.AuthMode == "" || strings.ToLower(apiParameter.AuthMode) == "conduktor" {
			jwtToken, err := result.Login(apiParameter.CdkUser, apiParameter.CdkPassword)
			if err != nil {
				return nil, fmt.Errorf("Could not login: %s", err)
			}
			bearer := BearerToken{jwtToken.AccessToken}
			result.authMethod = &bearer
		} else {
			return nil, fmt.Errorf("CDK_AUTH_MODE was: \"%s\". Accepted values are \"conduktor\" or \"external\".", apiParameter.AuthMode)
		}
	}

	if result.authMethod != nil {
		result.setAuthMethodInRestClient()
	} else {
		//it will be set later only when really needed
		//so aim is not fail when auth method is not set before printing the cmd help
	}

	err := result.initKindFromApi()
	if err != nil {
		if apiParameter.Debug {
			fmt.Fprintf(os.Stderr, "Cannot access the Conduktor API: %s\nUsing offline defaults.\n", err)
		}
		result.kinds = schema.ConsoleDefaultKind()
	}

	return result, nil
}

func MakeFromEnv() (*Client, error) {
	apiParameter := ApiParameter{
		BaseUrl:     os.Getenv("CDK_BASE_URL"),
		Debug:       utils.CdkDebug(),
		Key:         os.Getenv("CDK_KEY"),
		Cert:        os.Getenv("CDK_CERT"),
		Cacert:      os.Getenv("CDK_CACERT"),
		ApiKey:      os.Getenv("CDK_API_KEY"),
		CdkUser:     os.Getenv("CDK_USER"),
		CdkPassword: os.Getenv("CDK_PASSWORD"),
		AuthMode:    os.Getenv("CDK_AUTH_MODE"),
		Insecure:    strings.ToLower(os.Getenv("CDK_INSECURE")) == "true",
	}

	client, err := Make(apiParameter)
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

func (c *Client) setAuthMethodFromEnvIfNeeded() {
	if c.authMethod == nil {
		authMode := strings.ToLower(os.Getenv("CDK_AUTH_MODE"))
		apiKey := os.Getenv("CDK_API_KEY")

		if authMode == "external" {
			user := os.Getenv("CDK_USER")
			password := os.Getenv("CDK_PASSWORD")

			if apiKey == "" && user == "" {
				fmt.Fprintln(os.Stderr, "Please set CDK_API_KEY or CDK_USER/CDK_PASSWORD")
				os.Exit(1)
			}

			if apiKey != "" && user != "" {
				fmt.Fprintln(os.Stderr, "Can't set both CDK_API_KEY and CDK_USER")
				os.Exit(1)
			}

			if user != "" && password == "" {
				fmt.Fprintln(os.Stderr, "Please set CDK_PASSWORD when using CDK_USER")
				os.Exit(1)
			}

			if apiKey != "" {
				c.authMethod = BearerToken{apiKey}
			} else {
				c.authMethod = BasicAuth{user, password}
			}
		} else if authMode == "" || authMode == "conduktor" {
			if apiKey == "" {
				fmt.Fprintln(os.Stderr, "Please set CDK_API_KEY")
				os.Exit(1)
			}

			c.authMethod = BearerToken{apiKey}
		} else {
			fmt.Fprintf(os.Stderr, "CDK_AUTH_MODE was: \"%s\". Accepted values are \"conduktor\" or \"external\"\n.", authMode)
			os.Exit(1)
		}

		c.setAuthMethodInRestClient()
	}
}

func (c *Client) setAuthMethodInRestClient() {
	if c.authMethod == nil {
		fmt.Fprintln(os.Stderr, "No authentication method defined. Please set CDK_API_KEY or CDK_USER/CDK_PASSWORD")
		os.Exit(1)
	}
	c.client = c.client.SetHeader("Authorization", c.authMethod.AuthorizationHeader())
}

func (c *Client) SetApiKey(apiKey string) {
	c.authMethod = BearerToken{apiKey}
	c.setAuthMethodInRestClient()
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
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) Get(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, queryParams map[string]string) ([]resource.Resource, error) {
	var result []resource.Resource
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) Login(username, password string) (LoginResult, error) {
	url := client.baseUrl + "/login"
	resp, err := client.client.R().SetBody(map[string]string{"username": username, "password": password}).Post(url)
	if err != nil {
		return LoginResult{}, err
	} else if resp.IsError() {
		if resp.StatusCode() == 401 {
			return LoginResult{}, fmt.Errorf("Invalid username or password")
		} else {

			return LoginResult{}, fmt.Errorf(extractApiError(resp))
		}
	}
	result := LoginResult{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return LoginResult{}, err
	}
	return result, nil
}

func (client *Client) Describe(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) (resource.Resource, error) {
	var result resource.Resource
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) Delete(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) error {
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) DeleteResource(resource *resource.Resource) error {
	client.setAuthMethodFromEnvIfNeeded()
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
	client.kinds, err = schema.GetConsoleKinds(strict)
	if err != nil {
		fmt.Errorf("Cannot extract kinds from openapi: %s", err)
	}
	return nil
}

func (client *Client) ListAdminToken() ([]Token, error) {
	result := make([]Token, 0)
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseUrl + "/token/v1/admin_tokens"
	resp, err := client.client.R().Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractApiError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) ListApplicationInstanceToken(applicationInstanceName string) ([]Token, error) {
	result := make([]Token, 0)
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseUrl + "/token/v1/application_instance_tokens/" + applicationInstanceName
	resp, err := client.client.R().Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractApiError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) CreateAdminToken(name string) (CreatedToken, error) {
	var result CreatedToken
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseUrl + "/token/v1/admin_tokens"
	resp, err := client.client.R().SetBody(map[string]string{"name": name}).Post(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractApiError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) CreateApplicationInstanceToken(applicationInstanceName, name string) (CreatedToken, error) {
	var result CreatedToken
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseUrl + "/token/v1/application_instance_tokens/" + applicationInstanceName
	resp, err := client.client.R().SetBody(map[string]string{"name": name}).Post(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractApiError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

type SqlResult struct {
	Header []string        `json:"header"`
	Row    [][]interface{} `json:"row"`
}

func (client *Client) ExecuteSql(maxLine int, sql string) (SqlResult, error) {
	var result SqlResult
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseUrl + "/public/sql/v1/execute"
	resp, err := client.client.R().SetBody(sql).SetQueryParam("maxLine", fmt.Sprintf("%d", maxLine)).Post(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractApiError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) DeleteToken(uuid string) error {
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseUrl + "/token/v1/" + uuid
	resp, err := client.client.R().Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractApiError(resp))
	}
	return nil
}

func (client *Client) GetKinds() schema.KindCatalog {
	return client.kinds
}
