# Tasks: Cross-Terminal Input System

**Input**: Design documents from `/specs/001-input-system/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: TDD approach requested - tests MUST be written FIRST and FAIL before implementation (Constitution Principle IV)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- **Single Go library**: `input/`, `internal/`, `examples/`, `tests/` at repository root
- All paths shown are absolute or relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic Go module structure

- [x] T001 Create directory structure: input/, internal/backend/, internal/state/, examples/, tests/contract/, tests/integration/
- [x] T002 Initialize go.mod with module github.com/dshills/gokeys if not exists
- [x] T003 [P] Copy .golangci.yml configuration to ensure linting rules enforced
- [x] T004 [P] Create input/doc.go with package documentation per constitution

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and interfaces that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 [P] Define Key constants (KeyUnknown, KeyEscape, KeyEnter, KeyUp, KeyDown, KeyLeft, KeyRight, KeyA-Z, Key0-9, KeyCtrlA-Z, KeySpace, KeyF1-F12, etc.) in input/event.go
- [x] T006 [P] Define Modifier bitflags (ModNone, ModShift, ModAlt, ModCtrl) in input/event.go
- [x] T007 [P] Define Event struct (Key, Rune, Modifiers, Timestamp, Pressed, Repeat fields) in input/event.go
- [x] T008 [P] Define Input interface (Start, Stop, Poll, Next, IsPressed methods) in input/input.go
- [x] T009 [P] Define Backend interface (Init, Restore, ReadEvent methods) in internal/backend/backend.go
- [x] T010 Validate core types compile and pass golangci-lint: go build ./input && golangci-lint run ./input

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - CLI Tool Developer Captures Keyboard Input (Priority: P1) 🎯 MVP

**Goal**: Enable basic keyboard event capture with normalized key codes across platforms, supporting both blocking (Poll) and non-blocking (Next) patterns

**Independent Test**: Initialize input system, capture arrow keys and standard keys, verify identical normalized events across different terminals (iTerm2, Windows Terminal, xterm)

### Tests for User Story 1 (TDD - MUST FAIL before implementation) ⚠️

- [ ] T011 [P] [US1] Write contract test for escape sequence normalization (arrow keys: \x1b[A → KeyUp, \x1b[B → KeyDown, \x1b[C → KeyRight, \x1b[D → KeyLeft) in tests/contract/normalization_test.go
- [ ] T012 [P] [US1] Write contract test for Ctrl key normalization (Ctrl+C: \x03 → KeyCtrlC with ModCtrl) in tests/contract/normalization_test.go
- [ ] T013 [P] [US1] Write contract test for unknown sequence handling (\x1b[999~ → KeyUnknown) in tests/contract/normalization_test.go
- [ ] T014 [P] [US1] Write unit test for Event struct zero values and field validation in input/event_test.go
- [ ] T015 [P] [US1] Write unit test for Modifier bitflag operations (ModShift | ModCtrl) in input/event_test.go
- [ ] T016 [P] [US1] Write integration test for Unix backend terminal state save/restore in tests/integration/unix_test.go (build tag: !windows)
- [ ] T017 [P] [US1] Write unit test for Poll() blocking behavior and shutdown signal in input/input_test.go
- [ ] T018 [P] [US1] Write unit test for Next() non-blocking behavior (returns nil when no events) in input/input_test.go
- [ ] T019 Verify all User Story 1 tests FAIL: go test ./tests/contract ./input ./tests/integration (expect failures)

### Implementation for User Story 1

- [ ] T020 [P] [US1] Implement SequenceParser with trie structure (SequenceNode, addSequence, Parse methods) in internal/backend/parser.go
- [ ] T021 [P] [US1] Add Tier 1 escape sequences to parser trie (arrows, F1-F4, Home/End, PgUp/PgDn, Insert/Delete) in internal/backend/parser.go
- [ ] T022 [P] [US1] Implement UnixBackend struct (fd, parser, initialized, savedState fields) in internal/backend/unix.go (build tag: !windows)
- [ ] T023 [US1] Implement UnixBackend.Init() (tcgetattr, save state, set raw mode, tcsetattr) in internal/backend/unix.go
- [ ] T024 [US1] Implement UnixBackend.Restore() (restore saved termios, idempotent) in internal/backend/unix.go
- [ ] T025 [US1] Implement UnixBackend.ReadEvent() (syscall.Read, parser.Parse, timestamp) in internal/backend/unix.go
- [ ] T026 [P] [US1] Create platform detection factory createBackend() (runtime.GOOS check) in input/input.go
- [ ] T027 [P] [US1] Implement inputImpl struct (backend, events chan, done chan, once sync.Once fields) in input/input.go
- [ ] T028 [US1] Implement inputImpl.Start() (backend.Init, start capture goroutine, use sync.Once) in input/input.go
- [ ] T029 [US1] Implement inputImpl.Stop() (close done channel, backend.Restore) in input/input.go
- [ ] T030 [US1] Implement inputImpl.capture() goroutine (loop backend.ReadEvent, select on done, send to events channel) in input/input.go
- [ ] T031 [US1] Implement inputImpl.Poll() (select on events channel or done, return Event and bool) in input/input.go
- [ ] T032 [US1] Implement inputImpl.Next() (non-blocking select with default, return *Event or nil) in input/input.go
- [ ] T033 [US1] Implement New() factory function (create inputImpl, initialize fields, buffered channel cap 100) in input/input.go
- [ ] T034 [US1] Add godoc comments to all exported types and functions per constitution in input/event.go and input/input.go
- [ ] T035 [US1] Verify User Story 1 tests PASS: go test ./tests/contract ./input ./tests/integration
- [ ] T036 [US1] Run golangci-lint and fix any violations: golangci-lint run ./...
- [ ] T037 [US1] Create basic example demonstrating Poll() usage in examples/basic/main.go
- [ ] T038 [US1] Manual testing: run examples/basic/main.go and verify arrow key events captured correctly

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Game Developer Implements Real-Time Input (Priority: P2)

**Goal**: Add real-time key state tracking (IsPressed) and autorepeat detection for game development

**Independent Test**: Capture key press/release sequences, query IsPressed() state, validate repeat events flagged correctly

### Tests for User Story 2 (TDD - MUST FAIL before implementation) ⚠️

- [ ] T039 [P] [US2] Write unit test for StateTracker.Update() with press events in internal/state/tracker_test.go
- [ ] T040 [P] [US2] Write unit test for StateTracker.Update() with release events in internal/state/tracker_test.go
- [ ] T041 [P] [US2] Write unit test for StateTracker.IsPressed() concurrent reads in internal/state/tracker_test.go
- [ ] T042 [P] [US2] Write unit test for repeat flag detection (timing-based heuristic) in internal/backend/parser_test.go
- [ ] T043 [P] [US2] Write unit test for IsPressed() accuracy (press → true, release → false) in input/input_test.go
- [ ] T044 [P] [US2] Write integration test for rapid Next() calls returning nil in tests/integration/game_loop_test.go
- [ ] T045 Verify all User Story 2 tests FAIL: go test ./internal/state ./internal/backend ./input ./tests/integration

### Implementation for User Story 2

- [ ] T046 [P] [US2] Implement StateTracker struct (mu sync.RWMutex, pressed map[Key]bool) in internal/state/tracker.go
- [ ] T047 [P] [US2] Implement StateTracker.New() factory in internal/state/tracker.go
- [ ] T048 [US2] Implement StateTracker.Update() (lock, update map based on Pressed field) in internal/state/tracker.go
- [ ] T049 [US2] Implement StateTracker.IsPressed() (read lock, return map value) in internal/state/tracker.go
- [ ] T050 [US2] Add state *StateTracker field to inputImpl in input/input.go
- [ ] T051 [US2] Initialize StateTracker in New() factory in input/input.go
- [ ] T052 [US2] Call state.Update(event) in capture() goroutine before sending to channel in input/input.go
- [ ] T053 [US2] Implement inputImpl.IsPressed() delegation to state.IsPressed() in input/input.go
- [ ] T054 [P] [US2] Add repeat detection heuristic to parser (track last key, timestamp, 50ms threshold) in internal/backend/parser.go
- [ ] T055 [P] [US2] Set Event.Repeat field based on timing heuristic in internal/backend/parser.go
- [ ] T056 [P] [US2] Set Event.Pressed field (default true, Unix approximation for key-up) in internal/backend/parser.go
- [ ] T057 [US2] Add godoc comments to StateTracker types and methods in internal/state/tracker.go
- [ ] T058 [US2] Verify User Story 2 tests PASS: go test ./internal/state ./internal/backend ./input ./tests/integration
- [ ] T059 [US2] Run golangci-lint and fix any violations: golangci-lint run ./...
- [ ] T060 [US2] Create game loop example using IsPressed() and Next() in examples/game/main.go
- [ ] T061 [US2] Manual testing: run examples/game/main.go, hold keys, verify continuous movement

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Game Developer Uses Action Mapping (Priority: P3)

**Goal**: Provide high-level action mapping API allowing logical action names to be bound to physical keys

**Independent Test**: Bind actions to keys, trigger keys, verify IsActionPressed() returns correct values

### Tests for User Story 3 (TDD - MUST FAIL before implementation) ⚠️

- [ ] T062 [P] [US3] Write unit test for GameInput.Bind() single key binding in input/game_test.go
- [ ] T063 [P] [US3] Write unit test for GameInput.Bind() multiple keys per action in input/game_test.go
- [ ] T064 [P] [US3] Write unit test for GameInput.IsActionPressed() with single key bound in input/game_test.go
- [ ] T065 [P] [US3] Write unit test for GameInput.IsActionPressed() with multiple keys (OR logic) in input/game_test.go
- [ ] T066 [P] [US3] Write unit test for GameInput.Bind() replacing existing bindings in input/game_test.go
- [ ] T067 [P] [US3] Write unit test for GameInput.IsActionPressed() concurrent access in input/game_test.go
- [ ] T068 Verify all User Story 3 tests FAIL: go test ./input -run TestGameInput

### Implementation for User Story 3

- [ ] T069 [P] [US3] Define GameInput interface (Start, Stop, IsActionPressed, Bind methods) in input/game.go
- [ ] T070 [P] [US3] Implement gameInputImpl struct (input Input, bindings map[string][]Key, mu sync.RWMutex) in input/game.go
- [ ] T071 [P] [US3] Implement NewGameInput() factory in input/game.go
- [ ] T072 [US3] Implement gameInputImpl.Start() delegation to input.Start() in input/game.go
- [ ] T073 [US3] Implement gameInputImpl.Stop() delegation to input.Stop() in input/game.go
- [ ] T074 [US3] Implement gameInputImpl.Bind() (lock, replace bindings map entry) in input/game.go
- [ ] T075 [US3] Implement gameInputImpl.IsActionPressed() (read lock, iterate bound keys, check IsPressed) in input/game.go
- [ ] T076 [US3] Add godoc comments to GameInput interface and implementation in input/game.go
- [ ] T077 [US3] Verify User Story 3 tests PASS: go test ./input -run TestGameInput
- [ ] T078 [US3] Run golangci-lint and fix any violations: golangci-lint run ./...
- [ ] T079 [US3] Update game example to use action mapping (bind "move-up", "move-down", etc.) in examples/game/main.go
- [ ] T080 [US3] Manual testing: run examples/game/main.go, verify action-based controls work

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Cross-Platform Support (Windows Backend)

**Purpose**: Add Windows platform support using Console API

- [ ] T081 [P] Implement WindowsBackend struct (handle syscall.Handle, savedMode uint32, initialized bool) in internal/backend/windows.go (build tag: windows)
- [ ] T082 [P] Implement WindowsBackend.Init() (GetStdHandle, GetConsoleMode, save, SetConsoleMode with VT support) in internal/backend/windows.go
- [ ] T083 [P] Implement WindowsBackend.Restore() (restore saved console mode) in internal/backend/windows.go
- [ ] T084 [US1] Implement WindowsBackend.ReadEvent() (ReadConsoleInput, parse INPUT_RECORD, create Event) in internal/backend/windows.go
- [ ] T085 [P] Add Windows virtual key code to Key normalization mapping in internal/backend/windows.go
- [ ] T086 [US2] Set Event.Pressed from keyEvent.bKeyDown in internal/backend/windows.go
- [ ] T087 [US2] Set Event.Repeat from keyEvent.wRepeatCount in internal/backend/windows.go
- [ ] T088 Update createBackend() to return WindowsBackend on Windows in input/input.go
- [ ] T089 [P] Write Windows-specific integration tests in tests/integration/windows_test.go (build tag: windows)
- [ ] T090 Run contract tests on Windows to verify normalization equivalence: go test ./tests/contract
- [ ] T091 Manual testing on Windows: run examples/basic/main.go and examples/game/main.go

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T092 [P] Create key inspector debugging tool in examples/inspector/main.go
- [ ] T093 [P] Add extended Tier 2 escape sequences (Shift+Arrow, Alt+Key, Ctrl+Arrow) to parser in internal/backend/parser.go
- [ ] T094 [P] Add function keys F5-F12 escape sequences to parser in internal/backend/parser.go
- [ ] T095 [P] Verify 100+ Key constants defined (complete alphabet, numbers, function keys, navigation) in input/event.go
- [ ] T096 [P] Add Key.String() method for debugging output in input/event.go
- [ ] T097 [P] Add Modifier.String() method for debugging output in input/event.go
- [ ] T098 [P] Implement buffer overflow handling strategy (block vs drop) in inputImpl.capture() in input/input.go
- [ ] T099 [P] Add error handling and logging for backend.ReadEvent() failures in input/input.go
- [ ] T100 Run full test suite with coverage: go test -cover ./... | tee coverage.txt
- [ ] T101 Verify >80% coverage for public APIs per constitution
- [ ] T102 Run golangci-lint on entire codebase: golangci-lint run ./...
- [ ] T103 Verify all golangci-lint checks pass (errcheck, revive, cyclop <30, package avg <10)
- [ ] T104 Manual cross-terminal testing: verify on iTerm2, xterm, GNOME Terminal, Windows Terminal
- [ ] T105 Performance validation: verify <10ms event latency, <16ms IsPressed staleness
- [ ] T106 Create README.md with installation, usage examples, and API overview

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Cross-Platform (Phase 6)**: Can proceed after US1 core types exist
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Builds on US1 but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Wraps US1 & US2 but independently testable

### Within Each User Story

- Tests (TDD) MUST be written and FAIL before implementation (Constitution Principle IV)
- Parser before backends
- Backends before inputImpl
- Core implementation before examples
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models/types within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all contract tests for User Story 1 together:
Task: "Write contract test for escape sequence normalization in tests/contract/normalization_test.go"
Task: "Write contract test for Ctrl key normalization in tests/contract/normalization_test.go"
Task: "Write contract test for unknown sequence handling in tests/contract/normalization_test.go"

# Launch all unit tests for User Story 1 together:
Task: "Write unit test for Event struct in input/event_test.go"
Task: "Write unit test for Modifier bitflags in input/event_test.go"
Task: "Write unit test for Poll() in input/input_test.go"
Task: "Write unit test for Next() in input/input_test.go"

# Launch parallel implementation tasks:
Task: "Implement SequenceParser in internal/backend/parser.go"
Task: "Implement UnixBackend struct in internal/backend/unix.go"
Task: "Create platform detection factory in input/input.go"
Task: "Implement inputImpl struct in input/input.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (T011-T038)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

This gives you:
- ✅ Basic keyboard event capture
- ✅ Cross-platform normalization
- ✅ Blocking (Poll) and non-blocking (Next) APIs
- ✅ Terminal state restoration
- ✅ Working on Unix/macOS
- ❌ No state tracking yet (IsPressed)
- ❌ No action mapping yet
- ❌ No Windows support yet

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (adds IsPressed)
4. Add User Story 3 → Test independently → Deploy/Demo (adds action mapping)
5. Add Windows support (Phase 6) → Test on Windows → Deploy/Demo (full cross-platform)
6. Add Polish (Phase 7) → Final release
7. Each increment adds value without breaking previous functionality

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (T011-T038)
   - Developer B: User Story 2 (T039-T061) - requires US1 types but can write tests
   - Developer C: User Story 3 (T062-T080) - requires US1 types but can write tests
3. Stories integrate and validate independently
4. Team converges on Cross-Platform (Phase 6) together
5. Team converges on Polish (Phase 7) together

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD mandatory: Tests MUST be written first and FAIL before implementation (Constitution)
- Run golangci-lint after each phase to catch violations early
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

## Task Count Summary

- **Total Tasks**: 106
- **Setup Phase**: 4 tasks
- **Foundational Phase**: 6 tasks
- **User Story 1**: 28 tasks (19 tests + implementation)
- **User Story 2**: 23 tasks (7 tests + implementation)
- **User Story 3**: 19 tasks (7 tests + implementation)
- **Cross-Platform**: 11 tasks
- **Polish**: 15 tasks

**Parallel Opportunities**: 42 tasks marked [P] can run concurrently
**MVP Scope**: T001-T038 (Setup + Foundational + US1) = 38 tasks
