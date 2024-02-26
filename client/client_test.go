package client

import (
	"github.com/conduktor/ctl/resource"
	"github.com/jarcoal/httpmock"
	"testing"
)

func TestApplyShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, `NotChanged`)
	if err != nil {
		panic(err)
	}

	topic := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "topic",
		Name:    "toto",
		Version: "v1",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/topic",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	body, err := client.Apply(&topic)
	if err != nil {
		t.Error(err)
	}
	if body != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestApplyShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(400, "")
	if err != nil {
		panic(err)
	}

	topic := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "topic",
		Name:    "toto",
		Version: "v1",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/topic",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	_, err = client.Apply(&topic)
	if err == nil {
		t.Failed()
	}
}

func TestGetShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/application",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token),
		responder,
	)

	err = client.Get("application")
	if err != nil {
		t.Error(err)
	}
}

func TestGetShouldFailIfN2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(404, "")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/application",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token),
		responder,
	)

	err = client.Get("application")
	if err == nil {
		t.Failed()
	}
}

func TestDescribeShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token),
		responder,
	)

	err = client.Describe("application", "yo")
	if err != nil {
		t.Error(err)
	}
}

func TestDescribeShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(500, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/api/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token),
		responder,
	)

	err = client.Describe("application", "yo")
	if err == nil {
		t.Failed()
	}
}

func TestDeleteShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/api/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token),
		responder,
	)

	err = client.Delete("application", "yo")
	if err != nil {
		t.Error(err)
	}
}
func TestDeleteShouldFailOnNot2XX(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl, false)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(404, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/api/application/yo",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer "+token),
		responder,
	)

	err = client.Delete("application", "yo")
	if err == nil {
		t.Fail()
	}
}
