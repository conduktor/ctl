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
		return nil, NewStorageError(FileBackend, "failed to read state file", err, "Ensure that the file exists and is accessible by the CLI.")
	}
	var state *model.State
	err = json.Unmarshal(data, &state)
	if err != nil {
		tip := fmt.Sprintf("The state file may be corrupted or not in the expected format. You can try deleting or backing up the state file located at %s and rerun the command to generate a new state file.", b.FilePath)
		return nil, NewStorageError(FileBackend, "failed to unmarshal state JSON", err, tip)
	}
	return state, nil
}

func (b LocalFileBackend) SaveState(state *model.State, debug bool) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		tip := "Something went wrong while converting the state to JSON format. Check cause error and contact support if the issue persists."
		return NewStorageError(FileBackend, "failed to marshal state to JSON", err, tip)
	}

	err = os.MkdirAll(filepath.Dir(b.FilePath), os.ModePerm)
	if err != nil {
		tip := fmt.Sprintf("Ensure that the directory path to %s is correct and that you have the necessary permissions to create directories there.", b.FilePath)
		return NewStorageError(FileBackend, fmt.Sprintf("failed to create directories for %s", b.FilePath), err, tip)
	}

	err = os.WriteFile(b.FilePath, data, 0644)
	if err != nil {
		tip := fmt.Sprintf("Ensure that the file path %s is correct and that you have the necessary permissions to write to it.", b.FilePath)
		return NewStorageError(FileBackend, "failed to write state", err, tip)
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
