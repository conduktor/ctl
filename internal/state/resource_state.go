package state

import (
	"reflect"

	"github.com/conduktor/ctl/pkg/resource"
)

type ResourceState struct {
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   *map[string]any `json:"metadata"`
}

func NewResourceState(res resource.Resource) ResourceState {
	metadata := res.Metadata
	return ResourceState{
		APIVersion: res.Version,
		Kind:       res.Kind,
		Metadata:   &metadata,
	}
}

func (r *ResourceState) Equal(other *ResourceState) bool {
	if r.APIVersion != other.APIVersion || r.Kind != other.Kind {
		return false
	}
	if r.Metadata == nil || other.Metadata == nil {
		return false
	}
	if len(*r.Metadata) != len(*other.Metadata) {
		return false
	}
	for key, val1 := range *r.Metadata {
		if key != "labels" { // Ignore labels in comparison
			val2, exists := (*other.Metadata)[key]
			if !exists || !reflect.DeepEqual(val1, val2) {
				return false
			}
		}
	}
	return true
}
