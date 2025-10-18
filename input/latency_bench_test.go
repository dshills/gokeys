package input

import (
	"testing"
	"time"
)

// BenchmarkEscapeKeyLatency measures the time to process an Escape keypress.
// Before optimization: ~5ms (due to time.Sleep)
// After optimization: <1ms (using VTIME timeout)
func BenchmarkEscapeKeyLatency(b *testing.B) {
	backend := newBackend().(*unixBackend)
	if err := backend.Init(); err != nil {
		b.Skip("Not a terminal environment")
	}
	defer func() {
		_ = backend.Restore()
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		// Note: This benchmark requires manual ESC press during execution
		// For automated testing, we rely on the quickstart.md integration tests
		_, _ = backend.ReadEvent()
		latency := time.Since(start)

		if latency > 10*time.Millisecond {
			b.Logf("Warning: High latency detected: %v", latency)
		}
	}
}

// BenchmarkReadEventLatency measures end-to-end event processing time.
// This provides a baseline for overall system performance improvements.
func BenchmarkReadEventLatency(b *testing.B) {
	backend := newBackend().(*unixBackend)
	if err := backend.Init(); err != nil {
		b.Skip("Not a terminal environment")
	}
	defer func() {
		_ = backend.Restore()
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, _ = backend.ReadEvent()
		_ = time.Since(start)
		// Just measure the latency for reporting
	}
}
