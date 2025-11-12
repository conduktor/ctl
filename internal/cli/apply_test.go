package cli

import (
	"fmt"
	"sync"
	"testing"

	"github.com/conduktor/ctl/pkg/client"
	"github.com/conduktor/ctl/pkg/resource"
	"github.com/conduktor/ctl/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestApplyResources_MaxParallel(t *testing.T) {
	resources := []resource.Resource{
		{Kind: "A", Name: "1"},
		{Kind: "B", Name: "2"},
		{Kind: "C", Name: "3"},
	}

	for _, maxParallel := range []int{1, 2, 3, 10} {
		t.Run(fmt.Sprintf("maxParallel=%d", maxParallel), func(t *testing.T) {
			var mu sync.Mutex
			applyCount := 0
			maxConcurrent := 0
			currentConcurrent := 0

			// Create a handler with minimal setup
			debug := false
			handler := &ApplyHandler{rootCtx: RootContext{
				kinds:  schema.KindCatalog{},
				strict: true,
				debug:  &debug,
			}}

			applyFunc := func(r *resource.Resource, dryRun bool, printDiff bool) (client.Result, error) {
				mu.Lock()
				currentConcurrent++
				if currentConcurrent > maxConcurrent {
					maxConcurrent = currentConcurrent
				}
				mu.Unlock()
				// Simulate work
				temp := make(chan struct{})
				go func() {
					temp <- struct{}{}
				}()
				<-temp
				mu.Lock()
				applyCount++
				currentConcurrent--
				mu.Unlock()
				var result client.Result
				result.UpsertResult = "applied"
				return result, nil
			}

			cmdCtx := ApplyHandlerContext{
				DryRun:      false,
				PrintDiff:   false,
				MaxParallel: maxParallel,
			}

			results := handler.applyResources(resources, applyFunc, cmdCtx)

			assert.Equal(t, len(resources), len(results), "expected %d results, got %d", len(resources), len(results))
			assert.Equal(t, len(resources), applyCount, "expected applyFunc to be called %d times, got %d", len(resources), applyCount)

			if maxParallel > 1 {
				assert.LessOrEqual(t, maxConcurrent, maxParallel, "max concurrent goroutines %d exceeded maxParallel %d", maxConcurrent, maxParallel)
			}
			if maxParallel == 1 {
				assert.Equal(t, 1, maxConcurrent, "expected max concurrent goroutines to be 1, got %d", maxConcurrent)
			}
		})
	}
}

func TestNewApplyHandler(t *testing.T) {
	debug := false
	kinds := schema.KindCatalog{}
	rootCtx := RootContext{
		kinds:  kinds,
		strict: true,
		debug:  &debug,
	}

	handler := NewApplyHandler(rootCtx)

	assert.NotNil(t, handler)
	assert.Equal(t, rootCtx, handler.rootCtx)
}

func TestApplyHandler_applyResources_Sequential(t *testing.T) {
	debug := false
	handler := &ApplyHandler{rootCtx: RootContext{
		kinds:  schema.KindCatalog{},
		strict: true,
		debug:  &debug,
	}}

	resources := []resource.Resource{
		{Kind: "A", Name: "1"},
		{Kind: "B", Name: "2"},
	}

	applyFunc := func(r *resource.Resource, dryRun bool, printDiff bool) (client.Result, error) {
		return client.Result{UpsertResult: fmt.Sprintf("applied-%s-%s", r.Kind, r.Name)}, nil
	}

	cmdCtx := ApplyHandlerContext{
		DryRun:      false,
		PrintDiff:   false,
		MaxParallel: 1, // Sequential processing
	}

	results := handler.applyResources(resources, applyFunc, cmdCtx)

	assert.Len(t, results, 2)
	assert.Equal(t, "A", results[0].Resource.Kind)
	assert.Equal(t, "1", results[0].Resource.Name)
	assert.Equal(t, "applied-A-1", results[0].UpsertResult.UpsertResult)
	assert.NoError(t, results[0].Err)

	assert.Equal(t, "B", results[1].Resource.Kind)
	assert.Equal(t, "2", results[1].Resource.Name)
	assert.Equal(t, "applied-B-2", results[1].UpsertResult.UpsertResult)
	assert.NoError(t, results[1].Err)
}
