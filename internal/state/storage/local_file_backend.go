package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/conduktor/ctl/internal/state/model"
)

const StateFileName = "state.json"

type LocalFileBackend struct {
	FilePath string
}

func NewLocalFileBackend(filePath *string) *LocalFileBackend {
	var stateLocation = stateDefaultLocation()

	if filePath != nil && *filePath != "" {
		stateLocation = *filePath
	}
	return &LocalFileBackend{
		FilePath: stateLocation,
	}
}

func stateDefaultLocation() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".conduktor", "ctl", StateFileName)
}

func (b LocalFileBackend) SaveState(state *model.State) error {
	filePath := b.FilePath

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (b LocalFileBackend) LoadState() (*model.State, error) {
	var stateLocation = ""
	if b.FilePath != "" {
		stateLocation = b.FilePath
	}

	if stateLocation == "" {
		stateLocation = stateDefaultLocation()
	}
	_, err := os.Stat(stateLocation)
	if os.IsNotExist(err) {
		fmt.Println("State file does not exist, creating a new one at", stateLocation)
		return model.NewState(), nil
	}

	data, err := os.ReadFile(stateLocation)
	if err != nil {
		return nil, err
	}
	var state *model.State
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}
	return state, nil
}
