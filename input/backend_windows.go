//go:build windows
// +build windows

package input

import (
	"errors"
	"fmt"
)

// windowsBackend implements the Backend interface for Windows systems.
// Currently provides stub implementation.
type windowsBackend struct {
	initialized bool
	parser      *SequenceParser
}

// newBackend creates a new platform-specific backend.
// On Windows systems, this returns a Windows backend stub.
func newBackend() Backend {
	return &windowsBackend{
		parser: NewSequenceParser(),
	}
}

// Init initializes the backend.
// Currently returns an error as Windows support is not yet implemented.
func (b *windowsBackend) Init() error {
	if b.initialized {
		return nil
	}

	// TODO: Implement Windows console API support
	// - Save console mode
	// - Set raw mode using SetConsoleMode
	// - Configure input buffer
	return fmt.Errorf("windows backend not yet implemented - contributions welcome")
}

// Restore restores the terminal state.
// Safe to call even if Init failed.
func (b *windowsBackend) Restore() error {
	// Nothing to restore if not initialized
	if !b.initialized {
		return nil
	}

	// TODO: Implement console mode restoration
	return nil
}

// ReadEvent reads a keyboard event.
// Currently returns an error as Windows support is not yet implemented.
func (b *windowsBackend) ReadEvent() (Event, error) {
	// TODO: Implement Windows console input reading
	// - Use ReadConsoleInput
	// - Parse KEY_EVENT_RECORD
	// - Normalize to Event
	return Event{}, errors.New("windows backend not yet implemented")
}
