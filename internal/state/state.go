package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/conduktor/ctl/pkg/resource"
)

const StateVersion = "v1"
const StateFileName = "state.json"

type State struct {
	StateLocation string          `json:"-"`
	Version       string          `json:"version"`
	LastUpdated   string          `json:"lastUpdated"`
	Resources     []ResourceState `json:"resources"`
}

func StateDefaultLocation() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".conduktor", "ctl", StateFileName)
}

func NewState() *State {
	return &State{
		StateLocation: StateDefaultLocation(),
		Version:       StateVersion,
		LastUpdated:   "",
		Resources:     make([]ResourceState, 0),
	}
}

func LoadStateFromFile(filePath *string) (*State, error) {
	var stateLocation = ""
	if filePath != nil {
		stateLocation = *filePath
	}

	if stateLocation == "" {
		stateLocation = StateDefaultLocation()
	}
	_, err := os.Stat(stateLocation)
	if os.IsNotExist(err) {
		fmt.Println("State file does not exist, creating a new one at", stateLocation)
		return &State{
			StateLocation: stateLocation,
			Version:       StateVersion,
			LastUpdated:   "",
			Resources:     make([]ResourceState, 0),
		}, nil
	}

	data, err := os.ReadFile(stateLocation)
	if err != nil {
		return nil, err
	}
	var state *State
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}
	state.StateLocation = stateLocation
	return state, nil
}

func (s *State) SaveToFile() error {
	s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	filePath := s.StateLocation

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (s *State) AddManagedResource(res resource.Resource) {
	s.Resources = append(s.Resources, NewResourceState(res))
}

func (s *State) GetRemovedResources(activeResources []resource.Resource) []ResourceState {
	removed := make([]ResourceState, 0)
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
			removed = append(removed, stateRes)
		}
	}
	return removed
}

func (s *State) RemoveManagedResource(apiVersion, kind string, metadata *map[string]any) {
	for i, res := range s.Resources {
		if res.Equal(&ResourceState{APIVersion: apiVersion, Kind: kind, Metadata: metadata}) {
			// Remove the resource from the slice keeping order
			s.Resources = append(s.Resources[:i], s.Resources[i+1:]...)
			return
		}
	}
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
