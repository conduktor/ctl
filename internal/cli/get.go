package cli

import (
	"fmt"
	"sort"

	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
)

type GetAllHandlerContext struct {
	OnlyGateway *bool
	OnlyConsole *bool
}

type GetKindHandlerContext struct {
	Args                 []string
	ParentFlagValue      []*string
	ParentQueryFlagValue []*string
	QueryParams          map[string]string
}

func GetAllsHandler(rootCtx RootContext, cmdCtx GetAllHandlerContext) ([]resource.Resource, []error) {
	var allResources []resource.Resource
	var allErrors []error

	kindsByName := sortedKeys(rootCtx.Catalog.Kind)
	if rootCtx.gatewayAPIClientError != nil {
		if *rootCtx.Debug || *cmdCtx.OnlyGateway {
			return allResources, []error{fmt.Errorf("Cannot create Gateway client: %s\n", rootCtx.gatewayAPIClientError)}
		}
	}
	if rootCtx.consoleAPIClientError != nil {
		if *rootCtx.Debug || *cmdCtx.OnlyConsole {
			return allResources, []error{fmt.Errorf("Cannot create Console client: %s\n", rootCtx.consoleAPIClientError)}
		}
	}
	for _, key := range kindsByName {
		kind := rootCtx.Catalog.Kind[key]
		// keep only the Kinds where listing is provided TODO fix if config is provided
		if !kind.IsRootKind() {
			continue
		}
		var resources []resource.Resource
		var err error
		if kind.IsGatewayKind() && !*cmdCtx.OnlyConsole && rootCtx.gatewayAPIClientError == nil {
			resources, err = rootCtx.GatewayAPIClient().Get(&kind, []string{}, []string{}, map[string]string{})
		} else if kind.IsConsoleKind() && !*cmdCtx.OnlyGateway && rootCtx.consoleAPIClientError == nil {
			resources, err = rootCtx.ConsoleAPIClient().Get(&kind, []string{}, []string{}, map[string]string{})
		}
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("Error fetching resource %s: %s\n", kind.GetName(), err))
			continue
		}

		allResources = append(allResources, resources...)
	}
	return allResources, allErrors
}

func GetKindHandler(kind schema.Kind, rootCtx RootContext, cmdCtx GetKindHandlerContext) ([]resource.Resource, []error) {

	var result []resource.Resource
	var errors []error

	isGatewayKind := kind.IsGatewayKind()

	parentValue := make([]string, len(cmdCtx.ParentFlagValue))
	parentQueryValue := make([]string, len(cmdCtx.ParentQueryFlagValue))
	queryParams := cmdCtx.QueryParams
	for i, v := range cmdCtx.ParentFlagValue {
		parentValue[i] = *v
	}
	for i, v := range cmdCtx.ParentQueryFlagValue {
		parentQueryValue[i] = *v
	}

	var err error

	if len(cmdCtx.Args) == 0 {
		if isGatewayKind {
			result, err = rootCtx.GatewayAPIClient().Get(&kind, parentValue, parentQueryValue, queryParams)
		} else {
			result, err = rootCtx.ConsoleAPIClient().Get(&kind, parentValue, parentQueryValue, queryParams)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("Error fetching resources: %s", err))
		}
	} else if len(cmdCtx.Args) == 1 {
		var res resource.Resource
		if isGatewayKind {
			res, err = rootCtx.GatewayAPIClient().Describe(&kind, parentValue, parentQueryValue, cmdCtx.Args[0])
		} else {
			res, err = rootCtx.ConsoleAPIClient().Describe(&kind, parentValue, parentQueryValue, cmdCtx.Args[0])
		}
		result = append(result, res)
		if err != nil {
			errors = append(errors, fmt.Errorf("Error describing resources: %s", err))
		}

	}
	return result, errors
}

func sortedKeys(kinds schema.KindCatalog) []string {
	keys := make([]string, 0, len(kinds))
	for key := range kinds {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
