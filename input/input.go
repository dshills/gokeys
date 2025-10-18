package input

// Input defines the keyboard input API for cross-terminal event capture.
// Implementations provide normalized keyboard events across different terminals
// and operating systems.
//
// All methods are safe for concurrent use from multiple goroutines.
type Input interface {
	// Start initializes the input system and begins capturing keyboard events.
	// It puts the terminal into raw mode, starts the capture goroutine, and
	// prepares the event queue.
	//
	// Returns an error if:
	//   - Terminal initialization fails (permissions, unsupported terminal)
	//   - The input system is already started
	//   - Platform-specific backend initialization fails
	//
	// Start is safe to call from any goroutine. Use sync.Once internally to
	// ensure it only initializes once.
	Start() error

	// Stop restores the terminal to its original state and stops event capture.
	// It is safe to call multiple times (idempotent).
	//
	// All blocked Poll calls will return (zero event, false) after Stop.
	//
	// Best practice: defer input.Stop() after successful Start().
	//
	// Stop is safe to call from any goroutine.
	Stop()

	// Poll blocks until the next keyboard event is available or the input
	// system is shutting down.
	//
	// Returns:
	//   - (Event, true): Normal event
	//   - (zero, false): System shutting down (Stop was called)
	//
	// Poll is thread-safe. Multiple goroutines can call Poll, but each event
	// is delivered to only one caller (channel semantics).
	Poll() (Event, bool)

	// Next returns the next keyboard event immediately without blocking.
	//
	// Returns:
	//   - *Event: Pointer to the next event if available
	//   - nil: No event currently available
	//
	// Next is thread-safe and safe for concurrent calls from multiple goroutines.
	// Typical usage is in game loops or non-blocking event processing.
	Next() *Event

	// IsPressed returns true if the specified key is currently held down.
	// State is updated in real-time as events are processed.
	//
	// On platforms supporting key-up events, this reflects actual physical
	// key state. On platforms without key-up support, state is approximated
	// based on press events.
	//
	// IsPressed is thread-safe and safe for concurrent calls.
	IsPressed(k Key) bool
}

// Backend defines the internal contract for platform-specific terminal I/O.
// This interface is internal and used by the input implementation.
// It abstracts the differences between Unix termios and Windows Console API.
//
// Implementations must handle raw terminal mode, escape sequence parsing,
// and event normalization specific to their platform.
type Backend interface {
	// Init enters raw mode and saves the current terminal state.
	// Must be idempotent - calling multiple times should be a no-op after
	// the first successful call.
	//
	// Returns an error if:
	//   - Platform initialization fails
	//   - Insufficient permissions
	//   - Terminal is unsupported
	Init() error

	// Restore exits raw mode and restores the terminal to its original state.
	// Must be idempotent and safe to call even if Init failed.
	// Guaranteed to be called during cleanup.
	//
	// Returns an error if restoration fails, but implementations should
	// make best effort and not panic.
	Restore() error

	// ReadEvent blocks until a keyboard event is available, then parses
	// and returns a normalized Event.
	//
	// Returns an error if:
	//   - Read operation fails (terminal disconnected, etc.)
	//
	// Should NOT error on unparsable sequences - return Event with
	// Key=KeyUnknown instead.
	//
	// Thread-safety: Only called from a single capture goroutine.
	ReadEvent() (Event, error)
}
