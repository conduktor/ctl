package cli

import (
	"sync"

	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
)

type ApplyHandlerContext struct {
	FilePaths       []string
	DryRun          bool
	PrintDiff       bool
	RecursiveFolder bool
	MaxParallel     int
}

type ApplyResult struct {
	Resource     resource.Resource
	UpsertResult client.Result
	Err          error
}

type ApplyHandler struct {
	rootCtx RootContext
}

func NewApplyHandler(rootCtx RootContext) *ApplyHandler {
	return &ApplyHandler{
		rootCtx: rootCtx,
	}
}

func (h *ApplyHandler) Handle(cmdCtx ApplyHandlerContext) ([]ApplyResult, error) {
	// Load resources from files
	resources, err := LoadResourcesFromFiles(cmdCtx.FilePaths, h.rootCtx.Strict, cmdCtx.RecursiveFolder)
	if err != nil {
		return nil, err
	}

	// Sort resources for proper apply order
	schema.SortResourcesForApply(h.rootCtx.Catalog.Kind, resources, *h.rootCtx.Debug)

	// Group resources by kind
	kindGroups := make(map[string][]resource.Resource)
	var kindOrder []string
	for _, resrc := range resources {
		if _, exists := kindGroups[resrc.Kind]; !exists {
			kindOrder = append(kindOrder, resrc.Kind)
		}
		kindGroups[resrc.Kind] = append(kindGroups[resrc.Kind], resrc)
	}

	var allResults []ApplyResult

	// Process each kind group
	for _, kind := range kindOrder {
		kindResources := kindGroups[kind]
		if len(kindResources) == 0 {
			continue
		}

		var groupResults []ApplyResult
		if h.rootCtx.Catalog.IsGatewayResource(kindResources[0]) {
			groupResults = h.applyResources(kindResources, h.rootCtx.GatewayAPIClient().Apply, cmdCtx)
		} else {
			groupResults = h.applyResources(kindResources, h.rootCtx.ConsoleAPIClient().Apply, cmdCtx)
		}

		allResults = append(allResults, groupResults...)
	}

	return allResults, nil
}

func (h *ApplyHandler) applyResources(
	resources []resource.Resource,
	applyFunc func(*resource.Resource, bool, bool) (client.Result, error),
	cmdCtx ApplyHandlerContext,
) []ApplyResult {
	results := make([]ApplyResult, len(resources))

	if cmdCtx.MaxParallel > 1 {
		var wg sync.WaitGroup
		sem := make(chan struct{}, cmdCtx.MaxParallel)

		for i, resrc := range resources {
			wg.Add(1)
			sem <- struct{}{} // acquire a slot
			go func(i int, res resource.Resource) {
				defer func() {
					wg.Done()
					<-sem // release the slot
				}()
				upsertResult, err := applyFunc(&res, cmdCtx.DryRun, cmdCtx.PrintDiff)
				results[i] = ApplyResult{
					Resource:     res,
					UpsertResult: upsertResult,
					Err:          err,
				}
			}(i, resrc)
		}
		wg.Wait()
	} else {
		for i, res := range resources {
			upsertResult, err := applyFunc(&res, cmdCtx.DryRun, cmdCtx.PrintDiff)
			results[i] = ApplyResult{
				Resource:     res,
				UpsertResult: upsertResult,
				Err:          err,
			}
		}
	}

	return results
}
