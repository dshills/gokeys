// Package input provides cross-terminal, cross-platform keyboard input
// with normalized event handling.
//
// This package abstracts platform-specific terminal behavior behind a unified
// interface, enabling consistent keyboard input handling across different
// terminals (iTerm2, xterm, Windows Terminal, etc.) and operating systems
// (Linux, macOS, Windows).
//
// # Basic Usage
//
// The simplest way to capture keyboard events:
//
//	in := input.New()
//	if err := in.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer in.Stop()
//
//	for {
//	    event, ok := in.Poll()
//	    if !ok {
//	        break // System shutting down
//	    }
//
//	    if event.Key == input.KeyEscape || event.Key == input.KeyCtrlC {
//	        break
//	    }
//
//	    fmt.Printf("Key: %v\n", event.Key)
//	}
//
// # Features
//
//   - Normalized key codes across all platforms (KeyUp, KeyDown, KeyA, etc.)
//   - Blocking (Poll) and non-blocking (Next) event retrieval
//   - Real-time key state queries (IsPressed)
//   - Modifier key detection (Shift, Alt, Ctrl)
//   - Autorepeat event flagging
//   - Monotonic event timestamps
//   - Graceful terminal restoration
//
// # Platform Support
//
// The package automatically detects the platform and uses the appropriate
// backend:
//   - Unix/Linux/macOS: termios-based raw mode
//   - Windows: Console API with virtual terminal support
//
// All backends normalize escape sequences and key codes to produce identical
// Event values, ensuring cross-platform compatibility.
//
// # Thread Safety
//
// All Input methods are safe for concurrent use. The input capture runs in
// a separate goroutine and communicates via buffered channels.
//
// # GameInput - Action Mapping for Games
//
// For game development, the GameInput interface provides action mapping
// on top of the Input interface. Instead of checking physical keys,
// games can query logical actions:
//
//	game := input.NewGameInput(nil)
//	if err := game.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer game.Stop()
//
//	// Bind actions to keys
//	game.Bind("jump", input.KeySpace)
//	game.Bind("fire", input.KeyF, input.KeyEnter)  // Multiple keys
//	game.Bind("move-left", input.KeyLeft, input.KeyA)  // WASD + arrows
//
//	// Game loop
//	for {
//	    if game.IsActionPressed("jump") {
//	        player.Jump()
//	    }
//	    if game.IsActionPressed("fire") {
//	        player.Fire()
//	    }
//	    if game.IsActionPressed("move-left") {
//	        player.MoveLeft()
//	    }
//	}
//
// This enables:
//   - Rebindable controls (Bind replaces existing bindings)
//   - Alternative key schemes (multiple keys per action)
//   - Action-based game logic (decoupled from physical keys)
//
// Performance: ~9ns per IsActionPressed call, zero allocations, <1ms response time.
package input
