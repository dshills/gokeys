package main

import (
	"fmt"
	"time"

	"github.com/dshills/gokeys/input"
)

func main() {
	fmt.Println("=== GameInput Example - Action Mapping ===")
	fmt.Println()
	fmt.Println("User Story 1 (P1/MVP) - Basic Action Binding:")
	fmt.Println("  Space: Jump")
	fmt.Println("  F: Fire")
	fmt.Println("  ESC: Quit")
	fmt.Println()
	fmt.Println("User Story 2 (P2) - Multiple Keys Per Action:")
	fmt.Println("  Arrow Keys OR WASD: Movement")
	fmt.Println()
	fmt.Println("User Story 3 (P3) - Runtime Rebinding:")
	fmt.Println("  M: Open settings menu to rebind controls")
	fmt.Println()
	fmt.Println("Press any key to start...")
	fmt.Println()

	game := input.NewGameInput(nil)
	if err := game.Start(); err != nil {
		panic(err)
	}
	defer game.Stop()

	// P1: Basic single-key bindings
	game.Bind("jump", input.KeySpace)
	game.Bind("fire", input.KeyF)
	game.Bind("quit", input.KeyEscape)

	// P2: Multiple keys per action (WASD + arrows)
	game.Bind("move-up", input.KeyUp, input.KeyW)
	game.Bind("move-down", input.KeyDown, input.KeyS)
	game.Bind("move-left", input.KeyLeft, input.KeyA)
	game.Bind("move-right", input.KeyRight, input.KeyD)

	// P3: Menu key for rebinding
	game.Bind("menu", input.KeyM)

	// Game state
	x, y := 0, 0
	jumpCount := 0
	fireCount := 0

	// Simple game loop (~60fps)
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C

		// Handle movement (P2 - multiple keys)
		if game.IsActionPressed("move-up") {
			y--
		}
		if game.IsActionPressed("move-down") {
			y++
		}
		if game.IsActionPressed("move-left") {
			x--
		}
		if game.IsActionPressed("move-right") {
			x++
		}

		// Handle actions (P1 - basic bindings)
		switch {
		case game.IsActionPressed("jump"):
			jumpCount++
			fmt.Printf("\r🎮 Position: (%3d, %3d) | Jumps: %3d | Fires: %3d | [JUMP!]    ", x, y, jumpCount, fireCount)
			time.Sleep(100 * time.Millisecond) // Debounce
		case game.IsActionPressed("fire"):
			fireCount++
			fmt.Printf("\r🎮 Position: (%3d, %3d) | Jumps: %3d | Fires: %3d | [FIRE!]    ", x, y, jumpCount, fireCount)
			time.Sleep(100 * time.Millisecond) // Debounce
		default:
			fmt.Printf("\r🎮 Position: (%3d, %3d) | Jumps: %3d | Fires: %3d               ", x, y, jumpCount, fireCount)
		}

		// P3: Settings menu (rebinding)
		if game.IsActionPressed("menu") {
			fmt.Println()
			showSettingsMenu(game)
		}

		// Exit
		if game.IsActionPressed("quit") {
			fmt.Println("\n\n✅ Thanks for playing!")
			break
		}
	}
}

// showSettingsMenu demonstrates User Story 3 (P3) - runtime rebinding
func showSettingsMenu(game input.GameInput) {
	fmt.Println("\n\n=== Settings Menu - Rebind Controls ===")
	fmt.Println("1. Rebind 'jump' (currently Space)")
	fmt.Println("2. Rebind 'fire' (currently F)")
	fmt.Println("3. Rebind 'move-left' (currently Left/A)")
	fmt.Println("4. Rebind 'move-up' (currently Up/W)")
	fmt.Println("5. Back to game")
	fmt.Println()
	fmt.Print("Choose an option (1-5): ")

	// For simplicity, skip actual rebinding in this example
	// A full implementation would capture key input and rebind
	time.Sleep(2 * time.Second)
	fmt.Println("\n[Settings menu - rebinding not implemented in this basic example]")
	fmt.Println("Returning to game...")
	fmt.Println()
}
