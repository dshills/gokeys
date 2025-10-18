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
package input
