# Tasks: GameInput Action Mapping API

**Input**: Design documents from `/specs/004-game-input-api/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: TDD approach requested - tests MUST be written FIRST and FAIL before implementation (Constitution Principle IV)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- **Single Go library**: `input/`, `examples/`, `tests/` at repository root
- All paths shown are absolute or relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify existing Input system and prepare for GameInput addition

- [X] T001 Verify existing Input interface is functional by running `go test ./input -v`
- [X] T002 [P] Verify golangci-lint configuration is working by running `golangci-lint run ./input`
- [X] T003 [P] Create examples/game/ directory for action mapping example
- [X] T004 [P] Create tests/contract/ directory if it doesn't exist

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core GameInput interface and types that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 [P] Define GameInput interface in input/game.go with Start, Stop, Bind, IsActionPressed methods
- [X] T006 [P] Define gameInputImpl struct in input/game_impl.go with input, bindings map, and sync.RWMutex fields
- [X] T007 [P] Implement NewGameInput factory function in input/game.go (creates default Input if nil parameter)
- [X] T008 Verify foundational code compiles by running `go build ./input`
- [X] T009 Verify golangci-lint passes on new files by running `golangci-lint run ./input`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic Action Binding (Priority: P1) 🎯 MVP

**Goal**: Enable single actions bound to single keys with basic IsActionPressed() queries

**Independent Test**: Create GameInput instance, bind "jump" → Space, press Space, verify IsActionPressed("jump") returns true when held and false when released

### Tests for User Story 1 (TDD - MUST FAIL before implementation) ⚠️

- [X] T010 [P] [US1] Write unit test TestNewGameInput in input/game_test.go (tests factory with nil and non-nil Input)
- [X] T011 [P] [US1] Write unit test TestBindSingleKey in input/game_test.go (tests single action to single key binding)
- [X] T012 [P] [US1] Write unit test TestIsActionPressedUnbound in input/game_test.go (unbound action returns false, not error)
- [X] T013 [P] [US1] Write unit test TestStartStopDelegation in input/game_test.go (Start/Stop delegate to Input)
- [X] T014 [P] [US1] Write contract test TestBasicActionBinding in tests/contract/game_input_test.go (end-to-end test with real Input)
- [X] T015 [US1] Run tests to verify they FAIL by running `go test ./input -run TestGameInput -v` and `go test ./tests/contract -run TestBasicActionBinding -v`

### Implementation for User Story 1

- [X] T016 [P] [US1] Implement gameInputImpl.Start() method in input/game_impl.go (delegates to g.input.Start())
- [X] T017 [P] [US1] Implement gameInputImpl.Stop() method in input/game_impl.go (delegates to g.input.Stop())
- [X] T018 [US1] Implement gameInputImpl.Bind() method in input/game_impl.go (lock, update bindings map, unlock)
- [X] T019 [US1] Implement gameInputImpl.IsActionPressed() method in input/game_impl.go (RLock, lookup action, iterate keys calling Input.IsPressed, RUnlock)
- [X] T020 [US1] Add godoc comments to all exported types and functions in input/game.go and input/game_impl.go
- [X] T021 [US1] Verify User Story 1 tests PASS by running `go test ./input -run TestGameInput -v` and `go test ./tests/contract -run TestBasicActionBinding -v`
- [X] T022 [US1] Run golangci-lint and fix any violations by running `golangci-lint run ./input`
- [X] T023 [US1] Create basic example main.go in examples/game/ demonstrating single-key binding (jump→Space, fire→F, quit→ESC)
- [X] T024 [US1] Manual testing: run `go run examples/game/main.go` and verify Space, F, ESC work correctly

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Multiple Keys Per Action (Priority: P2)

**Goal**: Support alternative control schemes by allowing multiple keys to trigger the same action (WASD + arrows)

**Independent Test**: Bind "confirm" → Enter, Space, Y; press any of those keys; verify IsActionPressed("confirm") returns true for all keys

### Tests for User Story 2 (TDD - MUST FAIL before implementation) ⚠️

- [X] T025 [P] [US2] Write unit test TestBindMultipleKeys in input/game_test.go (bind 2+ keys to one action)
- [X] T026 [P] [US2] Write unit test TestMultipleKeysOrLogic in input/game_test.go (any key pressed → action returns true)
- [X] T027 [P] [US2] Write contract test TestMultipleKeysPerAction in tests/contract/game_input_test.go (WASD + arrows test)
- [X] T028 [US2] Run tests to verify they FAIL by running `go test ./input -run TestBindMultipleKeys|TestMultipleKeysOrLogic -v`

### Implementation for User Story 2

- [X] T029 [US2] Verify gameInputImpl.Bind() correctly stores slice of multiple keys (already implemented in US1, verify it works)
- [X] T030 [US2] Verify gameInputImpl.IsActionPressed() iterates all bound keys with early return (already implemented in US1, verify OR logic)
- [X] T031 [US2] Update examples/game/main.go to demonstrate multiple keys per action (add WASD + arrow key bindings for movement)
- [X] T032 [US2] Verify User Story 2 tests PASS by running `go test ./input -run TestBindMultipleKeys|TestMultipleKeysOrLogic -v`
- [X] T033 [US2] Run golangci-lint to ensure no new violations by running `golangci-lint run ./input`
- [X] T034 [US2] Manual testing: run `go run examples/game/main.go` and verify both arrow keys and WASD work for movement

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Dynamic Action Rebinding (Priority: P3)

**Goal**: Allow runtime key rebinding for player control customization

**Independent Test**: Bind "jump" → Space, then rebind to J, verify only J triggers jump (Space no longer works)

### Tests for User Story 3 (TDD - MUST FAIL before implementation) ⚠️

- [X] T035 [P] [US3] Write unit test TestBindReplace in input/game_test.go (rebind action to different key, old key no longer works)
- [X] T036 [P] [US3] Write unit test TestUnbind in input/game_test.go (bind action with empty keys removes binding)
- [X] T037 [P] [US3] Write unit test TestRebindingTakesEffectImmediately in input/game_test.go (no restart required)
- [X] T038 [P] [US3] Write contract test TestRuntimeRebinding in tests/contract/game_input_test.go (full rebinding flow)
- [X] T039 [US3] Run tests to verify they FAIL by running `go test ./input -run TestBindReplace|TestUnbind|TestRebindingTakesEffectImmediately -v`

### Implementation for User Story 3

- [X] T040 [US3] Verify gameInputImpl.Bind() replaces entire slice (already implemented in US1, no changes needed)
- [X] T041 [US3] Verify gameInputImpl.Bind() with zero keys deletes action from map (already implemented in US1, verify with empty slice)
- [X] T042 [US3] Add rebindAction helper function to examples/game/main.go (prompts for key, waits for input, rebinds)
- [X] T043 [US3] Add settingsMenu function to examples/game/main.go (shows rebind options, calls rebindAction)
- [X] T044 [US3] Update examples/game/main.go game loop to detect menu key and call settingsMenu
- [X] T045 [US3] Verify User Story 3 tests PASS by running `go test ./input -run TestBindReplace|TestUnbind|TestRebindingTakesEffectImmediately -v`
- [X] T046 [US3] Run golangci-lint to ensure no new violations by running `golangci-lint run ./input`
- [X] T047 [US3] Manual testing: run `go run examples/game/main.go`, enter settings, rebind controls, verify new bindings work

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Thread Safety & Performance Validation

**Purpose**: Validate concurrency and performance requirements from spec

- [X] T048 [P] Create input/game_concurrent_test.go with TestConcurrentIsActionPressed (10 goroutines, 1000 calls each)
- [X] T049 [P] Add TestConcurrentBindAndQuery to input/game_concurrent_test.go (1 writer, 5 readers, verify no race conditions)
- [X] T050 [P] Create input/game_bench_test.go with BenchmarkIsActionPressed_SingleKey
- [X] T051 [P] Add BenchmarkIsActionPressed_MultipleKeys to input/game_bench_test.go (3-5 keys bound)
- [X] T052 [P] Add BenchmarkBind to input/game_bench_test.go
- [X] T053 Run concurrent tests with race detector by running `go test ./input -race -run TestConcurrent -v`
- [X] T054 Run benchmarks and verify <1ms per IsActionPressed by running `go test ./input -bench=BenchmarkGameInput -benchmem`
- [X] T055 Verify all tests pass including race detector by running `go test ./... -race -v`

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and final validation

- [X] T056 [P] Update input/doc.go to include GameInput interface documentation and usage examples
- [X] T057 [P] Add GameInput section to README.md (if exists) or CLAUDE.md with basic usage example
- [X] T058 [P] Verify all exported types have godoc comments by running `go doc input.GameInput` and related commands
- [X] T059 [P] Add example of accessing underlying Input from GameInput to examples/game/main.go (for rebinding menu)
- [X] T060 Run full test suite with coverage by running `go test -cover ./... | tee coverage.txt`
- [X] T061 Verify >80% coverage for GameInput public APIs per constitution
- [X] T062 Run golangci-lint on entire codebase by running `golangci-lint run ./...`
- [X] T063 Verify all golangci-lint checks pass (errcheck, revive, cyclop <30)
- [X] T064 Manual cross-platform testing: verify example works on available platforms (Linux/macOS/Windows)
- [X] T065 Performance validation: verify SC-002 (<1ms), SC-003 (100 actions), SC-007 (60fps) from spec
- [X] T066 Update CLAUDE.md GameInput features section with performance characteristics and examples

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Thread Safety (Phase 6)**: Can proceed after any user story completes, but should follow US1 at minimum
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Builds on US1 implementation but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Uses US1/US2 implementation but independently testable

**NOTE**: US2 and US3 reuse US1's Bind() and IsActionPressed() implementations. They verify the implementations handle multiple keys and rebinding correctly, but don't require code changes.

### Within Each User Story

- Tests (TDD) MUST be written and FAIL before implementation (Constitution Principle IV)
- All unit tests before implementation
- Implementation tasks follow tests
- Contract tests validate after implementation
- Examples demonstrate the story
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- All implementation tasks marked [P] can run in parallel (different files)
- Thread safety tests (Phase 6) can run in parallel
- Polish tasks (Phase 7) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all unit tests for User Story 1 together:
Task: "Write unit test TestNewGameInput in input/game_test.go"
Task: "Write unit test TestBindSingleKey in input/game_test.go"
Task: "Write unit test TestIsActionPressedUnbound in input/game_test.go"
Task: "Write unit test TestStartStopDelegation in input/game_test.go"
Task: "Write contract test TestBasicActionBinding in tests/contract/game_input_test.go"

# After tests fail, launch all implementation tasks together:
Task: "Implement gameInputImpl.Start() in input/game_impl.go"
Task: "Implement gameInputImpl.Stop() in input/game_impl.go"
# (Bind and IsActionPressed have dependencies, run after these)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T009) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 (T010-T024)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

This gives you:
- ✅ Basic action mapping (single key per action)
- ✅ IsActionPressed queries
- ✅ Start/Stop lifecycle
- ✅ Working example
- ❌ No multiple keys per action yet
- ❌ No runtime rebinding yet

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (adds multi-key support)
4. Add User Story 3 → Test independently → Deploy/Demo (adds rebinding)
5. Add Thread Safety validation (Phase 6) → Performance validated
6. Add Polish (Phase 7) → Final release
7. Each increment adds value without breaking previous functionality

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (T010-T024)
   - Developer B: User Story 2 (T025-T034) - can start tests, waits for US1 implementation
   - Developer C: User Story 3 (T035-T047) - can start tests, waits for US1 implementation
3. Stories integrate (US2 and US3 verify US1 handles their cases)
4. Team converges on Thread Safety (Phase 6) together
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

- **Total Tasks**: 66
- **Setup Phase**: 4 tasks
- **Foundational Phase**: 5 tasks
- **User Story 1**: 15 tasks (6 tests + 9 implementation)
- **User Story 2**: 10 tasks (4 tests + 6 implementation)
- **User Story 3**: 13 tasks (5 tests + 8 implementation)
- **Thread Safety & Performance**: 8 tasks
- **Polish**: 11 tasks

**Parallel Opportunities**: 32 tasks marked [P] can run concurrently
**MVP Scope**: T001-T024 (Setup + Foundational + US1) = 24 tasks
**TDD Compliance**: 15 test tasks written before corresponding implementation

---

## Success Criteria Mapping

From specification → Task validation:

- **SC-001**: 5 lines of code → Validated in T023 (basic example)
- **SC-002**: <1ms queries → Validated in T054 (benchmarks)
- **SC-003**: 100 actions → Validated in T065 (stress test)
- **SC-004**: <20ms rebinding → Validated in T037, T045 (immediate effect)
- **SC-005**: Zero hardcoded keys → Validated in T023, T031, T044 (examples)
- **SC-006**: Thread-safe → Validated in T053 (race detector)
- **SC-007**: 60fps support → Validated in T054, T065 (benchmarks)

All success criteria have corresponding validation tasks.
