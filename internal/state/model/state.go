package model

import (
	"github.com/conduktor/ctl/pkg/schema"
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
	if !s.IsResourceManaged(res) {
		s.Resources = append(s.Resources, NewResourceState(res))
		s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	}
}

func (s *State) RemoveManagedResource(res resource.Resource) {
	s.RemoveManagedResourceVKM(res.Version, res.Kind, &res.Metadata)
}

func (s *State) RemoveManagedResourceKindName(kind schema.Kind, name string) {
	for i, res := range s.Resources {
		if res.Kind == kind.GetName() {
			if res.Metadata != nil {
				if resName, ok := (*res.Metadata)["name"].(string); ok && resName == name {
					// Remove the resource from the slice keeping order
					s.Resources = append(s.Resources[:i], s.Resources[i+1:]...)
					s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
					return
				}
			}
		}
	}
}

func (s *State) RemoveManagedResourceVKM(apiVersion, kind string, metadata *map[string]any) {
	searchResState := ResourceState{APIVersion: apiVersion, Kind: kind, Metadata: metadata}
	for i, res := range s.Resources {
		if res.Equal(&searchResState) {
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
