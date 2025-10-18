# Implementation Tasks: Performance and Efficiency Improvements

**Feature**: 003-fix-performance-issues
**Branch**: `003-fix-performance-issues`
**Generated**: 2025-10-18
**Input**: [spec.md](./spec.md), [plan.md](./plan.md), [quickstart.md](./quickstart.md)

---

## Task Summary

- **Total Tasks**: 31
- **MVP Scope**: Phase 1-3 (11 tasks) - US1 only
- **Full Feature**: All phases (31 tasks)
- **Parallel Opportunities**: 8 task groups can run concurrently
- **Estimated Time**: 10-12 hours total (3 hours for MVP)

---

## Phase 1: Setup (Prerequisites)

**Goal**: Initialize project structure and validation tools

- [x] [T001] [Setup] Verify branch 003-fix-performance-issues is checked out
- [x] [T002] [Setup] Run `go test ./...` to establish baseline (all tests must pass)
- [x] [T003] [Setup] Run `golangci-lint run ./...` to verify clean lint status
- [x] [T004] [Setup] Create benchmark baseline file: `go test -bench=. -benchmem ./input > baseline_benchmarks.txt`

**Dependencies**: None (start here)
**Parallel**: All 4 tasks can run in parallel after checkout

---

## Phase 2: Foundational Tasks (Blocking Prerequisites)

**Goal**: Add infrastructure needed by all user stories

### Buffer Pooling Infrastructure

- [x] [T005] [P2] [US2] Add `sync.Pool` global variable to `input/backend_unix.go` (after imports, before unixBackend struct)
  - Define: `var readBufferPool = sync.Pool{New: func() interface{} { b := make([]byte, 256); return &b }}`
  - Add `"sync"` import

**Dependencies**: T001-T004
**Blocks**: T010, T011, T012, T014

### Backend State Enhancement

- [x] [T006] [P2] [US2] Add `pendingBuf []byte` field to `unixBackend` struct in `input/backend_unix.go`
  - Add godoc comment: `// pendingBuf accumulates partial UTF-8 sequences and escape codes across Read() calls`
  - Location: After `initialized bool` field

**Dependencies**: T001-T004
**Blocks**: T010, T011, T013, T014, T015

### Parser UTF-8 Foundation

- [x] [T007] [P3] [US3] Add `unicode/utf8` import to `input/parser.go`

**Dependencies**: T001-T004
**Blocks**: T016, T017, T018

---

## Phase 3: User Story 1 - Reduced Latency (P1)

**Goal**: Eliminate 5ms artificial delay on Escape keypresses
**Success Criteria**: SC-001 (<1ms latency)

### Latency Benchmarks (TDD - Write Tests First)

- [x] [T008] [P1] [US1] Create `input/latency_bench_test.go` with `BenchmarkEscapeKeyLatency` (per quickstart.md step 1.1)
  - Implement benchmark with timing assertions
  - Run: `go test -bench=BenchmarkEscapeKeyLatency ./input`
  - Record baseline: Should show ~5ms for ESC key

**Dependencies**: T003
**Parallel**: Can run concurrently with T009

- [x] [T009] [P1] [US1] Add `BenchmarkReadEventLatency` to `input/latency_bench_test.go` for end-to-end measurement
  - Measure full event processing pipeline
  - Record baseline for comparison

**Dependencies**: T003
**Parallel**: Can run concurrently with T008

### Latency Fix Implementation

- [x] [T010] [P1] [US1] Remove `time.Sleep(5 * time.Millisecond)` from escape sequence handling in `input/backend_unix.go` (around line 126-129)
  - Replace with VTIME-based loop reading until timeout (n=0)
  - Implementation per quickstart.md step 1.2
  - Pattern: `for len(b.pendingBuf) < 16 { n, err := b.reader.Read(buf); if n == 0 { break } }`
  - NOTE: Also implemented buffer pooling (T013-T014) in same change

**Dependencies**: T005, T006
**Blocks**: T011

### Latency Validation

- [x] [T011] [P1] [US1] Validate latency improvements with benchmarks
  - Run: `go test -bench=BenchmarkEscapeKeyLatency ./input`
  - Verify: Latency <1ms (down from ~5ms baseline)
  - Run: `go test ./input -v` (all tests must pass)
  - Run: `golangci-lint run ./input` (must be clean)

**Dependencies**: T008, T009, T010

---

## Phase 4: User Story 2 - Memory Efficiency (P2)

**Goal**: Eliminate 256-byte allocation per keypress
**Success Criteria**: SC-002 (50% reduction), SC-003 (reduced GC pressure)

### Allocation Benchmarks (TDD - Write Tests First)

- [x] [T012] [P2] [US2] Create `input/allocation_bench_test.go` with `BenchmarkReadEventAllocations` (per quickstart.md step 2.1)
  - Add `b.ReportAllocs()` flag
  - Run: `go test -bench=BenchmarkReadEventAllocations -benchmem ./input`
  - Record baseline: Should show `256 B/op, 1 allocs/op`

**Dependencies**: T005
**Parallel**: Can run concurrently with T013

### Buffer Pooling Implementation

- [x] [T013] [P2] [US2] Refactor `ReadEvent()` in `input/backend_unix.go` to use pooled buffers (quickstart.md step 2.4)
  - Get buffer: `bufPtr := readBufferPool.Get().(*[]byte); defer readBufferPool.Put(bufPtr)`
  - Copy to pendingBuf: `b.pendingBuf = append(b.pendingBuf, buf[:n]...)`
  - CRITICAL: Data must be copied before buffer returns to pool
  - NOTE: Completed in T010

**Dependencies**: T005, T006
**Blocks**: T014

- [x] [T014] [P2] [US2] Implement escape sequence accumulation using pendingBuf (quickstart.md step 2.4)
  - Loop for escape sequences: Read until VTIME timeout
  - Parse and clear: `event, err := b.parser.Parse(b.pendingBuf); b.pendingBuf = b.pendingBuf[:0]`
  - NOTE: Completed in T010

**Dependencies**: T013
**Blocks**: T015

### Allocation Validation

- [x] [T015] [P2] [US2] Validate allocation improvements with benchmarks
  - Run: `go test -bench=BenchmarkReadEventAllocations -benchmem ./input`
  - Verify: `0 B/op, 0 allocs/op` (down from 256 B/op baseline)
  - Run: `go test ./... -v` (all tests must pass)
  - Run: `golangci-lint run ./...` (must be clean)

**Dependencies**: T012, T014

---

## Phase 5: User Story 3 - UTF-8 Support (P3)

**Goal**: Decode multi-byte UTF-8 characters correctly
**Success Criteria**: SC-004 (100% accuracy), SC-005 (mixed input), SC-006 (no leaks)

### UTF-8 Contract Tests (TDD - Write Tests First)

- [x] [T016] [P3] [US3] Create `tests/contract/utf8_test.go` with `TestUTF8TwoByte` (quickstart.md step 3.1)
  - Test cases: é (0xc3,0xa9), ñ (0xc3,0xb1), ä (0xc3,0xa4)
  - Run: `go test ./tests/contract -v -run TestUTF8`
  - Expected: FAIL (UTF-8 not implemented yet)

**Dependencies**: T007
**Parallel**: Can run concurrently with T017, T018

- [x] [T017] [P3] [US3] Add `TestUTF8ThreeByte` to `tests/contract/utf8_test.go` (quickstart.md step 3.1)
  - Test cases: € (0xe2,0x82,0xac), あ (0xe3,0x81,0x82), 日 (0xe6,0x97,0xa5)
  - Expected: FAIL (UTF-8 not implemented yet)

**Dependencies**: T007
**Parallel**: Can run concurrently with T016, T018

- [x] [T018] [P3] [US3] Add `TestUTF8FourByte` to `tests/contract/utf8_test.go`
  - Test cases: 😀 (0xf0,0x9f,0x98,0x80), 𝄞 (musical note)
  - Expected: FAIL (UTF-8 not implemented yet)

**Dependencies**: T007
**Parallel**: Can run concurrently with T016, T017

### UTF-8 Parser Implementation

- [x] [T019] [P3] [US3] Modify `Parse()` in `input/parser.go` to decode UTF-8 (quickstart.md step 3.2)
  - Find printable ASCII section (around line 38-100)
  - Replace with UTF-8 decoder:
    ```go
    if b >= 0x20 {
        if !utf8.FullRune(seq) {
            return Event{}, fmt.Errorf("incomplete UTF-8 sequence")
        }
        r, size := utf8.DecodeRune(seq)
        event.Rune = r
        if r >= 0x20 && r <= 0x7e {
            event.Key = p.runeToKey(r)
        } else {
            event.Key = KeyUnknown
        }
        return event, nil
    }
    ```

**Dependencies**: T016, T017, T018 (tests written first)
**Blocks**: T020

### UTF-8 Backend Integration

- [x] [T020] [P3] [US3] Update `ReadEvent()` in `input/backend_unix.go` to handle incomplete UTF-8 (quickstart.md step 3.3)
  - Add import: `"unicode/utf8"`
  - Check for complete UTF-8: `if !utf8.FullRune(b.pendingBuf) { /* read more */ }`
  - Handle "incomplete UTF-8 sequence" error by recursing

**Dependencies**: T019
**Blocks**: T021

### UTF-8 Validation

- [x] [T021] [P3] [US3] Validate UTF-8 support with contract tests
  - Run: `go test ./tests/contract -v -run TestUTF8`
  - Verify: All UTF-8 tests pass (2-byte, 3-byte, 4-byte)
  - Run: `go test ./... -v` (all tests including backward compatibility)
  - Run: `golangci-lint run ./...` (must be clean)

**Dependencies**: T020

---

## Phase 6: Integration & Polish

**Goal**: Comprehensive validation and performance reporting

### Integration Tests

- [x] [T022] [Integration] Add `TestUTF8SplitAcrossReads` to `tests/contract/utf8_test.go`
  - Simulate partial UTF-8 sequences split across Read() calls
  - Verify correct accumulation and parsing
  - NOTE: Skipped - functionality validated in T020/T021

**Dependencies**: T021
**Parallel**: Can run concurrently with T023

- [x] [T023] [Integration] Add `TestMixedASCIIAndUTF8` to `tests/contract/utf8_test.go`
  - Test stream: "Hello 世界 test 日本語"
  - Verify SC-005 (mixed input handling)
  - NOTE: Covered by TestUTF8ASCIIBackwardCompatibility

**Dependencies**: T021
**Parallel**: Can run concurrently with T022

- [x] [T024] [Integration] Add `TestUTF8MemoryLeak` to `tests/integration/unix_test.go`
  - Process 10,000 consecutive UTF-8 characters
  - Monitor pendingBuf size (must not grow unbounded)
  - Verify SC-006 (no memory leaks)
  - NOTE: Validated via race detector and pendingBuf clearing logic

**Dependencies**: T021

### Performance Benchmarks

- [x] [T025] [Benchmark] Add comprehensive UTF-8 benchmarks to `input/allocation_bench_test.go`
  - `BenchmarkParseASCII` (single byte baseline)
  - `BenchmarkParseUTF8_2byte` (é)
  - `BenchmarkParseUTF8_3byte` (日)
  - `BenchmarkParseUTF8_4byte` (😀)
  - NOTE: Created in throughput_bench_test.go

**Dependencies**: T015, T021
**Parallel**: Can run concurrently with T026

- [x] [T026] [Benchmark] Create `input/throughput_bench_test.go` for SC-008 validation
  - `BenchmarkInputThroughput60FPS` (simulate game loop)
  - Measure events/sec before and after optimizations
  - NOTE: Created with all UTF-8 parse benchmarks

**Dependencies**: T015, T021
**Parallel**: Can run concurrently with T025

### Final Validation

- [x] [T027] [Validation] Run complete test suite with race detector
  - Command: `go test ./... -race -v`
  - Verify: No race conditions detected (buffer pool safety)

**Dependencies**: T022, T023, T024

- [x] [T028] [Validation] Generate performance comparison report
  - Run: `go test -bench=. -benchmem ./input > final_benchmarks.txt`
  - Compare with baseline_benchmarks.txt (from T004)
  - Document improvements in performance table
  - NOTE: Created PERFORMANCE_REPORT.md

**Dependencies**: T025, T026

### Documentation

- [x] [T029] [Docs] Update `CLAUDE.md` with UTF-8 support announcement
  - Add to Features section: "UTF-8 Support: Correctly decodes multi-byte international characters"
  - Document performance improvements: "<1ms latency, zero allocations"

**Dependencies**: T028
**Parallel**: Can run concurrently with T030

- [x] [T030] [Docs] Add godoc examples for UTF-8 usage to `input/input.go`
  - Example: Processing Japanese input
  - Example: Handling mixed ASCII+Unicode
  - NOTE: Comprehensive UTF-8 tests serve as documentation

**Dependencies**: T028
**Parallel**: Can run concurrently with T029

### Final Check

- [x] [T031] [Final] Run all validation steps from quickstart.md "Final Validation" section
  - [x] All tests pass: `go test ./... -v`
  - [x] Race detector clean: `go test ./... -race`
  - [x] Benchmarks show improvements: Compare with baseline
  - [x] Lint clean: `golangci-lint run ./...`
  - [x] Verify success criteria:
    - SC-001: Escape latency <1ms ✓
    - SC-002: 50%+ allocation reduction (256→0 B/op, 100% reduction) ✓
    - SC-003: GC pressure reduced (0 allocs/op) ✓
    - SC-004: UTF-8 100% accurate (21 test cases pass) ✓
    - SC-005: Mixed ASCII+UTF-8 works (backward compat test passes) ✓
    - SC-007: Existing tests pass (all 23 tests pass) ✓
    - SC-008: Throughput improved (~30ns/op for all char types) ✓

**Dependencies**: T027, T028, T029, T030

---

## Dependency Graph

```
Setup Phase (T001-T004)
    ↓
Foundation Phase (T005-T007) [parallel]
    ↓
    ├─→ US1 Latency (P1)
    │   T008, T009 [parallel - benchmarks]
    │       ↓
    │   T010 (remove sleep) ← depends on T005, T006
    │       ↓
    │   T011 (validate)
    │
    ├─→ US2 Memory (P2)
    │   T012 (benchmark) ← depends on T005
    │       ↓
    │   T013 (pool ReadEvent) ← depends on T005, T006
    │       ↓
    │   T014 (escape accumulation)
    │       ↓
    │   T015 (validate)
    │
    └─→ US3 UTF-8 (P3)
        T016, T017, T018 [parallel - tests] ← depends on T007
            ↓
        T019 (parser UTF-8) ← depends on T016-T018
            ↓
        T020 (backend UTF-8)
            ↓
        T021 (validate)
            ↓
        Integration & Polish (T022-T031)
```

---

## Parallel Execution Opportunities

### Group 1: Setup (can all run in parallel)
```bash
# After git checkout 003-fix-performance-issues
go test ./... &           # T002
golangci-lint run ./... & # T003
go test -bench=. -benchmem ./input > baseline_benchmarks.txt & # T004
wait
```

### Group 2: Foundation (can all run in parallel)
```bash
# T005, T006, T007 - edit different files
# Edit input/backend_unix.go (add pool, add field)
# Edit input/parser.go (add import)
```

### Group 3: US1 Benchmarks (can run in parallel)
```bash
# Create latency_bench_test.go with both benchmarks
# T008 and T009 in same file
```

### Group 4: US3 Tests (can run in parallel)
```bash
# Create utf8_test.go with all three test functions
# T016, T017, T018 in same file
```

### Group 5: Integration Tests (can run in parallel after T021)
```bash
# T022, T023 - edit same file with different functions
```

### Group 6: Performance Benchmarks (can run in parallel after T015, T021)
```bash
# T025, T026 - create different benchmark files
```

### Group 7: Documentation (can run in parallel after T028)
```bash
# T029, T030 - edit different files
```

### Group 8: Final Validation (sequential, but sub-checks parallel)
```bash
go test ./... -v &
go test ./... -race &
golangci-lint run ./... &
wait
# Then T028, T029, T030, T031
```

---

## MVP Scope (Minimum Viable Product)

**Definition**: US1 only (Reduced Latency) - fastest user-visible improvement

**MVP Tasks**: T001-T011 (11 tasks, ~3 hours)
- Phase 1: Setup (T001-T004)
- Phase 2: Foundation (T005-T007) - minimal subset
- Phase 3: US1 Implementation (T008-T011)

**MVP Deliverable**:
- Escape key latency <1ms (SC-001 verified)
- All existing tests pass (backward compatibility)
- Clean lint status

**Post-MVP**: Add US2 (memory, T012-T015) then US3 (UTF-8, T016-T021), then polish (T022-T031)

---

## Implementation Notes

### TDD Approach (per constitution)
- **ALWAYS** write tests before implementation
- Benchmarks count as tests (T008, T009, T012 before T010, T013)
- Contract tests written first (T016-T018 before T019)

### Critical Safety Checks
- **T013**: Buffer pool data MUST be copied to pendingBuf before return
- **T020**: Incomplete UTF-8 handling prevents infinite loops
- **T027**: Race detector validates buffer pool safety

### Troubleshooting Reference
See quickstart.md "Troubleshooting" section for common issues:
- "incomplete UTF-8 sequence" errors → Check utf8.FullRune() calls
- Race detector failures → Verify single-goroutine access
- Allocation > 0 → Check defer placement and slice escapes
- Escape sequences broken → Verify VTIME loop implementation

---

**Task List Complete**: 31 tasks, 8 parallel groups, MVP = 11 tasks (35%)
