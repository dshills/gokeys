package input

import (
	"sync"
	"testing"
)

// TestConcurrentIsActionPressed verifies thread-safe concurrent reads
func TestConcurrentIsActionPressed(t *testing.T) {
	game := NewGameInput(nil)
	game.Bind("test", KeySpace, KeyEnter, KeyA)

	// 10 goroutines, 1000 calls each
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				game.IsActionPressed("test")
			}
		}()
	}
	wg.Wait()
}

// TestConcurrentBindAndQuery verifies thread-safe concurrent write and reads
func TestConcurrentBindAndQuery(t *testing.T) {
	game := NewGameInput(nil)

	var wg sync.WaitGroup

	// 1 writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			game.Bind("test", KeySpace)
			game.Bind("test", KeyEnter)
			game.Bind("test", KeyA, KeyB, KeyC)
		}
	}()

	// 5 reader goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				game.IsActionPressed("test")
				game.IsActionPressed("undefined")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentMultipleActions tests concurrent access to different actions
func TestConcurrentMultipleActions(t *testing.T) {
	game := NewGameInput(nil)

	// Bind multiple actions
	game.Bind("action1", KeySpace)
	game.Bind("action2", KeyEnter)
	game.Bind("action3", KeyA, KeyB, KeyC)

	var wg sync.WaitGroup

	// Multiple goroutines querying different actions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			var action string
			switch n % 3 {
			case 0:
				action = "action1"
			case 1:
				action = "action2"
			case 2:
				action = "action3"
			}
			for j := 0; j < 500; j++ {
				game.IsActionPressed(action)
			}
		}(i)
	}

	wg.Wait()
}
