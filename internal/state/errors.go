package state

import (
	"fmt"
)

// StateError represents a general state-related error.
type StateError struct {
	Cause   error
	Message string
}

func NewStateError(message string, cause error) *StateError {
	return &StateError{
		Message: message,
		Cause:   cause,
	}
}

func (e *StateError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *StateError) Unwrap() error {
	return e.Cause
}
