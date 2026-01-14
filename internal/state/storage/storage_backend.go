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
	LoadState(debug bool) (*model.State, error)
	SaveState(state *model.State, debug bool) error
	DebugString() string
}
