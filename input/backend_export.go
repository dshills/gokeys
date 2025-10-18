//go:build !windows
// +build !windows

package input

// NewTestBackend creates a backend instance for testing purposes.
// This is exported for use in integration tests.
func NewTestBackend() Backend {
	return newBackend()
}
