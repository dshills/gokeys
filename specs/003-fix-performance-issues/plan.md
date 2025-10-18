# Implementation Plan: Performance and Efficiency Improvements

**Branch**: `003-fix-performance-issues` | **Date**: 2025-10-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-fix-performance-issues/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature optimizes the gokeys input system by addressing three non-critical performance issues: eliminating artificial 5ms latency on Escape keypresses, reducing memory allocations by reusing buffers, and adding UTF-8 support for international characters. The improvements maintain full backward compatibility while enhancing performance for high-frequency input scenarios (games, real-time editors).

## Technical Context

**Language/Version**: Go 1.25.3+
**Primary Dependencies**: golang.org/x/sys/unix (existing), unicode/utf8 (standard library)
**Storage**: N/A (in-memory event processing)
**Testing**: go test (existing test suite), benchmarking with go test -bench
**Target Platform**: Unix-like systems (Linux, macOS, BSD) - existing Unix backend only
**Project Type**: Single library package (input/)
**Performance Goals**: <1ms Escape key latency, 50% reduction in allocations, 100% UTF-8 accuracy
**Constraints**: Zero external dependencies (beyond existing golang.org/x/sys), maintain backward compatibility
**Scale/Scope**: Affects input/backend_unix.go and input/parser.go (2 files, ~300 lines modified)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Cross-Platform Abstraction
**Status**: PASS ✅
- Changes are isolated to Unix backend (input/backend_unix.go)
- No platform-specific types exposed in public API
- UTF-8 decoder logic is platform-agnostic (applies to parser.go)
- Public Event struct already supports rune field for Unicode

### Principle II: Dual API Design
**Status**: PASS ✅
- No changes to Poll() or Next() interface
- Both methods continue to share unified event queue
- Performance improvements benefit both blocking and non-blocking paths equally

### Principle III: Code Quality Standards
**Status**: PASS ✅
- Must maintain golangci-lint compliance (currently 0 issues)
- UTF-8 decoder logic must stay under cyclomatic complexity <30
- All new code requires godoc comments
- Error handling must check all errors (buffer operations, UTF-8 decoding)

### Principle IV: Testing Requirements
**Status**: PASS ✅
- Must write tests FIRST (TDD approach)
- New tests required for:
  - UTF-8 character decoding (contract tests)
  - Buffer reuse safety (unit tests)
  - Escape sequence timing (benchmark tests)
- All existing tests must continue to pass (backward compatibility)

### Principle V: Platform Normalization
**Status**: PASS ✅
- UTF-8 characters will be normalized to Event.Rune field (existing)
- KeyUnknown handling unchanged for unparsable sequences
- Timestamps remain monotonic (no changes to Event.Timestamp)
- No changes to modifier handling or autorepeat detection

**Overall Assessment**: All constitution principles PASS. No violations requiring justification.

## Project Structure

### Documentation (this feature)

```
specs/003-fix-performance-issues/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
input/
├── backend_unix.go      # Modified: remove sleep, add buffer reuse
├── parser.go            # Modified: add UTF-8 decoding
├── event.go             # Unchanged (already has rune field)
├── impl.go              # Unchanged
└── input.go             # Unchanged

tests/
├── contract/
│   └── normalization_test.go    # Modified: add UTF-8 tests
├── integration/
│   └── unix_test.go             # Modified: add performance tests
└── benchmarks/
    └── input_bench_test.go      # New: latency and allocation benchmarks
```

**Structure Decision**: Single library project structure. Modifications are confined to the existing input package, specifically the Unix backend (backend_unix.go) for buffer optimization and latency fixes, and the parser (parser.go) for UTF-8 support. No new packages or architectural changes needed.

## Complexity Tracking

*No constitution violations - this section intentionally left empty.*

---

## Constitution Re-Check (Post Phase 1 Design)

**Status**: PASS ✅ (Re-validated after completing research and design artifacts)

### Updated Assessment

All five constitution principles continue to PASS after detailed design:

1. **Cross-Platform Abstraction**: Changes remain isolated to Unix backend
2. **Dual API Design**: No interface changes to Poll()/Next()
3. **Code Quality Standards**: Research confirms UTF-8 decoder complexity <30
4. **Testing Requirements**: Comprehensive test plan in quickstart.md
5. **Platform Normalization**: UTF-8 runes map to Event.Rune (existing field)

### Phase 1 Artifacts Completed

- ✅ research.md - All technical decisions documented with rationale
- ✅ data-model.md - Modified entities and state transitions defined
- ✅ contracts/backend-interface.md - ReadEvent() contract specified
- ✅ quickstart.md - Three-phase implementation guide created
- ✅ CLAUDE.md - Agent context updated with dependencies

**Ready for Phase 2**: `/speckit.tasks` can now generate implementation task list
