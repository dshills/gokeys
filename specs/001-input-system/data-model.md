# Data Model: Cross-Terminal Input System

**Feature**: 001-input-system
**Date**: 2025-10-17
**Source**: spec.md FR-001 through FR-020

## Overview

This document defines the data structures for the cross-terminal input system. The model consists of public types (exposed via input package), internal types (implementation details), and state structures for tracking key states and action bindings.

## Public Types (input package)

### Key (Enumerated Constant)

**Purpose**: Normalized key code representing a physical or logical key, platform-independent

**Type**: `int` (iota-based enumeration)

**Fields**: N/A (enumerated constants)

**Key Constants** (100+ total, organized by category):

```go
const (
    KeyUnknown Key = iota  // Unparsable sequence

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

    // Navigation keys
    KeyHome
    KeyEnd
    KeyPageUp
    KeyPageDown

    // Function keys
    KeyF1
    KeyF2
    // ... through KeyF12

    // Alphanumeric keys
    KeyA
    KeyB
    // ... through KeyZ
    Key0
    Key1
    // ... through Key9

    // Modifier combinations (represented as separate keys)
    KeyCtrlA
    KeyCtrlB
    // ... through KeyCtrlZ
    KeyCtrlC  // Common exit signal

    // Special keys
    KeySpace
    // Punctuation keys as needed
)
```

**Validation Rules**:
- Read-only constants (enforced by Go const)
- Never expose raw escape sequences to consumers
- KeyUnknown (0) is default zero value

**State Transitions**: N/A (immutable constants)

**Relationships**:
- Referenced by Event.Key
- Used as key in StateTracker.pressed map
- Used in GameInput bindings

### Modifier (Bitflag)

**Purpose**: Composable modifier keys (Shift, Alt, Ctrl) for key combinations

**Type**: `int` (bitflag enumeration)

**Fields**: N/A (bitflag constants)

**Modifier Constants**:

```go
const (
    ModNone  Modifier = 0
    ModShift Modifier = 1 << iota  // 0b001
    ModAlt                          // 0b010
    ModCtrl                         // 0b100
)
```

**Validation Rules**:
- Bitflags can be combined: `ModShift | ModCtrl`
- Zero value (ModNone) means no modifiers
- Invalid combinations filtered during parsing

**Operations**:
```go
// Check if Shift is pressed
if event.Modifiers & ModShift != 0 { ... }

// Set multiple modifiers
modifiers := ModShift | ModCtrl
```

**Relationships**:
- Referenced by Event.Modifiers
- Populated by escape sequence parser

### Event (Structure)

**Purpose**: Represents a single keyboard event with all associated metadata

**Type**: `struct`

**Fields**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| Key | Key | Normalized key code | Must be valid Key constant |
| Rune | rune | Printable character (if applicable) | 0 for non-printable keys |
| Modifiers | Modifier | Active modifier keys (bitflags) | Combination of ModShift/Alt/Ctrl |
| Timestamp | time.Time | Event capture time (monotonic) | Set by capture goroutine via time.Now() |
| Pressed | bool | true=key down, false=key up | Default true; false only on platforms with key-up support |
| Repeat | bool | true if OS autorepeat event | Default false; true for subsequent events of held key |

**Validation Rules**:
- Key must not be nil/invalid
- Timestamp must be set (non-zero)
- Rune populated only for printable keys (KeyA-KeyZ, Key0-Key9, etc.)
- Repeat can only be true if Pressed is true
- Modifiers only valid for certain key combinations

**Example Values**:

```go
// Ctrl+C press
Event{
    Key:       KeyCtrlC,
    Rune:      0,
    Modifiers: ModCtrl,
    Timestamp: time.Now(),
    Pressed:   true,
    Repeat:    false,
}

// Letter 'A' with Shift (capital A)
Event{
    Key:       KeyA,
    Rune:      'A',
    Modifiers: ModShift,
    Timestamp: time.Now(),
    Pressed:   true,
    Repeat:    false,
}

// Arrow Up autorepeat
Event{
    Key:       KeyUp,
    Rune:      0,
    Modifiers: ModNone,
    Timestamp: time.Now(),
    Pressed:   true,
    Repeat:    true,
}
```

**Relationships**:
- Produced by Backend.ReadEvent()
- Consumed via Input.Poll() or Input.Next()
- Used to update StateTracker

### Input (Interface)

**Purpose**: Primary public API for keyboard event capture

**Type**: `interface`

**Methods**:

| Method | Signature | Description | Error Conditions |
|--------|-----------|-------------|------------------|
| Start | `Start() error` | Initialize backend, start capture goroutine | Platform initialization failure, already started |
| Stop | `Stop()` | Restore terminal, stop capture goroutine | None (graceful, idempotent) |
| Poll | `Poll() (Event, bool)` | Block until next event or shutdown | Returns (zero, false) on shutdown |
| Next | `Next() *Event` | Non-blocking event retrieval | Returns nil if no event available |
| IsPressed | `IsPressed(k Key) bool` | Query current key state | None (returns false for unpressed) |

**Lifecycle**:
1. Create via `input.New()`
2. Call `Start()` to initialize
3. Poll or Next in event loop
4. Call `Stop()` before exit (or defer)

**Concurrency**:
- Start/Stop safe from any goroutine
- Poll blocks calling goroutine
- Next safe for concurrent calls
- IsPressed safe for concurrent reads

**Relationships**:
- Implemented by inputImpl (internal)
- Wraps Backend (internal)
- Uses StateTracker (internal)

### GameInput (Interface)

**Purpose**: High-level action mapping API for game development (optional)

**Type**: `interface`

**Methods**:

| Method | Signature | Description | Error Conditions |
|--------|-----------|-------------|------------------|
| Start | `Start() error` | Initialize underlying Input | Delegates to Input.Start() |
| Stop | `Stop()` | Cleanup | Delegates to Input.Stop() |
| IsActionPressed | `IsActionPressed(action string) bool` | Query if any bound key is pressed | Returns false for unbound actions |
| Bind | `Bind(action string, keys ...Key)` | Bind keys to action | None (replaces existing bindings) |

**Usage Example**:

```go
game := input.NewGameInput()
game.Bind("jump", input.KeySpace)
game.Bind("fire", input.KeySpace, input.KeyEnter)  // Multiple keys
game.Bind("move-left", input.KeyLeft, input.KeyA)

if game.IsActionPressed("jump") {
    player.Jump()
}
```

**Relationships**:
- Wraps Input interface
- Manages action→keys bindings (internal map)

## Internal Types (not exported)

### Backend (Interface - internal/backend)

**Purpose**: Platform-specific terminal I/O abstraction

**Type**: `interface`

**Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| Init | `Init() error` | Enter raw mode, save terminal state |
| Restore | `Restore() error` | Exit raw mode, restore original state |
| ReadEvent | `ReadEvent() (Event, error)` | Read and parse next event (blocking) |

**Implementations**:
- UnixBackend: termios-based (Linux, macOS, BSD)
- WindowsBackend: Console API-based (Windows)

### StateTracker (Structure - internal/state)

**Purpose**: Thread-safe key state tracking for IsPressed

**Type**: `struct`

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| mu | sync.RWMutex | Protects pressed map |
| pressed | map[Key]bool | Currently pressed keys |

**Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| Update | `Update(e Event)` | Update state from event (press/release) |
| IsPressed | `IsPressed(k Key) bool` | Query key state |

**Concurrency**: RWMutex allows concurrent reads, exclusive writes

### inputImpl (Structure - internal)

**Purpose**: Concrete implementation of Input interface

**Type**: `struct`

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| backend | Backend | Platform-specific I/O handler |
| events | chan Event | Buffered event queue (cap 100) |
| done | chan struct{} | Shutdown signal |
| state | *StateTracker | Key state tracker |
| once | sync.Once | Ensures single Start() |

**Methods**: Implements Input interface

### SequenceNode (Structure - internal/backend)

**Purpose**: Trie node for escape sequence parsing

**Type**: `struct`

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| children | map[byte]*SequenceNode | Next byte in sequence |
| key | Key | Matched key (if terminal node) |
| modifiers | Modifier | Modifiers for this sequence |

**Usage**: Build trie of escape sequences for efficient parsing

## State Transitions

### Input Lifecycle

```
[Uninitialized]
    ↓ New()
[Created]
    ↓ Start()
[Running] ←→ Poll()/Next()/IsPressed()
    ↓ Stop()
[Stopped]
    ↓ (optional) Start() again
[Running]
```

### Event Flow

```
[Terminal Input]
    ↓ Backend.ReadEvent()
[Raw Bytes]
    ↓ Parser (escape sequence → Key)
[Event Created]
    ↓ StateTracker.Update()
[State Updated]
    ↓ Channel Send
[Event Queue]
    ↓ Poll() or Next()
[Consumer Receives Event]
```

### Key State Tracking

```
Initial: pressed = {}

Press Event (Key=A, Pressed=true):
    pressed = {A: true}

Release Event (Key=A, Pressed=false):
    pressed = {}

Multiple Keys:
    Press A: pressed = {A: true}
    Press B: pressed = {A: true, B: true}
    Release A: pressed = {B: true}
```

## Validation Summary

| Entity | Critical Validations |
|--------|---------------------|
| Key | Must be valid constant, not out of range |
| Modifier | Must be valid bitflag combination |
| Event | Non-zero Timestamp, valid Key, Repeat implies Pressed |
| Input | Start() only once, Stop() idempotent |
| StateTracker | Thread-safe updates, accurate press/release tracking |

## Relationships Diagram

```
┌─────────────┐
│   Consumer  │
└──────┬──────┘
       │ Uses
       ↓
┌─────────────────────┐
│  Input (interface)  │
└──────┬──────────────┘
       │ Implemented by
       ↓
┌─────────────────────┐      ┌──────────────┐
│   inputImpl         │─────→│ StateTracker │
└──────┬──────────────┘      └──────────────┘
       │ Uses
       ↓
┌─────────────────────┐
│ Backend (interface) │
└──────┬──────────────┘
       │ Implemented by
       ↓
┌───────────────┬─────────────────┐
│  UnixBackend  │  WindowsBackend │
└───────────────┴─────────────────┘
       │                 │
       └────────┬────────┘
                ↓ Produces
         ┌─────────────┐
         │    Event    │
         └─────────────┘
              Contains
         ┌─────────────┐
         │     Key     │
         │  Modifier   │
         └─────────────┘
```

## Future Extensions

**Not included in initial implementation, reserved for later versions**:

- Mouse event support (Event.Mouse field)
- Resize event handling (Event.Resize field)
- Paste bracket mode (Event.Paste field)
- Focus events (Event.Focus field)
- Extended key set (F13-F24, multimedia keys)
- Configuration options (buffer size, timeout values)

These extensions maintain backward compatibility with current model.
