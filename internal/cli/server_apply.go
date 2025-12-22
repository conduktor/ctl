package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/resource"
)

// ServerApplyHandlerContext contains the configuration for a server apply operation
type ServerApplyHandlerContext struct {
	FilePaths       []string
	DryRun          bool
	PrintDiff       bool
	RecursiveFolder bool
	Strategy        string // "fail-fast" or "continue-on-error"
	NoProgress      bool   // suppress progress output (for CI)
	AssumeYes       bool   // skip safety prompt on large batches
}

// ServerApplyResult represents the result of applying a single resource
type ServerApplyResult struct {
	Resource     resource.Resource
	UpsertResult client.Result
	Err          error
}

// ServerApplyHandler handles server-side apply operations
type ServerApplyHandler struct {
	rootCtx RootContext
}

// NewServerApplyHandler creates a new ServerApplyHandler
func NewServerApplyHandler(rootCtx RootContext) *ServerApplyHandler {
	return &ServerApplyHandler{rootCtx: rootCtx}
}

// ResourceDefinition represents a resource to be applied
type ResourceDefinition struct {
	OriginalPath string `json:"originalPath"`
	Content      string `json:"content"`
}

// ServerApplyRequest is the request body for the server apply endpoint
type ServerApplyRequest struct {
	Resources []ResourceDefinition `json:"resources"`
	DryRun    bool                 `json:"dryRun"`
	PrintDiff bool                 `json:"printDiff"`
	Strategy  string               `json:"strategy"`
}

// ServerApplyResponse is the initial response from the server apply endpoint
type ServerApplyResponse struct {
	Token string `json:"token"`
}

// ServerApplyStatusResponse is the response from polling the server apply status
type ServerApplyStatusResponse struct {
	Token              string                    `json:"token"`
	Status             string                    `json:"status"`
	Results            []ServerApplyResultServer `json:"results"`
	Error              *string                   `json:"error"`
	Outcome            *string                   `json:"outcome"`
	TotalResources     int                       `json:"totalResources"`
	ProcessedResources int                       `json:"processedResources"`
	SuccessCount       int                       `json:"successCount"`
	FailureCount       int                       `json:"failureCount"`
}

// ServerApplyResultServer represents a single result from the server
type ServerApplyResultServer struct {
	ResourceName  string  `json:"resourceName"`
	ResourceKind  string  `json:"resourceKind"`
	Status        string  `json:"status"`
	Diff          *string `json:"diff"`
	Error         *string `json:"error"`
	DiffTruncated bool    `json:"diffTruncated"`
	RetryCount    int     `json:"retryCount"`
}

// ErrCancelled is returned when the operation is cancelled by the user
var ErrCancelled = fmt.Errorf("operation cancelled by user")

// Handle executes the server apply operation
func (h *ServerApplyHandler) Handle(cmdCtx ServerApplyHandlerContext) ([]ServerApplyResult, error) {
	resources, err := LoadResourcesFromFiles(cmdCtx.FilePaths, h.rootCtx.Strict, cmdCtx.RecursiveFolder)
	if err != nil {
		return nil, err
	}

	if len(resources) == 0 {
		return []ServerApplyResult{}, nil
	}

	var defs []ResourceDefinition
	for _, res := range resources {
		defs = append(defs, ResourceDefinition{
			OriginalPath: res.FilePath,
			Content:      string(res.Json),
		})
	}

	if !cmdCtx.AssumeYes && len(defs) > 50 {
		return nil, fmt.Errorf("refusing to apply %d resources without --yes; re-run with --yes to proceed", len(defs))
	}

	apiClient := h.rootCtx.ConsoleAPIClient()
	httpClient := apiClient.Resty()
	baseURL := apiClient.BaseURL()

	strategy := cmdCtx.Strategy
	if strategy == "" {
		strategy = "fail-fast"
	}

	if err := ValidateStrategy(strategy); err != nil {
		return nil, err
	}

	req := ServerApplyRequest{
		Resources: defs,
		DryRun:    cmdCtx.DryRun,
		PrintDiff: cmdCtx.PrintDiff,
		Strategy:  strategy,
	}

	var resp ServerApplyResponse
	postCtx, cancelPost := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelPost()

	// baseURL ends with /api, but batch-apply is at /public/v1/resources/batch-apply
	// So we need to strip /api and use /public/v1/resources/batch-apply
	batchApplyURL := strings.TrimSuffix(baseURL, "/api") + "/public/v1/resources/batch-apply"

	r, err := httpClient.R().
		SetContext(postCtx).
		SetHeader(ApiVersionHeader, ApiVersion).
		SetBody(req).
		SetResult(&resp).
		Post(batchApplyURL)

	if err != nil {
		return nil, err
	}

	if serverVersion := r.Header().Get(ApiVersionHeader); serverVersion != "" && serverVersion != ApiVersion {
		fmt.Fprintf(os.Stderr, "Warning: Server API version (%s) differs from CLI version (%s). Consider upgrading.\n", serverVersion, ApiVersion)
	}

	if r.IsError() {
		if strings.Contains(r.String(), "Unsupported API version") {
			return nil, fmt.Errorf("API version mismatch: %s\nPlease upgrade your CLI to a compatible version", r.String())
		}
		return nil, fmt.Errorf("server error: %s", r.String())
	}

	token := resp.Token
	if cmdCtx.DryRun {
		fmt.Printf("Server apply (DRY RUN) started with token: %s\n", token)
	} else {
		fmt.Printf("Server apply started with token: %s\n", token)
	}
	fmt.Printf("Strategy: %s\n", strategy)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var finalResults []ServerApplyResult
	lastProcessed := 0
	cancelled := false

	pollInterval := InitialPollInterval

	for {
		if ctx.Err() != nil && !cancelled {
			cancelled = true
			fmt.Println("\nCancelling server apply...")
			cancelCtx, cancelCancel := context.WithTimeout(context.Background(), CancelTimeout)
			_, _ = httpClient.R().
				SetContext(cancelCtx).
				SetHeader(ApiVersionHeader, ApiVersion).
				Delete(batchApplyURL + "/" + token)
			cancelCancel()
		}

		var status ServerApplyStatusResponse
		pollCtx, cancelPoll := context.WithTimeout(ctx, DefaultTimeout)
		r, err := httpClient.R().
			SetContext(pollCtx).
			SetHeader(ApiVersionHeader, ApiVersion).
			SetResult(&status).
			Get(batchApplyURL + "/" + token)
		cancelPoll()

		if err != nil {
			if cancelled {
				return finalResults, ErrCancelled
			}
			return nil, err
		}
		if r.IsError() {
			if cancelled {
				return finalResults, ErrCancelled
			}
			return nil, fmt.Errorf("server error polling status: %s", r.String())
		}

		if !cmdCtx.NoProgress {
			if status.TotalResources > 0 {
				progress := float64(status.ProcessedResources) / float64(status.TotalResources) * 100
				fmt.Printf("\033[2K\rProgress: %d/%d (%.0f%%) - Status: %s",
					status.ProcessedResources, status.TotalResources, progress, status.Status)

				if status.ProcessedResources > lastProcessed {
					for i := lastProcessed; i < len(status.Results) && i < status.ProcessedResources; i++ {
						res := status.Results[i]
						icon := "+"
						if res.Error != nil {
							icon = "x"
						}
						fmt.Printf("\n  %s %s/%s: %s", icon, res.ResourceKind, res.ResourceName, res.Status)
						if res.RetryCount > 0 {
							fmt.Printf(" (retries: %d)", res.RetryCount)
						}
						if res.Diff != nil && *res.Diff != "" {
							fmt.Printf("\n    Diff:\n%s", indentString(*res.Diff, "      "))
							if res.DiffTruncated {
								fmt.Printf("\n    [diff truncated]")
							}
						}
						if res.Error != nil {
							fmt.Printf(" - %s", *res.Error)
						}
					}
					lastProcessed = status.ProcessedResources
					pollInterval = InitialPollInterval
				}
			} else {
				fmt.Printf("\033[2K\rStatus: %s", status.Status)
			}
		}

		if status.Status == "Completed" || status.Status == "Cancelled" {
			fmt.Println()

			if !cmdCtx.NoProgress {
				if status.Outcome != nil {
					fmt.Printf("\nOutcome: %s\n", *status.Outcome)
				}
				fmt.Printf("Summary: %d succeeded, %d failed out of %d total\n",
					status.SuccessCount, status.FailureCount, status.TotalResources)
			}

			for _, res := range status.Results {
				result := ServerApplyResult{
					Resource: resource.Resource{Kind: res.ResourceKind, Name: res.ResourceName},
				}
				if res.Error != nil {
					result.Err = fmt.Errorf("%s", *res.Error)
				} else {
					result.UpsertResult = client.Result{UpsertResult: res.Status}
					if res.Diff != nil {
						result.UpsertResult.Diff = *res.Diff
					}
				}
				finalResults = append(finalResults, result)
			}

			if status.Status == "Cancelled" {
				return finalResults, ErrCancelled
			}
			if status.Error != nil {
				return finalResults, fmt.Errorf("server apply failed: %s", *status.Error)
			}
			break
		}

		time.Sleep(pollInterval)
		if pollInterval < MaxPollInterval {
			pollInterval = pollInterval * 2
			if pollInterval > MaxPollInterval {
				pollInterval = MaxPollInterval
			}
		}
	}

	return finalResults, nil
}

func indentString(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}
