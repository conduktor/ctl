package storage

import (
	"github.com/conduktor/ctl/internal/state/model"
)

type StorageBackend interface {
	SaveState(state *model.State) error
	LoadState() (*model.State, error)
}
