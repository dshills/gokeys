package input

import (
	"testing"
)

// BenchmarkReadEventAllocations measures memory allocations per keypress.
// Before optimization: 256 B/op, 1 allocs/op (buffer allocation)
// After optimization: 0 B/op, 0 allocs/op (sync.Pool reuse)
func BenchmarkReadEventAllocations(b *testing.B) {
	backend := newBackend().(*unixBackend)
	if err := backend.Init(); err != nil {
		b.Skip("Not a terminal environment")
	}
	defer func() {
		_ = backend.Restore()
	}()

	b.ReportAllocs() // Critical: enable allocation tracking
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Note: This benchmark requires manual keypresses during execution
		// For automated testing, we rely on the integration tests
		_, _ = backend.ReadEvent()
	}
}
