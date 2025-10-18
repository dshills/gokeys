# Quickstart Guide: Cross-Terminal Input System

**Feature**: 001-input-system
**Audience**: Developers implementing the input library
**Purpose**: Step-by-step implementation and validation guide

## Overview

This guide walks through implementing the cross-terminal input system from scratch, with validation checkpoints at each phase. Follow these steps sequentially to build a working implementation that passes all constitution checks and contract tests.

## Prerequisites

**Before You Begin**:
- [ ] Read `spec.md` (functional requirements)
- [ ] Read `plan.md` (architecture and constitution compliance)
- [ ] Read `research.md` (technical decisions)
- [ ] Review `contracts/` (API contracts)
- [ ] Have Go 1.25.3+ installed
- [ ] Have golangci-lint installed
- [ ] Understand termios (Unix) or Console API (Windows)

**Repository Structure Check**:
```bash
# Ensure you're in the project root
cd /path/to/gokeys

# Verify directory structure exists
ls -la  # Should see go.mod, .golangci.yml

# If not already created, initialize:
mkdir -p input internal/{backend,state,queue} examples/{basic,game,inspector} tests/{contract,integration}
```

## Phase 1: Core Types (P1 - Foundation)

**Objective**: Define public types (Event, Key, Modifier) with zero dependencies

**Files to Create**:
- `input/event.go` - Event struct, Key enum, Modifier flags
- `input/doc.go` - Package documentation

### Step 1.1: Define Key Constants

**File**: `input/event.go`

```go
package input

// Key represents a normalized key code.
type Key int

const (
    KeyUnknown Key = iota

    // Control keys
    KeyEscape
    KeyEnter
    KeyBackspace
    KeyTab
    KeyDelete
    KeyInsert

    // Arrow keys
    KeyUp
    KeyDown
    KeyLeft
    KeyRight

    // Navigation
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown

    // Function keys
    KeyF1
    KeyF2
    KeyF3
    KeyF4
    KeyF5
    KeyF6
    KeyF7
    KeyF8
    KeyF9
    KeyF10
    KeyF11
    KeyF12

    // Alphanumeric
    KeyA
    KeyB
    KeyC
    // ... through KeyZ

    Key0
    Key1
    // ... through Key9

    // Ctrl combinations
    KeyCtrlA
    KeyCtrlB
    KeyCtrlC
    // ... through KeyCtrlZ

    KeySpace
)
```

**Validation**:
```bash
go build ./input
golangci-lint run ./input

# Should compile with no errors
# Should pass linter checks
```

### Step 1.2: Define Modifier and Event

**Add to**: `input/event.go`

```go
import "time"

// Modifier represents key modifiers (Shift, Alt, Ctrl).
type Modifier int

const (
    ModNone  Modifier = 0
    ModShift Modifier = 1 << iota
    ModAlt
    ModCtrl
)

// Event represents a keyboard event.
type Event struct {
    Key        Key
    Rune       rune
    Modifiers  Modifier
    Timestamp  time.Time
    Pressed    bool
    Repeat     bool
}
```

**Validation**:
```bash
go test -run TestEventStructure ./input

# Write test:
# - Verify Event zero value
# - Verify Modifier bitflags work (ModShift | ModCtrl)
# - Verify Key constants are sequential
```

### Step 1.3: Add Package Documentation

**File**: `input/doc.go`

```go
// Package input provides cross-terminal, cross-platform keyboard input
// with normalized event handling.
//
// Basic usage:
//
//     in := input.New()
//     defer in.Stop()
//     in.Start()
//
//     for {
//         event, ok := in.Poll()
//         if !ok {
//             break
//         }
//         // Process event
//     }
package input
```

**Validation**:
```bash
go doc input
# Should display package documentation

golangci-lint run ./input
# Should pass revive exported rule
```

**Checkpoint 1**: ✅ Core types defined, documented, linting passes

---

## Phase 2: Input Interface (P1 - API Contract)

**Objective**: Define Input interface and factory function

**Files to Create**:
- `input/input.go` - Input interface, New() factory, inputImpl skeleton

### Step 2.1: Define Input Interface

**File**: `input/input.go`

```go
package input

// Input defines the keyboard input API.
type Input interface {
    Start() error
    Stop()
    Poll() (Event, bool)
    Next() *Event
    IsPressed(k Key) bool
}

// New creates a new Input instance with platform-appropriate backend.
func New() Input {
    // TODO: Detect platform, create backend
    return &inputImpl{}
}
```

**Validation**:
```bash
go build ./input
# Should compile
```

### Step 2.2: Create inputImpl Skeleton

**Add to**: `input/input.go`

```go
type inputImpl struct {
    // TODO: Add fields after backend is ready
}

func (i *inputImpl) Start() error {
    return nil // TODO
}

func (i *inputImpl) Stop() {
    // TODO
}

func (i *inputImpl) Poll() (Event, bool) {
    return Event{}, false // TODO
}

func (i *inputImpl) Next() *Event {
    return nil // TODO
}

func (i *inputImpl) IsPressed(k Key) bool {
    return false // TODO
}
```

**Validation**:
```bash
# Verify interface compliance
go build ./input

# Write interface compliance test
# input/input_test.go:
var _ Input = (*inputImpl)(nil)
```

**Checkpoint 2**: ✅ Input interface defined, skeleton compiles

---

## Phase 3: Backend Interface (P1 - Platform Abstraction)

**Objective**: Define internal Backend interface for platform isolation

**Files to Create**:
- `internal/backend/backend.go` - Backend interface
- `internal/backend/parser.go` - Escape sequence parser (shared)

### Step 3.1: Define Backend Interface

**File**: `internal/backend/backend.go`

```go
package backend

import "github.com/dshills/gokeys/input"

// Backend defines platform-specific terminal I/O.
type Backend interface {
    Init() error
    Restore() error
    ReadEvent() (input.Event, error)
}
```

**Validation**:
```bash
go build ./internal/backend
```

### Step 3.2: Create Escape Sequence Parser

**File**: `internal/backend/parser.go`

```go
package backend

import "github.com/dshills/gokeys/input"

// SequenceParser parses terminal escape sequences to Key codes.
type SequenceParser struct {
    trie *SequenceNode
}

type SequenceNode struct {
    children  map[byte]*SequenceNode
    key       input.Key
    modifiers input.Modifier
}

func NewSequenceParser() *SequenceParser {
    p := &SequenceParser{
        trie: &SequenceNode{
            children: make(map[byte]*SequenceNode),
        },
    }
    p.buildTrie()
    return p
}

func (p *SequenceParser) buildTrie() {
    // Add common sequences (Tier 1 from research.md)
    p.addSequence([]byte{0x1b, '[', 'A'}, input.KeyUp, input.ModNone)
    p.addSequence([]byte{0x1b, '[', 'B'}, input.KeyDown, input.ModNone)
    p.addSequence([]byte{0x1b, '[', 'C'}, input.KeyRight, input.ModNone)
    p.addSequence([]byte{0x1b, '[', 'D'}, input.KeyLeft, input.ModNone)
    // ... add more sequences
}

func (p *SequenceParser) addSequence(seq []byte, key input.Key, mod input.Modifier) {
    // TODO: Implement trie insertion
}

func (p *SequenceParser) Parse(buf []byte) (input.Event, error) {
    // TODO: Implement trie lookup
    return input.Event{}, nil
}
```

**Validation**:
```bash
# Write parser tests (table-driven)
go test ./internal/backend -run TestSequenceParser

# Test cases:
# - Arrow keys: \x1b[A → KeyUp
# - Ctrl+C: \x03 → KeyCtrlC
# - Unknown: \x1b[999~ → KeyUnknown
```

**Checkpoint 3**: ✅ Backend interface defined, parser skeleton ready

---

## Phase 4: Unix Backend (P1 - Platform Implementation)

**Objective**: Implement UnixBackend using termios

**Files to Create**:
- `internal/backend/unix.go` (build tag: `!windows`)

### Step 4.1: Implement UnixBackend

**File**: `internal/backend/unix.go`

```go
//go:build !windows
// +build !windows

package backend

import (
    "os"
    "syscall"
    "github.com/dshills/gokeys/input"
)

type UnixBackend struct {
    initialized bool
    savedState  syscall.Termios
    fd          int
    parser      *SequenceParser
}

func NewUnixBackend() *UnixBackend {
    return &UnixBackend{
        fd:     int(os.Stdin.Fd()),
        parser: NewSequenceParser(),
    }
}

func (b *UnixBackend) Init() error {
    if b.initialized {
        return nil
    }

    // Save current state
    if err := syscall.Tcgetattr(b.fd, &b.savedState); err != nil {
        return err
    }

    // Enter raw mode
    raw := b.savedState
    raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ISIG
    raw.Cc[syscall.VMIN] = 1
    raw.Cc[syscall.VTIME] = 0

    if err := syscall.Tcsetattr(b.fd, syscall.TCSANOW, &raw); err != nil {
        return err
    }

    b.initialized = true
    return nil
}

func (b *UnixBackend) Restore() error {
    if !b.initialized {
        return nil
    }
    syscall.Tcsetattr(b.fd, syscall.TCSANOW, &b.savedState)
    b.initialized = false
    return nil
}

func (b *UnixBackend) ReadEvent() (input.Event, error) {
    buf := make([]byte, 32)
    n, err := syscall.Read(b.fd, buf)
    if err != nil {
        return input.Event{}, err
    }

    return b.parser.Parse(buf[:n])
}
```

**Validation**:
```bash
# On Unix system only
go test ./internal/backend -run TestUnixBackend

# Manual test:
go run examples/inspector/main.go
# Press keys, verify correct Key codes printed
```

**Checkpoint 4**: ✅ Unix backend implemented, manual testing passes

---

## Phase 5: Event Queue & Capture (P1 - Concurrency)

**Objective**: Implement capture goroutine and event queue

**Files to Update**:
- `input/input.go` - Complete inputImpl with channels

### Step 5.1: Add Fields to inputImpl

**Update**: `input/input.go`

```go
import (
    "sync"
    "github.com/dshills/gokeys/internal/backend"
)

type inputImpl struct {
    backend backend.Backend
    events  chan Event
    done    chan struct{}
    once    sync.Once
}

func New() Input {
    return &inputImpl{
        backend: createBackend(), // Platform-specific
        events:  make(chan Event, 100),
        done:    make(chan struct{}),
    }
}

func createBackend() backend.Backend {
    // Platform detection
    if runtime.GOOS == "windows" {
        return backend.NewWindowsBackend()
    }
    return backend.NewUnixBackend()
}
```

### Step 5.2: Implement Start/Stop

**Update**: `input/input.go`

```go
func (i *inputImpl) Start() error {
    var err error
    i.once.Do(func() {
        if err = i.backend.Init(); err != nil {
            return
        }
        go i.capture()
    })
    return err
}

func (i *inputImpl) Stop() {
    close(i.done)
    i.backend.Restore()
}

func (i *inputImpl) capture() {
    for {
        select {
        case <-i.done:
            return
        default:
            event, err := i.backend.ReadEvent()
            if err != nil {
                continue
            }

            select {
            case i.events <- event:
            case <-i.done:
                return
            }
        }
    }
}
```

### Step 5.3: Implement Poll/Next

**Update**: `input/input.go`

```go
func (i *inputImpl) Poll() (Event, bool) {
    select {
    case event := <-i.events:
        return event, true
    case <-i.done:
        return Event{}, false
    }
}

func (i *inputImpl) Next() *Event {
    select {
    case event := <-i.events:
        return &event
    default:
        return nil
    }
}
```

**Validation**:
```bash
# Integration test
go test ./input -run TestEventCapture

# Manual test
go run examples/basic/main.go
# Press keys, verify events captured
```

**Checkpoint 5**: ✅ Event capture working, Poll/Next functional

---

## Phase 6: State Tracking (P2 - IsPressed)

**Objective**: Implement key state tracking for IsPressed()

**Files to Create**:
- `internal/state/tracker.go` - StateTracker implementation

### Step 6.1: Implement StateTracker

**File**: `internal/state/tracker.go`

```go
package state

import (
    "sync"
    "github.com/dshills/gokeys/input"
)

type Tracker struct {
    mu      sync.RWMutex
    pressed map[input.Key]bool
}

func New() *Tracker {
    return &Tracker{
        pressed: make(map[input.Key]bool),
    }
}

func (t *Tracker) Update(e input.Event) {
    t.mu.Lock()
    defer t.mu.Unlock()

    if e.Pressed {
        t.pressed[e.Key] = true
    } else {
        delete(t.pressed, e.Key)
    }
}

func (t *Tracker) IsPressed(k input.Key) bool {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.pressed[k]
}
```

### Step 6.2: Integrate StateTracker

**Update**: `input/input.go`

```go
import "github.com/dshills/gokeys/internal/state"

type inputImpl struct {
    // ... existing fields
    state *state.Tracker
}

func New() Input {
    return &inputImpl{
        // ... existing initialization
        state: state.New(),
    }
}

func (i *inputImpl) capture() {
    for {
        select {
        case <-i.done:
            return
        default:
            event, err := i.backend.ReadEvent()
            if err != nil {
                continue
            }

            i.state.Update(event) // Update state

            select {
            case i.events <- event:
            case <-i.done:
                return
            }
        }
    }
}

func (i *inputImpl) IsPressed(k Key) bool {
    return i.state.IsPressed(k)
}
```

**Validation**:
```bash
# Test state tracking
go test ./internal/state -run TestTracker

# Manual test with game example
go run examples/game/main.go
# Hold keys, verify continuous movement
```

**Checkpoint 6**: ✅ IsPressed() working, state tracking accurate

---

## Phase 7: Action Mapping (P3 - GameInput)

**Objective**: Implement optional GameInput interface

**Files to Create**:
- `input/game.go` - GameInput implementation

### Step 7.1: Implement GameInput

**File**: `input/game.go`

```go
package input

import "sync"

type GameInput interface {
    Start() error
    Stop()
    IsActionPressed(action string) bool
    Bind(action string, keys ...Key)
}

type gameInputImpl struct {
    input    Input
    bindings map[string][]Key
    mu       sync.RWMutex
}

func NewGameInput() GameInput {
    return &gameInputImpl{
        input:    New(),
        bindings: make(map[string][]Key),
    }
}

func (g *gameInputImpl) Start() error {
    return g.input.Start()
}

func (g *gameInputImpl) Stop() {
    g.input.Stop()
}

func (g *gameInputImpl) IsActionPressed(action string) bool {
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

func (g *gameInputImpl) Bind(action string, keys ...Key) {
    g.mu.Lock()
    g.bindings[action] = keys
    g.mu.Unlock()
}
```

**Validation**:
```bash
# Test action mapping
go test ./input -run TestGameInput

# Manual test
go run examples/game/main.go
# Verify action-based controls work
```

**Checkpoint 7**: ✅ GameInput implemented, action mapping works

---

## Phase 8: Contract Tests (Cross-Platform Validation)

**Objective**: Write contract tests to validate normalization across backends

**Files to Create**:
- `tests/contract/normalization_test.go`

### Step 8.1: Create Contract Tests

**File**: `tests/contract/normalization_test.go`

```go
package contract_test

import (
    "testing"
    "github.com/dshills/gokeys/input"
    "github.com/dshills/gokeys/internal/backend"
)

func TestKeyNormalization(t *testing.T) {
    tests := []struct {
        name     string
        sequence []byte
        want     input.Key
        wantMod  input.Modifier
    }{
        {"ArrowUp", []byte{0x1b, '[', 'A'}, input.KeyUp, input.ModNone},
        {"ArrowDown", []byte{0x1b, '[', 'B'}, input.KeyDown, input.ModNone},
        {"Ctrl+C", []byte{0x03}, input.KeyCtrlC, input.ModCtrl},
        {"Escape", []byte{0x1b}, input.KeyEscape, input.ModNone},
    }

    parser := backend.NewSequenceParser()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            event, err := parser.Parse(tt.sequence)
            if err != nil {
                t.Fatalf("Parse error: %v", err)
            }

            if event.Key != tt.want {
                t.Errorf("Key = %v, want %v", event.Key, tt.want)
            }
            if event.Modifiers != tt.wantMod {
                t.Errorf("Modifiers = %v, want %v", event.Modifiers, tt.wantMod)
            }
        })
    }
}
```

**Validation**:
```bash
go test ./tests/contract/...
# All tests should pass
```

**Checkpoint 8**: ✅ Contract tests passing, normalization validated

---

## Phase 9: Examples & Documentation

**Objective**: Create working examples and finalize documentation

**Files to Create**:
- `examples/basic/main.go` - Simple event loop
- `examples/game/main.go` - Game input with actions
- `examples/inspector/main.go` - Key code inspector
- `README.md` - Project README
- `CHANGELOG.md` - Version history

### Step 9.1: Create Examples

**File**: `examples/basic/main.go`

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

    fmt.Println("Press keys (Escape to exit):")

    for {
        event, ok := in.Poll()
        if !ok || event.Key == input.KeyEscape {
            break
        }
        fmt.Printf("Key: %v, Mods: %v, Pressed: %v, Repeat: %v\n",
            event.Key, event.Modifiers, event.Pressed, event.Repeat)
    }
}
```

**Validation**:
```bash
go run examples/basic/main.go
# Verify output matches key presses

go run examples/game/main.go
# Verify action mapping works

go run examples/inspector/main.go
# Debug tool for escape sequences
```

**Checkpoint 9**: ✅ Examples working, ready for documentation

---

## Phase 10: Final Validation

**Objective**: Comprehensive testing and linting

### Step 10.1: Run Full Test Suite

```bash
# Unit tests
go test ./...

# Contract tests
go test ./tests/contract/...

# Integration tests (platform-specific)
go test ./tests/integration/... -tags=integration

# Coverage
go test -cover ./... | tee coverage.txt
# Target: >80% for public APIs
```

### Step 10.2: Lint Check

```bash
golangci-lint run ./...
# Must pass all checks (per constitution)

# Specific checks:
# - errcheck: All errors handled
# - revive: All exports documented
# - cyclop: Complexity <30
```

### Step 10.3: Constitution Compliance Review

**Checklist**:
- [ ] Cross-Platform Abstraction: Backends isolated ✅
- [ ] Dual API Design: Poll + Next implemented ✅
- [ ] Code Quality: golangci-lint passes ✅
- [ ] Testing: TDD followed, contract tests exist ✅
- [ ] Platform Normalization: Escape sequences normalized ✅
- [ ] Thread Safety: Goroutine + channels used ✅
- [ ] State Tracking: IsPressed() accurate ✅

### Step 10.4: Manual Platform Testing

**Unix/macOS**:
```bash
# Test on actual terminal
go run examples/basic/main.go

# Test different terminals
# - iTerm2
# - Terminal.app
# - gnome-terminal
# - xterm

# Verify consistent Key codes
```

**Windows**:
```bash
# Test on Windows Terminal
go run examples/basic/main.go

# Test on cmd.exe
go run examples/basic/main.go

# Verify key-up events work
```

**Checkpoint 10**: ✅ All tests passing, ready for production

---

## Common Issues & Solutions

### Issue: Terminal Not Restoring

**Symptom**: Terminal stays in raw mode after program exit

**Solution**:
```go
// Always defer Stop()
func main() {
    in := input.New()
    if err := in.Start(); err != nil {
        panic(err)
    }
    defer in.Stop() // Critical!

    // ... rest of program
}
```

### Issue: Events Not Captured

**Symptom**: Poll() blocks forever

**Debug**:
```bash
# Check if backend initialized
if err := in.Start(); err != nil {
    log.Fatal(err) // Print error
}

# Check if capture goroutine started
# Add debug logging to capture() function
```

### Issue: IsPressed() Always False

**Symptom**: State tracking not working

**Debug**:
- Verify Update() called in capture loop
- Check if Pressed field set correctly
- Verify platform supports key-up events

### Issue: Linter Failures

**Symptom**: golangci-lint reports errors

**Common Fixes**:
```go
// Missing godoc
// Add:

// KeyUp represents the up arrow key.
const KeyUp = ...

// Unchecked error
// Change:
backend.Init()

// To:
if err := backend.Init(); err != nil {
    return err
}
```

---

## Success Criteria Validation

**From spec.md Success Criteria**:

- [ ] SC-001: <10 lines of code for basic usage ✅
- [ ] SC-002: Identical key codes across 4+ terminals ✅
- [ ] SC-003: >95% normalization accuracy ✅
- [ ] SC-004: <10ms event latency ✅
- [ ] SC-005: <16ms state staleness ✅
- [ ] SC-006: 100-event buffer capacity ✅
- [ ] SC-007: Terminal restored 100% ✅
- [ ] SC-008: >95% repeat detection accuracy ✅
- [ ] SC-009: Works on 3+ platforms ✅
- [ ] SC-010: Supports 30+ fps game loops ✅

---

## Next Steps

After completing this quickstart:

1. **Generate Tasks**: Run `/speckit.tasks` to create detailed implementation tasks
2. **Begin Implementation**: Start with Phase 1 (Core Types)
3. **Iterate**: Follow TDD - write tests before implementation
4. **Validate Continuously**: Run tests and linter after each phase
5. **Document**: Add godoc comments as you go

**Good luck!** 🚀
