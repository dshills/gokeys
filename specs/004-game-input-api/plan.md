# Implementation Plan: GameInput Action Mapping API

**Branch**: `004-game-input-api` | **Date**: 2025-10-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-game-input-api/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

The GameInput API provides a higher-level action mapping abstraction on top of the existing Input interface. Game developers can bind logical action names (e.g., "jump", "fire", "move-left") to one or more physical keys, enabling rebindable controls and simplifying game logic by decoupling it from specific key bindings. The implementation wraps the Input interface with a thread-safe binding map and delegates key state queries to the underlying Input.IsPressed() method.

## Technical Context

**Language/Version**: Go 1.25.3+ (per go.mod)
**Primary Dependencies**:
- Existing `input` package (Input interface, Key types, Event types)
- Standard library only: `sync` (RWMutex for thread safety)
**Storage**: N/A (in-memory binding map only)
**Testing**: `go test` with unit tests, contract tests
**Target Platform**: Cross-platform (Linux, macOS, Windows) - inherits from Input interface
**Project Type**: Single Go library package extending existing `input` package
**Performance Goals**:
- IsActionPressed queries <1ms response time
- Support 100+ concurrent actions without degradation
- 60fps game loop support (10+ action queries per frame)
**Constraints**:
- Zero external dependencies beyond standard library
- Thread-safe for concurrent Bind/IsActionPressed calls
- Must delegate to existing Input.IsPressed() for key state
**Scale/Scope**:
- 100+ unique action names
- 10+ keys per action
- Production-ready library for game development

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Cross-Platform Abstraction
- ✅ **PASS**: GameInput wraps existing Input interface, which already provides platform abstraction
- ✅ **PASS**: No platform-specific code required in GameInput (delegates to Input)
- ✅ **PASS**: Public API exposes only action names (strings) and Key types (already normalized)

### Principle II: Dual API Design
- ✅ **PASS**: GameInput does not replace blocking/non-blocking patterns - it extends them
- ✅ **PASS**: Developers can still use Poll()/Next() on underlying Input if needed
- ℹ️ **N/A**: GameInput is state-query focused, not event-stream focused

### Principle III: Code Quality Standards
- ✅ **MUST VERIFY**: All code will pass golangci-lint (to be enforced during implementation)
- ✅ **MUST VERIFY**: Cyclomatic complexity <30 per function (simple map lookups expected)
- ✅ **MUST VERIFY**: All errors checked (minimal error surfaces - Start/Stop only)
- ✅ **MUST VERIFY**: Godoc on all exported types (GameInput interface, NewGameInput factory)
- ✅ **MUST VERIFY**: Error messages follow conventions (lowercase, no punctuation)

### Principle IV: Testing Requirements
- ✅ **PLANNED**: Unit tests for Bind(), IsActionPressed(), Start(), Stop()
- ✅ **PLANNED**: Contract tests for single-key, multi-key, and rebinding scenarios
- ✅ **PLANNED**: Thread-safety tests for concurrent Bind/IsActionPressed
- ✅ **PLANNED**: TDD approach - tests written first
- ✅ **PASS**: Standard `go test ./...` execution

### Principle V: Platform Normalization
- ✅ **PASS**: GameInput uses existing normalized Key types
- ✅ **PASS**: No new escape sequences or platform-specific code
- ℹ️ **N/A**: Action mapping is a layer above normalization

### Architecture Constraints - Thread Safety
- ✅ **PLANNED**: RWMutex for binding map protection
- ✅ **PLANNED**: Start/Stop delegate to Input (already goroutine-safe)
- ✅ **PASS**: No direct backend access - all via Input interface

### Architecture Constraints - State Tracking
- ✅ **PASS**: IsActionPressed queries Input.IsPressed for each bound key
- ✅ **PASS**: Delegates state tracking to existing Input implementation

**GATE STATUS**: ✅ **PASSED** - All constitution requirements met or N/A for this feature

---

## Post-Design Constitution Re-Check

*Re-evaluated after Phase 1 (Design & Contracts) completion*

### Updated Verification

**Principle III: Code Quality Standards**:
- ✅ **VERIFIED**: Interface and implementation designed with simple methods (<30 complexity)
- ✅ **VERIFIED**: Error handling limited to Start() delegation (propagates Input errors)
- ✅ **VERIFIED**: Godoc complete in contract specification
- ✅ **VERIFIED**: No error messages to format (void methods or propagated errors)

**Principle IV: Testing Requirements**:
- ✅ **VERIFIED**: Unit tests planned for all methods in quickstart.md
- ✅ **VERIFIED**: Contract tests planned for all user stories (P1, P2, P3)
- ✅ **VERIFIED**: Concurrency tests with race detector planned
- ✅ **VERIFIED**: TDD approach documented in quickstart (tests before implementation)

**Architecture Constraints - Thread Safety**:
- ✅ **VERIFIED**: RWMutex design validated in research.md (R1)
- ✅ **VERIFIED**: Lock granularity defined in data-model.md
- ✅ **VERIFIED**: Concurrent test strategy in quickstart.md Step 4.1

**Architecture Constraints - State Tracking**:
- ✅ **VERIFIED**: IsActionPressed delegates to Input.IsPressed (per contract)
- ✅ **VERIFIED**: No direct backend access (composition with Input)

**FINAL GATE STATUS**: ✅ **PASSED** - Design artifacts validate all constitution requirements

## Project Structure

### Documentation (this feature)

```
specs/004-game-input-api/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── game-input-interface.md  # GameInput API contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
input/                   # Existing package - GameInput will be added here
├── event.go            # Existing: Key, Modifier, Event types
├── input.go            # Existing: Input interface
├── impl.go             # Existing: inputImpl implementation
├── backend_unix.go     # Existing: Unix backend
├── parser.go           # Existing: Escape sequence parser
├── game.go             # NEW: GameInput interface definition
├── game_impl.go        # NEW: gameInputImpl implementation
├── game_test.go        # NEW: Unit tests for GameInput
└── doc.go              # Existing: Package documentation

examples/                # Existing examples directory
├── basic/              # Existing: Basic input example
├── poll/               # Existing: Poll example
├── next/               # Existing: Next example
└── game/               # NEW: Game with action mapping example
    └── main.go

tests/                   # Existing test directory
└── contract/           # Existing: Contract tests
    └── game_input_test.go  # NEW: GameInput contract tests
```

**Structure Decision**: This is a single Go library project. GameInput is added as new files in the existing `input/` package alongside the Input interface. This maintains package cohesion and allows GameInput to directly reference Input types without import cycles. The contract interface already exists at `specs/001-input-system/contracts/game-input-interface.md` which we'll reference for consistency.

## Complexity Tracking

*No violations - this section is not needed.*

All constitution requirements are satisfied without exceptions.

