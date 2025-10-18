# gokeys

[![Go Reference](https://pkg.go.dev/badge/github.com/dshills/gokeys.svg)](https://pkg.go.dev/github.com/dshills/gokeys)
[![Go Report Card](https://goreportcard.com/badge/github.com/dshills/gokeys)](https://goreportcard.com/report/github.com/dshills/gokeys)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A vendor-grade, cross-terminal keyboard input system for Go. Provides clean, normalized keyboard input handling across different terminals and platforms, with optional high-level action mapping for game development.

## Features

### Core Input System

- 🎯 **Normalized Key Codes** - Unified key representation across all platforms (KeyUp, KeyDown, KeyA, etc.)
- 🔄 **Dual API** - Both blocking (`Poll()`) and non-blocking (`Next()`) event retrieval
- 🎮 **State Tracking** - Real-time key state queries with `IsPressed(Key)`
- 🔧 **Modifier Support** - Bitflag-based detection for Shift, Alt, Ctrl combinations
- 🔁 **Autorepeat Detection** - OS autorepeat events flagged via `Event.Repeat`
- 🌍 **UTF-8 Support** - Full multi-byte character decoding (2, 3, 4-byte chars including emoji)
- ⚡ **High Performance** - Zero-allocation input processing, <1ms latency
- 🔒 **Thread Safe** - Safe for concurrent use across goroutines
- 🖥️ **Cross-Platform Design** - Architected for Linux, macOS, Windows (macOS tested)
- 📦 **Zero Dependencies** - Standard library only

### GameInput API (Optional)

- 🎮 **Action Mapping** - Map logical actions (e.g., "jump", "fire") to physical keys
- 🔀 **Multiple Keys** - Bind multiple keys to one action (WASD + arrow keys)
- 🔄 **Runtime Rebinding** - Dynamic control customization during gameplay
- ⚡ **Ultra-Fast** - ~10ns per action query, zero allocations on hot path
- 🔒 **Thread Safe** - Concurrent bind and query operations
- 🎯 **Game-Optimized** - Designed for 60fps+ game loops

## Installation

```bash
go get github.com/dshills/gokeys
```

## Quick Start

### Basic Input Handling

```go
package main

import (
    "fmt"
    "log"

    "github.com/dshills/gokeys/input"
)

func main() {
    // Create input system
    in := input.New()
    if err := in.Start(); err != nil {
        log.Fatal(err)
    }
    defer in.Stop()

    fmt.Println("Press keys (ESC to quit)...")

    // Blocking event loop
    for {
        event, ok := in.Poll()
        if !ok {
            break // System shutting down
        }

        if event.Key == input.KeyEscape {
            break
        }

        fmt.Printf("Key: %v, Rune: %c, Modifiers: %v\n",
            event.Key, event.Rune, event.Modifiers)
    }
}
```

### Non-Blocking Input

```go
// Non-blocking event retrieval
for {
    event := in.Next()
    if event == nil {
        // No events available, do other work
        time.Sleep(16 * time.Millisecond)
        continue
    }

    handleEvent(event)
}
```

### Key State Queries

```go
// Real-time key state checking
if in.IsPressed(input.KeySpace) {
    player.Jump()
}

if in.IsPressed(input.KeyW) && in.IsPressed(input.KeyShift) {
    player.Sprint()
}
```

### Game Input with Action Mapping

```go
package main

import (
    "log"
    "time"

    "github.com/dshills/gokeys/input"
)

func main() {
    // Create game input with action mapping
    game := input.NewGameInput(nil)
    if err := game.Start(); err != nil {
        log.Fatal(err)
    }
    defer game.Stop()

    // Bind actions to keys
    game.Bind("jump", input.KeySpace)
    game.Bind("fire", input.KeyF, input.KeyEnter)        // Multiple keys
    game.Bind("move-left", input.KeyLeft, input.KeyA)    // WASD + arrows
    game.Bind("move-right", input.KeyRight, input.KeyD)
    game.Bind("move-up", input.KeyUp, input.KeyW)
    game.Bind("move-down", input.KeyDown, input.KeyS)
    game.Bind("quit", input.KeyEscape)

    // Game loop (60fps)
    ticker := time.NewTicker(16 * time.Millisecond)
    defer ticker.Stop()

    for {
        <-ticker.C

        // Query actions instead of keys
        if game.IsActionPressed("jump") {
            player.Jump()
        }
        if game.IsActionPressed("fire") {
            player.Fire()
        }
        if game.IsActionPressed("move-left") {
            player.MoveLeft()
        }
        if game.IsActionPressed("move-right") {
            player.MoveRight()
        }
        if game.IsActionPressed("quit") {
            break
        }

        update()
        render()
    }
}
```

## API Reference

### Input Interface

The core input interface for low-level keyboard event handling.

#### Methods

```go
// Start initializes the input system and puts terminal in raw mode
Start() error

// Stop restores terminal state and shuts down the input system
Stop()

// Poll blocks until a keyboard event is available (blocking API)
Poll() (Event, bool)

// Next returns the next available event or nil (non-blocking API)
Next() *Event

// IsPressed returns true if the specified key is currently pressed
IsPressed(key Key) bool
```

#### Event Structure

```go
type Event struct {
    Key        Key        // Normalized key code
    Rune       rune       // Unicode character (0 for non-printable keys)
    Modifiers  Modifier   // Modifier keys (Shift, Alt, Ctrl)
    Timestamp  time.Time  // Monotonic event timestamp
    Pressed    bool       // True for key-down, false for key-up
    Repeat     bool       // True if this is an OS autorepeat event
}
```

#### Key Codes

Common key codes (see [full list](https://pkg.go.dev/github.com/dshills/gokeys/input#Key)):

```go
// Special keys
KeyEscape, KeyEnter, KeyBackspace, KeyTab, KeySpace
KeyUp, KeyDown, KeyLeft, KeyRight
KeyHome, KeyEnd, KeyPageUp, KeyPageDown
KeyInsert, KeyDelete

// Function keys
KeyF1, KeyF2, ..., KeyF12

// Letters
KeyA, KeyB, ..., KeyZ

// Numbers
Key0, Key1, ..., Key9

// Modifiers
KeyShift, KeyAlt, KeyCtrl

// And many more...
```

#### Modifiers

```go
type Modifier uint8

const (
    ModShift Modifier = 1 << iota
    ModAlt
    ModCtrl
)

// Usage
if event.Modifiers & ModShift != 0 {
    // Shift is pressed
}
```

### GameInput Interface

High-level action mapping API for game development.

#### Methods

```go
// Start initializes the underlying Input system
Start() error

// Stop cleans up and restores terminal state
Stop()

// Bind associates one or more keys with a logical action name
// Passing no keys unbinds the action
Bind(action string, keys ...Key)

// IsActionPressed returns true if any key bound to the action is pressed
// Returns false if action has no bound keys or none are pressed
IsActionPressed(action string) bool
```

#### Factory Function

```go
// NewGameInput creates a new GameInput instance
// If input is nil, creates a default Input via input.New()
func NewGameInput(input Input) GameInput
```

#### Action Binding Examples

```go
// Single key binding
game.Bind("jump", input.KeySpace)

// Multiple keys for one action (OR logic)
game.Bind("confirm", input.KeyEnter, input.KeySpace, input.KeyY)

// Alternative control schemes
game.Bind("move-left", input.KeyLeft, input.KeyA)  // Arrows OR WASD

// Runtime rebinding
game.Bind("jump", input.KeyJ)  // Replaces previous binding

// Unbinding
game.Bind("jump")  // No keys = unbind
```

## Examples

Complete working examples are available in the [`examples/`](examples/) directory:

- **[basic](examples/basic/)** - Simple event printing
- **[poll](examples/poll/)** - Blocking Poll() API usage
- **[next](examples/next/)** - Non-blocking Next() API usage
- **[game](examples/game/)** - Complete game input with action mapping

Run examples:

```bash
cd examples/basic
go run main.go
```

## Performance

### Core Input System

- **Zero Allocations**: Buffer pooling eliminates per-keypress allocations (0 B/op)
- **Sub-millisecond Latency**: Escape key processing <1ms
- **UTF-8 Efficiency**: Multi-byte character parsing ~30ns/op

```
BenchmarkParseASCII-10          41,120,065    29.86 ns/op    0 B/op    0 allocs/op
BenchmarkParseUTF8_2byte-10     39,646,921    30.92 ns/op    0 B/op    0 allocs/op
BenchmarkParseUTF8_3byte-10     40,314,056    30.18 ns/op    0 B/op    0 allocs/op
BenchmarkParseUTF8_4byte-10     40,837,856    30.24 ns/op    0 B/op    0 allocs/op
```

### GameInput API

- **Ultra-fast queries**: ~10ns per IsActionPressed call
- **Zero allocations** on hot path (queries)
- **Minimal overhead**: 1 allocation per Bind operation

```
BenchmarkIsActionPressed_SingleKey-10      115,165,269    10.27 ns/op     0 B/op    0 allocs/op
BenchmarkIsActionPressed_MultipleKeys-10    36,302,398    31.65 ns/op    48 B/op    1 allocs/op
BenchmarkBind-10                            60,530,014    19.73 ns/op     8 B/op    1 allocs/op
```

**Game Loop Performance**: Supports 60fps+ with 10+ action queries per frame (well under 1ms budget).

## Platform Support

### Currently Tested

- ✅ **macOS** - Actively tested on macOS (darwin/arm64)

### Designed For (Not Yet Tested)

The codebase includes platform-specific backends designed to support:

- 🔶 **Linux** - termios-based implementation (needs testing)
- 🔶 **Windows** - Console API implementation (needs testing)

**⚠️ Community Testing Needed**: While the code is architected for cross-platform support with separate backends for Unix and Windows, comprehensive testing on Linux and Windows systems is still needed. Contributions and test reports from these platforms are welcome!

### Terminal Compatibility

The implementation is designed to work with standard terminal emulators that support:
- VT100/ANSI escape sequences
- Raw mode / cbreak mode
- UTF-8 encoding

**Examples**: iTerm2, Terminal.app, xterm, gnome-terminal, Windows Terminal, etc.

**Note**: Actual compatibility should be verified on your specific terminal.

### Backend Implementation

- **Unix/Linux/macOS**: termios-based raw mode
- **Windows**: Console API with virtual terminal support

All backends are designed to normalize escape sequences and key codes to produce identical Event values across platforms.

## Testing

### Run Tests

```bash
# All tests
go test ./...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...
```

### Test Categories

- **Unit Tests**: Core functionality and edge cases
- **Contract Tests**: End-to-end behavior validation
- **Integration Tests**: Cross-component interaction
- **Concurrency Tests**: Thread safety and race detection
- **Benchmarks**: Performance validation

### Code Quality

```bash
# Run linter
golangci-lint run
```

The project uses comprehensive linting with:
- Security: gosec
- Complexity: cyclop (max 30)
- Code Quality: gocritic, staticcheck, govet
- Error Handling: errcheck
- Maintainability: dupl, ineffassign

## Architecture

### Core Layer: Input Package

Provides low-level, cross-platform keyboard event handling:

```
┌─────────────────────────────────────┐
│      Application Code               │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│      Input Interface                │
│  Poll() / Next() / IsPressed()      │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│      Platform Backend               │
│  (unixReader / windowsReader)       │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│      Terminal / OS                  │
└─────────────────────────────────────┘
```

### GameInput Layer (Optional)

Higher-level action mapping for games:

```
┌─────────────────────────────────────┐
│      Game Logic                     │
│  IsActionPressed("jump")            │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│      GameInput Interface            │
│  Action → Keys mapping              │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│      Input Interface                │
│  IsPressed(KeySpace)                │
└─────────────────────────────────────┘
```

### Thread Safety

- **Input System**: Goroutine-safe, runs input capture in separate goroutine
- **GameInput**: Thread-safe concurrent Bind/IsActionPressed with RWMutex
- **State Tracking**: Concurrent-safe key state management

## Design Principles

1. **Cross-Platform Abstraction** - Same API works everywhere
2. **Dual API Design** - Both blocking and non-blocking patterns
3. **Code Quality** - Comprehensive linting and testing
4. **Zero Dependencies** - Standard library only
5. **Performance First** - Zero allocations, <1ms latency
6. **Thread Safety** - Safe for concurrent use

## Use Cases

### Terminal Applications

- CLI tools with keyboard navigation
- Terminal-based editors
- Interactive shells
- System utilities

### Games

- Terminal-based games (roguelikes, puzzles)
- Retro-style games
- Educational games
- Game prototypes

### Interactive Applications

- Terminal UIs (TUIs)
- Real-time monitoring dashboards
- Interactive forms
- Command-line interfaces

## FAQ

### Q: How do I handle Ctrl+C gracefully?

```go
if event.Key == input.KeyCtrlC {
    // Cleanup and exit
    in.Stop()
    os.Exit(0)
}
```

### Q: Can I use both Poll() and Next()?

Yes, but not simultaneously. Choose one pattern for your application. Poll() blocks, Next() doesn't.

### Q: How do I detect key releases?

Check the `event.Pressed` field:

```go
if !event.Pressed {
    // Key was released
}
```

### Q: Does GameInput support key combinations?

Not directly. For Shift+A, check modifiers in the underlying Input system:

```go
if event.Key == input.KeyA && event.Modifiers&input.ModShift != 0 {
    // Shift+A pressed
}
```

### Q: Can I rebind controls at runtime?

Yes, with GameInput:

```go
// Initial binding
game.Bind("jump", input.KeySpace)

// Player changes controls
game.Bind("jump", input.KeyJ)  // Replaces Space with J
```

### Q: What about Windows support?

The code includes a Windows backend using the Console API with virtual terminal sequences, but it has not been tested yet. Community testing and feedback on Windows platforms would be greatly appreciated!

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (see commit message format below)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Commit Message Format

```
Brief description (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.

- Feature details
- Bug fixes
- Breaking changes

Closes #123
```

### Code Standards

- All code must pass `golangci-lint run`
- Maintain test coverage >80%
- Add tests for new features
- Update documentation
- Follow existing code style

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

Created by Davin Hills ([@dshills](https://github.com/dshills))

## Related Projects

- [tcell](https://github.com/gdamore/tcell) - Terminal cell-based UI
- [termbox-go](https://github.com/nsf/termbox-go) - Terminal rendering
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework

## Support

- 📚 [Documentation](https://pkg.go.dev/github.com/dshills/gokeys)
- 🐛 [Issue Tracker](https://github.com/dshills/gokeys/issues)
- 💬 [Discussions](https://github.com/dshills/gokeys/discussions)

---

**Built with ❤️ for the Go community**
