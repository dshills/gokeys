# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gokeys is a vendor-grade, cross-terminal key input system for Go. It provides a clean, normalized interface for keyboard input handling across different terminals and platforms, abstracting away platform-specific details like blocking reads, escape-sequence parsing, buffering, and OS key repeat behavior.

## Architecture

The project is designed around two main abstraction layers:

### Core Layer: `input` package
Provides low-level, cross-platform keyboard event handling with:
- **Event System**: Normalized `Event` type containing Key, Rune, Modifiers, Timestamp, Pressed state, and Repeat flag
- **Key Abstraction**: Unified `Key` type representing normalized key codes (e.g., KeyUp, KeyDown, KeyA, KeyEscape)
- **Modifier Support**: Bitflag-based `Modifier` type for Shift, Alt, Ctrl combinations
- **Dual API**: Both blocking (`Poll()`) and non-blocking (`Next()`) event retrieval
- **State Tracking**: `IsPressed(Key)` for querying current key states
- **Backend System**: Platform-specific implementations (unixReader, windowsReader, etc.) auto-selected via `input.New()`
- **UTF-8 Support**: Full multi-byte UTF-8 character decoding for international input (2-byte, 3-byte, 4-byte characters including emoji)
- **Performance Optimizations**: Zero-allocation input processing using sync.Pool buffer reuse, <1ms latency for all key events

### Higher Layer: Game/Application Input
Optional higher-level API that provides:
- Action mapping: Bind logical actions (e.g., "jump", "fire") to physical keys
- Action queries: `IsActionPressed(action string)` for game logic

### Key Design Decisions
- **Thread Safety**: Input system maintains its own goroutine feeding a buffered channel
- **Normalization**: All terminal escape sequences converted to unified Key codes
- **Portability**: Works with tcell, termbox, Windows console API, WASM, SSH clients
- **Explicit Repeat**: OS autorepeat indicated via `Event.Repeat` flag for consistent behavior
- **Optional Key-Up Events**: `Event.Pressed` field supports both press and release tracking where available

## Development Commands

### Linting
```bash
golangci-lint run
```

The project uses golangci-lint with a comprehensive set of linters configured in `.golangci.yml`:
- Security: gosec
- Complexity: cyclop (max 30), revive
- Code Quality: gocritic, staticcheck, govet
- Error Handling: errcheck
- Maintainability: dupl, ineffassign, unconvert

Note: Cognitive complexity (gocognit) and unused checks are disabled.

### Building
```bash
go build ./...
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestName ./path/to/package

# Run tests with coverage
go test -cover ./...
```

## Code Standards

When implementing features:
- **Error Handling**: All errors must be checked (errcheck enforcer enabled)
- **Complexity Limits**: Functions should stay under cyclomatic complexity of 30
- **Context Usage**: Pass context as first argument (revive rule enforced)
- **Error Messages**: Start with lowercase, no punctuation (revive error-strings rule)
- **Exported Names**: All exported types/functions must have doc comments (revive exported rule)

## Performance Characteristics

The input system has been optimized for high-performance scenarios:
- **Zero Allocations**: Buffer pooling eliminates per-keypress allocations (0 B/op, 0 allocs/op)
- **Sub-millisecond Latency**: Escape key processing <1ms (removed artificial 5ms delay)
- **UTF-8 Efficiency**: Multi-byte character parsing ~30ns/op (comparable to ASCII)
- **Thread Safety**: Concurrent-safe buffer pooling and state management

## Implementation Status

The core input system is implemented with:
- ✅ Cross-platform backend (Unix/termios complete)
- ✅ Full escape sequence parsing
- ✅ UTF-8 multi-byte character support
- ✅ Zero-allocation buffer management
- ✅ Comprehensive test coverage (contract, integration, benchmarks)
- ✅ Race condition testing and validation

See `spec/basic_spec.md` for the original design and `specs/003-fix-performance-issues/` for performance optimization details.
