package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
)

type DeleteFileHandlerContext struct {
	FilePaths       []string
	RecursiveFolder bool
	DryRun          bool
	StateEnabled    bool
	StateRef        *model.State
}

type DeleteKindHandlerContext struct {
	Args                 []string
	ParentFlagValue      []*string
	ParentQueryFlagValue []*string
}

type DeleteByVClusterAndNameHandlerContext struct {
	Name     string
	VCluster string
}

type DeleteInterceptorHandlerContext struct {
	Name     string
	VCluster string
	Group    string
	Username string
}

type DeleteResult struct {
	Resource resource.Resource
	Err      error
}

type DeleteHandler struct {
	rootCtx RootContext
}

func NewDeleteHandler(rootCtx RootContext) *DeleteHandler {
	return &DeleteHandler{
		rootCtx: rootCtx,
	}
}

func (h *DeleteHandler) HandleFromFiles(cmdCtx DeleteFileHandlerContext) ([]DeleteResult, error) {
	debug := *h.rootCtx.Debug
	dryRun := cmdCtx.DryRun
	stateRef := cmdCtx.StateRef

	// Load resources from files
	resources, err := LoadResourcesFromFiles(cmdCtx.FilePaths, h.rootCtx.Strict, cmdCtx.RecursiveFolder)
	if err != nil {
		return nil, err
	}

	return h.HandleFromList(resources, stateRef, dryRun, debug)
}

func (h *DeleteHandler) HandleFromList(resources []resource.Resource, stateRef *model.State, dryRun bool, debug bool) ([]DeleteResult, error) {
	// Sort resources for proper delete order
	schema.SortResourcesForDelete(h.rootCtx.Catalog.Kind, resources, *h.rootCtx.Debug)

	var results []DeleteResult

	// Process each resource
	for _, res := range resources {
		var err error
		if dryRun {
			fmt.Printf("%s/%s: Deleted (dry-run)\n", res.Kind, res.Name)
		} else {
			if h.rootCtx.Catalog.IsGatewayResource(res) {
				if h.rootCtx.gatewayAPIClientError != nil && h.rootCtx.gatewayAPIClient == nil {
					// fail early if client is not initialized
					return results, fmt.Errorf("cannot delete Gateway API resource %s/%s: %s", res.Kind, res.Name, h.rootCtx.gatewayAPIClientError)
				}
				gatewayClient := h.rootCtx.gatewayAPIClient

				if isResourceIdentifiedByName(res) {
					err = gatewayClient.DeleteResourceByName(&res)
				} else if isResourceIdentifiedByNameAndVCluster(res) {
					err = gatewayClient.DeleteResourceByNameAndVCluster(&res)
				} else if isResourceInterceptor(res) {
					err = gatewayClient.DeleteResourceInterceptors(&res)
				}
			} else {
				if h.rootCtx.consoleAPIClientError != nil && h.rootCtx.consoleAPIClient == nil {
					// fail early if client is not initialized
					return results, fmt.Errorf("cannot delete Console API resource %s/%s: %s", res.Kind, res.Name, h.rootCtx.consoleAPIClientError)
				}

				err = h.rootCtx.consoleAPIClient.DeleteResource(&res)
			}

			// Remove successful deletions from state
			if err == nil && stateRef != nil {
				if debug {
					fmt.Fprintf(os.Stderr, "Remove resource %s/%s from state\n", res.Kind, res.Name)
				}
				stateRef.RemoveManagedResource(res)
			}
		}

		results = append(results, DeleteResult{
			Resource: res,
			Err:      err,
		})
	}

	return results, nil
}

func (h *DeleteHandler) HandleKind(kind schema.Kind, cmdCtx DeleteKindHandlerContext) error {
	parentValue := make([]string, len(cmdCtx.ParentFlagValue))
	parentQueryValue := make([]string, len(cmdCtx.ParentQueryFlagValue))
	for i, v := range cmdCtx.ParentFlagValue {
		parentValue[i] = *v
	}
	for i, v := range cmdCtx.ParentQueryFlagValue {
		parentQueryValue[i] = *v
	}

	if kind.IsGatewayKind() {
		if h.rootCtx.gatewayAPIClientError != nil && h.rootCtx.gatewayAPIClient == nil {
			// fail early if client is not initialized
			return fmt.Errorf("cannot delete Gateway API resource of kind %s: %s", kind.GetName(), h.rootCtx.gatewayAPIClientError)
		}
		return h.rootCtx.gatewayAPIClient.Delete(&kind, parentValue, parentQueryValue, cmdCtx.Args[0])
	} else {
		if h.rootCtx.consoleAPIClientError != nil && h.rootCtx.consoleAPIClient == nil {
			// fail early if client is not initialized
			return fmt.Errorf("cannot delete Console API resource of kind %s: %s", kind.GetName(), h.rootCtx.consoleAPIClientError)
		}
		return h.rootCtx.consoleAPIClient.Delete(&kind, parentValue, parentQueryValue, cmdCtx.Args[0])
	}
}

func (h *DeleteHandler) HandleByVClusterAndName(kind schema.Kind, cmdCtx DeleteByVClusterAndNameHandlerContext) error {
	bodyParams := make(map[string]string)
	if cmdCtx.Name != "" {
		bodyParams["name"] = cmdCtx.Name
	}
	bodyParams["vCluster"] = cmdCtx.VCluster

	if h.rootCtx.gatewayAPIClientError != nil && h.rootCtx.gatewayAPIClient == nil {
		// fail early if client is not initialized
		return fmt.Errorf("cannot delete Gateway API resource of kind %s: %s", kind.GetName(), h.rootCtx.gatewayAPIClientError)
	}

	return h.rootCtx.gatewayAPIClient.DeleteKindByNameAndVCluster(&kind, bodyParams)
}

func (h *DeleteHandler) HandleInterceptor(kind schema.Kind, cmdCtx DeleteInterceptorHandlerContext) error {
	bodyParams := make(map[string]string)
	if cmdCtx.VCluster != "" {
		bodyParams["vCluster"] = cmdCtx.VCluster
	}
	if cmdCtx.Group != "" {
		bodyParams["group"] = cmdCtx.Group
	}
	if cmdCtx.Username != "" {
		bodyParams["username"] = cmdCtx.Username
	}

	if h.rootCtx.gatewayAPIClientError != nil && h.rootCtx.gatewayAPIClient == nil {
		// fail early if client is not initialized
		return fmt.Errorf("cannot delete Gateway API resource of kind %s: %s", kind.GetName(), h.rootCtx.gatewayAPIClientError)
	}

	return h.rootCtx.gatewayAPIClient.DeleteInterceptor(&kind, cmdCtx.Name, bodyParams)
}

func IsKindIdentifiedByNameAndVCluster(res schema.Kind) bool {
	return isIdentifiedByNameAndVCluster(res.GetName())
}

func IsKindInterceptor(res schema.Kind) bool {
	return isInterceptor(res.GetName())
}

func isResourceIdentifiedByName(res resource.Resource) bool {
	return isIdentifiedByName(res.Kind)
}

func isResourceIdentifiedByNameAndVCluster(res resource.Resource) bool {
	return isIdentifiedByNameAndVCluster(res.Kind)
}

func isIdentifiedByNameAndVCluster(kind string) bool {
	return strings.Contains(strings.ToLower(kind), "aliastopic") ||
		strings.Contains(strings.ToLower(kind), "gatewayserviceaccount") ||
		strings.Contains(strings.ToLower(kind), "concentrationrule")
}

func isIdentifiedByName(kind string) bool {
	return strings.Contains(strings.ToLower(kind), "virtualcluster") ||
		strings.Contains(strings.ToLower(kind), "group")
}

func isResourceInterceptor(res resource.Resource) bool {
	return isInterceptor(res.Kind)
}

func isInterceptor(kind string) bool {
	return strings.Contains(strings.ToLower(kind), "interceptor")
}
