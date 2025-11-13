package cli

import (
	"strings"

	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
)

type DeleteFileHandlerContext struct {
	FilePaths       []string
	RecursiveFolder bool
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
	// Load resources from files
	resources, err := LoadResourcesFromFiles(cmdCtx.FilePaths, h.rootCtx.Strict, cmdCtx.RecursiveFolder)
	if err != nil {
		return nil, err
	}

	// Sort resources for proper delete order
	schema.SortResourcesForDelete(h.rootCtx.Catalog.Kind, resources, *h.rootCtx.Debug)

	var results []DeleteResult

	// Process each resource
	for _, res := range resources {
		var err error
		if h.rootCtx.Catalog.IsGatewayResource(res) {
			if isResourceIdentifiedByName(res) {
				err = h.rootCtx.GatewayAPIClient().DeleteResourceByName(&res)
			} else if isResourceIdentifiedByNameAndVCluster(res) {
				err = h.rootCtx.GatewayAPIClient().DeleteResourceByNameAndVCluster(&res)
			} else if isResourceInterceptor(res) {
				err = h.rootCtx.GatewayAPIClient().DeleteResourceInterceptors(&res)
			}
		} else {
			err = h.rootCtx.ConsoleAPIClient().DeleteResource(&res)
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
		return h.rootCtx.GatewayAPIClient().Delete(&kind, parentValue, parentQueryValue, cmdCtx.Args[0])
	} else {
		return h.rootCtx.ConsoleAPIClient().Delete(&kind, parentValue, parentQueryValue, cmdCtx.Args[0])
	}
}

func (h *DeleteHandler) HandleByVClusterAndName(kind schema.Kind, cmdCtx DeleteByVClusterAndNameHandlerContext) error {
	bodyParams := make(map[string]string)
	if cmdCtx.Name != "" {
		bodyParams["name"] = cmdCtx.Name
	}
	bodyParams["vCluster"] = cmdCtx.VCluster

	return h.rootCtx.GatewayAPIClient().DeleteKindByNameAndVCluster(&kind, bodyParams)
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

	return h.rootCtx.GatewayAPIClient().DeleteInterceptor(&kind, cmdCtx.Name, bodyParams)
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
