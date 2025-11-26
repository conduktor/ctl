package storage

import (
	"fmt"
)

// StorageError represents a general storage-related error.
type StorageError struct {
	BackendType StorageBackendType
	Cause       error
	Message     string
}

func NewStorageError(backendType StorageBackendType, msg string, cause error) *StorageError {
	return &StorageError{
		BackendType: backendType,
		Message:     msg,
		Cause:       cause,
	}
}

func (e *StorageError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s storage error: %s.\n  Cause: %v", e.BackendType, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s storage error: %s.", e.BackendType, e.Message)
}

func (e *StorageError) Unwrap() error {
	return e.Cause
}
