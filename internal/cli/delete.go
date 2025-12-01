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
	IgnoreMissing   bool // don't fail if resource is already missing
	DryRun          bool
	StateEnabled    bool
	StateRef        *model.State
}

type DeleteKindHandlerContext struct {
	Args                 []string
	ParentFlagValue      []*string
	ParentQueryFlagValue []*string
	IgnoreMissing        bool // don't fail if resource is already missing
	DryRun               bool
	StateEnabled         bool
	StateRef             *model.State
}

type DeleteByVClusterAndNameHandlerContext struct {
	Name          string
	VCluster      string
	IgnoreMissing bool // don't fail if resource is already missing
	DryRun        bool
	StateEnabled  bool
	StateRef      *model.State
}

type DeleteInterceptorHandlerContext struct {
	Name          string
	VCluster      string
	Group         string
	Username      string
	IgnoreMissing bool // don't fail if resource is already missing
	DryRun        bool
	StateEnabled  bool
	StateRef      *model.State
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
	ignoreMissing := cmdCtx.IgnoreMissing
	stateRef := cmdCtx.StateRef

	// Load resources from files
	resources, err := LoadResourcesFromFiles(cmdCtx.FilePaths, h.rootCtx.Strict, cmdCtx.RecursiveFolder)
	if err != nil {
		return nil, err
	}

	return h.HandleFromList(resources, stateRef, ignoreMissing, dryRun, debug)
}

func (h *DeleteHandler) HandleFromList(resources []resource.Resource, stateRef *model.State, ignoreMissing, dryRun, debug bool) ([]DeleteResult, error) {
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
					err = gatewayClient.DeleteResourceByName(&res, ignoreMissing)
				} else if isResourceIdentifiedByNameAndVCluster(res) {
					err = gatewayClient.DeleteResourceByNameAndVCluster(&res, ignoreMissing)
				} else if isResourceInterceptor(res) {
					err = gatewayClient.DeleteResourceInterceptors(&res, ignoreMissing)
				}
			} else {
				if h.rootCtx.consoleAPIClientError != nil && h.rootCtx.consoleAPIClient == nil {
					// fail early if client is not initialized
					return results, fmt.Errorf("cannot delete Console API resource %s/%s: %s", res.Kind, res.Name, h.rootCtx.consoleAPIClientError)
				}

				err = h.rootCtx.consoleAPIClient.DeleteResource(&res, ignoreMissing)
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
	debug := *h.rootCtx.Debug
	name := cmdCtx.Args[0]
	parentValue := make([]string, len(cmdCtx.ParentFlagValue))
	parentQueryValue := make([]string, len(cmdCtx.ParentQueryFlagValue))
	for i, v := range cmdCtx.ParentFlagValue {
		parentValue[i] = *v
	}
	for i, v := range cmdCtx.ParentQueryFlagValue {
		parentQueryValue[i] = *v
	}
	if cmdCtx.DryRun {
		fmt.Printf("%s/%s: Deleted (dry-run)\n", kind.GetName(), name)
		return nil
	} else {
		var err error
		if kind.IsGatewayKind() {
			if h.rootCtx.gatewayAPIClientError != nil && h.rootCtx.gatewayAPIClient == nil {
				// fail early if client is not initialized
				return fmt.Errorf("cannot delete Gateway API resource of kind %s: %s", kind.GetName(), h.rootCtx.gatewayAPIClientError)
			}
			err = h.rootCtx.gatewayAPIClient.Delete(&kind, parentValue, parentQueryValue, name)
		} else {
			if h.rootCtx.consoleAPIClientError != nil && h.rootCtx.consoleAPIClient == nil {
				// fail early if client is not initialized
				return fmt.Errorf("cannot delete Console API resource of kind %s: %s", kind.GetName(), h.rootCtx.consoleAPIClientError)
			}
			err = h.rootCtx.consoleAPIClient.Delete(&kind, parentValue, parentQueryValue, name, cmdCtx.IgnoreMissing)
		}

		// Remove successful deletions from state
		if err == nil && cmdCtx.StateRef != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "Remove resource %s/%s from state\n", kind.GetName(), name)
			}
			cmdCtx.StateRef.RemoveManagedResourceKindName(kind, name)
		}
		return err
	}
}

func (h *DeleteHandler) HandleByVClusterAndName(kind schema.Kind, cmdCtx DeleteByVClusterAndNameHandlerContext) error {
	debug := *h.rootCtx.Debug
	bodyParams := make(map[string]string)
	if cmdCtx.Name != "" {
		bodyParams["name"] = cmdCtx.Name
	}
	bodyParams["vCluster"] = cmdCtx.VCluster

	if cmdCtx.DryRun {
		fmt.Printf("%s/%s: Deleted (dry-run)\n", kind.GetName(), cmdCtx.Name)
		return nil
	} else {
		if h.rootCtx.gatewayAPIClientError != nil && h.rootCtx.gatewayAPIClient == nil {
			// fail early if client is not initialized
			return fmt.Errorf("cannot delete Gateway API resource of kind %s: %s", kind.GetName(), h.rootCtx.gatewayAPIClientError)
		}

		err := h.rootCtx.gatewayAPIClient.DeleteKindByNameAndVCluster(&kind, bodyParams, cmdCtx.IgnoreMissing)

		// Remove successful deletions from state
		if err == nil && cmdCtx.StateRef != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "Remove resource %s/%s from state\n", kind.GetName(), cmdCtx.Name)
			}
			scope := make(map[string]any)
			scope["vCluster"] = cmdCtx.VCluster
			metadata := make(map[string]any)
			metadata["name"] = cmdCtx.Name
			metadata["scope"] = scope
			cmdCtx.StateRef.RemoveManagedResourceVKM(kind.GetLatestKindVersion().GetName(), kind.GetName(), &metadata)
		}
		return err
	}
}

func (h *DeleteHandler) HandleInterceptor(kind schema.Kind, cmdCtx DeleteInterceptorHandlerContext) error {
	debug := *h.rootCtx.Debug
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

	if cmdCtx.DryRun {
		fmt.Printf("%s/%s: Deleted (dry-run)\n", kind.GetName(), cmdCtx.Name)
		return nil
	} else {
		if h.rootCtx.gatewayAPIClientError != nil && h.rootCtx.gatewayAPIClient == nil {
			// fail early if client is not initialized
			return fmt.Errorf("cannot delete Gateway API resource of kind %s: %s", kind.GetName(), h.rootCtx.gatewayAPIClientError)
		}

		err := h.rootCtx.gatewayAPIClient.DeleteInterceptor(&kind, cmdCtx.Name, bodyParams, cmdCtx.IgnoreMissing)

		// Remove successful deletions from state
		if err == nil && cmdCtx.StateRef != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "Remove resource %s/%s from state\n", kind.GetName(), cmdCtx.Name)
			}
			metadata := make(map[string]any)
			metadata["name"] = cmdCtx.Name
			metadata["scope"] = bodyParams
			cmdCtx.StateRef.RemoveManagedResourceVKM(kind.GetLatestKindVersion().GetName(), kind.GetName(), &metadata)
		}
		return err
	}
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
