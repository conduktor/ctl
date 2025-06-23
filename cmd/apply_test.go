package cmd

import (
	"fmt"
	"github.com/conduktor/ctl/resource"
	"sync"
	"testing"
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
			var logs []string
			applyCount := 0
			maxConcurrent := 0
			currentConcurrent := 0

			applyFunc := func(r *resource.Resource, dryRun bool) (string, error) {
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
				return "applied", nil
			}
			logFunc := func(res resource.Resource, upsertResult string, err error) {
				mu.Lock()
				logs = append(logs, fmt.Sprintf("%s/%s: %s", res.Kind, res.Name, upsertResult))
				mu.Unlock()
			}

			results := ApplyResources(resources, applyFunc, logFunc, false, maxParallel)

			if len(results) != len(resources) {
				t.Errorf("expected %d results, got %d", len(resources), len(results))
			}
			if applyCount != len(resources) {
				t.Errorf("expected applyFunc to be called %d times, got %d", len(resources), applyCount)
			}
			if len(logs) != len(resources) {
				t.Errorf("expected %d log entries, got %d", len(resources), len(logs))
			}
			if maxParallel > 1 && maxConcurrent > maxParallel {
				t.Errorf("max concurrent goroutines %d exceeded maxParallel %d", maxConcurrent, maxParallel)
			}
			if maxParallel == 1 && maxConcurrent != 1 {
				t.Errorf("expected max concurrent goroutines to be 1, got %d", maxConcurrent)
			}
		})
	}
}
