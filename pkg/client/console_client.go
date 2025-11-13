package client

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/conduktor/ctl/internal/utils"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
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
	authMethod    AuthMethod
	baseURL       string
	client        *resty.Client
	schemaCatalog *schema.Catalog
}

type APIParameter struct {
	APIKey      string
	BaseURL     string
	Debug       bool
	Key         string
	Cert        string
	Cacert      string
	CdkUser     string
	CdkPassword string
	AuthMode    string
	Insecure    bool
}

func uniformizeBaseURL(baseURL string) string {
	regex := regexp.MustCompile(`(/api)?/?$`)
	return regex.ReplaceAllString(baseURL, "/api")
}

func Make(apiParameter APIParameter) (*Client, error) {
	//apiKey is set later because it's not mandatory for getting the openapi and parsing different kind
	//or to get jwt token
	restyClient := resty.New().SetDebug(apiParameter.Debug).SetHeader("X-CDK-CLIENT", "CLI/"+utils.GetConduktorVersion())

	if apiParameter.BaseURL == "" {
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

	if apiParameter.CdkUser != "" && apiParameter.APIKey != "" {
		return nil, fmt.Errorf("Can't set both CDK_USER and CDK_API_KEY")
	}

	if apiParameter.Cacert != "" {
		restyClient.SetRootCertificate(apiParameter.Cacert)
	}

	result := &Client{
		authMethod:    nil,
		baseURL:       uniformizeBaseURL(apiParameter.BaseURL),
		client:        restyClient,
		schemaCatalog: nil,
	}

	if apiParameter.Insecure {
		result.IgnoreUntrustedCertificate()
	}

	if apiParameter.APIKey != "" {
		result.authMethod = BearerToken{apiParameter.APIKey}
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
		// set auth method in rest client only if defined to avoid failing on cmd help command print
		result.setAuthMethodInRestClient()
	}

	err := result.initKindFromAPI()
	if err != nil {
		if apiParameter.Debug {
			fmt.Fprintf(os.Stderr, "Cannot access the Conduktor API: %s\nUsing offline defaults.\n", err)
		}
		result.schemaCatalog = schema.ConsoleDefaultCatalog()
	}

	return result, nil
}

func MakeFromEnv() (*Client, error) {
	apiParameter := APIParameter{
		BaseURL:     os.Getenv("CDK_BASE_URL"),
		Debug:       utils.CdkDebug(),
		Key:         os.Getenv("CDK_KEY"),
		Cert:        os.Getenv("CDK_CERT"),
		Cacert:      os.Getenv("CDK_CACERT"),
		APIKey:      os.Getenv("CDK_API_KEY"),
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

type Result struct {
	UpsertResult string
	Diff         string
}

func (client *Client) IgnoreUntrustedCertificate() {
	client.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}

func (client *Client) setAuthMethodFromEnvIfNeeded() {
	if client.authMethod == nil {
		authMode := strings.ToLower(os.Getenv("CDK_AUTH_MODE"))
		apiKey := os.Getenv("CDK_API_KEY")

		switch authMode {
		case "external":
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
				client.authMethod = BearerToken{apiKey}
			} else {
				client.authMethod = BasicAuth{user, password}
			}
		case "", "conduktor":
			if apiKey == "" {
				fmt.Fprintln(os.Stderr, "Please set CDK_API_KEY")
				os.Exit(1)
			}

			client.authMethod = BearerToken{apiKey}
		default:
			fmt.Fprintf(os.Stderr, "CDK_AUTH_MODE was: \"%s\". Accepted values are \"conduktor\" or \"external\"\n.", authMode)
			os.Exit(1)
		}

		client.setAuthMethodInRestClient()
	}
}

func (client *Client) setAuthMethodInRestClient() {
	if client.authMethod == nil {
		fmt.Fprintln(os.Stderr, "No authentication method defined. Please set CDK_API_KEY or CDK_USER/CDK_PASSWORD")
		os.Exit(1)
	}
	client.client = client.client.SetHeader("Authorization", client.authMethod.AuthorizationHeader())
}

func (client *Client) SetAPIKey(apiKey string) {
	client.authMethod = BearerToken{apiKey}
	client.setAuthMethodInRestClient()
}

func extractAPIError(resp *resty.Response) string {
	var apiError APIError
	jsonError := json.Unmarshal(resp.Body(), &apiError)
	if jsonError != nil {
		return resp.String()
	} else {
		return apiError.String()
	}
}

//nolint:unused
func (client *Client) publicV1Url() string {
	return client.baseURL + "/public/v1"
}

func (client *Client) ActivateDebug() {
	client.client.SetDebug(true)
}

func (client *Client) Apply(resource *resource.Resource, dryMode bool, diffMode bool) (Result, error) {
	var result = Result{}

	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) Get(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, queryParams map[string]string) ([]resource.Resource, error) {
	var result []resource.Resource
	client.setAuthMethodFromEnvIfNeeded()
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

// GetFromResource fetches the current version of a resource from the server by name and kind.
// It builds the appropriate URL and query parameters based on the resource's kind configuration.
func (client *Client) GetFromResource(res *resource.Resource) (resource.Resource, error) {
	var results []resource.Resource
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) Run(run schema.Run, pathValue []string, queryParams map[string]string, body interface{}) ([]byte, error) {
	if run.BackendType != schema.CONSOLE {
		return nil, fmt.Errorf("Only console backend type is supported by console client")
	}
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) Login(username, password string) (LoginResult, error) {
	url := client.baseURL + "/login"
	resp, err := client.client.R().SetBody(map[string]string{"username": username, "password": password}).Post(url)
	if err != nil {
		return LoginResult{}, err
	} else if resp.IsError() {
		if resp.StatusCode() == 401 {
			return LoginResult{}, fmt.Errorf("Invalid username or password")
		} else {

			return LoginResult{}, fmt.Errorf(extractAPIError(resp))
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

func (client *Client) Delete(kind *schema.Kind, parentPathValue []string, parentQueryValue []string, name string) error {
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) DeleteResource(resource *resource.Resource) error {
	client.setAuthMethodFromEnvIfNeeded()
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

func (client *Client) GetOpenAPI() ([]byte, error) {
	url := client.baseURL + "/public/docs/docs.yaml"
	resp, err := client.client.R().Get(url)
	if err != nil {
		return nil, err
	} else if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	return resp.Body(), nil
}

func (client *Client) initKindFromAPI() error {
	data, err := client.GetOpenAPI()
	if err != nil {
		return fmt.Errorf("Cannot get openapi: %s", err)
	}
	schema, err := schema.NewOpenAPIParser(data)
	if err != nil {
		return fmt.Errorf("Cannot parse openapi: %s", err)
	}
	strict := false
	client.schemaCatalog, err = schema.GetConsoleCatalog(strict)
	if err != nil {
		return fmt.Errorf("Cannot extract schemaCatalog from openapi: %s", err)
	}
	return nil
}

func (client *Client) ListAdminToken() ([]Token, error) {
	result := make([]Token, 0)
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseURL + "/token/v1/admin_tokens"
	resp, err := client.client.R().Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractAPIError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) ListApplicationInstanceToken(applicationInstanceName string) ([]Token, error) {
	result := make([]Token, 0)
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseURL + "/token/v1/application_instance_tokens/" + applicationInstanceName
	resp, err := client.client.R().Get(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractAPIError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) CreateAdminToken(name string) (CreatedToken, error) {
	var result CreatedToken
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseURL + "/token/v1/admin_tokens"
	resp, err := client.client.R().SetBody(map[string]string{"name": name}).Post(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractAPIError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) CreateApplicationInstanceToken(applicationInstanceName, name string) (CreatedToken, error) {
	var result CreatedToken
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseURL + "/token/v1/application_instance_tokens/" + applicationInstanceName
	resp, err := client.client.R().SetBody(map[string]string{"name": name}).Post(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractAPIError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

type SQLResult struct {
	Header []string        `json:"header"`
	Row    [][]interface{} `json:"row"`
}

func (client *Client) ExecuteSQL(maxLine int, sql string) (SQLResult, error) {
	var result SQLResult
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseURL + "/public/sql/v1/execute"
	resp, err := client.client.R().SetBody(sql).SetQueryParam("maxLine", fmt.Sprintf("%d", maxLine)).Post(url)
	if err != nil {
		return result, err
	} else if resp.IsError() {
		return result, fmt.Errorf(extractAPIError(resp))
	} else {
		err = json.Unmarshal(resp.Body(), &result)
		return result, err
	}
}

func (client *Client) DeleteToken(uuid string) error {
	client.setAuthMethodFromEnvIfNeeded()
	url := client.baseURL + "/token/v1/" + uuid
	resp, err := client.client.R().Delete(url)
	if err != nil {
		return err
	} else if resp.IsError() {
		return fmt.Errorf(extractAPIError(resp))
	}
	return nil
}

func (client *Client) GetKinds() schema.KindCatalog {
	if client.schemaCatalog == nil {
		return map[string]schema.Kind{}
	}
	return client.schemaCatalog.Kind
}

func (client *Client) GetCatalog() *schema.Catalog {
	return client.schemaCatalog
}
