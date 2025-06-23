package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/resource"
	"sync"
	"testing"
)

func TestApplyResourcesInParallel_ParallelAndSequential(t *testing.T) {
	type testResource struct {
		kind string
		name string
	}
	resources := []resource.Resource{
		{Kind: "A", Name: "1"},
		{Kind: "B", Name: "2"},
		{Kind: "C", Name: "3"},
	}

	for _, runInParallel := range []bool{false, true} {
		t.Run(fmt.Sprintf("parallel=%v", runInParallel), func(t *testing.T) {
			var mu sync.Mutex
			var logs []string
			applyCount := 0

			applyFunc := func(r *resource.Resource, dryRun bool) (string, error) {
				mu.Lock()
				applyCount++
				mu.Unlock()
				return "applied", nil
			}
			logFunc := func(res resource.Resource, upsertResult string, err error) {
				mu.Lock()
				logs = append(logs, fmt.Sprintf("%s/%s: %s", res.Kind, res.Name, upsertResult))
				mu.Unlock()
			}

			results := ApplyResources(resources, applyFunc, false, runInParallel, logFunc)

			if len(results) != len(resources) {
				t.Errorf("expected %d results, got %d", len(resources), len(results))
			}
			if applyCount != len(resources) {
				t.Errorf("expected applyFunc to be called %d times, got %d", len(resources), applyCount)
			}
			if len(logs) != len(resources) {
				t.Errorf("expected %d log entries, got %d", len(resources), len(logs))
			}
		})
	}
}
