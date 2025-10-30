package client

import (
	"reflect"
	"testing"

	"github.com/conduktor/ctl/internal/schema"

	"github.com/conduktor/ctl/internal/resource"
	"github.com/jarcoal/httpmock"
)

var aResource = resource.Resource{
	Version:  "v1",
	Kind:     "Topic",
	Name:     "abc.myTopic",
	Metadata: map[string]interface{}{"name": "abc.myTopic"},
	Json:     []byte(`{"apiVersion":"v1","kind":"Topic","metadata":{"name":"abc.myTopic"},"spec":{"replicationFactor":1}}`),
}

func TestUniformizeBaseUrl(t *testing.T) {
	validUrls := []string{
		"http://baseUrl/api/",
		"http://baseUrl/api",
		"http://baseUrl/",
		"http://baseUrl",
	}
	expectedResult := "http://baseUrl/api"
	for _, url := range validUrls {
		finalURL := uniformizeBaseURL(url)
		if finalURL != expectedResult {
			t.Errorf("When uniformize %s got %s expected %s", url, finalURL, expectedResult)
		}
	}
}
func TestApplyShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	client.setAuthMethodFromEnvIfNeeded()
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder := httpmock.NewStringResponder(200, `{"upsertResult": "NotChanged"}`)

	topic := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "Topic",
		Name:    "toto",
		Version: "v2",
		Metadata: map[string]interface{}{
			"cluster": "local",
		},
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/public/kafka/v2/cluster/local/topic",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	body, err := client.Apply(&topic, false, false)
	if err != nil {
		t.Error(err)
	}
	if body.UpsertResult != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestApplyShouldWorkWithExternalAuthMode(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	user := "user"
	password := "password"
	client, err := Make(APIParameter{
		BaseURL:     baseURL,
		CdkUser:     user,
		CdkPassword: password,
		AuthMode:    "ExTerNaL",
	})
	if err != nil {
		panic(err)
	}
	client.setAuthMethodFromEnvIfNeeded()
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder := httpmock.NewStringResponder(200, `{"upsertResult": "NotChanged"}`)

	topic := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "Topic",
		Name:    "toto",
		Version: "v2",
		Metadata: map[string]interface{}{
			"cluster": "local",
		},
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/public/kafka/v2/cluster/local/topic",
		nil,
		httpmock.HeaderIs("Authorization", "Basic dXNlcjpwYXNzd29yZA==").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	body, err := client.Apply(&topic, false, false)
	if err != nil {
		t.Error(err)
	}
	if body.UpsertResult != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestRunShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	client.setAuthMethodFromEnvIfNeeded()
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder := httpmock.NewStringResponder(200, `somebody`)

	run := schema.Run{
		Path: "/path/{p1}/{p2}/end",
		Name: "run",
		Doc:  "who care",
		QueryParameter: map[string]schema.FlagParameterOption{
			"q1": {Required: true, Type: "string", FlagName: "q1"},
			"q2": {Required: true, Type: "int", FlagName: "q2"},
		},
		PathParameter: []string{"p1", "p2"},
		BodyFields: map[string]schema.FlagParameterOption{
			"b1": {Required: true, Type: "string", FlagName: "b1"},
			"b2": {Required: true, Type: "integer", FlagName: "b2"},
			"b3": {Required: true, Type: "boolean", FlagName: "b3"},
		},
		Method:      "POST",
		BackendType: schema.CONSOLE,
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"POST",
		"http://baseUrl/api/path/p1val/p2val/end",
		"q1=q1val&q2=42",
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsString(`{"b1":"b1","b2":2,"b3":true}`)),
		responder,
	)

	body, err := client.Run(run, []string{"p1val", "p2val"}, map[string]string{"q1": "q1val", "q2": "42"}, map[string]interface{}{"b1": "b1", "b2": 2, "b3": true})
	if err != nil {
		t.Error(err)
	}
	if string(body) != "somebody" {
		t.Errorf("Bad result expected somebody got: %s", body)
	}
}

func TestApplyShouldWorkWithQueryParams(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	client.setAuthMethodFromEnvIfNeeded()
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder := httpmock.NewStringResponder(200, `{"upsertResult": "NotChanged"}`)

	topic := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "Alert",
		Name:    "my-alert",
		Version: "v3",
		Metadata: map[string]interface{}{
			"appInstance": "my-app",
		},
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/public/monitoring/v3/alert?appInstance=my-app",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	body, err := client.Apply(&topic, false, false)
	if err != nil {
		t.Error(err)
	}
	if body.UpsertResult != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestApplyWithDryModeShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder := httpmock.NewStringResponder(200, `{"upsertResult": "NotChanged"}`)

	topic := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "Application",
		Name:    "toto",
		Version: "v1",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/public/self-serve/v1/application",
		"dryMode=true",
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	body, err := client.Apply(&topic, true, false)
	if err != nil {
		t.Error(err)
	}
	if body.UpsertResult != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestApplyShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(400, "")
	if err != nil {
		panic(err)
	}

	topic := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "Application",
		Name:    "toto",
		Version: "v1",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/public/self-serve/v1/application",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	_, err = client.Apply(&topic, false, false)
	if err == nil {
		t.Failed()
	}
}

func TestGetShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, []resource.Resource{aResource})
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/public/self-serve/v1/application",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	app := client.GetKinds()["Application"]
	result, err := client.Get(&app, []string{}, []string{}, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result[0].Json, aResource.Json) {
		t.Errorf("Bad result expected %v got: %v", aResource, result)
	}
}

func TestGetShouldFailIfN2xx(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(404, "")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/public/self-serve/v1/application",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey),
		responder,
	)

	app := client.GetKinds()["Application"]
	_, err = client.Get(&app, []string{}, []string{}, nil)
	if err == nil {
		t.Failed()
	}
}

func TestDescribeShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, aResource)
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/public/self-serve/v1/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	app := client.GetKinds()["Application"]
	result, err := client.Describe(&app, []string{}, []string{}, "yo")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result.Json, aResource.Json) {
		t.Errorf("Bad result expected %v got: %v", aResource, result)

	}
}

func TestDescribeShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl/api"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(500, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/public/self-serve/v1/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey),
		responder,
	)

	app := client.GetKinds()["Application"]
	_, err = client.Describe(&app, []string{}, []string{}, "yo")
	if err == nil {
		t.Failed()
	}
}

func TestDeleteShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/api/public/self-serve/v1/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	app := client.GetKinds()["Application"]
	err = client.Delete(&app, []string{}, []string{}, "yo")
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteResourceShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/api/public/kafka/v2/cluster/local/topic/toto",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	resource, err := resource.FromYamlByte([]byte(`{"apiVersion":"v2","kind":"Topic","metadata":{"name":"toto","cluster":"local"},"spec":{}}`), true)
	if err != nil {
		t.Error(err)
	}
	err = client.DeleteResource(&resource[0])
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteResourceWhenMetadataContainsQueryParameter(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/api/public/monitoring/v3/alert/alert1",
		"group=admin",
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey).
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	resource, err := resource.FromYamlByte([]byte(`{"apiVersion":"v3","kind":"Alert","metadata":{"name":"alert1","group":"admin"},"spec":{}}`), true)
	if err != nil {
		t.Error(err)
	}
	err = client.DeleteResource(&resource[0])
	if err != nil {
		t.Error(err)
	}
}
func TestDeleteShouldFailOnNot2XX(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(APIParameter{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(404, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/api/public/self_serve/v1/api/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+apiKey),
		responder,
	)

	app := client.GetKinds()["Application"]
	err = client.Delete(&app, []string{}, []string{}, "yo")
	if err == nil {
		t.Fail()
	}
}
