<!--
Sync Impact Report:
- Version: 0.0.0 → 1.0.0
- Initial constitution creation for gokeys project
- Added principles:
  1. Cross-Platform Abstraction
  2. Dual API Design
  3. Code Quality Standards
  4. Testing Requirements
  5. Platform Normalization
- Templates status:
  ✅ plan-template.md - Constitution Check section verified
  ✅ spec-template.md - Requirements alignment verified
  ✅ tasks-template.md - Task categorization verified
  ⚠ No command templates found in .specify/templates/commands/
- Follow-up: None - all placeholders resolved
-->

# gokeys Constitution

## Core Principles

### I. Cross-Platform Abstraction

**MUST**: All terminal/platform-specific behavior MUST be abstracted behind unified interfaces.

**MUST**: Platform-specific implementations (unixReader, windowsReader, etc.) MUST be isolated in backend modules.

**MUST**: Public APIs MUST NOT expose platform-specific types or escape sequences.

**MUST**: Auto-detection of optimal backend MUST be provided via factory functions (e.g., `input.New()`).

**Rationale**: The primary goal of gokeys is to eliminate cross-terminal and cross-platform inconsistencies. Leaking platform details to consumers defeats this purpose and creates maintenance burden. Backend isolation enables independent testing and platform-specific optimizations without breaking the public contract.

### II. Dual API Design (Blocking + Non-blocking)

**MUST**: Input systems MUST provide both blocking (`Poll()`) and non-blocking (`Next()`) event retrieval.

**MUST**: `Poll()` MUST block until an event is available or shutdown occurs.

**MUST**: `Next()` MUST return immediately with `nil` if no event is available.

**MUST**: Both methods MUST return events from the same unified queue (consistent ordering).

**Rationale**: Different use cases require different concurrency patterns. Game loops need non-blocking checks within render cycles. CLI tools need blocking waits without manual polling. Providing both patterns eliminates the need for consumers to build wrapper abstractions and prevents busy-wait anti-patterns.

### III. Code Quality Standards (NON-NEGOTIABLE)

**MUST**: All code MUST pass `golangci-lint run` with the project's configured linters before commit.

**MUST**: Cyclomatic complexity MUST NOT exceed 30 per function (enforced by cyclop linter).

**MUST**: All errors MUST be checked and handled (enforced by errcheck).

**MUST**: All exported types, functions, and constants MUST have godoc comments (enforced by revive).

**MUST**: Error messages MUST start with lowercase and contain no trailing punctuation (enforced by revive error-strings rule).

**MUST**: Context parameters MUST be the first argument in function signatures (enforced by revive context-as-argument rule).

**MUST**: Package average complexity MUST NOT exceed 10.0 (enforced by cyclop package-average).

**Rationale**: gokeys is a vendor-grade library targeting production use in games, TUIs, and critical CLI tools. Inconsistent quality or undocumented behavior undermines trust. Automated enforcement via golangci-lint ensures maintainability and reduces review burden. Complexity limits prevent unmaintainable god-functions that are common in input handling code.

### IV. Testing Requirements

**MUST**: Every public API function MUST have corresponding unit tests.

**MUST**: Backend implementations MUST have platform-specific integration tests.

**MUST**: Cross-backend normalization MUST be validated via contract tests (e.g., all backends produce identical Key codes for common sequences).

**MUST**: Tests MUST be written FIRST and FAIL before implementation (TDD).

**MUST**: Tests MUST run via standard `go test ./...` without external dependencies.

**SHOULD**: Platform-specific tests SHOULD be skipped gracefully on unsupported platforms using build tags.

**Rationale**: Input handling is notorious for platform-specific edge cases and escape sequence ambiguity. Contract tests ensure normalization correctness across backends. TDD prevents implementation-driven test design that misses edge cases. Build tag guards enable cross-platform development without requiring all platforms for every contributor.

### V. Platform Normalization

**MUST**: All terminal escape sequences MUST be normalized to unified `Key` constants.

**MUST**: Modifier keys (Shift, Alt, Ctrl) MUST use bitflag composition for combinations.

**MUST**: Autorepeat events MUST be explicitly flagged via `Event.Repeat` field.

**MUST**: Key-up events MUST be supported where platform allows via `Event.Pressed` field.

**MUST**: Timestamps MUST be recorded in `Event.Timestamp` using monotonic clock source.

**MUST**: Unparsable sequences MUST map to `KeyUnknown` rather than panic or drop events.

**Rationale**: Normalization is the core value proposition. Incomplete normalization forces consumers to handle platform differences, negating the library's purpose. Explicit repeat/pressed flags enable consistent behavior across terminals with different autorepeat policies. Monotonic timestamps prevent time-travel bugs during system clock adjustments. `KeyUnknown` provides graceful degradation for terminal emulators with non-standard sequences.

## Architecture Constraints

### Thread Safety

**MUST**: Input capture MUST run in a dedicated goroutine separate from event consumers.

**MUST**: Event queue MUST use buffered channels to prevent blocking input capture.

**MUST**: `Start()` and `Stop()` MUST be safe to call from any goroutine.

**MUST NOT**: Consumers MUST NOT directly access backend state; all state queries MUST go through Input interface methods.

**Rationale**: Asynchronous input capture prevents event loss during consumer processing delays. Goroutine isolation simplifies cancellation and cleanup. Buffered channels prevent backpressure from slow consumers dropping system input events.

### State Tracking

**MUST**: `IsPressed(Key)` MUST reflect current physical key state, not event stream position.

**MUST**: Key state MUST be updated on both Press and Release events where supported.

**SHOULD**: Key state SHOULD gracefully degrade to event-based approximation on platforms without key-up events.

**Rationale**: Real-time state queries are essential for game input (e.g., "is spacebar held?"). Event-only tracking creates race conditions between event processing and state queries. Graceful degradation maintains API compatibility across platform capabilities.

## Development Workflow

### Linting

**MUST**: Run `golangci-lint run` before every commit.

**MUST**: Address all linter errors; warnings MAY be suppressed with inline justification.

**MUST NOT**: Disable linters globally without constitution amendment.

### Building

**MUST**: Use `go build ./...` for build verification.

**MUST**: Support Go 1.25.3+ (as specified in go.mod).

### Testing

**MUST**: Run `go test ./...` to execute all tests.

**MUST**: Achieve minimum 80% code coverage for public APIs.

**SHOULD**: Use `go test -run TestName ./path/to/package` for targeted test execution during development.

**SHOULD**: Generate coverage reports via `go test -cover ./...` for validation.

## Governance

**Amendment Procedure**:
1. Proposed changes MUST be documented in a constitution amendment PR.
2. Amendment PR MUST include rationale, affected code areas, and migration plan.
3. Constitution version MUST be incremented per semantic versioning:
   - MAJOR: Backward-incompatible principle removals or redefinitions.
   - MINOR: New principles or materially expanded guidance.
   - PATCH: Clarifications, wording fixes, non-semantic refinements.
4. Amendments MUST update dependent templates (plan, spec, tasks) in the same PR.
5. LAST_AMENDED_DATE MUST be updated to amendment merge date.

**Versioning Policy**:
- Constitution version is independent of library semantic version.
- Constitution version tracks governance evolution, not implementation milestones.

**Compliance Review**:
- All PRs MUST verify constitution compliance before merge.
- Architecture decisions violating principles MUST be justified in "Complexity Tracking" section of plan.md.
- Unjustified violations are grounds for PR rejection.

**Runtime Guidance**:
- Agents MUST consult `CLAUDE.md` for development workflow and project structure.
- Constitution defines "what and why"; `CLAUDE.md` defines "how".

**Version**: 1.0.0 | **Ratified**: 2025-10-17 | **Last Amended**: 2025-10-17
