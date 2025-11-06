package state

import (
	"github.com/conduktor/ctl/internal/state/model"
	"github.com/conduktor/ctl/internal/state/storage"
)

type StateService struct {
	backend storage.StorageBackend
}

func NewStateService(config storage.StorageConfig) *StateService {
	// future backends can be added here
	var backend storage.StorageBackend = storage.NewLocalFileBackend(config.StateFilePath)

	return &StateService{
		backend: backend,
	}
}

func (s *StateService) LoadState() (*model.State, error) {
	return s.backend.LoadState()
}

func (s *StateService) SaveState(state *model.State) error {
	return s.backend.SaveState(state)
}
