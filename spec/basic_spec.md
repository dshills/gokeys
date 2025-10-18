# gokeys
If someone wanted to make a vendor-grade, cross-terminal key input system for Go, it would have to abstract away all the nonsense you’ve just run into: blocking reads, escape-sequence parsing, buffering, and OS repeat behavior.

Here’s what such a clean, general-purpose interface would look like.

CORE Interface Design example
```go
// Package input provides cross-terminal, cross-platform keyboard input
// with normalized event handling and optional mouse/resize support.
package input

import "time"

// Key represents a normalized key code (e.g. Up, Down, A, Escape).
type Key int

const (
	KeyUnknown Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEscape
	KeyCtrlC
	KeyA
	KeyB
	KeyC
	// ...
)

// Modifier represents optional key modifiers (Shift, Alt, Ctrl).
type Modifier int

const (
	ModNone Modifier = 0
	ModShift Modifier = 1 << iota
	ModAlt
	ModCtrl
)

// Event represents a single key event, with optional modifiers and timestamps.
type Event struct {
	Key        Key
	Rune       rune       // printable character, if any
	Modifiers  Modifier
	Timestamp  time.Time
	Pressed    bool       // true = key down, false = key up
	Repeat     bool       // true if OS autorepeat
}

// Input defines the high-level cross-terminal input API.
type Input interface {
	// Start begins capturing input events (initializes the backend).
	Start() error

	// Poll blocks until the next event is available.
	// Returns false if the input system is shutting down.
	Poll() (Event, bool)

	// Next returns immediately; returns nil if no event available.
	Next() *Event

	// IsPressed returns true if the given key is currently held down.
	IsPressed(k Key) bool

	// Stop restores the terminal and stops all background input handling.
	Stop()
}
```

## Higher-Level Interface (for games / engines) example
```go
type GameInput interface {
	Start() error
	Stop()

	// Actions are high-level logical inputs mapped to keys.
	IsActionPressed(action string) bool
	Bind(action string, keys ...Key)
}
```

Concern: Solution
Blocking vs. non-blocking: Two methods: Poll() (blocking) and Next() (non-blocking)
Key repeat consistency: Explicit Event.Repeat flag
Cross-terminal weirdness: Normalize sequences into unified Key codes
Key-up events: Optional Pressed flag gives “up/down” state if available
Platform differences: Backend implementations per OS/terminal: unixReader, windowsReader, etc.
Thread safety: Input maintains its own goroutine feeding a buffered channel
Portability: Works under tcell, termbox, Windows console API, even WASM or SSH clients

## Example Usage
```go
in := input.New() // auto-detects best backend
defer in.Stop()

for {
    ev, ok := in.Poll()
    if !ok {
        break
    }

    if ev.Key == input.KeyEscape || ev.Key == input.KeyCtrlC {
        break
    }

    if ev.Pressed {
        switch ev.Key {
        case input.KeyUp:
            player.Move(0, -1)
        case input.KeyDown:
            player.Move(0, 1)
        }
    }
}
```
