package storage

import (
	"github.com/conduktor/ctl/internal/state/model"
)

type StorageBackendType string

const (
	FileBackend StorageBackendType = "file"
)

type StorageBackend interface {
	Type() StorageBackendType
	SaveState(state *model.State) error
	LoadState() (*model.State, error)
}
