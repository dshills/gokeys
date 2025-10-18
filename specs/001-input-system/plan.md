# Implementation Plan: Cross-Terminal Input System

**Branch**: `001-input-system` | **Date**: 2025-10-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-input-system/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature implements a vendor-grade, cross-terminal keyboard input system for Go that abstracts platform-specific terminal behavior behind a unified interface. The system provides normalized key event capture with both blocking (Poll) and non-blocking (Next) APIs, real-time key state tracking (IsPressed), and optional high-level action mapping for game development. Core technical approach: separate goroutine for input capture feeding buffered channel, platform-specific backend selection via factory, escape sequence normalization to unified Key constants, monotonic timestamps, and graceful degradation for platform capability differences.

## Technical Context

**Language/Version**: Go 1.25.3+
**Primary Dependencies**: Standard library only (os, syscall, time, sync); optional compatibility with tcell/termbox for advanced use cases
**Storage**: N/A (in-memory state tracking only)
**Testing**: Go standard testing (`go test`), table-driven tests for normalization, build tags for platform-specific tests
**Target Platform**: Cross-platform library supporting Linux, macOS, Windows terminals; Unix-like systems via termios, Windows via Console API
**Project Type**: Single Go library project
**Performance Goals**: <10ms event capture latency, <16ms state query staleness (60fps), 100-event buffer capacity, support 30+ fps game loops
**Constraints**: Zero external dependencies for core functionality, <200ms p95 initialization time, graceful terminal restoration on abnormal termination
**Scale/Scope**: Library supporting 100+ key codes, 3+ platform backends, 4+ terminal emulators, suitable for production CLI tools and terminal games

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Cross-Platform Abstraction ✅

- ✅ Platform-specific backends (UnixBackend, WindowsBackend) isolated in internal/backend/
- ✅ Public API (input.Input interface) exposes only normalized types (Key, Event, Modifier)
- ✅ Factory function `input.New()` auto-detects platform
- ✅ No platform-specific types leak to consumers

**Compliance**: PASS - Design inherently follows abstraction principle

### Principle II: Dual API Design ✅

- ✅ Poll() method: blocking, returns (Event, bool)
- ✅ Next() method: non-blocking, returns *Event (nil if empty)
- ✅ Both consume from same buffered channel
- ✅ Consistent event ordering guaranteed

**Compliance**: PASS - Core requirement from spec

### Principle III: Code Quality Standards ✅

- ✅ golangci-lint integration required
- ✅ All public types will have godoc comments
- ✅ Error handling mandatory (errcheck)
- ✅ Complexity limits: functions <30, package avg <10.0

**Compliance**: PASS - Will be enforced during implementation via CI

### Principle IV: Testing Requirements ✅

- ✅ TDD: Contract tests for normalization written first
- ✅ Unit tests for all public APIs
- ✅ Platform-specific integration tests with build tags
- ✅ Contract tests validate backend equivalence

**Compliance**: PASS - Test structure defined in Phase 1

### Principle V: Platform Normalization ✅

- ✅ Escape sequences → unified Key constants
- ✅ Modifier bitflags (Shift|Alt|Ctrl)
- ✅ Event.Repeat for autorepeat
- ✅ Event.Pressed for key up/down
- ✅ Monotonic timestamps (time.Now() with monotonic clock)
- ✅ KeyUnknown for unparsable sequences

**Compliance**: PASS - Core value proposition

### Thread Safety ✅

- ✅ Input capture in dedicated goroutine
- ✅ Buffered channel (100 events)
- ✅ Start()/Stop() goroutine-safe
- ✅ No direct backend state access

**Compliance**: PASS - Architecture requirement

### State Tracking ✅

- ✅ IsPressed() reflects physical state
- ✅ Updated on press/release events
- ✅ Graceful degradation on platforms without key-up

**Compliance**: PASS - FR-009, FR-010 requirement

**GATE RESULT: ✅ ALL CHECKS PASSED - Proceed to Phase 0**

## Project Structure

### Documentation (this feature)

```
specs/001-input-system/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── input-interface.md
│   ├── game-input-interface.md
│   └── backend-interface.md
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Single Go library project

input/                   # Main package (public API)
├── event.go            # Event, Key, Modifier types
├── input.go            # Input interface and factory
├── game.go             # GameInput interface (action mapping)
└── doc.go              # Package documentation

internal/
├── backend/            # Platform-specific implementations
│   ├── backend.go      # Backend interface (internal contract)
│   ├── unix.go         # Unix/Linux/macOS backend (build tag: !windows)
│   ├── windows.go      # Windows backend (build tag: windows)
│   └── parser.go       # Escape sequence parser (shared)
├── state/              # Key state tracking
│   └── tracker.go      # IsPressed implementation
└── queue/              # Event queue management
    └── buffer.go       # Buffered channel wrapper

examples/               # Example programs
├── basic/              # Simple event loop example
├── game/               # Game input with action mapping
└── inspector/          # Key inspector tool (debugging)

tests/                  # Test organization
├── contract/           # Cross-backend normalization tests
│   └── normalization_test.go
├── integration/        # Platform-specific integration tests
│   ├── unix_test.go    # Unix integration (build tag)
│   └── windows_test.go # Windows integration (build tag)
└── unit/               # Unit tests co-located with source
    # (Go convention: *_test.go alongside source files)
```

**Structure Decision**: Single Go library project (Option 1) chosen because this is a standalone library package, not a web/mobile application. Source organized by public API (input/) vs internal implementation (internal/) following Go best practices. Platform-specific code isolated via build tags. Tests organized by type (contract/integration/unit) to enable selective execution.

## Complexity Tracking

*No constitution violations - this section intentionally left empty.*

All design decisions align with constitution principles. No complexity justifications required.
