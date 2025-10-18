package input

import (
	"testing"
)

func TestNewGameInput(t *testing.T) {
	// Test with nil input
	game := NewGameInput(nil)
	if game == nil {
		t.Fatal("NewGameInput(nil) returned nil")
	}

	// Test with existing input
	inp := New()
	game = NewGameInput(inp)
	if game == nil {
		t.Fatal("NewGameInput(input) returned nil")
	}
}

func TestBindSingleKey(t *testing.T) {
	game := NewGameInput(nil)

	// Bind should not panic
	game.Bind("jump", KeySpace)

	// Verify binding exists by checking implementation
	impl, ok := game.(*gameInputImpl)
	if !ok {
		t.Fatal("NewGameInput did not return *gameInputImpl")
	}

	impl.mu.RLock()
	keys, exists := impl.bindings["jump"]
	impl.mu.RUnlock()

	if !exists {
		t.Error("Bind did not create binding for 'jump'")
	}
	if len(keys) != 1 {
		t.Errorf("Expected 1 key bound, got %d", len(keys))
	}
	if keys[0] != KeySpace {
		t.Errorf("Expected KeySpace, got %v", keys[0])
	}
}

func TestIsActionPressedUnbound(t *testing.T) {
	game := NewGameInput(nil)

	// Unbound action should return false, not error
	if game.IsActionPressed("undefined") {
		t.Error("IsActionPressed on unbound action returned true")
	}
}

func TestStartStopDelegation(t *testing.T) {
	// Test that Start/Stop delegate correctly
	game := NewGameInput(nil)

	// Start/Stop should not panic
	// Note: Start may fail in test environment, that's okay
	_ = game.Start()
	game.Stop()
}

func TestBindReplace(t *testing.T) {
	game := NewGameInput(nil)

	// Initial binding
	game.Bind("jump", KeySpace)

	// Verify initial binding
	impl := game.(*gameInputImpl)
	impl.mu.RLock()
	keys := impl.bindings["jump"]
	impl.mu.RUnlock()
	if len(keys) != 1 || keys[0] != KeySpace {
		t.Fatal("Initial bind failed")
	}

	// Replace binding
	game.Bind("jump", KeyEnter)

	// Verify replacement
	impl.mu.RLock()
	keys = impl.bindings["jump"]
	impl.mu.RUnlock()
	if len(keys) != 1 {
		t.Errorf("Expected 1 key after replace, got %d", len(keys))
	}
	if keys[0] != KeyEnter {
		t.Errorf("Expected KeyEnter after replace, got %v", keys[0])
	}
}

func TestUnbind(t *testing.T) {
	game := NewGameInput(nil)

	// Bind then unbind
	game.Bind("jump", KeySpace)
	game.Bind("jump") // Empty keys = unbind

	// Verify unbind
	impl := game.(*gameInputImpl)
	impl.mu.RLock()
	_, exists := impl.bindings["jump"]
	impl.mu.RUnlock()

	if exists {
		t.Error("Unbind did not remove action from bindings")
	}

	if game.IsActionPressed("jump") {
		t.Error("Unbound action returned true")
	}
}

func TestBindMultipleKeys(t *testing.T) {
	game := NewGameInput(nil)

	// Bind two keys to one action
	game.Bind("fire", KeySpace, KeyEnter)

	// Verify both keys are bound
	impl := game.(*gameInputImpl)
	impl.mu.RLock()
	keys := impl.bindings["fire"]
	impl.mu.RUnlock()

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys bound, got %d", len(keys))
	}
	if keys[0] != KeySpace || keys[1] != KeyEnter {
		t.Errorf("Expected [KeySpace, KeyEnter], got %v", keys)
	}
}

func TestMultipleKeysOrLogic(t *testing.T) {
	// This test requires mocking Input.IsPressed
	// For now, just verify the binding structure is correct
	game := NewGameInput(nil)
	game.Bind("confirm", KeyEnter, KeySpace, KeyY)

	impl := game.(*gameInputImpl)
	impl.mu.RLock()
	keys := impl.bindings["confirm"]
	impl.mu.RUnlock()

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys bound, got %d", len(keys))
	}
}

func TestRebindingTakesEffectImmediately(t *testing.T) {
	game := NewGameInput(nil)

	// Initial binding
	game.Bind("jump", KeySpace)

	// Rebind to different key
	game.Bind("jump", KeyJ)

	// Verify new binding immediately
	impl := game.(*gameInputImpl)
	impl.mu.RLock()
	keys := impl.bindings["jump"]
	impl.mu.RUnlock()

	if len(keys) != 1 || keys[0] != KeyJ {
		t.Error("Rebinding did not take effect immediately")
	}
}
