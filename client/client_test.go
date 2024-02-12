package client

import (
	"github.com/conduktor/ctl/resource"
	"github.com/jarcoal/httpmock"
	"testing"
)

func TestApplyShouldWork(t *testing.T) {
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "")
	if err != nil {
		panic(err)
	}

	topic := resource.Resource{
		Json:       []byte(`{"yolo": "data"}`),
		Kind:       "topic",
		Name:       "toto",
		ApiVersion: "v1",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/topic",
		nil,
		httpmock.HeaderIs("Authentication", "Bearer "+token).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	err = client.Apply(&topic)
	if err != nil {
		t.Error(err)
	}
}

func TestApplyShouldFailIfNo2xx(t *testing.T) {
	baseUrl := "http://baseUrl/api"
	token := "aToken"
	client := Make(token, baseUrl)
	httpmock.ActivateNonDefault(
		client.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(400, "")
	if err != nil {
		panic(err)
	}

	topic := resource.Resource{
		Json:       []byte(`{"yolo": "data"}`),
		Kind:       "topic",
		Name:       "toto",
		ApiVersion: "v1",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/topic",
		nil,
		httpmock.HeaderIs("Authentication", "Bearer "+token).
			And(httpmock.BodyContainsBytes(topic.Json)),
		responder,
	)

	err = client.Apply(&topic)
	if err == nil {
		t.Failed()
	}
}
