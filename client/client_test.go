package client

import (
	"reflect"
	"testing"

	"github.com/conduktor/ctl/resource"
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
		finalUrl := uniformizeBaseUrl(url)
		if finalUrl != expectedResult {
			t.Errorf("When uniformize %s got %s expected %s", url, finalUrl, expectedResult)
		}
	}
}
func TestApplyShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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

	body, err := client.Apply(&topic, false)
	if err != nil {
		t.Error(err)
	}
	if body != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestApplyShouldWorkWithExternalAuthMode(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	user := "user"
	password := "password"
	client, err := Make(ApiParameter{
		BaseUrl:     baseUrl,
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

	body, err := client.Apply(&topic, false)
	if err != nil {
		t.Error(err)
	}
	if body != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestApplyWithDryModeShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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

	body, err := client.Apply(&topic, true)
	if err != nil {
		t.Error(err)
	}
	if body != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestApplyShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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

	_, err = client.Apply(&topic, false)
	if err == nil {
		t.Failed()
	}
}

func TestGetShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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
	result, err := client.Get(&app, []string{}, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result[0].Json, aResource.Json) {
		t.Errorf("Bad result expected %v got: %v", aResource, result)
	}
}

func TestGetShouldFailIfN2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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
	_, err = client.Get(&app, []string{}, nil)
	if err == nil {
		t.Failed()
	}
}

func TestDescribeShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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
	result, err := client.Describe(&app, []string{}, "yo")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result.Json, aResource.Json) {
		t.Errorf("Bad result expected %v got: %v", aResource, result)

	}
}

func TestDescribeShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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
	_, err = client.Describe(&app, []string{}, "yo")
	if err == nil {
		t.Failed()
	}
}

func TestDeleteShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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
	err = client.Delete(&app, []string{}, "yo")
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteResourceShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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
func TestDeleteShouldFailOnNot2XX(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	apiKey := "aToken"
	client, err := Make(ApiParameter{
		ApiKey:  apiKey,
		BaseUrl: baseUrl,
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
	err = client.Delete(&app, []string{}, "yo")
	if err == nil {
		t.Fail()
	}
}
