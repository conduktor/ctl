package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Tests for the new server apply flow ===

func TestServerApplyRequest_Serialization(t *testing.T) {
	req := ServerApplyRequest{
		Resources: []ResourceDefinition{
			{OriginalPath: "test.yaml", Content: "kind: Topic\nmetadata:\n  name: test"},
		},
		DryRun:    true,
		PrintDiff: true,
		Strategy:  "continue-on-error",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed ServerApplyRequest
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, req.Resources[0].OriginalPath, parsed.Resources[0].OriginalPath)
	assert.Equal(t, req.Resources[0].Content, parsed.Resources[0].Content)
	assert.Equal(t, req.DryRun, parsed.DryRun)
	assert.Equal(t, req.PrintDiff, parsed.PrintDiff)
	assert.Equal(t, req.Strategy, parsed.Strategy)
}

func TestServerApplyStatusResponse_Deserialization(t *testing.T) {
	// Contract test - this JSON must match Scala server output
	jsonStr := `{
		"token": "abc-123",
		"status": "Completed",
		"results": [
			{"resourceName": "my-topic", "resourceKind": "Topic", "status": "Created", "diff": null, "error": null}
		],
		"error": null,
		"outcome": "Success",
		"totalResources": 1,
		"processedResources": 1,
		"successCount": 1,
		"failureCount": 0
	}`

	var status ServerApplyStatusResponse
	err := json.Unmarshal([]byte(jsonStr), &status)
	require.NoError(t, err)

	assert.Equal(t, "abc-123", status.Token)
	assert.Equal(t, "Completed", status.Status)
	assert.Equal(t, 1, len(status.Results))
	assert.Equal(t, "my-topic", status.Results[0].ResourceName)
	assert.Equal(t, "Topic", status.Results[0].ResourceKind)
	assert.Equal(t, "Created", status.Results[0].Status)
	assert.Nil(t, status.Results[0].Error)
	assert.NotNil(t, status.Outcome)
	assert.Equal(t, "Success", *status.Outcome)
	assert.Equal(t, 1, status.TotalResources)
	assert.Equal(t, 1, status.ProcessedResources)
	assert.Equal(t, 1, status.SuccessCount)
	assert.Equal(t, 0, status.FailureCount)
}

func TestServerApplyStatusResponse_WithError(t *testing.T) {
	errorMsg := "HTTP 500: Internal Server Error"
	jsonStr := fmt.Sprintf(`{
		"token": "abc-123",
		"status": "Completed",
		"results": [
			{"resourceName": "my-topic", "resourceKind": "Topic", "status": "Failed", "diff": null, "error": %q}
		],
		"error": null,
		"outcome": "Failure",
		"totalResources": 1,
		"processedResources": 1,
		"successCount": 0,
		"failureCount": 1
	}`, errorMsg)

	var status ServerApplyStatusResponse
	err := json.Unmarshal([]byte(jsonStr), &status)
	require.NoError(t, err)

	assert.NotNil(t, status.Results[0].Error)
	assert.Equal(t, errorMsg, *status.Results[0].Error)
	assert.Equal(t, 0, status.SuccessCount)
	assert.Equal(t, 1, status.FailureCount)
}

func TestServerApplyStatusResponse_PartialSuccess(t *testing.T) {
	jsonStr := `{
		"token": "abc-123",
		"status": "Completed",
		"results": [
			{"resourceName": "topic1", "resourceKind": "Topic", "status": "Created", "diff": null, "error": null},
			{"resourceName": "topic2", "resourceKind": "Topic", "status": "Failed", "diff": null, "error": "HTTP 400: Bad Request"}
		],
		"error": null,
		"outcome": "PartialSuccess",
		"totalResources": 2,
		"processedResources": 2,
		"successCount": 1,
		"failureCount": 1
	}`

	var status ServerApplyStatusResponse
	err := json.Unmarshal([]byte(jsonStr), &status)
	require.NoError(t, err)

	assert.Equal(t, "PartialSuccess", *status.Outcome)
	assert.Equal(t, 1, status.SuccessCount)
	assert.Equal(t, 1, status.FailureCount)
}

func TestDefaultTimeout_IsReasonable(t *testing.T) {
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.True(t, DefaultTimeout >= 10*time.Second, "timeout should be at least 10 seconds")
	assert.True(t, DefaultTimeout <= 120*time.Second, "timeout should be at most 120 seconds")
}

func TestValidateStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantErr  bool
	}{
		{"empty is valid", "", false},
		{"fail-fast is valid", "fail-fast", false},
		{"continue-on-error is valid", "continue-on-error", false},
		{"invalid strategy", "invalid", true},
		{"typo in strategy", "fail-fsat", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStrategy(tt.strategy)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewServerApplyHandler(t *testing.T) {
	debug := false
	rootCtx := RootContext{
		Catalog: schema.Catalog{
			Kind: schema.KindCatalog{},
		},
		Strict: true,
		Debug:  &debug,
	}

	handler := NewServerApplyHandler(rootCtx)

	assert.NotNil(t, handler)
	assert.Equal(t, rootCtx, handler.rootCtx)
}

func TestServerApplyResultServer_WithRetryCount(t *testing.T) {
	jsonStr := `{
		"resourceName": "my-topic",
		"resourceKind": "Topic",
		"status": "Created",
		"diff": null,
		"error": null,
		"diffTruncated": false,
		"retryCount": 2
	}`

	var result ServerApplyResultServer
	err := json.Unmarshal([]byte(jsonStr), &result)
	require.NoError(t, err)

	assert.Equal(t, 2, result.RetryCount)
	assert.False(t, result.DiffTruncated)
}

func TestServerApplyResultServer_WithTruncatedDiff(t *testing.T) {
	diff := "some diff content"
	jsonStr := fmt.Sprintf(`{
		"resourceName": "my-topic",
		"resourceKind": "Topic",
		"status": "Updated",
		"diff": %q,
		"error": null,
		"diffTruncated": true,
		"retryCount": 0
	}`, diff)

	var result ServerApplyResultServer
	err := json.Unmarshal([]byte(jsonStr), &result)
	require.NoError(t, err)

	assert.NotNil(t, result.Diff)
	assert.Equal(t, diff, *result.Diff)
	assert.True(t, result.DiffTruncated)
}

func TestIndentString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		indent   string
		expected string
	}{
		{
			name:     "single line",
			input:    "hello",
			indent:   "  ",
			expected: "  hello",
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			indent:   "    ",
			expected: "    line1\n    line2\n    line3",
		},
		{
			name:     "empty lines preserved",
			input:    "line1\n\nline3",
			indent:   "  ",
			expected: "  line1\n\n  line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indentString(tt.input, tt.indent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock server for integration-style tests
func TestServerApplyHandler_MockServer(t *testing.T) {
	pollCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/public/v1/resources/batch-apply" && r.Method == "POST" {
			// Initial request - return token
			json.NewEncoder(w).Encode(ServerApplyResponse{Token: "test-token-123"})
			return
		}

		if r.URL.Path == "/public/v1/resources/batch-apply/test-token-123" && r.Method == "GET" {
			pollCount++
			// Return completed status
			outcome := "Success"
			json.NewEncoder(w).Encode(ServerApplyStatusResponse{
				Token:              "test-token-123",
				Status:             "Completed",
				Outcome:            &outcome,
				TotalResources:     1,
				ProcessedResources: 1,
				SuccessCount:       1,
				FailureCount:       0,
				Results: []ServerApplyResultServer{
					{
						ResourceName: "test-resource",
						ResourceKind: "Topic",
						Status:       "Created",
					},
				},
			})
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	// Create a mock client that points to our test server
	restyClient := resty.New().SetBaseURL(server.URL)

	// Verify the server responds correctly
	var resp ServerApplyResponse
	r, err := restyClient.R().
		SetHeader(ApiVersionHeader, ApiVersion).
		SetBody(ServerApplyRequest{
			Resources: []ResourceDefinition{{Content: "test"}},
			Strategy:  "fail-fast",
		}).
		SetResult(&resp).
		Post("/public/v1/resources/batch-apply")

	require.NoError(t, err)
	assert.False(t, r.IsError())
	assert.Equal(t, "test-token-123", resp.Token)

	// Verify polling works
	var status ServerApplyStatusResponse
	r, err = restyClient.R().
		SetHeader(ApiVersionHeader, ApiVersion).
		SetResult(&status).
		Get("/public/v1/resources/batch-apply/test-token-123")

	require.NoError(t, err)
	assert.False(t, r.IsError())
	assert.Equal(t, "Completed", status.Status)
	assert.Equal(t, 1, status.SuccessCount)
}

func TestApiVersionConstants(t *testing.T) {
	assert.Equal(t, "1.0", ApiVersion)
	assert.Equal(t, "X-Conduktor-API-Version", ApiVersionHeader)
}

func TestPollIntervalConstants(t *testing.T) {
	assert.Equal(t, 100*time.Millisecond, InitialPollInterval)
	assert.Equal(t, 2*time.Second, MaxPollInterval)
	assert.True(t, InitialPollInterval < MaxPollInterval, "initial interval should be less than max")
}

// === Additional edge case tests ===

func TestServerApplyRequest_EmptyResources(t *testing.T) {
	req := ServerApplyRequest{
		Resources: []ResourceDefinition{},
		DryRun:    false,
		PrintDiff: false,
		Strategy:  "fail-fast",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed ServerApplyRequest
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Empty(t, parsed.Resources)
}

func TestServerApplyRequest_MultipleResources(t *testing.T) {
	req := ServerApplyRequest{
		Resources: []ResourceDefinition{
			{OriginalPath: "file1.yaml", Content: "kind: Topic\nname: topic1"},
			{OriginalPath: "file2.yaml", Content: "kind: Topic\nname: topic2"},
			{OriginalPath: "file3.yaml", Content: "kind: User\nname: user1"},
		},
		DryRun:    false,
		PrintDiff: true,
		Strategy:  "continue-on-error",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed ServerApplyRequest
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed.Resources, 3)
	assert.Equal(t, "file1.yaml", parsed.Resources[0].OriginalPath)
	assert.Equal(t, "file2.yaml", parsed.Resources[1].OriginalPath)
	assert.Equal(t, "file3.yaml", parsed.Resources[2].OriginalPath)
}

func TestServerApplyStatusResponse_InProgress(t *testing.T) {
	jsonStr := `{
		"token": "abc-123",
		"status": "InProgress",
		"results": [
			{"resourceName": "topic1", "resourceKind": "Topic", "status": "Created", "diff": null, "error": null}
		],
		"error": null,
		"outcome": null,
		"totalResources": 3,
		"processedResources": 1,
		"successCount": 1,
		"failureCount": 0
	}`

	var status ServerApplyStatusResponse
	err := json.Unmarshal([]byte(jsonStr), &status)
	require.NoError(t, err)

	assert.Equal(t, "InProgress", status.Status)
	assert.Nil(t, status.Outcome)
	assert.Equal(t, 3, status.TotalResources)
	assert.Equal(t, 1, status.ProcessedResources)
	assert.Len(t, status.Results, 1)
}

func TestServerApplyStatusResponse_Cancelled(t *testing.T) {
	jsonStr := `{
		"token": "abc-123",
		"status": "Cancelled",
		"results": [
			{"resourceName": "topic1", "resourceKind": "Topic", "status": "Created", "diff": null, "error": null}
		],
		"error": null,
		"outcome": "Cancelled",
		"totalResources": 5,
		"processedResources": 1,
		"successCount": 1,
		"failureCount": 0
	}`

	var status ServerApplyStatusResponse
	err := json.Unmarshal([]byte(jsonStr), &status)
	require.NoError(t, err)

	assert.Equal(t, "Cancelled", status.Status)
	assert.NotNil(t, status.Outcome)
	assert.Equal(t, "Cancelled", *status.Outcome)
}

func TestServerApplyStatusResponse_WithServerError(t *testing.T) {
	serverError := "Connection timeout"
	jsonStr := fmt.Sprintf(`{
		"token": "abc-123",
		"status": "Completed",
		"results": [],
		"error": %q,
		"outcome": "Failure",
		"totalResources": 0,
		"processedResources": 0,
		"successCount": 0,
		"failureCount": 0
	}`, serverError)

	var status ServerApplyStatusResponse
	err := json.Unmarshal([]byte(jsonStr), &status)
	require.NoError(t, err)

	assert.NotNil(t, status.Error)
	assert.Equal(t, serverError, *status.Error)
}

func TestResourceDefinition_Serialization(t *testing.T) {
	def := ResourceDefinition{
		OriginalPath: "/path/to/file.yaml",
		Content:      "apiVersion: v1\nkind: Topic\nmetadata:\n  name: test-topic",
	}

	data, err := json.Marshal(def)
	require.NoError(t, err)

	var parsed ResourceDefinition
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, def.OriginalPath, parsed.OriginalPath)
	assert.Equal(t, def.Content, parsed.Content)
}

func TestServerApplyResponse_Serialization(t *testing.T) {
	resp := ServerApplyResponse{Token: "unique-token-12345"}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed ServerApplyResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "unique-token-12345", parsed.Token)
}

func TestServerApplyHandlerContext_DefaultValues(t *testing.T) {
	ctx := ServerApplyHandlerContext{}

	// Verify zero values
	assert.Empty(t, ctx.FilePaths)
	assert.False(t, ctx.DryRun)
	assert.False(t, ctx.PrintDiff)
	assert.False(t, ctx.RecursiveFolder)
	assert.Empty(t, ctx.Strategy)
	assert.False(t, ctx.NoProgress)
	assert.False(t, ctx.AssumeYes)
}

func TestServerApplyHandlerContext_AllFieldsSet(t *testing.T) {
	ctx := ServerApplyHandlerContext{
		FilePaths:       []string{"file1.yaml", "file2.yaml"},
		DryRun:          true,
		PrintDiff:       true,
		RecursiveFolder: true,
		Strategy:        "continue-on-error",
		NoProgress:      true,
		AssumeYes:       true,
	}

	assert.Len(t, ctx.FilePaths, 2)
	assert.True(t, ctx.DryRun)
	assert.True(t, ctx.PrintDiff)
	assert.True(t, ctx.RecursiveFolder)
	assert.Equal(t, "continue-on-error", ctx.Strategy)
	assert.True(t, ctx.NoProgress)
	assert.True(t, ctx.AssumeYes)
}

func TestServerApplyResult_WithError(t *testing.T) {
	result := ServerApplyResult{
		Err: fmt.Errorf("failed to apply resource"),
	}

	assert.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "failed to apply resource")
}

func TestServerApplyResult_Success(t *testing.T) {
	result := ServerApplyResult{
		UpsertResult: client.Result{
			UpsertResult: "Created",
			Diff:         "+ added line",
		},
	}

	assert.NoError(t, result.Err)
	assert.Equal(t, "Created", result.UpsertResult.UpsertResult)
	assert.Equal(t, "+ added line", result.UpsertResult.Diff)
}

func TestErrCancelled(t *testing.T) {
	assert.Error(t, ErrCancelled)
	assert.Equal(t, "operation cancelled by user", ErrCancelled.Error())
}

// Test mock server with progress polling
func TestServerApplyHandler_MockServer_WithProgress(t *testing.T) {
	pollCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ApiVersionHeader, ApiVersion)

		if r.URL.Path == "/public/v1/resources/batch-apply" && r.Method == "POST" {
			json.NewEncoder(w).Encode(ServerApplyResponse{Token: "progress-token"})
			return
		}

		if r.URL.Path == "/public/v1/resources/batch-apply/progress-token" && r.Method == "GET" {
			pollCount++
			if pollCount < 3 {
				// Return in-progress status
				json.NewEncoder(w).Encode(ServerApplyStatusResponse{
					Token:              "progress-token",
					Status:             "InProgress",
					TotalResources:     2,
					ProcessedResources: pollCount,
					SuccessCount:       pollCount,
					FailureCount:       0,
					Results: []ServerApplyResultServer{
						{ResourceName: "topic1", ResourceKind: "Topic", Status: "Created"},
					},
				})
			} else {
				// Return completed status
				outcome := "Success"
				json.NewEncoder(w).Encode(ServerApplyStatusResponse{
					Token:              "progress-token",
					Status:             "Completed",
					Outcome:            &outcome,
					TotalResources:     2,
					ProcessedResources: 2,
					SuccessCount:       2,
					FailureCount:       0,
					Results: []ServerApplyResultServer{
						{ResourceName: "topic1", ResourceKind: "Topic", Status: "Created"},
						{ResourceName: "topic2", ResourceKind: "Topic", Status: "Created"},
					},
				})
			}
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	restyClient := resty.New().SetBaseURL(server.URL)

	// Start batch apply
	var resp ServerApplyResponse
	r, err := restyClient.R().
		SetHeader(ApiVersionHeader, ApiVersion).
		SetBody(ServerApplyRequest{
			Resources: []ResourceDefinition{
				{Content: "topic1"},
				{Content: "topic2"},
			},
			Strategy: "fail-fast",
		}).
		SetResult(&resp).
		Post("/public/v1/resources/batch-apply")

	require.NoError(t, err)
	assert.False(t, r.IsError())
	assert.Equal(t, "progress-token", resp.Token)

	// Poll until complete
	var status ServerApplyStatusResponse
	for i := 0; i < 5; i++ {
		r, err = restyClient.R().
			SetHeader(ApiVersionHeader, ApiVersion).
			SetResult(&status).
			Get("/public/v1/resources/batch-apply/progress-token")

		require.NoError(t, err)
		if status.Status == "Completed" {
			break
		}
	}

	assert.Equal(t, "Completed", status.Status)
	assert.Equal(t, 2, status.SuccessCount)
	assert.Len(t, status.Results, 2)
}

// Test mock server with API version mismatch warning
func TestServerApplyHandler_MockServer_VersionMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(ApiVersionHeader, "2.0") // Different version

		json.NewEncoder(w).Encode(ServerApplyResponse{Token: "version-test"})
	}))
	defer server.Close()

	restyClient := resty.New().SetBaseURL(server.URL)

	var resp ServerApplyResponse
	r, err := restyClient.R().
		SetHeader(ApiVersionHeader, ApiVersion).
		SetBody(ServerApplyRequest{Resources: []ResourceDefinition{{Content: "test"}}}).
		SetResult(&resp).
		Post("/public/v1/resources/batch-apply")

	require.NoError(t, err)
	assert.False(t, r.IsError())
	// Verify we can detect version mismatch from header
	assert.Equal(t, "2.0", r.Header().Get(ApiVersionHeader))
	assert.NotEqual(t, ApiVersion, r.Header().Get(ApiVersionHeader))
}

// Test mock server error handling
func TestServerApplyHandler_MockServer_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	restyClient := resty.New().SetBaseURL(server.URL)

	r, err := restyClient.R().
		SetHeader(ApiVersionHeader, ApiVersion).
		SetBody(ServerApplyRequest{Resources: []ResourceDefinition{{Content: "test"}}}).
		Post("/public/v1/resources/batch-apply")

	require.NoError(t, err)
	assert.True(t, r.IsError())
	assert.Equal(t, http.StatusInternalServerError, r.StatusCode())
}

func TestIndentString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		indent   string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			indent:   "  ",
			expected: "",
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			indent:   "  ",
			expected: "\n\n\n",
		},
		{
			name:     "tab indent",
			input:    "line1\nline2",
			indent:   "\t",
			expected: "\tline1\n\tline2",
		},
		{
			name:     "no indent",
			input:    "line1\nline2",
			indent:   "",
			expected: "line1\nline2",
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			indent:   "  ",
			expected: "  line1\n  line2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indentString(tt.input, tt.indent)
			assert.Equal(t, tt.expected, result)
		})
	}
}
