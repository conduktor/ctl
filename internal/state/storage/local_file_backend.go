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

	fmt.Fprintf(os.Stderr, "Using local file state storage at: %s\n", stateLocation)
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

func (b LocalFileBackend) Type() StorageBackendType {
	return FileBackend
}

func (b LocalFileBackend) SaveState(state *model.State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return NewStorageError(FileBackend, "failed to marshal state to JSON", err)
	}

	err = os.MkdirAll(filepath.Dir(b.FilePath), os.ModePerm)
	if err != nil {
		return NewStorageError(FileBackend, fmt.Sprintf("failed to create directories for %s", b.FilePath), err)
	}

	err = os.WriteFile(b.FilePath, data, 0644)
	if err != nil {
		return NewStorageError(FileBackend, fmt.Sprintf("failed to write state to %s", b.FilePath), err)
	}

	return nil
}

func (b LocalFileBackend) LoadState() (*model.State, error) {
	_, err := os.Stat(b.FilePath)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "State file does not exist, creating a new one")
		return model.NewState(), nil
	}

	data, err := os.ReadFile(b.FilePath)
	if err != nil {
		return nil, NewStorageError(FileBackend, "failed to read state file", err)
	}
	var state *model.State
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, NewStorageError(FileBackend, "failed to unmarshal state JSON", err)
	}
	return state, nil
}
