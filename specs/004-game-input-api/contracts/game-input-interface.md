# Contract: GameInput Interface

**Feature**: 004-game-input-api
**Package**: `input`
**Type**: Interface
**Purpose**: High-level action mapping API for game development

**Note**: This contract is defined in the original input system specification at `specs/001-input-system/contracts/game-input-interface.md`. This feature implements that contract.

## Interface Definition

```go
package input

// GameInput provides a higher-level abstraction over Input for game development.
// It allows mapping logical actions (e.g., "jump", "fire") to physical keys,
// enabling key rebinding and simplifying game input logic.
type GameInput interface {
    // Start initializes the underlying Input system.
    // Delegates to the wrapped Input.Start().
    //
    // Returns error if underlying Input fails to start.
    Start() error

    // Stop cleans up and restores terminal state.
    // Delegates to the wrapped Input.Stop().
    Stop()

    // IsActionPressed returns true if any key bound to the action is currently pressed.
    // Returns false if:
    //   - Action has no bound keys
    //   - Action name not recognized
    //   - None of the bound keys are pressed
    //
    // Thread-safe: Safe for concurrent calls.
    IsActionPressed(action string) bool

    // Bind associates one or more keys with a logical action name.
    // If the action already has bindings, they are replaced.
    // Multiple keys can be bound to the same action (e.g., "jump" → Space, Enter).
    //
    // Thread-safe: Safe for concurrent calls.
    Bind(action string, keys ...Key)
}
```

## Factory Function

```go
// NewGameInput creates a new GameInput instance.
// If input is nil, creates a default Input via input.New().
//
// The provided Input should not be started yet - GameInput.Start()
// will initialize it.
func NewGameInput(input Input) GameInput
```

## Method Contracts

### Start()

**Preconditions**:
- GameInput instance created via NewGameInput()
- Underlying Input not yet started (or Start() not yet called)

**Postconditions**:
- Underlying Input system started
- Action bindings ready to use
- IsActionPressed() queries will work correctly

**Error Cases**:
- Returns error if underlying Input.Start() fails (terminal initialization errors)
- Returns error if already started (idempotency check)

**Delegation**:
```go
func (g *gameInputImpl) Start() error {
    return g.input.Start()
}
```

---

### Stop()

**Preconditions**: None (idempotent)

**Postconditions**:
- Underlying Input stopped
- Terminal restored to original state
- Action bindings preserved in memory (can Start() again)

**Error Cases**: None (void return)

**Delegation**:
```go
func (g *gameInputImpl) Stop() {
    g.input.Stop()
}
```

---

### IsActionPressed(action string)

**Preconditions**:
- Start() successfully called
- Action name is any string (including empty string "")

**Behavior**:
```go
// Pseudocode
mu.RLock()
keys := bindings[action]
mu.RUnlock()

for _, key := range keys {
    if input.IsPressed(key) {
        return true  // Short-circuit: any key pressed → action pressed
    }
}
return false
```

**Return Values**:
- `true`: At least one key bound to action is currently pressed
- `false`: No keys bound, or all bound keys are released

**Lookup Semantics**:
- **Case-sensitive**: "jump" ≠ "Jump" (FR-005)
- **Empty string valid**: "" is a valid action name
- **Unbound action**: Returns false (not error) (FR-007)
- **Multiple keys**: OR relationship - any key triggers (FR-004)

**Performance**:
- O(k) where k = number of keys bound to action (typically 1-3)
- RLock allows concurrent calls from multiple goroutines
- Target: <1ms per call (SC-002)

**Thread Safety**: Safe for concurrent calls (RWMutex read lock)

**Examples**:
```go
// Single key
game.Bind("jump", input.KeySpace)
pressed := game.IsActionPressed("jump")  // true if Space held

// Multiple keys (OR)
game.Bind("fire", input.KeySpace, input.KeyEnter)
pressed := game.IsActionPressed("fire")  // true if Space OR Enter held

// Unbound action
pressed := game.IsActionPressed("undefined")  // false (not an error)
```

---

### Bind(action string, keys ...Key)

**Preconditions**:
- None (can be called before or after Start())

**Postconditions**:
- Action associated with provided keys (replaces previous bindings)
- If keys is empty, action is unbound (removed from map)
- Binding takes effect immediately for subsequent IsActionPressed() calls

**Parameters**:
- `action`: Case-sensitive action name (any string, including "")
- `keys`: Variadic list of Key values (0-N keys)

**Behavior**:
```go
mu.Lock()
if len(keys) == 0 {
    delete(bindings, action)  // Unbind
} else {
    bindings[action] = keys   // Bind/rebind
}
mu.Unlock()
```

**Semantics**:
- **Replace**: Existing bindings are completely replaced (FR-006)
- **Unbind**: Empty keys list removes action (FR-012)
- **Multiple keys**: All keys trigger action via OR (FR-004)
- **Duplicates**: Not deduplicated (harmless)
- **Same key, multiple actions**: Allowed (FR-008)

**Thread Safety**: Safe for concurrent calls (RWMutex write lock)

**Performance**: O(1) map operation

**Examples**:
```go
// Bind single key
game.Bind("jump", input.KeySpace)

// Bind multiple keys (WASD + arrows)
game.Bind("move-left", input.KeyLeft, input.KeyA)
game.Bind("move-right", input.KeyRight, input.KeyD)

// Rebind (replace)
game.Bind("jump", input.KeySpace)  // Initially Space
game.Bind("jump", input.KeyJ)      // Now only J (Space no longer works)

// Unbind
game.Bind("jump")  // No keys = remove binding
```

---

## Usage Examples

### Basic Game Loop (User Story 1 - P1)

```go
package main

import (
    "github.com/dshills/gokeys/input"
)

func main() {
    game := input.NewGameInput(nil)  // Creates default Input
    if err := game.Start(); err != nil {
        panic(err)
    }
    defer game.Stop()

    // Bind actions
    game.Bind("jump", input.KeySpace)
    game.Bind("fire", input.KeyF)
    game.Bind("quit", input.KeyEscape)

    // Game loop
    for {
        if game.IsActionPressed("jump") {
            player.Jump()
        }
        if game.IsActionPressed("fire") {
            player.Fire()
        }
        if game.IsActionPressed("quit") {
            break
        }

        update()
        render()
    }
}
```

---

### Multiple Keys Per Action (User Story 2 - P2)

```go
func setupControls(game input.GameInput) {
    // Support both arrow keys and WASD
    game.Bind("move-up", input.KeyUp, input.KeyW)
    game.Bind("move-down", input.KeyDown, input.KeyS)
    game.Bind("move-left", input.KeyLeft, input.KeyA)
    game.Bind("move-right", input.KeyRight, input.KeyD)

    // Multiple confirm keys for accessibility
    game.Bind("confirm", input.KeyEnter, input.KeySpace, input.KeyY)
    game.Bind("cancel", input.KeyEscape, input.KeyBackspace, input.KeyN)
}
```

---

### Runtime Rebinding (User Story 3 - P3)

```go
func rebindAction(game input.GameInput, inp input.Input, actionName string) {
    fmt.Printf("Press new key for action '%s'...\n", actionName)

    // Wait for key press
    event, ok := inp.Poll()
    if !ok {
        return
    }

    // Rebind action to new key
    game.Bind(actionName, event.Key)
    fmt.Printf("Action '%s' now bound to %v\n", actionName, event.Key)
}

func settingsMenu(game input.GameInput, inp input.Input) {
    actions := []string{"jump", "fire", "move-left", "move-right"}

    for i, action := range actions {
        fmt.Printf("%d. Rebind '%s'\n", i+1, action)
    }

    event, ok := inp.Poll()
    if !ok {
        return
    }

    // User selected action by number key
    if event.Key >= input.Key1 && event.Key <= input.Key4 {
        idx := int(event.Key - input.Key1)
        rebindAction(game, inp, actions[idx])
    }
}
```

---

## Design Rationale

### Why Action Mapping?

**Problem**: Hardcoded key checks scatter game logic:
```go
// BAD: Tight coupling to physical keys
if input.IsPressed(KeySpace) {
    player.Jump()
}
```

**Issues**:
- Cannot rebind controls
- No support for alternate keys (WASD vs arrows)
- Game logic tied to input implementation

**Solution**: Decouple actions from keys:
```go
// GOOD: Logical actions decoupled from keys
if game.IsActionPressed("jump") {
    player.Jump()
}
```

**Benefits**:
- Key rebinding support (FR-017)
- Multiple keys per action (FR-004, FR-016)
- Game logic independent of input (testable)

---

### Why Composition Over Embedding?

**Decision**: GameInput wraps Input, does not embed it

**Rationale**:
- GameInput "has-a" Input (composition)
- Avoids exposing Poll()/Next() through GameInput (API confusion)
- Clear separation: Input = events, GameInput = actions

---

### Why Variadic Bind()?

**Decision**: `Bind(action string, keys ...Key)` instead of `Bind(action string, keys []Key)`

**Benefits**:
- Natural syntax: `Bind("jump", KeySpace)` vs `Bind("jump", []Key{KeySpace})`
- Supports zero keys for unbinding: `Bind("jump")`
- Consistent with Go idioms (fmt.Printf, append, etc.)

---

## Performance Characteristics

| Operation | Complexity | Blocking | Thread-Safe |
|-----------|------------|----------|-------------|
| Start() | O(1) | No | Yes |
| Stop() | O(1) | No | Yes |
| IsActionPressed() | O(k) | No | Yes (concurrent reads) |
| Bind() | O(1) | No | Yes (blocks reads) |

**Notes**:
- k = keys per action (typically 1-3, max ~10)
- RWMutex enables concurrent IsActionPressed calls
- Bind blocks briefly but is rare (initialization/rebinding)

---

## Relationship to Input Interface

```
┌─────────────────────┐
│   GameInput         │  High-level: Action queries
│   (wrapper)         │  Methods: Bind, IsActionPressed
└──────────┬──────────┘
           │ Wraps (composition)
           ↓
┌─────────────────────┐
│   Input             │  Low-level: Event stream, key state
│   (interface)       │  Methods: Poll, Next, IsPressed
└─────────────────────┘
```

**When to Use**:
- **Input**: CLI tools, direct key handling, full control
- **GameInput**: Games, rebindable controls, action-based logic

---

## Thread Safety Guarantees (FR-011)

**Concurrent Operations**:
- ✅ Multiple IsActionPressed() calls (RLock allows parallel reads)
- ✅ Bind() + IsActionPressed() (Write lock blocks reads temporarily)
- ✅ Start()/Stop() from any goroutine (delegates to Input)

**Lock Strategy**:
- RWMutex protects bindings map
- Short critical sections (map access only)
- No locks held during Input.IsPressed() calls

**Typical Pattern**:
```go
// Game loop goroutine
for {
    if game.IsActionPressed("jump") { ... }  // RLock
    if game.IsActionPressed("fire") { ... }  // RLock
}

// UI thread
game.Bind("jump", newKey)  // Lock (blocks reads briefly)
```

---

## Version Compatibility

**Current Version**: 1.0.0 (initial release)

**Stability**: Stable interface - no breaking changes planned

**Future Extensions** (backward compatible):
- Key combination support (e.g., "Shift+A")
- Action priority/overrides
- Event callbacks for actions
- Binding persistence (save/load)

All extensions will be additive (new methods/types) without breaking existing code.

---

## Reference

**Original Contract**: `specs/001-input-system/contracts/game-input-interface.md`

This implementation contract is consistent with the original design and adds implementation-specific details from research and data model phases.
