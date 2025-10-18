# Contract: Input Interface

**Package**: `input`
**Type**: Interface
**Purpose**: Primary public API for cross-platform keyboard event capture

## Interface Definition

```go
package input

// Input defines the core keyboard input API for cross-terminal event capture.
// Implementations provide normalized keyboard events across different terminals
// and operating systems.
type Input interface {
    // Start initializes the input system and begins capturing keyboard events.
    // It puts the terminal into raw mode, starts the capture goroutine, and
    // prepares the event queue.
    //
    // Returns error if:
    //   - Terminal initialization fails (permissions, unsupported terminal)
    //   - Already started (call Stop() first)
    //   - Platform-specific backend initialization fails
    //
    // Safe to call from any goroutine.
    Start() error

    // Stop restores the terminal to its original state and stops event capture.
    // It is safe to call multiple times (idempotent).
    // All blocked Poll() calls will return (zero event, false) after Stop().
    //
    // Best practice: defer input.Stop() after successful Start()
    //
    // Safe to call from any goroutine.
    Stop()

    // Poll blocks until the next keyboard event is available or the input
    // system is shutting down.
    //
    // Returns:
    //   (Event, true)  - Normal event
    //   (zero, false)  - System shutting down (Stop() was called)
    //
    // Thread-safe: Multiple goroutines can call Poll(), but each event is
    // delivered to only one caller.
    Poll() (Event, bool)

    // Next returns the next keyboard event immediately without blocking.
    //
    // Returns:
    //   *Event - Pointer to next event if available
    //   nil    - No event currently available
    //
    // Thread-safe: Safe for concurrent calls from multiple goroutines.
    // Typical usage in game loops or non-blocking event processing.
    Next() *Event

    // IsPressed returns true if the specified key is currently held down.
    // State is updated in real-time as events are processed.
    //
    // On platforms supporting key-up events, this reflects actual physical
    // key state. On platforms without key-up support, state is approximated
    // based on press events (see documentation for limitations).
    //
    // Thread-safe: Safe for concurrent calls.
    IsPressed(k Key) bool
}
```

## Method Contracts

### Start()

**Preconditions**:
- Input system not already started
- Terminal supports raw mode
- Sufficient permissions to modify terminal settings

**Postconditions** (on success):
- Terminal in raw mode (canonical and echo disabled)
- Capture goroutine running
- Event channel ready to receive events
- Ready to call Poll(), Next(), IsPressed()

**Error Cases**:
```go
// Permission denied
if err == ErrPermissionDenied { ... }

// Already started
if err == ErrAlreadyStarted { ... }

// Unsupported terminal
if err == ErrUnsupportedTerminal { ... }
```

**Side Effects**:
- Modifies terminal settings (termios on Unix, Console mode on Windows)
- Launches background goroutine
- Allocates event buffer channel

### Stop()

**Preconditions**: None (safe to call anytime)

**Postconditions**:
- Terminal restored to original state
- Capture goroutine terminated
- All blocked Poll() calls return false
- Subsequent Next() calls return nil

**Idempotency**: Safe to call multiple times, no-op after first call

**Cleanup Guarantees**:
- Terminal always restored (even if Start() failed mid-initialization)
- No goroutine leaks
- Channel properly closed

### Poll()

**Preconditions**:
- Start() successfully called
- Not yet called Stop()

**Blocking Behavior**:
- Blocks calling goroutine until event available or Stop() called
- Uses channel receive (efficient Go blocking primitive)
- Wakes immediately on Stop()

**Return Value Contract**:
```go
event, ok := input.Poll()
if !ok {
    // System shutting down, exit event loop
    return
}
// Process event
```

**Concurrency**:
- Multiple goroutines can Poll() concurrently
- Each event delivered to exactly one caller (channel semantics)
- FIFO ordering guaranteed

### Next()

**Preconditions**:
- Start() successfully called

**Non-Blocking Guarantee**:
- Returns immediately (within microseconds)
- Never blocks calling goroutine
- Suitable for game loops at 60+ fps

**Return Value Contract**:
```go
if event := input.Next(); event != nil {
    // Process event
} else {
    // No event, continue with other game logic
}
```

**Concurrency**: Safe for concurrent calls, but event delivered to only one caller

### IsPressed()

**Preconditions**:
- Start() successfully called
- Key argument is valid Key constant

**State Accuracy**:
- **With key-up support** (Unix xterm, Windows): Reflects actual physical state
- **Without key-up support** (some terminals): Approximation based on press events

**Approximation Strategy** (platforms without key-up):
```
1. Key press → mark as pressed
2. Different key press → clear previous (heuristic)
3. Document limitation in user-facing docs
```

**Concurrency**:
- Lock-free reads via RWMutex (read lock)
- No blocking under normal load
- Consistent snapshot of state

**Return Value**:
```go
if input.IsPressed(KeySpace) {
    // Spacebar currently held
}
// Returns false for:
//   - Never pressed keys
//   - Released keys
//   - Invalid Key constants
```

## Usage Examples

### Basic CLI Tool (Blocking)

```go
package main

import (
    "fmt"
    "github.com/dshills/gokeys/input"
)

func main() {
    in := input.New()
    if err := in.Start(); err != nil {
        panic(err)
    }
    defer in.Stop()

    fmt.Println("Press Escape or Ctrl+C to exit")

    for {
        event, ok := in.Poll()
        if !ok {
            break // Shutdown
        }

        if event.Key == input.KeyEscape || event.Key == input.KeyCtrlC {
            break
        }

        fmt.Printf("Key: %v, Modifiers: %v\n", event.Key, event.Modifiers)
    }
}
```

### Game Loop (Non-Blocking)

```go
package main

import (
    "time"
    "github.com/dshills/gokeys/input"
)

func main() {
    in := input.New()
    if err := in.Start(); err != nil {
        panic(err)
    }
    defer in.Stop()

    ticker := time.NewTicker(16 * time.Millisecond) // 60 fps
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Process input
            for {
                event := in.Next()
                if event == nil {
                    break
                }
                handleEvent(event)
            }

            // Update game state
            if in.IsPressed(input.KeyUp) {
                player.MoveUp()
            }
            if in.IsPressed(input.KeyDown) {
                player.MoveDown()
            }

            // Render
            render()

            // Exit check
            if in.IsPressed(input.KeyEscape) {
                return
            }
        }
    }
}
```

### Signal Handling Integration

```go
package main

import (
    "os"
    "os/signal"
    "syscall"
    "github.com/dshills/gokeys/input"
)

func main() {
    in := input.New()
    if err := in.Start(); err != nil {
        panic(err)
    }
    defer in.Stop()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    for {
        select {
        case <-sigChan:
            // OS signal received, cleanup
            return
        default:
            event := in.Next()
            if event == nil {
                time.Sleep(10 * time.Millisecond)
                continue
            }

            if event.Key == input.KeyCtrlC {
                // User requested exit
                return
            }

            processEvent(event)
        }
    }
}
```

## Error Handling

### Recommended Pattern

```go
in := input.New()

// Start with error handling
if err := input.Start(); err != nil {
    switch {
    case errors.Is(err, input.ErrPermissionDenied):
        log.Fatal("Need terminal permissions")
    case errors.Is(err, input.ErrUnsupportedTerminal):
        log.Fatal("Terminal not supported")
    default:
        log.Fatalf("Failed to start input: %v", err)
    }
}

// Ensure cleanup even on panic
defer in.Stop()
```

## Performance Characteristics

| Operation | Time Complexity | Blocking | Thread-Safe |
|-----------|----------------|----------|-------------|
| Start() | O(1) | No | Yes |
| Stop() | O(1) | No | Yes |
| Poll() | O(1) amortized | Yes (until event) | Yes |
| Next() | O(1) | No | Yes |
| IsPressed() | O(1) | No | Yes |

**Buffer Characteristics**:
- Capacity: 100 events
- Overflow behavior: Blocks capture until space available (no event loss)
- Latency: <1ms from capture to Poll/Next availability

## Platform-Specific Behavior

### Unix/Linux/macOS

- Raw mode via termios
- Escape sequence parsing for key codes
- Key-up events: Limited (approximated via heuristics)
- Modifier detection: Via escape sequences
- Repeat flag: Detected via timing heuristics

### Windows

- Raw mode via Console API
- Native key event parsing (INPUT_RECORD)
- Key-up events: Full support (Pressed field accurate)
- Modifier detection: Native key state query
- Repeat flag: Provided by Console API

## Version Compatibility

**Current Version**: 1.0.0

**Stability**: This interface is stable and will not have breaking changes in 1.x releases.

**Future Extensions** (backward compatible):
- Additional methods for mouse events (when added)
- Configuration options (buffer size, timeout)
- Extended key sets (multimedia keys)

All extensions will be optional and backward compatible.
