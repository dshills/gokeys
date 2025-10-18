package input

import "testing"

// BenchmarkIsActionPressed_SingleKey benchmarks action query with single key
func BenchmarkIsActionPressed_SingleKey(b *testing.B) {
	game := NewGameInput(nil)
	game.Bind("test", KeySpace)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.IsActionPressed("test")
	}
}

// BenchmarkIsActionPressed_MultipleKeys benchmarks action query with multiple keys
func BenchmarkIsActionPressed_MultipleKeys(b *testing.B) {
	game := NewGameInput(nil)
	game.Bind("test", KeySpace, KeyEnter, KeyA, KeyB, KeyC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.IsActionPressed("test")
	}
}

// BenchmarkIsActionPressed_Unbound benchmarks action query for unbound action
func BenchmarkIsActionPressed_Unbound(b *testing.B) {
	game := NewGameInput(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.IsActionPressed("undefined")
	}
}

// BenchmarkBind benchmarks binding operation
func BenchmarkBind(b *testing.B) {
	game := NewGameInput(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Bind("test", KeySpace)
	}
}

// BenchmarkBindMultipleKeys benchmarks binding multiple keys
func BenchmarkBindMultipleKeys(b *testing.B) {
	game := NewGameInput(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Bind("test", KeySpace, KeyEnter, KeyA, KeyB, KeyC)
	}
}

// BenchmarkUnbind benchmarks unbinding operation
func BenchmarkUnbind(b *testing.B) {
	game := NewGameInput(nil)
	game.Bind("test", KeySpace) // Initial binding

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.Bind("test") // Unbind
		game.Bind("test", KeySpace) // Rebind for next iteration
	}
}
