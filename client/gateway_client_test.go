package client

import (
	"reflect"
	"testing"

	"github.com/conduktor/ctl/resource"
	"github.com/jarcoal/httpmock"
)

var aGwResource = resource.Resource{
	Version:  "gateway/v2",
	Kind:     "VClusters",
	Name:     "vcluster1",
	Metadata: map[string]interface{}{"name": "vcluster1"},
	Json:     []byte(`{"apiVersion":"v1","kind":"VClusters","metadata":{"name":"vcluster1"},"spec":{"prefix":"vcluster1_"}}`),
}

func TestUniformizeGwBaseUrl(t *testing.T) {
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

	vCluster := resource.Resource{
		Json:    []byte(`{"": "data"}`),
		Kind:    "VClusters",
		Name:    "vcluster1",
		Version: "gateway/v2",
		Metadata: map[string]interface{}{
			"name": "titi",
		},
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/gateway/v2/vclusters",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsBytes(vCluster.Json)),
		responder,
	)

	body, err := gatewayClient.Apply(&vCluster, false)
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

	vCluster := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "VClusters",
		Name:    "toto",
		Version: "gateway/v2",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/gateway/v2/vclusters",
		"dryMode=true",
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.BodyContainsBytes(vCluster.Json)),
		responder,
	)

	body, err := gatewayClient.Apply(&vCluster, true)
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

	vCluster := resource.Resource{
		Json:    []byte(`{"yolo": "data"}`),
		Kind:    "Application",
		Name:    "toto",
		Version: "v1",
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/gateway/v2/vclusters",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.BodyContainsBytes(vCluster.Json)),
		responder,
	)

	_, err = gatewayClient.Apply(&vCluster, false)
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
	responder, err := httpmock.NewJsonResponder(200, []resource.Resource{aGwResource})
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

	vCluster := gatewayClient.GetKinds()["VClusters"]
	result, err := gatewayClient.Get(&vCluster, []string{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result[0].Json, aGwResource.Json) {
		t.Errorf("Bad result expected %v got: %v", aGwResource, result)
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
