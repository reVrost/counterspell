package orion

// This file will be populated as needed during development.
// For now, errors are defined inline where they occur.

import (
	"errors"
)

var (
	// ErrMessageNotFound is returned when a message cannot be found.
	ErrMessageNotFound = errors.New("message not found")

	// ErrSessionNotFound is returned when a session cannot be found.
	ErrSessionNotFound = errors.New("session not found")
)
