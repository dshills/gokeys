package input

import (
	"sync"
	"testing"
	"time"
)

// TestStopConcurrency validates that Stop() is safe to call concurrently
// from multiple goroutines and is truly idempotent.
func TestStopConcurrency(t *testing.T) {
	in := New().(*inputImpl)

	// Don't actually start it (we're just testing Stop idempotency)
	in.started = true
	in.done = make(chan struct{})

	// Call Stop() from multiple goroutines concurrently
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			in.Stop() // Should never panic
		}()
	}

	wg.Wait()

	// If we get here without panicking, the test passes
}

// TestStopAfterStart validates that Stop() works correctly after a real Start().
func TestStopAfterStart(t *testing.T) {
	// This test requires a terminal, but we can test the logic with a mock
	// For now, we'll just verify the state transitions work correctly

	in := New().(*inputImpl)

	// Simulate started state
	in.started = true
	in.done = make(chan struct{})
	in.wg.Add(1)

	// Start a goroutine that simulates captureLoop
	go func() {
		defer in.wg.Done()
		<-in.done
	}()

	// Stop should work without blocking
	done := make(chan struct{})
	go func() {
		in.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Stop() blocked for too long")
	}

	// Second Stop() should be safe
	in.Stop()
}

// TestMultipleStopCalls validates that calling Stop() multiple times
// sequentially is safe.
func TestMultipleStopCalls(t *testing.T) {
	in := New().(*inputImpl)

	// Call Stop() multiple times without ever starting
	for i := 0; i < 10; i++ {
		in.Stop() // Should be safe to call on never-started instance
	}
}
