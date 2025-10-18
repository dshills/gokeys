package input

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// inputImpl is the concrete implementation of the Input interface.
// It manages a background goroutine for event capture and maintains
// a buffered channel for event delivery.
type inputImpl struct {
	backend   Backend
	events    chan Event
	done      chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
	keyState  map[Key]bool
	started   bool
	stopping  bool
	stopOnce  sync.Once
}

// New creates a new Input instance with the appropriate backend
// for the current platform.
func New() Input {
	return &inputImpl{
		backend:  newBackend(),
		events:   make(chan Event, 100),
		done:     make(chan struct{}),
		keyState: make(map[Key]bool),
	}
}

// Start initializes the input system and begins event capture.
// It enters raw mode and spawns a background goroutine to read events.
func (in *inputImpl) Start() error {
	in.mu.Lock()
	defer in.mu.Unlock()

	if in.started {
		return fmt.Errorf("input already started")
	}

	// Initialize backend (enter raw mode)
	if err := in.backend.Init(); err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}

	// Start capture goroutine
	in.wg.Add(1)
	go in.captureLoop()

	in.started = true
	return nil
}

// Stop gracefully shuts down the input system.
// It signals the capture goroutine to exit, waits for it to finish,
// and restores the terminal to its original state.
// Safe to call multiple times (idempotent).
func (in *inputImpl) Stop() {
	// Use sync.Once to prevent double-close panics and race conditions
	in.stopOnce.Do(func() {
		in.mu.Lock()
		if !in.started {
			in.mu.Unlock()
			return
		}
		in.stopping = true
		in.mu.Unlock()

		// Signal shutdown - safe because stopOnce ensures single execution
		close(in.done)

		// Wait for capture goroutine to exit
		in.wg.Wait()

		// Now we can safely clean up with the lock
		in.mu.Lock()
		defer in.mu.Unlock()

		// Restore terminal state
		_ = in.backend.Restore()

		// Mark as stopped
		in.started = false

		// Close events channel and drain
		close(in.events)
		for range in.events {
		}
	})
}

// Poll returns the next available event, blocking until one is available
// or the system is shutting down.
// Returns (event, true) if an event is available, or (zero, false) on shutdown.
func (in *inputImpl) Poll() (Event, bool) {
	select {
	case event, ok := <-in.events:
		if !ok {
			return Event{}, false
		}
		in.updateKeyState(event)
		return event, true
	case <-in.done:
		return Event{}, false
	}
}

// Next returns the next available event without blocking.
// Returns nil if no event is available.
func (in *inputImpl) Next() *Event {
	select {
	case event := <-in.events:
		in.updateKeyState(event)
		return &event
	default:
		return nil
	}
}

// IsPressed returns true if the specified key is currently pressed.
func (in *inputImpl) IsPressed(k Key) bool {
	in.mu.RLock()
	defer in.mu.RUnlock()
	return in.keyState[k]
}

// captureLoop is the background goroutine that reads events from the backend
// and feeds them into the event channel.
func (in *inputImpl) captureLoop() {
	defer in.wg.Done()

	const (
		maxConsecutiveErrors = 10
		errorBackoff         = 100 * time.Millisecond
	)

	consecutiveErrors := 0

	for {
		// Check if we should exit
		select {
		case <-in.done:
			return
		default:
		}

		// Read event from backend (blocking)
		event, err := in.backend.ReadEvent()
		if err != nil {
			if err == io.EOF {
				// Backend closed, exit gracefully
				return
			}

			// Handle other errors with backoff to prevent CPU spin
			consecutiveErrors++
			if consecutiveErrors >= maxConsecutiveErrors {
				// Too many errors, something is seriously wrong
				// Exit gracefully to prevent infinite error loop
				return
			}

			// Back off to avoid CPU spin on persistent errors
			select {
			case <-time.After(errorBackoff):
				continue
			case <-in.done:
				return
			}
		}

		// Reset error counter on successful read
		consecutiveErrors = 0

		// Try to send event to channel
		select {
		case in.events <- event:
			// Event sent successfully
		case <-in.done:
			// Shutdown signal received
			return
		}
	}
}

// updateKeyState updates the internal key state tracking.
func (in *inputImpl) updateKeyState(event Event) {
	in.mu.Lock()
	defer in.mu.Unlock()

	if event.Pressed {
		in.keyState[event.Key] = true
	} else {
		in.keyState[event.Key] = false
	}
}
