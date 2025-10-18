# Research: GameInput Action Mapping API

**Feature**: 004-game-input-api
**Date**: 2025-10-18
**Status**: Complete

## Overview

This research phase validates technical decisions for implementing the GameInput action mapping API. Since this feature extends the existing Input interface with minimal new surface area, research focuses on design patterns, thread safety, and API consistency.

---

## R1: Thread-Safe Action Binding Map

**Question**: What synchronization mechanism should protect the action binding map for concurrent Bind() writes and IsActionPressed() reads?

**Decision**: `sync.RWMutex` for read-write lock on the bindings map

**Rationale**:
- IsActionPressed() is read-heavy (called every frame in game loops)
- Bind() is write-rare (only during initialization or settings changes)
- RWMutex allows multiple concurrent readers without blocking
- Standard library pattern with zero dependencies
- Consistent with existing Input implementation patterns in the codebase

**Alternatives Considered**:
- **sync.Mutex**: Would serialize all reads, degrading game loop performance. Rejected because 60fps games query 10+ actions per frame.
- **sync.Map**: Overkill for this use case - designed for append-only or partition-keyed scenarios. Adds complexity without measurable benefit for small binding maps.
- **Lock-free atomic maps**: Requires unsafe operations or third-party libraries. Violates zero-dependency constraint and adds complexity.

**Implementation Notes**:
- Use RLock() for IsActionPressed() (read path)
- Use Lock() for Bind() (write path)
- Hold lock only during map access, not during Input.IsPressed() calls

---

## R2: Action Binding Storage Structure

**Question**: How should action-to-keys bindings be stored internally?

**Decision**: `map[string][]Key` - action name to key slice

**Rationale**:
- Direct mapping from action name (string) to list of bound keys
- Go maps provide O(1) lookup by action name
- Slices allow variable-length key lists (1-10 keys typical)
- Simple iteration for IsActionPressed() OR logic
- No additional indexing or reverse lookups needed

**Alternatives Considered**:
- **Bidirectional map (action→keys + key→actions)**: Would enable faster "what actions use this key?" queries, but that's not a feature requirement. Adds complexity and memory overhead.
- **map[string]map[Key]bool**: Using nested map for key membership would be faster for large key sets, but typical actions have 1-3 keys where slice iteration is faster due to cache locality.

**Implementation Notes**:
- Replace entire slice on Bind() call (don't mutate in place)
- Empty slice for unbound actions (removed from map)
- No need to deduplicate keys - that's caller responsibility

---

## R3: Delegation to Input Interface

**Question**: How should GameInput interact with the underlying Input interface?

**Decision**: Composition - GameInput wraps an Input instance and delegates lifecycle and state queries

**Rationale**:
- GameInput "has-a" Input, not "is-a" Input - composition is the correct relationship
- Allows GameInput to be a separate interface without polluting Input interface
- Start/Stop delegate directly to wrapped Input
- IsActionPressed() loops through bound keys calling Input.IsPressed()
- Clear separation of concerns: Input handles keys, GameInput handles actions

**Alternatives Considered**:
- **Embedding Input interface**: Would expose all Input methods through GameInput, creating API confusion. Users wouldn't know whether to use Poll()/Next() or action queries.
- **Subclassing/extending Input**: Go doesn't have inheritance. Would require wrapping all Input methods.

**Implementation Notes**:
- `type gameInputImpl struct { input Input; ... }`
- NewGameInput() accepts an Input instance or creates one via input.New()
- No need to expose wrapped Input - encapsulation principle

---

## R4: Factory Function Design

**Question**: Should NewGameInput() create its own Input instance or accept one as a parameter?

**Decision**: Accept optional Input parameter, create default if nil

**Rationale**:
- Flexibility: Advanced users can customize Input configuration (buffer size, backend)
- Simplicity: Basic users can call NewGameInput() with no arguments
- Testability: Tests can inject mock Input implementations
- Follows existing NewGameInput() signature in original contract spec

**Signature**:
```go
func NewGameInput(input Input) GameInput
```

**Alternatives Considered**:
- **Always create new Input**: Simpler API but prevents customization. Rejected because some games may need custom Input configuration.
- **Separate factory functions**: NewGameInput() and NewGameInputWithCustomInput(). Verbose and un-idiomatic in Go.

**Implementation Notes**:
- If input parameter is nil, create via input.New()
- Validate that Input is not already started (or document that it should not be)

---

## R5: Error Handling Strategy

**Question**: How should GameInput handle errors from the wrapped Input's Start() method?

**Decision**: Propagate errors directly from Input.Start()

**Rationale**:
- GameInput has no failure modes of its own (binding is infallible)
- Start() errors come from terminal initialization in Input
- Callers already expect and handle Input.Start() errors
- No value in wrapping or transforming errors

**Error Cases**:
- Start() → returns Input.Start() error
- Stop() → delegates to Input.Stop() (no error return)
- Bind() → no error (infallible)
- IsActionPressed() → no error (returns false for unbound actions)

**Implementation Notes**:
- Use `return g.input.Start()` with no wrapping
- Document that errors originate from underlying Input

---

## R6: Action Name Case Sensitivity

**Question**: Should action names be case-sensitive or normalized to lowercase?

**Decision**: Case-sensitive (treat "jump" and "Jump" as different actions)

**Rationale**:
- Principle of least surprise: Go maps are case-sensitive
- Developers control action names - they're code constants, not user input
- No benefit to case normalization when developers define names
- Simpler implementation (no string.ToLower() overhead)
- Consistent with spec requirement FR-005

**Alternatives Considered**:
- **Case-insensitive**: Would require strings.ToLower() on every Bind/IsActionPressed call. Adds overhead for no clear benefit when action names are developer-defined constants.

**Implementation Notes**:
- Document case sensitivity in godoc
- Example code should demonstrate consistent naming (e.g., "move-left", not "Move-Left")

---

## R7: Unbinding Actions

**Question**: How should developers unbind an action (remove all key bindings)?

**Decision**: Bind(action) with empty key list removes the binding

**Signature**: `Bind(action string, keys ...Key)` - variadic allows zero keys

**Rationale**:
- Natural extension of binding API
- No separate Unbind() method needed
- Consistent with spec requirement FR-012
- Empty slice semantics: "bind action to nothing" = "unbind action"

**Alternatives Considered**:
- **Separate Unbind(action string) method**: More explicit but adds API surface. Rejected because Bind with empty keys is self-documenting.
- **Bind(action, nil)**: Confusing - nil slice vs empty slice semantics.

**Implementation Notes**:
- Delete action from map when keys slice is empty
- IsActionPressed() returns false for deleted actions (map lookup returns zero value)

---

## R8: Performance Optimization

**Question**: What performance optimizations are necessary for 60fps game loops?

**Decision**: Minimize lock contention and optimize hot path (IsActionPressed)

**Hot Path Analysis**:
- IsActionPressed() called 10+ times per frame @ 60fps = 600+ calls/sec
- Each call: RLock, map lookup, iterate 1-3 keys, call Input.IsPressed() per key, RUnlock
- Critical: RLock must be fast, key iteration must be cache-friendly

**Optimizations**:
- Use RWMutex (not Mutex) for concurrent reads
- Keep key slices small (1-10 keys) for fast iteration
- No allocations in hot path (read from existing map/slice)
- Early return on first pressed key (short-circuit OR logic)

**Benchmarking Plan**:
- Benchmark IsActionPressed() with 1, 3, 10 bound keys
- Benchmark concurrent IsActionPressed() from multiple goroutines
- Target: <1ms per call (SC-002 requirement)

**Implementation Notes**:
- Document performance characteristics in godoc
- Keep RLock duration minimal (release before Input.IsPressed() calls if beneficial)

---

## R9: API Consistency with Existing Input Interface

**Question**: Should GameInput method signatures mirror Input where applicable?

**Decision**: Use consistent patterns but adapt to action-mapping semantics

**Consistency Checklist**:
- ✅ Start() error vs Start() - both return error
- ✅ Stop() vs Stop() - both void return
- ✅ IsPressed(Key) bool vs IsActionPressed(string) bool - consistent naming pattern
- ✅ Factory: New() vs NewGameInput() - clear differentiation

**Rationale**:
- Familiar patterns reduce learning curve
- IsActionPressed mirrors IsPressed naming
- Start/Stop lifecycle matches exactly

**Implementation Notes**:
- Godoc should reference Input interface for lifecycle semantics
- Examples should show both Input and GameInput usage side-by-side

---

## R10: Example Game Implementation

**Question**: What should the example game demonstrate?

**Decision**: Simple movement + action game loop showing all three user stories

**Example Features**:
- P1: Basic single-key bindings (arrows → movement)
- P2: Multiple keys per action (WASD + arrows)
- P3: Runtime rebinding (press 'R' to rebind controls)

**Implementation**:
```
examples/game/main.go
- Initialize GameInput
- Bind default actions
- Game loop: query IsActionPressed, update position, render
- Press R: enter rebind mode, capture key, update binding
- Press ESC: exit
```

**Rationale**:
- Demonstrates all priority levels from spec
- Shows typical game loop pattern
- Provides copy-paste starting point for developers

---

## Summary of Decisions

| ID | Topic | Decision | Impact |
|----|-------|----------|--------|
| R1 | Thread Safety | sync.RWMutex | Enables concurrent reads, minimal contention |
| R2 | Storage | map[string][]Key | O(1) action lookup, simple iteration |
| R3 | Delegation | Composition with Input | Clear separation, delegates lifecycle |
| R4 | Factory | Accept optional Input param | Flexible initialization |
| R5 | Error Handling | Propagate Input.Start() errors | Consistent with Input semantics |
| R6 | Case Sensitivity | Case-sensitive action names | Least surprise, no overhead |
| R7 | Unbinding | Bind(action) with empty keys | Natural API extension |
| R8 | Performance | RLock hot path, early return | <1ms target achievable |
| R9 | API Consistency | Mirror Input patterns | Familiar to Input users |
| R10 | Example | Multi-feature game loop | Demonstrates P1-P3 |

---

## Open Questions: None

All technical decisions resolved. Ready for Phase 1 (Design & Contracts).
