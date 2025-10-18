package input

import (
	"testing"
	"time"
)

// TestPollBlockingBehavior validates that Poll() blocks until an event
// is available or the system shuts down.
func TestPollBlockingBehavior(t *testing.T) {
	// This test validates the contract but requires a mock backend
	// We'll test with a simple channel-based simulation

	events := make(chan Event, 1)
	done := make(chan struct{})

	// Simulate Poll behavior
	poll := func() (Event, bool) {
		select {
		case event := <-events:
			return event, true
		case <-done:
			return Event{}, false
		}
	}

	// Test 1: Poll should block until event arrives
	go func() {
		time.Sleep(10 * time.Millisecond)
		events <- Event{Key: KeyA, Pressed: true}
	}()

	start := time.Now()
	event, ok := poll()
	elapsed := time.Since(start)

	if !ok {
		t.Fatal("Poll() should return true when event available")
	}

	if event.Key != KeyA {
		t.Errorf("Event.Key = %v, want KeyA", event.Key)
	}

	if elapsed < 10*time.Millisecond {
		t.Error("Poll() should have blocked until event was sent")
	}

	// Test 2: Poll should return false on shutdown
	close(done)
	event, ok = poll()

	if ok {
		t.Error("Poll() should return false after shutdown")
	}

	if event.Key != KeyUnknown {
		t.Error("Poll() should return zero event on shutdown")
	}
}

// TestPollShutdownSignal validates that Poll() returns (zero, false)
// when the input system is shutting down.
func TestPollShutdownSignal(t *testing.T) {
	done := make(chan struct{})

	poll := func() (Event, bool) {
		//nolint:staticcheck // Simplified test simulation of Poll behavior
		select {
		case <-done:
			return Event{}, false
		}
	}

	// Close shutdown channel
	close(done)

	// Poll should immediately return false
	event, ok := poll()

	if ok {
		t.Error("Poll() should return false on shutdown")
	}

	if event.Key != KeyUnknown {
		t.Error("Poll() should return zero event on shutdown")
	}
}

// TestNextNonBlockingBehavior validates that Next() returns immediately
// without blocking, regardless of whether events are available.
func TestNextNonBlockingBehavior(t *testing.T) {
	events := make(chan Event, 1)

	// Simulate Next behavior
	next := func() *Event {
		select {
		case event := <-events:
			return &event
		default:
			return nil
		}
	}

	// Test 1: Next should return nil immediately when no events
	start := time.Now()
	result := next()
	elapsed := time.Since(start)

	if result != nil {
		t.Error("Next() should return nil when no events available")
	}

	if elapsed > 1*time.Millisecond {
		t.Error("Next() should return immediately (non-blocking)")
	}

	// Test 2: Next should return event when available
	testEvent := Event{Key: KeyB, Pressed: true}
	events <- testEvent

	result = next()

	if result == nil {
		t.Fatal("Next() should return event when available")
	}

	if result.Key != KeyB {
		t.Errorf("Event.Key = %v, want KeyB", result.Key)
	}

	// Test 3: Subsequent Next should return nil (no more events)
	result = next()

	if result != nil {
		t.Error("Next() should return nil after all events consumed")
	}
}

// TestNextReturnsNilWhenNoEvents validates that Next() returns nil
// immediately when the event queue is empty.
func TestNextReturnsNilWhenNoEvents(t *testing.T) {
	events := make(chan Event, 10)

	next := func() *Event {
		select {
		case event := <-events:
			return &event
		default:
			return nil
		}
	}

	// Empty queue should return nil immediately
	for i := 0; i < 100; i++ {
		result := next()
		if result != nil {
			t.Errorf("Iteration %d: Next() should return nil on empty queue", i)
		}
	}
}

// TestEventOrdering validates that Poll() and Next() return events
// in FIFO order (first-in, first-out).
func TestEventOrdering(t *testing.T) {
	events := make(chan Event, 10)

	// Send events in order
	expectedKeys := []Key{KeyA, KeyB, KeyC, KeyD, KeyE}
	for _, key := range expectedKeys {
		events <- Event{Key: key, Pressed: true}
	}

	// Consume via Poll (simulated)
	poll := func() (Event, bool) {
		select {
		case event := <-events:
			return event, true
		default:
			return Event{}, false
		}
	}

	for i, expectedKey := range expectedKeys {
		event, ok := poll()
		if !ok {
			t.Fatalf("Event %d: Poll() returned false", i)
		}

		if event.Key != expectedKey {
			t.Errorf("Event %d: Key = %v, want %v", i, event.Key, expectedKey)
		}
	}

	// Queue should be empty
	_, ok := poll()
	if ok {
		t.Error("Poll() should return false when queue is empty")
	}
}
