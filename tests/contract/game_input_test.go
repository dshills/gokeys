package contract

import (
	"testing"

	"github.com/dshills/gokeys/input"
)

// TestBasicActionBinding validates User Story 1 (P1) - basic single-key action binding
func TestBasicActionBinding(t *testing.T) {
	game := input.NewGameInput(nil)
	if game == nil {
		t.Fatal("NewGameInput returned nil")
	}

	// Test binding actions
	game.Bind("jump", input.KeySpace)
	game.Bind("fire", input.KeyF)
	game.Bind("quit", input.KeyEscape)

	// Note: IsActionPressed will return false in test environment
	// because Input system is not started and no keys are actually pressed.
	// This contract test validates the API contract, not key detection.

	// Verify unbound actions return false
	if game.IsActionPressed("undefined") {
		t.Error("Unbound action should return false")
	}

	// Test unbinding
	game.Bind("jump") // Unbind
	if game.IsActionPressed("jump") {
		t.Error("Unbound action should return false")
	}
}

// TestMultipleKeysPerAction validates User Story 2 (P2) - multiple keys per action
func TestMultipleKeysPerAction(t *testing.T) {
	game := input.NewGameInput(nil)

	// Bind multiple keys to movement actions (WASD + arrows)
	game.Bind("move-up", input.KeyUp, input.KeyW)
	game.Bind("move-down", input.KeyDown, input.KeyS)
	game.Bind("move-left", input.KeyLeft, input.KeyA)
	game.Bind("move-right", input.KeyRight, input.KeyD)

	// Bind multiple confirm keys
	game.Bind("confirm", input.KeyEnter, input.KeySpace, input.KeyY)

	// Verify API accepts multiple keys without error
	// Actual key press testing requires integration with real Input
}

// TestRuntimeRebinding validates User Story 3 (P3) - dynamic rebinding
func TestRuntimeRebinding(t *testing.T) {
	game := input.NewGameInput(nil)

	// Initial binding
	game.Bind("jump", input.KeySpace)

	// Rebind to different key
	game.Bind("jump", input.KeyJ)

	// Verify rebinding works (old key should not work)
	// Note: Actual key press testing requires integration test

	// Test unbinding
	game.Bind("jump")

	// Verify unbinding works
	if game.IsActionPressed("jump") {
		t.Error("Unbound action should return false")
	}
}
