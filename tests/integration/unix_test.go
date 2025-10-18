//go:build !windows
// +build !windows

package integration_test

import (
	"os"
	"testing"

	"github.com/dshills/gokeys/input"
	"golang.org/x/sys/unix"
)

// TestUnixBackendTerminalStateSaveRestore validates that the Unix backend
// correctly saves and restores terminal state.
//
// This test requires a real terminal (tty). It will be skipped if stdin
// is not a terminal.
func TestUnixBackendTerminalStateSaveRestore(t *testing.T) {
	// Check if we're running with a real terminal
	if !isTerminal() {
		t.Skip("Skipping integration test: not running in a terminal")
	}

	fd := int(os.Stdin.Fd())

	// Get original terminal state
	originalState, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		t.Fatalf("Failed to get original terminal state: %v", err)
	}

	// Create and initialize backend
	b := input.NewTestBackend()

	// Initialize (should enter raw mode)
	if err := b.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify we're in raw mode
	rawState, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		t.Fatalf("Failed to get raw state: %v", err)
	}

	// Check that canonical mode is disabled
	if rawState.Lflag&unix.ICANON != 0 {
		t.Error("Terminal should have ICANON disabled in raw mode")
	}

	// Check that echo is disabled
	if rawState.Lflag&unix.ECHO != 0 {
		t.Error("Terminal should have ECHO disabled in raw mode")
	}

	// Restore terminal
	if err := b.Restore(); err != nil {
		t.Fatalf("Restore() failed: %v", err)
	}

	// Verify terminal state was restored
	restoredState, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		t.Fatalf("Failed to get restored state: %v", err)
	}

	// Compare critical flags
	if restoredState.Lflag != originalState.Lflag {
		t.Errorf("Lflag not restored: got %v, want %v", restoredState.Lflag, originalState.Lflag)
	}

	if restoredState.Iflag != originalState.Iflag {
		t.Errorf("Iflag not restored: got %v, want %v", restoredState.Iflag, originalState.Iflag)
	}

	if restoredState.Oflag != originalState.Oflag {
		t.Errorf("Oflag not restored: got %v, want %v", restoredState.Oflag, originalState.Oflag)
	}

	if restoredState.Cflag != originalState.Cflag {
		t.Errorf("Cflag not restored: got %v, want %v", restoredState.Cflag, originalState.Cflag)
	}
}

// TestUnixBackendIdempotent validates that Init and Restore are idempotent.
func TestUnixBackendIdempotent(t *testing.T) {
	if !isTerminal() {
		t.Skip("Skipping integration test: not running in a terminal")
	}

	b := input.NewTestBackend()

	// Multiple Init calls should be safe
	if err := b.Init(); err != nil {
		t.Fatalf("First Init() failed: %v", err)
	}

	if err := b.Init(); err != nil {
		t.Fatalf("Second Init() failed: %v", err)
	}

	// Multiple Restore calls should be safe
	if err := b.Restore(); err != nil {
		t.Fatalf("First Restore() failed: %v", err)
	}

	if err := b.Restore(); err != nil {
		t.Fatalf("Second Restore() failed: %v", err)
	}

	// Restore without Init should be safe
	b2 := input.NewTestBackend()
	if err := b2.Restore(); err != nil {
		t.Fatalf("Restore() without Init() failed: %v", err)
	}
}

// isTerminal checks if stdin is a terminal.
func isTerminal() bool {
	fd := int(os.Stdin.Fd())
	_, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	return err == nil
}
