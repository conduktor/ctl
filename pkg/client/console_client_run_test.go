package client

import (
	"strings"
	"testing"

	"github.com/conduktor/ctl/pkg/schema"
	"github.com/jarcoal/httpmock"
)

func makeMockedClient(t *testing.T) *Client {
	t.Helper()
	client, err := Make(APIParameter{
		APIKey:  "aToken",
		BaseURL: "http://baseUrl",
	})
	if err != nil {
		t.Fatalf("failed creating client: %s", err)
	}
	client.setAuthMethodFromEnvIfNeeded()
	httpmock.ActivateNonDefault(client.client.GetClient())
	return client
}

var connectorOffsetsRun = schema.Run{
	BackendType:    schema.CONSOLE,
	Name:           "connectorAlterOffsets",
	Path:           "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/offsets",
	PathParameter:  []string{"cluster", "connectCluster", "connector-name"},
	QueryParameter: map[string]schema.FlagParameterOption{},
	BodyFields: map[string]schema.FlagParameterOption{
		"offsets": {FlagName: "offsets", Type: "json"},
	},
	Method: "PATCH",
}

// connectorStop / pause / resume / restart are all empty-body PUTs that differ
// only by the trailing path segment.
func TestRunConnectorPutVerbs(t *testing.T) {
	verbs := []struct {
		name    string
		segment string
	}{
		{"connectorStop", "stop"},
		{"connectorPause", "pause"},
		{"connectorResume", "resume"},
		{"connectorRestart", "restart"},
	}
	for _, verb := range verbs {
		t.Run(verb.name, func(t *testing.T) {
			defer httpmock.Reset()
			client := makeMockedClient(t)

			run := schema.Run{
				BackendType:   schema.CONSOLE,
				Name:          verb.name,
				Path:          "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/" + verb.segment,
				PathParameter: []string{"cluster", "connectCluster", "connector-name"},
				Method:        "PUT",
			}
			httpmock.RegisterMatcherResponderWithQuery(
				"PUT",
				"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/connect/my-connect/connector/my-connector/"+verb.segment,
				nil,
				httpmock.HeaderIs("Authorization", "Bearer aToken"),
				httpmock.NewStringResponder(204, ""),
			)

			_, err := client.Run(run, []string{"my-cluster", "my-connect", "my-connector"}, map[string]string{}, nil)
			if err != nil {
				t.Errorf("expected no error, got: %s", err)
			}
		})
	}
}

func TestRunConnectorGetOffsets(t *testing.T) {
	defer httpmock.Reset()
	client := makeMockedClient(t)

	run := schema.Run{
		BackendType:   schema.CONSOLE,
		Name:          "connectorGetOffsets",
		Path:          "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/offsets",
		PathParameter: []string{"cluster", "connectCluster", "connector-name"},
		Method:        "GET",
	}
	httpmock.RegisterResponder(
		"GET",
		"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/connect/my-connect/connector/my-connector/offsets",
		httpmock.NewStringResponder(200, `{"offsets":[]}`),
	)

	body, err := client.Run(run, []string{"my-cluster", "my-connect", "my-connector"}, map[string]string{}, nil)
	if err != nil {
		t.Errorf("expected no error, got: %s", err)
	}
	if string(body) != `{"offsets":[]}` {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestRunConnectorResetOffsets(t *testing.T) {
	defer httpmock.Reset()
	client := makeMockedClient(t)

	run := schema.Run{
		BackendType:   schema.CONSOLE,
		Name:          "connectorResetOffsets",
		Path:          "/public/kafka/v2/cluster/{cluster}/connect/{connectCluster}/connector/{connector-name}/offsets",
		PathParameter: []string{"cluster", "connectCluster", "connector-name"},
		Method:        "DELETE",
	}
	httpmock.RegisterResponder(
		"DELETE",
		"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/connect/my-connect/connector/my-connector/offsets",
		httpmock.NewStringResponder(200, ""),
	)

	_, err := client.Run(run, []string{"my-cluster", "my-connect", "my-connector"}, map[string]string{}, nil)
	if err != nil {
		t.Errorf("expected no error, got: %s", err)
	}
}

func TestRunConnectorAlterOffsetsSendsJSONBody(t *testing.T) {
	defer httpmock.Reset()
	client := makeMockedClient(t)

	// The decoded JSON value the CLI would pass after parsing --offsets.
	body := map[string]interface{}{
		"offsets": []interface{}{
			map[string]interface{}{"partition": map[string]interface{}{"kafka_topic": "t"}, "offset": map[string]interface{}{"kafka_offset": 42}},
		},
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PATCH",
		"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/connect/my-connect/connector/my-connector/offsets",
		nil,
		httpmock.BodyContainsString(`"offsets"`).
			And(httpmock.BodyContainsString(`"kafka_topic":"t"`)),
		httpmock.NewStringResponder(200, ""),
	)

	_, err := client.Run(connectorOffsetsRun, []string{"my-cluster", "my-connect", "my-connector"}, map[string]string{}, body)
	if err != nil {
		t.Errorf("expected no error, got: %s", err)
	}
}

// The API returns 400 when the connector is not STOPPED; the CLI must surface
// the API's message rather than a generic failure.
func TestRunConnectorAlterOffsetsNotStoppedError(t *testing.T) {
	defer httpmock.Reset()
	client := makeMockedClient(t)

	httpmock.RegisterResponder(
		"PATCH",
		"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/connect/my-connect/connector/my-connector/offsets",
		httpmock.NewStringResponder(400, `{"title":"The request is invalid (e.g. the connector is not in STOPPED state)"}`),
	)

	_, err := client.Run(connectorOffsetsRun, []string{"my-cluster", "my-connect", "my-connector"}, map[string]string{}, map[string]interface{}{"offsets": []interface{}{}})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "not in STOPPED state") {
		t.Errorf("expected error to surface the API message, got: %s", err)
	}
}

var topicPathParams = []string{"cluster", "topic-name"}

var topicAddPartitionsRun = schema.Run{
	BackendType:   schema.CONSOLE,
	Name:          "topicAddPartitions",
	Path:          "/public/kafka/v2/cluster/{cluster}/topic/{topic-name}/partitions",
	PathParameter: topicPathParams,
	BodyFields: map[string]schema.FlagParameterOption{
		"partitionCount": {FlagName: "partition-count", Required: true, Type: "integer"},
	},
	Method: "PUT",
}

// topicAddPartitions sends the new partition count in a small JSON object body.
func TestRunTopicAddPartitionsSendsJSONBody(t *testing.T) {
	defer httpmock.Reset()
	client := makeMockedClient(t)

	// The body the CLI assembles from the --partition-count flag.
	body := map[string]interface{}{"partitionCount": 6}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/topic/my-topic/partitions",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer aToken").
			And(httpmock.BodyContainsString(`"partitionCount":6`)),
		httpmock.NewStringResponder(200, ""),
	)

	_, err := client.Run(topicAddPartitionsRun, []string{"my-cluster", "my-topic"}, map[string]string{}, body)
	if err != nil {
		t.Errorf("expected no error, got: %s", err)
	}
}

// The API returns 400 when the requested count is not greater than the current
// one; the CLI must surface the API's message rather than a generic failure.
func TestRunTopicAddPartitionsError(t *testing.T) {
	defer httpmock.Reset()
	client := makeMockedClient(t)

	httpmock.RegisterResponder(
		"PUT",
		"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/topic/my-topic/partitions",
		httpmock.NewStringResponder(400, `{"title":"The number of partitions can only be increased"}`),
	)

	_, err := client.Run(topicAddPartitionsRun, []string{"my-cluster", "my-topic"}, map[string]string{}, map[string]interface{}{"partitionCount": 1})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "can only be increased") {
		t.Errorf("expected error to surface the API message, got: %s", err)
	}
}

// topicEmpty is a body-less PUT. The optional partition is sent as a query param.
func TestRunTopicEmpty(t *testing.T) {
	emptyRun := schema.Run{
		BackendType:    schema.CONSOLE,
		Name:           "topicEmpty",
		Path:           "/public/kafka/v2/cluster/{cluster}/topic/{topic-name}/empty",
		PathParameter:  topicPathParams,
		QueryParameter: map[string]schema.FlagParameterOption{"partition": {FlagName: "partition", Type: "integer"}},
		Method:         "PUT",
	}

	t.Run("whole topic (no partition query)", func(t *testing.T) {
		defer httpmock.Reset()
		client := makeMockedClient(t)

		httpmock.RegisterResponder(
			"PUT",
			"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/topic/my-topic/empty",
			httpmock.NewStringResponder(200, ""),
		)

		_, err := client.Run(emptyRun, []string{"my-cluster", "my-topic"}, map[string]string{}, nil)
		if err != nil {
			t.Errorf("expected no error, got: %s", err)
		}
	})

	t.Run("single partition (partition query set)", func(t *testing.T) {
		defer httpmock.Reset()
		client := makeMockedClient(t)

		httpmock.RegisterResponderWithQuery(
			"PUT",
			"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/topic/my-topic/empty",
			"partition=2",
			httpmock.NewStringResponder(200, ""),
		)

		_, err := client.Run(emptyRun, []string{"my-cluster", "my-topic"}, map[string]string{"partition": "2"}, nil)
		if err != nil {
			t.Errorf("expected no error, got: %s", err)
		}
	})
}

// topicSetLowWatermark sends a per-partition offset map as its JSON body.
func TestRunTopicSetLowWatermarkSendsJSONBody(t *testing.T) {
	defer httpmock.Reset()
	client := makeMockedClient(t)

	setLowWatermarkRun := schema.Run{
		BackendType:   schema.CONSOLE,
		Name:          "topicSetLowWatermark",
		Path:          "/public/kafka/v2/cluster/{cluster}/topic/{topic-name}/low-watermark",
		PathParameter: topicPathParams,
		BodyFields: map[string]schema.FlagParameterOption{
			"offsets": {FlagName: "offsets", Required: true, Type: "json"},
		},
		Method: "PUT",
	}

	// The decoded JSON value the CLI would pass after parsing --offsets.
	body := map[string]interface{}{
		"offsets": map[string]interface{}{"0": 42, "1": 1337},
	}

	httpmock.RegisterMatcherResponderWithQuery(
		"PUT",
		"http://baseUrl/api/public/kafka/v2/cluster/my-cluster/topic/my-topic/low-watermark",
		nil,
		httpmock.HeaderIs("Authorization", "Bearer aToken").
			And(httpmock.BodyContainsString(`"offsets"`)).
			And(httpmock.BodyContainsString(`"0":42`)),
		httpmock.NewStringResponder(200, ""),
	)

	_, err := client.Run(setLowWatermarkRun, []string{"my-cluster", "my-topic"}, map[string]string{}, body)
	if err != nil {
		t.Errorf("expected no error, got: %s", err)
	}
}
