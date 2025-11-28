package storage

import (
	"fmt"
)

// StorageError represents a general storage-related error.
type StorageError struct {
	BackendType StorageBackendType
	Cause       error
	Message     string
	Tip         string
}

func NewStorageError(backendType StorageBackendType, msg string, cause error, tip string) *StorageError {
	return &StorageError{
		BackendType: backendType,
		Message:     msg,
		Cause:       cause,
		Tip:         tip,
	}
}

func (e *StorageError) Error() string {
	causeStr := ""
	if e.Cause != nil {
		causeStr = fmt.Sprintf("\n  Cause: %v", e.Cause)
	}
	tipStr := ""
	if e.Tip != "" {
		tipStr = fmt.Sprintf("\n%s", e.Tip)
	}
	return fmt.Sprintf("%s storage error: %s.%s%s", e.BackendType, e.Message, causeStr, tipStr)
}

func (e *StorageError) Unwrap() error {
	return e.Cause
}
