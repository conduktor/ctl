package model

import (
	"time"

	"github.com/conduktor/ctl/pkg/resource"
)

const StateFileVersion = "v1"

type State struct {
	Version     string          `json:"version"`
	LastUpdated string          `json:"lastUpdated"`
	Resources   []ResourceState `json:"resources"`
}

func NewState() *State {
	return &State{
		Version:     StateFileVersion,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Resources:   make([]ResourceState, 0),
	}
}

func (s *State) AddManagedResource(res resource.Resource) {
	s.Resources = append(s.Resources, NewResourceState(res))
	s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
}

func (s *State) RemoveManagedResource(apiVersion, kind string, metadata *map[string]any) {
	for i, res := range s.Resources {
		if res.Equal(&ResourceState{APIVersion: apiVersion, Kind: kind, Metadata: metadata}) {
			// Remove the resource from the slice keeping order
			s.Resources = append(s.Resources[:i], s.Resources[i+1:]...)
			s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
			return
		}
	}
}

func (s *State) GetRemovedResources(activeResources []resource.Resource) []resource.Resource {
	removed := make([]resource.Resource, 0)
	for _, stateRes := range s.Resources {
		found := false
		for _, currRes := range activeResources {
			currResState := NewResourceState(currRes)
			if stateRes.Equal(&currResState) {
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, stateRes.ToResource())
		}
	}
	return removed
}

func (s *State) IsResourceManaged(ressource resource.Resource) bool {
	asResState := NewResourceState(ressource)
	for _, res := range s.Resources {
		if res.Equal(&asResState) {
			return true
		}
	}
	return false
}
