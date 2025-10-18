# Research: Cross-Terminal Input System

**Feature**: 001-input-system
**Date**: 2025-10-17
**Status**: Complete

## Executive Summary

This document consolidates technical research for implementing a cross-platform terminal input library in Go. Key decisions: use standard library only (no external deps), termios for Unix/Linux/macOS, Windows Console API for Windows, table-driven escape sequence parser, buffered channels for event queue, sync.RWMutex for key state tracking, and build tags for platform isolation.

## Research Areas

### 1. Terminal Raw Mode Implementation

**Decision**: Use `golang.org/x/term` patterns but implement directly via syscall to avoid external dependencies

**Rationale**:
- Standard library syscall package provides access to termios on Unix and Console API on Windows
- Minimizes dependencies while maintaining cross-platform support
- Proven pattern used by golang.org/x/term (which we'll reference but not import)

**Implementation Approach**:

**Unix/Linux/macOS**:
```
1. Save current terminal state via tcgetattr
2. Create new termios struct with raw mode flags:
   - Disable canonical mode (ICANON)
   - Disable echo (ECHO)
   - Disable signal generation (ISIG for interrupt handling)
   - Set VMIN=1, VTIME=0 for blocking single-byte reads
3. Apply via tcsetattr
4. Restore original state in Stop() via saved termios
```

**Windows**:
```
1. Save current console mode via GetConsoleMode
2. Set new mode flags:
   - ENABLE_VIRTUAL_TERMINAL_INPUT for ANSI sequences
   - Disable ENABLE_LINE_INPUT (raw mode)
   - Disable ENABLE_ECHO_INPUT
   - Enable ENABLE_WINDOW_INPUT for key events
3. Read via ReadConsoleInput for native events
4. Restore original mode in Stop()
```

**Alternatives Considered**:
- **tcell library**: Too heavyweight, includes full TUI framework
- **termbox-go**: Unmaintained, last update 2015
- **golang.org/x/term**: Adds external dependency, but provides good reference patterns

**References**:
- Go syscall package: termios constants, ioctl
- POSIX termios specification
- Windows Console API documentation

### 2. Escape Sequence Parsing Strategy

**Decision**: Table-driven parser with prefix tree (trie) structure for sequence matching

**Rationale**:
- Escape sequences vary by terminal emulator but follow patterns
- Trie enables efficient longest-prefix matching
- Table-driven design allows easy addition of new sequences
- Centralized normalization logic (single source of truth)

**Key Sequences to Support** (Priority Order):

**Tier 1 (MVP - Common across all terminals)**:
```
Arrow keys:     \x1b[A (Up), \x1b[B (Down), \x1b[C (Right), \x1b[D (Left)
Function keys:  \x1bOP (F1), \x1bOQ (F2), \x1bOR (F3), \x1bOS (F4)
                \x1b[15~ (F5), \x1b[17~ (F6), ... \x1b[24~ (F12)
Home/End:       \x1b[H or \x1bOH (Home), \x1b[F or \x1bOF (End)
Page Up/Down:   \x1b[5~ (PgUp), \x1b[6~ (PgDn)
Insert/Delete:  \x1b[2~ (Insert), \x1b[3~ (Delete)
Ctrl+Key:       \x01-\x1a (Ctrl+A through Ctrl+Z)
Escape:         \x1b alone (with timeout for CSI sequences)
Enter:          \r or \n
Backspace:      \x7f or \x08
Tab:            \t
```

**Tier 2 (Extended - Terminal-specific variants)**:
```
Shift+Arrow:    \x1b[1;2A (xterm), varies by terminal
Alt+Key:        \x1b<key> or sequences with ;3 modifier
Ctrl+Arrow:     \x1b[1;5A (xterm)
F13-F24:        Extended function keys
```

**Tier 3 (Advanced - Key up/down where supported)**:
```
xterm:          Mouse tracking sequences repurposed
Windows:        Native key up/down events via Console API
```

**Parser Architecture**:
```
1. Read bytes from terminal
2. Detect escape prefix (\x1b)
3. Collect sequence with timeout (50ms for multi-byte)
4. Match against trie (longest prefix)
5. If match: return normalized Key
6. If no match: return KeyUnknown with raw bytes
7. Update Event fields (Modifiers, Pressed, Repeat, Timestamp)
```

**Alternatives Considered**:
- **Regex matching**: Too slow, doesn't handle partial sequences
- **State machine**: Complex to maintain, harder to extend
- **Hardcoded if/else**: Not scalable, violates DRY

**Timeout Strategy**:
- 50ms window for multi-byte sequences (covers network latency)
- Standalone Escape vs Escape prefix disambiguation

### 3. Event Queue Design

**Decision**: Buffered channel with capacity 100, separate goroutine for capture

**Rationale**:
- Go channels provide built-in concurrency primitives
- Buffering prevents event loss during consumer processing spikes
- Goroutine isolation simplifies cancellation and cleanup
- Standard Go pattern for producer-consumer

**Architecture**:
```go
type inputImpl struct {
    events  chan Event      // Buffered channel (cap 100)
    done    chan struct{}   // Shutdown signal
    backend Backend         // Platform-specific reader
    state   *StateTracker   // Key state tracking
}

// Capture goroutine (started by Start())
func (i *inputImpl) capture() {
    for {
        select {
        case <-i.done:
            return
        default:
            event, err := i.backend.ReadEvent()
            if err != nil {
                continue // Log and continue
            }
            i.state.Update(event) // Update IsPressed state
            select {
            case i.events <- event:
                // Event queued
            case <-i.done:
                return
            default:
                // Buffer full - drop oldest (or block?)
                // Decision: Block to ensure no loss
            }
        }
    }
}
```

**Buffer Sizing**:
- 100 events = 1.6s of buffering at 60fps
- Sufficient for GC pauses, rendering spikes
- Configurable in future if needed

**Alternatives Considered**:
- **Ring buffer**: More complex, no GC benefit over channel
- **Unbounded queue**: Memory leak risk
- **Smaller buffer (10)**: Insufficient for rendering spikes

### 4. Key State Tracking (IsPressed)

**Decision**: sync.RWMutex-protected map[Key]bool, updated by capture goroutine

**Rationale**:
- RWMutex allows concurrent reads (IsPressed) with exclusive writes (capture)
- Map provides O(1) lookup
- Updated inline during event processing (no delay)
- Graceful degradation: platforms without key-up still track via press events

**Architecture**:
```go
type StateTracker struct {
    mu      sync.RWMutex
    pressed map[Key]bool
}

func (s *StateTracker) Update(e Event) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if e.Pressed {
        s.pressed[e.Key] = true
    } else {
        delete(s.pressed, e.Key)
    }
}

func (s *StateTracker) IsPressed(k Key) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.pressed[k]
}
```

**Degradation Strategy** (platforms without key-up):
- Press event → mark as pressed
- Next different key press → clear previous (approximation)
- Document limitation in godoc

**Alternatives Considered**:
- **Atomic operations**: Map[Key]bool doesn't support atomic ops
- **Lock-free data structures**: Over-engineered for this use case
- **Event-based state**: Race condition between event processing and IsPressed query

### 5. Monotonic Timestamps

**Decision**: Use `time.Now()` directly - Go 1.9+ includes monotonic clock component

**Rationale**:
- time.Now() automatically includes monotonic reading in Go 1.9+
- Survives system clock adjustments (NTP, manual changes)
- No additional syscall overhead
- Duration calculations use monotonic component automatically

**Verification**:
```go
t1 := time.Now()
// System clock adjusted backwards
t2 := time.Now()
d := t2.Sub(t1) // Still positive due to monotonic component
```

**Alternatives Considered**:
- **Manual monotonic clock via syscall**: Unnecessary, Go handles it
- **Relative timestamps**: Complicates consumer usage

### 6. Cross-Platform Build Strategy

**Decision**: Build tags for platform-specific files, shared interface in backend.go

**Rationale**:
- Go build tags are standard for platform code
- Clear separation of Unix vs Windows implementations
- Shared Backend interface enforces contract
- Testing can verify interface compliance

**File Structure**:
```
internal/backend/
├── backend.go           // Backend interface (no build tag - always included)
├── unix.go              // +build !windows
├── windows.go           // +build windows
└── parser.go            // Shared escape sequence parser (no build tag)
```

**Build Tags**:
```go
// unix.go
//go:build !windows
// +build !windows

package backend

// UnixBackend implements Backend for Unix-like systems
type UnixBackend struct { ... }
```

**Factory Pattern** (in input/input.go):
```go
func New() (Input, error) {
    var backend internal.Backend
    if runtime.GOOS == "windows" {
        backend = &backend.WindowsBackend{}
    } else {
        backend = &backend.UnixBackend{}
    }
    return newInput(backend), nil
}
```

**Alternatives Considered**:
- **Runtime switching**: Less efficient, requires dead code on all platforms
- **Separate packages**: Breaks Go package conventions

### 7. Action Mapping Design (GameInput)

**Decision**: Map[string][]Key with simple lookup, separate package optional import

**Rationale**:
- Simple map structure covers 99% of game use cases
- Multiple keys per action via slice
- O(n) scan of bound keys acceptable (n typically <5 per action)
- Optional import keeps core library lightweight

**Architecture**:
```go
type GameInput struct {
    input    Input
    bindings map[string][]Key
    mu       sync.RWMutex
}

func (g *GameInput) IsActionPressed(action string) bool {
    g.mu.RLock()
    keys := g.bindings[action]
    g.mu.RUnlock()

    for _, key := range keys {
        if g.input.IsPressed(key) {
            return true
        }
    }
    return false
}

func (g *GameInput) Bind(action string, keys ...Key) {
    g.mu.Lock()
    g.bindings[action] = keys
    g.mu.Unlock()
}
```

**Alternatives Considered**:
- **Reverse map (Key→[]Action)**: Complicates binding updates
- **Event-based actions**: Adds complexity, polling is idiomatic for games
- **Built into core**: Violates single responsibility, increases core complexity

## Summary of Key Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| Dependencies | Standard library only | Zero external deps, self-contained |
| Unix Raw Mode | termios via syscall | Direct control, proven pattern |
| Windows Raw Mode | Console API via syscall | Native key events, ANSI support |
| Escape Parsing | Trie-based table lookup | Efficient, extensible, maintainable |
| Event Queue | Buffered channel (cap 100) | Go-idiomatic, built-in concurrency |
| State Tracking | RWMutex + map[Key]bool | Concurrent reads, O(1) lookup |
| Timestamps | time.Now() (monotonic) | Built-in Go 1.9+ support |
| Platform Isolation | Build tags | Standard Go practice |
| Action Mapping | Optional map[string][]Key | Simple, sufficient, decoupled |

## Implementation Priorities

**Phase 1 (P1 - MVP Core)**:
1. Event, Key, Modifier types (input/event.go)
2. Input interface (input/input.go)
3. Backend interface (internal/backend/backend.go)
4. Unix backend with termios (internal/backend/unix.go)
5. Escape sequence parser - Tier 1 keys (internal/backend/parser.go)
6. Event queue + capture goroutine (input/input.go)
7. Start/Stop lifecycle (input/input.go)

**Phase 2 (P2 - State Tracking)**:
8. StateTracker implementation (internal/state/tracker.go)
9. IsPressed integration (input/input.go)
10. Repeat detection logic (internal/backend/parser.go)
11. Pressed field for key up/down (Unix approximation)

**Phase 3 (P3 - Action Mapping)**:
12. GameInput interface (input/game.go)
13. Bind/IsActionPressed implementation

**Phase 4 (Cross-Platform)**:
14. Windows backend (internal/backend/windows.go)
15. Platform contract tests (tests/contract/)
16. Cross-platform integration tests with build tags

## Open Questions & Deferred Decisions

**Resolved - No open questions remain**

All NEEDS CLARIFICATION items from Technical Context have been resolved:
- ✅ Dependencies: Standard library only
- ✅ Testing: Go test, table-driven, build tags
- ✅ Platform support: Unix via termios, Windows via Console API
- ✅ Performance approach: Buffered channels, RWMutex, monotonic time
- ✅ Build strategy: Build tags for platform code
