package client

import (
	"reflect"
	"testing"

	"github.com/conduktor/ctl/resource"
	"github.com/jarcoal/httpmock"
)

var aVClusterResource = resource.Resource{
	Version:  "gateway/v2",
	Kind:     "VClusters",
	Name:     "vcluster1",
	Metadata: map[string]interface{}{"name": "vcluster1"},
	Json:     []byte(`{"apiVersion":"v1","kind":"VClusters","metadata":{"name":"vcluster1"},"spec":{"prefix":"vcluster1_"}}`),
}

func TestGwApplyShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	gatewayClient, err := MakeGateaway(GatewayApiParameter{
		BaseUrl:            baseUrl,
		Debug:              false,
		CdkGatewayUser:     "admin",
		CdkGatewayPassword: "conduktor",
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		gatewayClient.client.GetClient(),
	)
	responder := httpmock.NewStringResponder(200, `{"upsertResult": "NotChanged"}`)

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/gateway/v2/vclusters",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsBytes(aVClusterResource.Json)),
		responder,
	)

	body, err := gatewayClient.Apply(&aVClusterResource, false)
	if err != nil {
		t.Error(err)
	}
	if body != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestGwApplyWithDryModeShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	gatewayClient, err := MakeGateaway(GatewayApiParameter{
		BaseUrl:            baseUrl,
		Debug:              false,
		CdkGatewayUser:     "admin",
		CdkGatewayPassword: "conduktor",
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		gatewayClient.client.GetClient(),
	)
	responder := httpmock.NewStringResponder(200, `{"upsertResult": "NotChanged"}`)

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/gateway/v2/vclusters",
		"dryMode=true",
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.BodyContainsBytes(aVClusterResource.Json)),
		responder,
	)

	body, err := gatewayClient.Apply(&aVClusterResource, true)
	if err != nil {
		t.Error(err)
	}
	if body != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestGwApplyShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	gatewayClient, err := MakeGateaway(GatewayApiParameter{
		BaseUrl:            baseUrl,
		Debug:              false,
		CdkGatewayUser:     "admin",
		CdkGatewayPassword: "conduktor",
	})

	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		gatewayClient.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(400, "")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/gateway/v2/vclusters",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.BodyContainsBytes(aVClusterResource.Json)),
		responder,
	)

	_, err = gatewayClient.Apply(&aVClusterResource, false)
	if err == nil {
		t.Failed()
	}
}

func TestGwGetShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	gatewayClient, err := MakeGateaway(GatewayApiParameter{
		BaseUrl:            baseUrl,
		Debug:              false,
		CdkGatewayUser:     "admin",
		CdkGatewayPassword: "conduktor",
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		gatewayClient.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, []resource.Resource{aVClusterResource})
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/gateway/v2/vclusters",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	vClusterKind := gatewayClient.GetKinds()["VClusters"]
	result, err := gatewayClient.Get(&vClusterKind, []string{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result[0].Json, aVClusterResource.Json) {
		t.Errorf("Bad result expected %v got: %v", aVClusterResource, result)
	}
}

func TestGwGetShouldFailIfN2xx(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	gatewayClient, err := MakeGateaway(GatewayApiParameter{
		BaseUrl:            baseUrl,
		Debug:              false,
		CdkGatewayUser:     "admin",
		CdkGatewayPassword: "conduktor",
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		gatewayClient.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(404, "")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"GET",
		"http://baseUrl/gateway/v2/vclusters",
		nil,
		httpmock.HeaderIs("Authorization", "Basic changeme"),
		responder,
	)

	vClusterKind := gatewayClient.GetKinds()["VClusters"]
	_, err = gatewayClient.Get(&vClusterKind, []string{})
	if err == nil {
		t.Failed()
	}
}

func TestGwDeleteShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	gatewayClient, err := MakeGateaway(GatewayApiParameter{
		BaseUrl:            baseUrl,
		Debug:              false,
		CdkGatewayUser:     "admin",
		CdkGatewayPassword: "conduktor",
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		gatewayClient.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(200, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/gateway/v2/vclusters/vcluster1",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	vClusters := gatewayClient.GetKinds()["VClusters"]
	err = gatewayClient.Delete(&vClusters, []string{}, "vcluster1")
	if err != nil {
		t.Error(err)
	}
}

func TestGwDeleteShouldFailOnNot2XX(t *testing.T) {
	defer httpmock.Reset()
	baseUrl := "http://baseUrl"
	gatewayClient, err := MakeGateaway(GatewayApiParameter{
		BaseUrl:            baseUrl,
		Debug:              false,
		CdkGatewayUser:     "admin",
		CdkGatewayPassword: "conduktor",
	})
	if err != nil {
		panic(err)
	}
	httpmock.ActivateNonDefault(
		gatewayClient.client.GetClient(),
	)
	responder, err := httpmock.NewJsonResponder(404, "[]")
	if err != nil {
		panic(err)
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"DELETE",
		"http://baseUrl/gateway/v2/vclusters/vcluster1",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y"),
		responder,
	)

	vClusterKind := gatewayClient.GetKinds()["VClusters"]
	err = gatewayClient.Delete(&vClusterKind, []string{}, "vcluster1")
	if err == nil {
		t.Fail()
	}
}
