# Contract: GameInput Interface

**Package**: `input`
**Type**: Interface
**Purpose**: High-level action mapping API for game development

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

## Method Contracts

### Start()

**Preconditions**: Same as Input.Start()

**Postconditions**:
- Underlying Input system started
- Action bindings ready to use
- IsActionPressed() queries will work

**Error Cases**: Same as Input.Start()

**Delegation**:
```go
func (g *GameInput) Start() error {
    return g.input.Start()
}
```

### Stop()

**Preconditions**: None (idempotent)

**Postconditions**:
- Underlying Input stopped
- Terminal restored
- Action bindings preserved (can Start() again)

**Delegation**:
```go
func (g *GameInput) Stop() {
    g.input.Stop()
}
```

### IsActionPressed()

**Preconditions**:
- Start() successfully called
- Action name is string (any value accepted)

**Behavior**:
```go
// Check each bound key via underlying Input.IsPressed()
for _, key := range bindings[action] {
    if input.IsPressed(key) {
        return true
    }
}
return false
```

**Lookup Semantics**:
- Case-sensitive action names ("jump" ≠ "Jump")
- Empty string action valid (can bind to empty string)
- Unbound action returns false (not error)

**Performance**:
- O(n) where n = number of keys bound to action
- Typically n < 5, so effectively O(1)
- Each IsPressed() call is O(1)

**Concurrency**:
- RWMutex read lock during binding lookup
- No blocking under normal load
- Safe with concurrent Bind() calls

### Bind()

**Preconditions**: None

**Behavior**:
- Replaces existing bindings for action
- Empty keys list clears bindings
- Duplicate keys in list deduplicated
- Order of keys does not matter

**Examples**:
```go
// Single key binding
game.Bind("jump", input.KeySpace)

// Multiple keys (OR relationship)
game.Bind("fire", input.KeySpace, input.KeyEnter, input.KeyCtrlF)

// Unbind action
game.Bind("jump") // No keys = unbind

// Replace binding
game.Bind("jump", input.KeySpace)  // Initially Space
game.Bind("jump", input.KeyEnter)  // Now only Enter
```

**Concurrency**:
- RWMutex write lock during update
- Atomic replacement of binding list
- Safe with concurrent IsActionPressed() calls

## Usage Examples

### Basic Game Controls

```go
package main

import (
    "github.com/dshills/gokeys/input"
)

func main() {
    game := input.NewGameInput()
    if err := game.Start(); err != nil {
        panic(err)
    }
    defer game.Stop()

    // Bind actions to keys
    game.Bind("jump", input.KeySpace)
    game.Bind("fire", input.KeySpace, input.KeyEnter)
    game.Bind("move-left", input.KeyLeft, input.KeyA)
    game.Bind("move-right", input.KeyRight, input.KeyD)
    game.Bind("pause", input.KeyEscape, input.KeyP)

    // Game loop
    for {
        if game.IsActionPressed("jump") {
            player.Jump()
        }
        if game.IsActionPressed("fire") {
            player.Fire()
        }
        if game.IsActionPressed("move-left") {
            player.Move(-1, 0)
        }
        if game.IsActionPressed("move-right") {
            player.Move(1, 0)
        }
        if game.IsActionPressed("pause") {
            pauseMenu()
        }

        update()
        render()
    }
}
```

### Key Rebinding Menu

```go
package main

import (
    "fmt"
    "github.com/dshills/gokeys/input"
)

func rebindAction(game input.GameInput, action string) {
    fmt.Printf("Press new key for action '%s'...\n", action)

    event, ok := game.input.Poll() // Access underlying Input
    if !ok {
        return
    }

    game.Bind(action, event.Key)
    fmt.Printf("Action '%s' now bound to %v\n", action, event.Key)
}

func showRebindMenu(game input.GameInput) {
    actions := []string{"jump", "fire", "move-left", "move-right"}

    for i, action := range actions {
        fmt.Printf("%d. Rebind '%s'\n", i+1, action)
    }

    event, ok := game.input.Poll()
    if !ok {
        return
    }

    // User selected action based on number key
    if event.Key >= input.Key1 && event.Key <= input.Key4 {
        idx := int(event.Key - input.Key1)
        rebindAction(game, actions[idx])
    }
}
```

### Multiple Key Alternatives

```go
package main

import (
    "github.com/dshills/gokeys/input"
)

func main() {
    game := input.NewGameInput()
    if err := game.Start(); err != nil {
        panic(err)
    }
    defer game.Stop()

    // Support both arrow keys and WASD
    game.Bind("move-up", input.KeyUp, input.KeyW)
    game.Bind("move-down", input.KeyDown, input.KeyS)
    game.Bind("move-left", input.KeyLeft, input.KeyA)
    game.Bind("move-right", input.KeyRight, input.KeyD)

    // Support multiple confirm keys
    game.Bind("confirm", input.KeyEnter, input.KeySpace, input.KeyY)

    // Support multiple cancel keys
    game.Bind("cancel", input.KeyEscape, input.KeyBackspace, input.KeyN)

    // Game loop
    for {
        dx, dy := 0, 0
        if game.IsActionPressed("move-up") {
            dy = -1
        }
        if game.IsActionPressed("move-down") {
            dy = 1
        }
        if game.IsActionPressed("move-left") {
            dx = -1
        }
        if game.IsActionPressed("move-right") {
            dx = 1
        }

        if dx != 0 || dy != 0 {
            player.Move(dx, dy)
        }

        update()
        render()
    }
}
```

### Action Groups

```go
package main

import (
    "github.com/dshills/gokeys/input"
)

type Controls struct {
    game input.GameInput
}

func (c *Controls) SetupMovement() {
    c.game.Bind("move-up", input.KeyUp, input.KeyW)
    c.game.Bind("move-down", input.KeyDown, input.KeyS)
    c.game.Bind("move-left", input.KeyLeft, input.KeyA)
    c.game.Bind("move-right", input.KeyRight, input.KeyD)
}

func (c *Controls) SetupActions() {
    c.game.Bind("jump", input.KeySpace)
    c.game.Bind("fire", input.KeyF, input.KeyCtrlF)
    c.game.Bind("reload", input.KeyR)
}

func (c *Controls) SetupUI() {
    c.game.Bind("menu", input.KeyEscape)
    c.game.Bind("inventory", input.KeyI, input.KeyTab)
    c.game.Bind("map", input.KeyM)
}

func main() {
    game := input.NewGameInput()
    if err := game.Start(); err != nil {
        panic(err)
    }
    defer game.Stop()

    controls := &Controls{game: game}
    controls.SetupMovement()
    controls.SetupActions()
    controls.SetupUI()

    // Game loop using action names
    for {
        if game.IsActionPressed("jump") {
            player.Jump()
        }
        // ... etc
    }
}
```

## Design Rationale

### Why Action Mapping?

**Problem**: Games hardcoding specific keys:
```go
// Bad: Hardcoded keys
if input.IsPressed(KeySpace) {
    player.Jump()
}
if input.IsPressed(KeyF) {
    player.Fire()
}
```

**Issues**:
- Cannot rebind controls
- No support for alternate keys (e.g., WASD vs arrows)
- Game logic coupled to physical keys

**Solution**: Action mapping decouples:
```go
// Good: Logical actions
if game.IsActionPressed("jump") {
    player.Jump()
}
if game.IsActionPressed("fire") {
    player.Fire()
}
```

**Benefits**:
- Key rebinding support
- Multiple keys per action
- Game logic independent of input
- Easier to reason about game behavior

### Why Multiple Keys Per Action?

**Use Case 1**: Alternative control schemes
```go
// Support both arrow keys and WASD
game.Bind("move-left", input.KeyLeft, input.KeyA)
```

**Use Case 2**: Accessibility
```go
// Multiple ways to confirm
game.Bind("confirm", input.KeyEnter, input.KeySpace, input.KeyY)
```

**Use Case 3**: Convenience
```go
// Exit game multiple ways
game.Bind("quit", input.KeyEscape, input.KeyCtrlC, input.KeyQ)
```

## Performance Characteristics

| Operation | Time Complexity | Blocking | Thread-Safe |
|-----------|----------------|----------|-------------|
| Start() | O(1) | No | Yes |
| Stop() | O(1) | No | Yes |
| IsActionPressed() | O(k) where k=keys bound | No | Yes |
| Bind() | O(k) where k=keys bound | No | Yes |

**Typical Values**:
- k (keys per action): 1-5
- Effective complexity: O(1) for all operations

## Relationship to Input Interface

```
┌──────────────────┐
│   GameInput      │
│   (higher-level) │
└────────┬─────────┘
         │ Wraps
         ↓
┌──────────────────┐
│   Input          │
│   (lower-level)  │
└──────────────────┘
```

**Access to Underlying Input**:
```go
type GameInput interface {
    // ... GameInput methods

    // If needed, expose underlying Input
    Input() Input
}
```

**When to Use Each**:
- **Input**: CLI tools, simple key detection, full control
- **GameInput**: Games, applications with rebindable controls, logical actions

## Version Compatibility

**Current Version**: 1.0.0

**Stability**: Stable interface, no breaking changes in 1.x

**Future Extensions** (backward compatible):
- Action priority (e.g., "jump" overrides "fire")
- Key combination support (e.g., "Shift+A")
- Action event callbacks
- Binding persistence (save/load config)

All extensions will be optional and maintain backward compatibility.
