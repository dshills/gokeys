# Data Model: GameInput Action Mapping API

**Feature**: 004-game-input-api
**Date**: 2025-10-18
**Input**: [spec.md](./spec.md), [research.md](./research.md)

## Overview

The GameInput API has a minimal data model consisting of action bindings stored in memory. There are no persistent entities or database schemas - all state is transient and managed in the gameInputImpl struct.

---

## Core Entities

### GameInput Interface

**Purpose**: Public API for action mapping

**Methods**:
- `Start() error` - Initialize underlying Input system
- `Stop()` - Shutdown and restore terminal
- `Bind(action string, keys ...Key)` - Associate keys with action name
- `IsActionPressed(action string) bool` - Query if any bound key is pressed

**Relationships**:
- Wraps one Input interface instance (composition)
- Maintains binding map internally (encapsulated)

**Lifecycle**:
1. Created via NewGameInput(input)
2. Start() delegates to wrapped Input
3. Bind() calls update internal map
4. IsActionPressed() queries during runtime
5. Stop() delegates to wrapped Input

---

## Internal State (gameInputImpl)

### Binding Map

**Type**: `map[string][]Key`

**Purpose**: Maps action names to lists of keys that trigger that action

**Structure**:
```go
{
  "jump":       [KeySpace],
  "fire":       [KeySpace, KeyEnter],
  "move-left":  [KeyLeft, KeyA],
  "move-right": [KeyRight, KeyD],
  "move-up":    [KeyUp, KeyW],
  "move-down":  [KeyDown, KeyS]
}
```

**Constraints**:
- **Key**: Action name (case-sensitive string)
- **Value**: Slice of 0-10 Key values (typically 1-3)
- **Size**: Support 100+ unique action names
- **Empty entry**: Action with no keys is removed from map

**Access Patterns**:
- **Write**: Bind() replaces entire slice for action (rare, during initialization or rebinding)
- **Read**: IsActionPressed() looks up action and iterates keys (frequent, every frame)

**Thread Safety**: Protected by sync.RWMutex
- Read lock for IsActionPressed() lookups
- Write lock for Bind() updates

---

## State Transitions

### Action Binding Lifecycle

```
[Unbound] --Bind(action, keys...)--> [Bound]
[Bound]   --Bind(action, newKeys)--> [Bound] (updated)
[Bound]   --Bind(action)-----------→ [Unbound] (removed)
```

**States**:
- **Unbound**: Action not in map, IsActionPressed returns false
- **Bound**: Action in map with 1+ keys, IsActionPressed checks key states

**Transitions**:
- Bind with keys → add/update map entry
- Bind with no keys → delete map entry

---

## Validation Rules

### Action Names (FR-005)

**Rule**: Case-sensitive strings, no normalization

**Valid**:
- "jump", "fire", "move-left"
- "confirm", "cancel", "menu"
- Empty string "" (treated as valid action name)

**Invalid**: None - all strings accepted

**Rationale**: Developers control action names in code, no need for validation

---

### Key Lists (FR-015)

**Rule**: 0-10 keys per action (soft limit)

**Valid**:
- [] (empty - unbinds action)
- [KeySpace] (single key)
- [KeyLeft, KeyA] (multiple keys)
- [KeyEnter, KeySpace, KeyY, KeyReturn] (4 keys)

**Invalid**: None - any slice length accepted, 10+ keys supported but not recommended

**Rationale**: Large key lists degrade IsActionPressed performance (linear iteration)

---

### Key Duplicates

**Rule**: No automatic deduplication

**Behavior**:
- Bind("jump", KeySpace, KeySpace) → stores [KeySpace, KeySpace]
- IsActionPressed checks both (redundant but harmless)

**Rationale**: Deduplication adds complexity, duplicates have no semantic difference

---

## Relationships

### GameInput → Input

**Type**: Composition (has-a)

**Cardinality**: 1:1 (one GameInput wraps exactly one Input)

**Lifetime**: GameInput owns the Input reference, delegates Start/Stop

**Interface**:
```go
type gameInputImpl struct {
    input    Input           // Wrapped Input instance
    bindings map[string][]Key // Action binding map
    mu       sync.RWMutex    // Protects bindings
}
```

**Dependencies**:
- GameInput.Start() → Input.Start()
- GameInput.Stop() → Input.Stop()
- GameInput.IsActionPressed() → Input.IsPressed() for each bound key

---

### Action → Keys

**Type**: Association (one-to-many)

**Cardinality**: 1 action : 0-N keys

**Semantics**: OR relationship - any key triggers action

**Implementation**: map[string][]Key

**Example**:
- Action "jump" → [KeySpace]
- Action "fire" → [KeySpace, KeyEnter, KeyF]
- Action "unbound" → [] (not in map)

---

## Memory Footprint

### Per-Action Overhead

**Map Entry**:
- Key: ~16 bytes (string header) + len(action_name)
- Value: ~24 bytes (slice header) + 4 bytes * num_keys

**Typical Action** ("move-left", 2 keys):
- ~16 + 9 = 25 bytes (key)
- ~24 + 8 = 32 bytes (value)
- **Total: ~57 bytes**

**100 Actions**:
- ~5.7 KB total (negligible)

**Conclusion**: Memory usage is not a concern for typical game action counts

---

## Concurrency Model

### Thread Safety Guarantees (FR-011)

**Protected Operations**:
- Bind() - write lock for map update
- IsActionPressed() - read lock for map lookup

**Lock Granularity**:
```go
// Bind - write path
mu.Lock()
bindings[action] = keys
mu.Unlock()

// IsActionPressed - read path
mu.RLock()
keys := bindings[action]
mu.RUnlock()
for _, key := range keys {
    if input.IsPressed(key) {
        return true
    }
}
return false
```

**Concurrency Patterns**:
- Multiple goroutines can call IsActionPressed concurrently (RLock allows parallel readers)
- Bind blocks all reads/writes during update (rare operation, acceptable latency)
- No goroutines spawned by GameInput (delegates to Input's goroutines)

---

## Performance Characteristics

### Time Complexity

| Operation | Best Case | Worst Case | Typical |
|-----------|-----------|------------|---------|
| Bind() | O(1) | O(1) | O(1) |
| IsActionPressed() | O(1) | O(k) | O(k) where k=2-3 |
| Start()/Stop() | O(1) | O(1) | O(1) |

**Notes**:
- k = number of keys bound to action (typically 1-3)
- Map lookup is O(1)
- Key iteration is O(k) but k is small (cache-friendly)

### Space Complexity

**Storage**: O(a * k) where a = actions, k = avg keys per action
- 100 actions * 3 keys = ~300 Key values = ~1.2 KB

**Overhead**: Negligible for game use cases

---

## Persistence: None

**Rationale**: Action bindings are runtime configuration, not persistent data

**If Persistence Needed** (future extension):
- Save bindings to JSON/TOML config file
- Load bindings on startup via Bind() calls
- Out of scope for this feature

---

## Summary

The GameInput data model is intentionally minimal:
- **Single entity**: In-memory binding map
- **Simple structure**: map[string][]Key
- **No persistence**: Runtime-only state
- **Thread-safe**: RWMutex protection
- **Efficient**: O(1) lookup, O(k) iteration with small k

This simplicity enables easy testing, low overhead, and straightforward implementation while meeting all functional requirements.
