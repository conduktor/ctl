package state

import (
	"errors"
	"fmt"
	"os"

	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/internal/state/storage"
)

type StateService struct {
	config  storage.StorageConfig
	backend storage.StorageBackend
}

// RunWithState is a helper function that initializes the state service,
// loads the state, executes the provided function with the state reference,
// and saves the state back if not a dry run.
// function f should accept a pointer to model.State and return an error and NEVER panic or Exit itself (except for fail fast strategy).
func RunWithState(stateCfg storage.StorageConfig, dryrun, debug bool, f func(stateRef *model.State) error) error {
	stateSvc := NewStateService(stateCfg, debug)

	// Load the state
	stateRef, err := stateSvc.LoadState(debug)
	if err != nil {
		// fail fast if state cannot be loaded
		return err
	}

	// Execute the provided function with the loaded state
	runErr := f(stateRef)

	// Save the state
	saveErr := stateSvc.SaveState(stateRef, dryrun, debug)

	// Close the backend if needed
	closeErr := stateSvc.backend.Close()

	// Combine run and save errors if both occurred
	return errors.Join(runErr, saveErr, closeErr)
}

func NewStateService(config storage.StorageConfig, debug bool) *StateService {
	var backend storage.StorageBackend

	// Determine which backend to use based on configuration
	if config.RemoteURI != nil && *config.RemoteURI != "" {
		// Use remote backend if RemoteURI is provided
		remoteBackend, err := storage.NewRemoteFileBackend(*config.RemoteURI, debug)
		if err != nil {
			// If remote backend initialization fails, log error and fallback to local
			fmt.Fprintf(os.Stderr, "Failed to initialize remote backend: %v\nFalling back to local file backend.\n", err)
			backend = storage.NewLocalFileBackend(config.FilePath, debug)
		} else {
			backend = remoteBackend
		}
	} else {
		// Use local file backend by default
		backend = storage.NewLocalFileBackend(config.FilePath, debug)
	}

	return &StateService{
		config:  config,
		backend: backend,
	}
}

func (s *StateService) LoadState(debug bool) (*model.State, error) {
	if !s.config.Enabled {
		if debug {
			fmt.Fprintln(os.Stderr, "State storage is disabled.")
		}
		return model.NewState(), nil
	}

	fmt.Fprintln(os.Stderr, "Loading state from", s.backend.DebugString())
	state, err := s.backend.LoadState(debug)
	if err != nil {
		return nil, NewStateError("could not load state", err)
	}
	return state, nil
}

func (s *StateService) SaveState(state *model.State, dryrun, debug bool) error {
	if !s.config.Enabled {
		if debug {
			fmt.Fprintln(os.Stderr, "State storage is disabled. Not saving state.")
		}
		return nil
	}

	if dryrun {
		if debug {
			fmt.Fprintln(os.Stderr, "Dry run enabled. Not saving state.")
		}
		return nil
	}

	fmt.Fprintln(os.Stderr, "Saving state into", s.backend.DebugString())
	err := s.backend.SaveState(state, debug)
	if err != nil {
		return NewStateError("could not save state", err)
	}
	return nil
}
