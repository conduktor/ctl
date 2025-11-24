package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
)

type ApplyHandlerContext struct {
	FilePaths       []string
	RecursiveFolder bool
	DryRun          bool
	PrintDiff       bool
	MaxParallel     int
	StateEnabled    bool
	StateRef        *model.State
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
	debug := *h.rootCtx.Debug
	dryRun := cmdCtx.DryRun
	stateRef := cmdCtx.StateRef

	// Load resources from files
	resources, err := LoadResourcesFromFiles(cmdCtx.FilePaths, h.rootCtx.Strict, cmdCtx.RecursiveFolder)
	if err != nil {
		return nil, err
	}

	// Sort resources for proper apply order
	schema.SortResourcesForApply(h.rootCtx.Catalog.Kind, resources, debug)

	if cmdCtx.StateEnabled && stateRef != nil {
		// Delete missing managed resources
		removedResources := stateRef.GetRemovedResources(resources)
		schema.SortResourcesForDelete(h.rootCtx.Catalog.Kind, removedResources, debug)
		if len(removedResources) > 0 {
			fmt.Fprintln(os.Stderr, "Deleting resources missing from state...")

			deleteHandler := NewDeleteHandler(h.rootCtx)
			deleteResult, err := deleteHandler.HandleFromList(removedResources, stateRef, dryRun, debug)
			if err != nil {
				return nil, fmt.Errorf("error deleting resources missing from state: %s", err)
			}

			deleteSuccess := true
			for _, res := range deleteResult {
				if res.Err != nil {
					deleteSuccess = false
					fmt.Fprintf(os.Stderr, "Could not delete resource %s/%s missing from state: %s\n", res.Resource.Kind, res.Resource.Name, res.Err)
				}
			}
			if !deleteSuccess {
				return nil, fmt.Errorf("one or more errors occurred while deleting resources missing from state")
			}
		}
	}

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

	// Update state and save it enabled
	if cmdCtx.StateEnabled && stateRef != nil {
		for _, result := range allResults {
			if result.Err == nil {
				stateRef.AddManagedResource(result.Resource)
			}
		}
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
