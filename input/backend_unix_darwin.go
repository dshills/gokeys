//go:build darwin
// +build darwin

package input

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// readBufferPool provides reusable 256-byte buffers for reading terminal input.
// This eliminates per-keypress allocations and reduces garbage collection pressure.
var readBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 256)
		return &b
	},
}

// unixBackend implements the Backend interface for Unix-like systems
// using termios for raw mode terminal control.
type unixBackend struct {
	fd            int
	originalState *unix.Termios
	parser        *SequenceParser
	file          *os.File
	initialized   bool

	// pendingBuf accumulates partial UTF-8 sequences and escape codes across Read() calls.
	// This is critical for handling multi-byte UTF-8 characters and escape sequences that
	// may be split across multiple terminal read operations (e.g., on slow SSH connections).
	pendingBuf []byte
}

// newBackend creates a new platform-specific backend.
// On Unix systems, this returns a Unix backend.
func newBackend() Backend {
	return &unixBackend{
		fd:     int(os.Stdin.Fd()),
		parser: NewSequenceParser(),
		file:   os.Stdin,
	}
}

// Init initializes the backend by saving the current terminal state
// and entering raw mode. This allows reading individual keypresses
// without line buffering or echo.
// Idempotent: calling multiple times is safe and does nothing after first call.
func (b *unixBackend) Init() error {
	// Idempotency check: if already initialized, do nothing
	if b.initialized {
		return nil
	}

	// Get current terminal state
	state, err := unix.IoctlGetTermios(b.fd, unix.TIOCGETA)
	if err != nil {
		return fmt.Errorf("failed to get terminal state: %w", err)
	}

	// Save original state for restoration
	b.originalState = state

	// Create raw mode state
	rawState := *state

	// Disable canonical mode (line buffering)
	rawState.Lflag &^= unix.ICANON

	// Disable echo
	rawState.Lflag &^= unix.ECHO

	// Disable signal generation (Ctrl+C, Ctrl+Z, etc.)
	rawState.Lflag &^= unix.ISIG

	// Disable extended input processing
	rawState.Lflag &^= unix.IEXTEN

	// Disable input parity checking
	rawState.Iflag &^= unix.INPCK

	// Disable stripping 8th bit
	rawState.Iflag &^= unix.ISTRIP

	// Disable CR-to-NL translation
	rawState.Iflag &^= unix.ICRNL

	// Disable output processing
	rawState.Oflag &^= unix.OPOST

	// Set character size to 8 bits
	rawState.Cflag &^= unix.CSIZE
	rawState.Cflag |= unix.CS8

	// Set minimum characters to 1 (blocking read - wait for at least 1 byte)
	rawState.Cc[unix.VMIN] = 1

	// Set timeout to 0 (no inter-byte timeout)
	// We'll use SetReadDeadline for timeouts on subsequent reads
	rawState.Cc[unix.VTIME] = 0

	// Apply raw mode
	if err := unix.IoctlSetTermios(b.fd, unix.TIOCSETA, &rawState); err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}

	// Mark as initialized to ensure idempotency
	b.initialized = true

	return nil
}

// Restore restores the original terminal state.
// This should be called when shutting down to return the terminal
// to its normal operating mode.
func (b *unixBackend) Restore() error {
	if b.originalState == nil {
		// Nothing to restore (Init was never called)
		return nil
	}

	if err := unix.IoctlSetTermios(b.fd, unix.TIOCSETA, b.originalState); err != nil {
		return fmt.Errorf("failed to restore terminal state: %w", err)
	}

	return nil
}

// ReadEvent reads a single event from the terminal.
// It performs blocking reads and handles multi-byte escape sequences.
// With VMIN=1, read() blocks until at least one byte is available.
func (b *unixBackend) ReadEvent() (Event, error) {
	// Get buffer from pool
	bufPtr := readBufferPool.Get().(*[]byte)
	defer readBufferPool.Put(bufPtr)
	buf := *bufPtr

	// Clear any previous read deadline
	_ = b.file.SetReadDeadline(time.Time{})

	// Read first byte (blocking until data available with VMIN=1)
	n, err := b.file.Read(buf)
	if err != nil {
		return Event{}, err
	}

	// CRITICAL: Copy data to persistent buffer before returning pooled buffer
	b.pendingBuf = append(b.pendingBuf, buf[:n]...)

	// Helper function for non-blocking reads with timeout
	readWithTimeout := func(timeout time.Duration) (int, error) {
		_ = b.file.SetReadDeadline(time.Now().Add(timeout))
		defer b.file.SetReadDeadline(time.Time{})
		return b.file.Read(buf)
	}

	// For UTF-8, check if we have a complete character
	if len(b.pendingBuf) > 0 && b.pendingBuf[0] >= 0x80 && b.pendingBuf[0] != 0x1b {
		// Non-ASCII UTF-8 character
		if !utf8.FullRune(b.pendingBuf) {
			// Incomplete UTF-8 - read more bytes (up to 4-byte UTF-8 max)
			if len(b.pendingBuf) < utf8.UTFMax {
				n, err := readWithTimeout(50 * time.Millisecond)
				if err == nil && n > 0 {
					b.pendingBuf = append(b.pendingBuf, buf[:n]...)
				}
			}
		}
	}

	// For escape sequences, continue reading until timeout
	if len(b.pendingBuf) > 0 && b.pendingBuf[0] == 0x1b {
		// Read additional bytes until timeout or max escape sequence length
		for len(b.pendingBuf) < 16 { // Max escape sequence length
			n, err := readWithTimeout(50 * time.Millisecond)
			if err != nil || n == 0 {
				break // Timeout or error - no more data
			}
			b.pendingBuf = append(b.pendingBuf, buf[:n]...)
		}
	}

	// Parse and clear buffer
	event, err := b.parser.Parse(b.pendingBuf)

	// Handle incomplete UTF-8 by trying to read more
	if err != nil && err.Error() == "incomplete UTF-8 sequence" {
		// Try one more read for incomplete UTF-8
		if len(b.pendingBuf) < utf8.UTFMax {
			n, readErr := readWithTimeout(50 * time.Millisecond)
			if readErr == nil && n > 0 {
				b.pendingBuf = append(b.pendingBuf, buf[:n]...)
				// Retry parse
				event, err = b.parser.Parse(b.pendingBuf)
			}
		}
	}

	b.pendingBuf = b.pendingBuf[:0] // Clear for next read

	return event, err
}
