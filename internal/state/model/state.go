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
	identity    func(kind string) []string
}

// IdentifyBy tells resources apart by the given metadata keys of their kind (see schema.Catalog.StateIdentity)
// instead of by all their metadata, so changing a description or labels does not make a resource look removed.
func (s *State) IdentifyBy(identity func(kind string) []string) {
	s.identity = identity
}

func (s *State) same(stored, other *ResourceState) bool {
	if s.identity == nil {
		return stored.Equal(other)
	}
	keys := s.identity(stored.Kind)
	if keys == nil {
		return stored.Equal(other)
	}
	return stored.SameIdentity(other, keys)
}

func NewState() *State {
	return &State{
		Version:     StateFileVersion,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Resources:   make([]ResourceState, 0),
	}
}

func (s *State) AddManagedResource(res resource.Resource) {
	asResState := NewResourceState(res)
	for i := range s.Resources {
		if s.same(&s.Resources[i], &asResState) {
			// The same resource, applied again: keep what it is now.
			s.Resources[i] = asResState
			s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
			return
		}
	}
	s.Resources = append(s.Resources, asResState)
	s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
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
		if s.same(&res, &searchResState) {
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
			if s.same(&stateRes, &currResState) {
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
		if s.same(&res, &asResState) {
			return true
		}
	}
	return false
}
