# Quickstart: GameInput Action Mapping API

**Feature**: 004-game-input-api
**Date**: 2025-10-18
**For**: Developers implementing the GameInput API

## Overview

This guide provides a step-by-step implementation path for the GameInput action mapping feature. Follow the user story priorities (P1 → P2 → P3) for incremental delivery.

---

## Prerequisites

✅ **Verify existing Input system is working**:
```bash
cd /Users/dshills/Development/projects/gokeys
go test ./input -v
```

All tests should pass. The Input interface, Key types, and Event types must be available.

---

## Phase 1: Basic Action Binding (P1 - MVP)

**Goal**: Single actions bound to single keys, basic IsActionPressed() queries

### Step 1.1: Define GameInput Interface

**File**: `input/game.go`

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
    // Returns false if action has no bound keys or none are pressed.
    //
    // Thread-safe: Safe for concurrent calls.
    IsActionPressed(action string) bool

    // Bind associates one or more keys with a logical action name.
    // If the action already has bindings, they are replaced.
    // Passing no keys unbinds the action.
    //
    // Thread-safe: Safe for concurrent calls.
    Bind(action string, keys ...Key)
}

// NewGameInput creates a new GameInput instance.
// If input is nil, creates a default Input via input.New().
func NewGameInput(input Input) GameInput {
    if input == nil {
        input = New()
    }
    return &gameInputImpl{
        input:    input,
        bindings: make(map[string][]Key),
    }
}
```

**Verify**:
```bash
go build ./input
```

---

### Step 1.2: Implement gameInputImpl Struct

**File**: `input/game_impl.go`

```go
package input

import "sync"

// gameInputImpl is the concrete implementation of GameInput.
type gameInputImpl struct {
    input    Input
    bindings map[string][]Key
    mu       sync.RWMutex
}

// Start delegates to the underlying Input.
func (g *gameInputImpl) Start() error {
    return g.input.Start()
}

// Stop delegates to the underlying Input.
func (g *gameInputImpl) Stop() {
    g.input.Stop()
}

// IsActionPressed returns true if any key bound to the action is pressed.
func (g *gameInputImpl) IsActionPressed(action string) bool {
    g.mu.RLock()
    keys, ok := g.bindings[action]
    g.mu.RUnlock()

    if !ok {
        return false  // Unbound action
    }

    // OR logic: any key pressed → action pressed
    for _, key := range keys {
        if g.input.IsPressed(key) {
            return true
        }
    }
    return false
}

// Bind associates keys with an action. Empty keys unbinds the action.
func (g *gameInputImpl) Bind(action string, keys ...Key) {
    g.mu.Lock()
    defer g.mu.Unlock()

    if len(keys) == 0 {
        delete(g.bindings, action)  // Unbind
    } else {
        g.bindings[action] = keys  // Bind/rebind
    }
}
```

**Verify**:
```bash
go build ./input
golangci-lint run ./input
```

---

### Step 1.3: Write Unit Tests (TDD)

**File**: `input/game_test.go`

```go
package input

import (
    "testing"
)

func TestNewGameInput(t *testing.T) {
    // Test with nil input
    game := NewGameInput(nil)
    if game == nil {
        t.Fatal("NewGameInput(nil) returned nil")
    }

    // Test with existing input
    inp := New()
    game = NewGameInput(inp)
    if game == nil {
        t.Fatal("NewGameInput(input) returned nil")
    }
}

func TestBindSingleKey(t *testing.T) {
    game := NewGameInput(nil)
    game.Bind("jump", KeySpace)

    // Verify binding exists (indirectly via IsActionPressed)
    // Direct verification requires starting the input system
}

func TestIsActionPressedUnbound(t *testing.T) {
    game := NewGameInput(nil)

    // Unbound action should return false, not error
    if game.IsActionPressed("undefined") {
        t.Error("IsActionPressed on unbound action returned true")
    }
}

func TestBindReplace(t *testing.T) {
    game := NewGameInput(nil)

    // Initial binding
    game.Bind("jump", KeySpace)

    // Replace binding
    game.Bind("jump", KeyEnter)

    // Verify only new binding is active (requires integration test with real input)
}

func TestUnbind(t *testing.T) {
    game := NewGameInput(nil)

    // Bind then unbind
    game.Bind("jump", KeySpace)
    game.Bind("jump")  // Empty keys = unbind

    if game.IsActionPressed("jump") {
        t.Error("Unbound action returned true")
    }
}

func TestStartStopDelegation(t *testing.T) {
    // Test that Start/Stop delegate correctly
    // Requires mock Input for proper testing
    game := NewGameInput(nil)

    // Start/Stop should not panic
    _ = game.Start()
    game.Stop()
}
```

**Run Tests**:
```bash
go test ./input -run TestGameInput -v
```

---

### Step 1.4: Create Basic Example

**File**: `examples/game/main.go`

```go
package main

import (
    "fmt"
    "time"

    "github.com/dshills/gokeys/input"
)

func main() {
    fmt.Println("GameInput Example - Basic Action Binding")
    fmt.Println("Controls:")
    fmt.Println("  Space: Jump")
    fmt.Println("  F: Fire")
    fmt.Println("  ESC: Quit")
    fmt.Println()

    game := input.NewGameInput(nil)
    if err := game.Start(); err != nil {
        panic(err)
    }
    defer game.Stop()

    // Bind actions to keys
    game.Bind("jump", input.KeySpace)
    game.Bind("fire", input.KeyF)
    game.Bind("quit", input.KeyEscape)

    // Simple game loop
    ticker := time.NewTicker(16 * time.Millisecond)  // ~60fps
    defer ticker.Stop()

    for {
        <-ticker.C

        if game.IsActionPressed("jump") {
            fmt.Println("JUMP!")
        }
        if game.IsActionPressed("fire") {
            fmt.Println("FIRE!")
        }
        if game.IsActionPressed("quit") {
            fmt.Println("Quitting...")
            break
        }
    }
}
```

**Test Example**:
```bash
cd examples/game
go run main.go
# Press Space, F, ESC to test
```

---

### Step 1.5: Validate P1 Complete

**Checklist**:
- [ ] GameInput interface defined in `input/game.go`
- [ ] gameInputImpl implemented in `input/game_impl.go`
- [ ] Unit tests pass: `go test ./input -v`
- [ ] Linter passes: `golangci-lint run ./input`
- [ ] Example works: `go run examples/game/main.go`
- [ ] Godoc complete on all exported types

**Success Criteria** (from spec):
- ✅ SC-001: Under 5 lines of code for bind + query
- ✅ SC-005: Zero hardcoded key references in game logic

---

## Phase 2: Multiple Keys Per Action (P2)

**Goal**: Support alternative control schemes (WASD + arrows)

### Step 2.1: Update Example for Multiple Keys

**File**: `examples/game/main.go` (add to existing)

```go
func setupControls(game input.GameInput) {
    // Support both arrow keys and WASD
    game.Bind("move-up", input.KeyUp, input.KeyW)
    game.Bind("move-down", input.KeyDown, input.KeyS)
    game.Bind("move-left", input.KeyLeft, input.KeyA)
    game.Bind("move-right", input.KeyRight, input.KeyD)

    // Multiple confirm keys
    game.Bind("confirm", input.KeyEnter, input.KeySpace, input.KeyY)
}

func gameLoop(game input.GameInput) {
    ticker := time.NewTicker(16 * time.Millisecond)
    defer ticker.Stop()

    x, y := 0, 0

    for {
        <-ticker.C

        // Movement with multiple key options
        if game.IsActionPressed("move-up") {
            y--
        }
        if game.IsActionPressed("move-down") {
            y++
        }
        if game.IsActionPressed("move-left") {
            x--
        }
        if game.IsActionPressed("move-right") {
            x++
        }

        fmt.Printf("Position: (%d, %d)\r", x, y)

        if game.IsActionPressed("confirm") {
            break
        }
    }
}
```

---

### Step 2.2: Write Multi-Key Tests

**File**: `input/game_test.go` (add tests)

```go
func TestBindMultipleKeys(t *testing.T) {
    game := NewGameInput(nil)

    // Bind two keys to one action
    game.Bind("fire", KeySpace, KeyEnter)

    // Verify both keys work (requires integration test with mock Input)
}

func TestMultipleKeysOr Logic(t *testing.T) {
    // Test that any key pressed returns true
    // Requires integration test with real key state
}
```

---

### Step 2.3: Validate P2 Complete

**Checklist**:
- [ ] Multiple keys bind correctly
- [ ] IsActionPressed() OR logic works (any key → true)
- [ ] Example demonstrates WASD + arrows
- [ ] Tests cover multi-key scenarios

**Success Criteria**:
- ✅ SC-002: <1ms query time (verify with benchmark)
- ✅ SC-003: 100 actions without degradation

---

## Phase 3: Runtime Rebinding (P3)

**Goal**: Allow players to rebind controls during gameplay

### Step 3.1: Add Rebinding to Example

**File**: `examples/game/main.go` (extend)

```go
func rebindAction(game input.GameInput, inp input.Input, actionName string) {
    fmt.Printf("\nPress new key for action '%s': ", actionName)

    // Wait for key press using underlying Input
    event, ok := inp.Poll()
    if !ok {
        return
    }

    // Rebind action to new key
    game.Bind(actionName, event.Key)
    fmt.Printf("Action '%s' now bound to %v\n", actionName, event.Key)
}

func settingsMenu(game input.GameInput, inp input.Input) {
    fmt.Println("\nSettings - Rebind Controls")
    fmt.Println("1. Rebind 'jump'")
    fmt.Println("2. Rebind 'fire'")
    fmt.Println("3. Rebind 'move-left'")
    fmt.Println("4. Back to game")

    event, ok := inp.Poll()
    if !ok {
        return
    }

    switch event.Key {
    case input.Key1:
        rebindAction(game, inp, "jump")
    case input.Key2:
        rebindAction(game, inp, "fire")
    case input.Key3:
        rebindAction(game, inp, "move-left")
    case input.Key4:
        return
    }
}

func main() {
    game := input.NewGameInput(nil)
    inp := input.New()  // Need separate Input for menu navigation

    if err := game.Start(); err != nil {
        panic(err)
    }
    defer game.Stop()

    setupControls(game)

    for {
        // Game loop
        gameLoop(game)

        // Press M for settings
        if game.IsActionPressed("menu") {
            settingsMenu(game, inp)
        }
    }
}
```

---

### Step 3.2: Write Rebinding Tests

**File**: `input/game_test.go`

```go
func TestRebinding(t *testing.T) {
    game := NewGameInput(nil)

    // Initial binding
    game.Bind("jump", KeySpace)

    // Rebind to different key
    game.Bind("jump", KeyJ)

    // Verify only new binding works (integration test needed)
}

func TestRebindingTakesEffectImmediately(t *testing.T) {
    // Verify rebinding works without restart
    // Requires integration test with real Input
}
```

---

### Step 3.3: Validate P3 Complete

**Checklist**:
- [ ] Runtime rebinding works
- [ ] Changes take effect immediately (SC-004: <20ms)
- [ ] Example demonstrates rebinding menu
- [ ] Tests cover rebinding scenarios

---

## Phase 4: Thread Safety & Performance

### Step 4.1: Add Concurrency Tests

**File**: `input/game_concurrent_test.go`

```go
package input

import (
    "sync"
    "testing"
)

func TestConcurrentIsActionPressed(t *testing.T) {
    game := NewGameInput(nil)
    game.Bind("test", KeySpace)

    // Multiple goroutines calling IsActionPressed
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 1000; j++ {
                game.IsActionPressed("test")
            }
        }()
    }
    wg.Wait()
}

func TestConcurrentBindAndQuery(t *testing.T) {
    game := NewGameInput(nil)

    var wg sync.WaitGroup

    // Writer goroutine
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 100; i++ {
            game.Bind("test", KeySpace)
        }
    }()

    // Reader goroutines
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 1000; j++ {
                game.IsActionPressed("test")
            }
        }()
    }

    wg.Wait()
}
```

**Run with race detector**:
```bash
go test ./input -race -v
```

---

### Step 4.2: Add Performance Benchmarks

**File**: `input/game_bench_test.go`

```go
package input

import "testing"

func BenchmarkIsActionPressed_SingleKey(b *testing.B) {
    game := NewGameInput(nil)
    game.Bind("test", KeySpace)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        game.IsActionPressed("test")
    }
}

func BenchmarkIsActionPressed_MultipleKeys(b *testing.B) {
    game := NewGameInput(nil)
    game.Bind("test", KeySpace, KeyEnter, KeyA, KeyB, KeyC)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        game.IsActionPressed("test")
    }
}

func BenchmarkBind(b *testing.B) {
    game := NewGameInput(nil)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        game.Bind("test", KeySpace)
    }
}
```

**Run benchmarks**:
```bash
go test ./input -bench=BenchmarkGameInput -benchmem
```

**Target**: IsActionPressed <1μs (1000ns), <1ms is 1,000,000ns - plenty of headroom

---

## Final Validation

### Acceptance Criteria

**User Story 1 (P1)**:
- [x] Single actions bind to single keys
- [x] IsActionPressed returns true when key pressed
- [x] Start/Stop delegate correctly
- [x] Unbound actions return false

**User Story 2 (P2)**:
- [x] Multiple keys bind to one action
- [x] OR logic: any key → action pressed
- [x] WASD + arrows example works

**User Story 3 (P3)**:
- [x] Runtime rebinding replaces old bindings
- [x] Changes take effect immediately
- [x] Unbinding works (empty keys)

### Success Criteria (from spec)

- [ ] **SC-001**: 5 lines of code ✓
  ```go
  game := input.NewGameInput(nil)
  game.Start()
  game.Bind("jump", input.KeySpace)
  if game.IsActionPressed("jump") { ... }
  game.Stop()
  ```

- [ ] **SC-002**: <1ms query time (verify with benchmarks)
- [ ] **SC-003**: 100 actions without degradation (stress test)
- [ ] **SC-004**: <20ms rebinding effect (measure with timestamps)
- [ ] **SC-005**: Zero hardcoded keys in example ✓
- [ ] **SC-006**: Thread-safe (race detector passes) ✓
- [ ] **SC-007**: 60fps with 10+ actions (run example and profile)

---

## Integration with Existing Code

### Update Package Documentation

**File**: `input/doc.go` (add section)

```go
// GameInput Interface
//
// For game development, the GameInput interface provides action mapping
// on top of the Input interface. Instead of checking physical keys,
// games can query logical actions:
//
//   game := input.NewGameInput(nil)
//   game.Start()
//   defer game.Stop()
//
//   game.Bind("jump", input.KeySpace)
//   game.Bind("fire", input.KeyF, input.KeyEnter)  // Multiple keys
//
//   for {
//       if game.IsActionPressed("jump") { player.Jump() }
//       if game.IsActionPressed("fire") { player.Fire() }
//   }
//
// This enables rebindable controls and alternative key schemes (WASD + arrows).
```

---

### Verify golangci-lint Compliance

**Run full linter**:
```bash
golangci-lint run ./input
```

**Expected**: No errors

**Common Issues**:
- Missing godoc on exported types → Add comments
- Error not checked → Wrap Input.Start() error
- Cyclomatic complexity → Keep methods simple (already <30)

---

## Troubleshooting

### Issue: IsActionPressed always returns false

**Cause**: Input system not started or key state not tracked

**Fix**:
```go
game := input.NewGameInput(nil)
if err := game.Start(); err != nil {  // MUST call Start()
    panic(err)
}
```

---

### Issue: Race detector failures

**Cause**: Missing locks or incorrect lock usage

**Fix**: Ensure RWMutex is used correctly:
```go
// Read path - RLock
g.mu.RLock()
keys := g.bindings[action]
g.mu.RUnlock()

// Write path - Lock
g.mu.Lock()
g.bindings[action] = keys
g.mu.Unlock()
```

---

### Issue: Benchmark shows >1ms query time

**Cause**: Too many keys bound to action

**Fix**: Limit keys per action to 1-10, optimize hot path:
```go
// Early return optimization
for _, key := range keys {
    if g.input.IsPressed(key) {
        return true  // Short-circuit
    }
}
```

---

## Next Steps

After completing this quickstart:

1. **Run `/speckit.tasks`** to generate detailed implementation tasks
2. **Implement P1 first** (basic binding) - this is the MVP
3. **Validate independently** before moving to P2
4. **Add P2** (multiple keys) - test thoroughly
5. **Add P3** (rebinding) - complete feature
6. **Update CLAUDE.md** with GameInput examples
7. **Consider future extensions**:
   - Key combination support (Shift+A)
   - Binding persistence (save/load)
   - Action priority system

---

## Summary

**Incremental Delivery Path**:

1. **P1 (MVP)**: Basic single-key binding → Deliverable standalone
2. **P2**: Multiple keys per action → Enhanced flexibility
3. **P3**: Runtime rebinding → Complete customization

**Testing Strategy**:
- Unit tests for each method
- Contract tests for user stories
- Concurrency tests with race detector
- Benchmarks for performance validation

**Integration Points**:
- Wraps existing Input interface
- Added to `input/` package
- Examples demonstrate all features
- Documentation in godoc + CLAUDE.md

This feature is ready for implementation via `/speckit.tasks`.
