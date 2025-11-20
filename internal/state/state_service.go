package state

import (
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
func RunWithState(stateCfg storage.StorageConfig, dryrun, debug bool, f func(stateRef *model.State)) {
	stateSvc := NewStateService(stateCfg, debug)
	// Load the state
	stateRef, err := stateSvc.LoadState(debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %s\n", err)
		os.Exit(1)
	}

	// Execute the provided function with the loaded state
	f(stateRef)

	// Save the state if not a dry run
	err = stateSvc.SaveState(stateRef, dryrun, debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving state: %s\n", err)
		os.Exit(1)
	}
}

func NewStateService(config storage.StorageConfig, debug bool) *StateService {
	// future backends can be added here
	var backend storage.StorageBackend = storage.NewLocalFileBackend(config.FilePath, debug)

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

	if debug {
		fmt.Fprintln(os.Stderr, "Loading state using backend:", s.backend.DebugString())
	}
	state, err := s.backend.LoadState()
	if err != nil {
		return nil, NewStateError("Could not load state", err)
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

	if debug {
		fmt.Fprintln(os.Stderr, "Saving state using backend:", s.backend.DebugString())
	}

	err := s.backend.SaveState(state)
	if err != nil {
		return NewStateError("Could not save state", err)
	}
	return nil
}
