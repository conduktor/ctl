package client

import (
	"reflect"
	"testing"

	"github.com/conduktor/ctl/internal/schema"

	"github.com/conduktor/ctl/internal/resource"
	"github.com/jarcoal/httpmock"
)

var aVClusterResource = resource.Resource{
	Version:  "gateway/v2",
	Kind:     "VirtualCluster",
	Name:     "vcluster1",
	Metadata: map[string]interface{}{"name": "vcluster1"},
	Json:     []byte(`{"apiVersion":"v1","kind":"VirtualCluster","metadata":{"name":"vcluster1"},"spec":{"prefix":"vcluster1_"}}`),
}

func TestGwApplyShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		"http://baseURL/gateway/v2/virtual-cluster",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsBytes(aVClusterResource.Json)),
		responder,
	)

	body, err := gatewayClient.Apply(&aVClusterResource, false, false)
	if err != nil {
		t.Error(err)
	}
	if body.UpsertResult != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestGwApplyWithDryModeShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		"http://baseURL/gateway/v2/virtual-cluster",
		"dryMode=true",
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.BodyContainsBytes(aVClusterResource.Json)),
		responder,
	)

	body, err := gatewayClient.Apply(&aVClusterResource, true, false)
	if err != nil {
		t.Error(err)
	}
	if body.UpsertResult != "NotChanged" {
		t.Errorf("Bad result expected NotChanged got: %s", body)
	}
}

func TestGwApplyShouldFailIfNo2xx(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		"http://baseURL/gateway/v2/virtual-cluster",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.BodyContainsBytes(aVClusterResource.Json)),
		responder,
	)

	_, err = gatewayClient.Apply(&aVClusterResource, false, false)
	if err == nil {
		t.Failed()
	}
}

func TestGwGetShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		"http://baseURL/gateway/v2/virtual-cluster",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	vClusterKind := gatewayClient.GetKinds()["VirtualCluster"]
	result, err := gatewayClient.Get(&vClusterKind, []string{}, []string{}, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result[0].Json, aVClusterResource.Json) {
		t.Errorf("Bad result expected %v got: %v", aVClusterResource, result)
	}
}

func TestGwGetShouldFailIfN2xx(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		"http://baseURL/gateway/v2/virtual-cluster",
		nil,
		httpmock.HeaderIs("Authorization", "Basic changeme"),
		responder,
	)

	vClusterKind := gatewayClient.GetKinds()["VirtualCluster"]
	_, err = gatewayClient.Get(&vClusterKind, []string{}, []string{}, nil)
	if err == nil {
		t.Failed()
	}
}

func TestGwDeleteShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		"http://baseURL/gateway/v2/virtual-cluster/vcluster1",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")),
		responder,
	)

	vClusters := gatewayClient.GetKinds()["VirtualCluster"]
	err = gatewayClient.Delete(&vClusters, []string{}, []string{}, "vcluster1")
	if err != nil {
		t.Error(err)
	}
}

func TestGwDeleteShouldFailOnNot2XX(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		"http://baseURL/gateway/v2/virtual-cluster/vcluster1",
		nil,
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y"),
		responder,
	)

	vClusterKind := gatewayClient.GetKinds()["VirtualCluster"]
	err = gatewayClient.Delete(&vClusterKind, []string{}, []string{}, "vcluster1")
	if err == nil {
		t.Fail()
	}
}

func TestGwRunShouldWork(t *testing.T) {
	defer httpmock.Reset()
	baseURL := "http://baseURL"
	gatewayClient, err := MakeGateway(GatewayAPIParameter{
		BaseURL:            baseURL,
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
		BackendType: schema.GATEWAY,
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"POST",
		"http://baseURL/path/p1val/p2val/end",
		"q1=q1val&q2=42",
		httpmock.HeaderIs("Authorization", "Basic YWRtaW46Y29uZHVrdG9y").
			And(httpmock.HeaderIs("X-CDK-CLIENT", "CLI/unknown")).
			And(httpmock.BodyContainsString(`{"b1":"b1","b2":2,"b3":true}`)),
		responder,
	)

	body, err := gatewayClient.Run(run, []string{"p1val", "p2val"}, map[string]string{"q1": "q1val", "q2": "42"}, map[string]interface{}{"b1": "b1", "b2": 2, "b3": true})
	if err != nil {
		t.Error(err)
	}
	if string(body) != "somebody" {
		t.Errorf("Bad result expected somebody got: %s", body)
	}
}
