package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/internal/utils"
)

const StateFileName = "cli-state.json"

type LocalFileBackend struct {
	FilePath string
}

func NewLocalFileBackend(filePath *string, debug bool) *LocalFileBackend {
	var stateLocation = stateDefaultLocation()

	if filePath != nil && *filePath != "" {
		stateLocation = *filePath
	}

	return &LocalFileBackend{
		FilePath: stateLocation,
	}
}

func (b LocalFileBackend) Type() StorageBackendType {
	return FileBackend
}

func (b LocalFileBackend) LoadState(debug bool) (*model.State, error) {
	_, err := os.Stat(b.FilePath)
	if os.IsNotExist(err) {
		if debug {
			fmt.Fprintf(os.Stderr, "State file does not exist, creating a new one\n")
		}
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

func (b LocalFileBackend) SaveState(state *model.State, debug bool) error {
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

func (b LocalFileBackend) DebugString() string {
	return fmt.Sprintf("Local File %s", b.FilePath)
}

func stateDefaultLocation() string {
	dataDir, err := utils.GetDataDir()
	if err != nil {
		return filepath.Join(dataDir, ".conduktor", StateFileName)
	}
	return filepath.Join(dataDir, StateFileName)
}
