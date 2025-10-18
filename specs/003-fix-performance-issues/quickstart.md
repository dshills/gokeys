# Quickstart: Performance and Efficiency Improvements

**Feature**: 003-fix-performance-issues
**Branch**: `003-fix-performance-issues`
**Date**: 2025-10-18

## Overview

This guide provides a step-by-step implementation plan for the performance optimizations. Follow the phases in order, as each builds on the previous.

## Prerequisites

- Branch `003-fix-performance-issues` checked out
- All existing tests passing: `go test ./...`
- golangci-lint clean: `golangci-lint run ./...`
- Familiarity with:
  - `unicode/utf8` package
  - `sync.Pool` usage
  - Unix termios (VMIN/VTIME)

---

## Phase 1: Remove Escape Key Latency (P1)

**Goal**: Eliminate 5ms artificial delay on Escape keypresses

**Estimated Time**: 2-3 hours

### Step 1.1: Add Latency Benchmark

**File**: `input/latency_bench_test.go` (new)

```go
package input

import (
    "testing"
    "time"
)

func BenchmarkEscapeKeyLatency(b *testing.B) {
    backend := newBackend().(*unixBackend)
    if err := backend.Init(); err != nil {
        b.Skip("Not a terminal")
    }
    defer backend.Restore()

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        start := time.Now()
        // Simulate ESC press (manual test)
        _, _ = backend.ReadEvent()
        latency := time.Since(start)

        if latency > 10*time.Millisecond {
            b.Logf("Warning: High latency %v", latency)
        }
    }
}
```

**Run**: `go test -bench=BenchmarkEscapeKeyLatency ./input`

**Baseline**: Record current latency (should show ~5ms for ESC)

### Step 1.2: Remove time.Sleep

**File**: `input/backend_unix.go`

**Find** (around line 126-129):
```go
if buf[0] == 0x1b {
    time.Sleep(5 * time.Millisecond)  // ← Remove this line
    n2, err := b.reader.Read(buf[1:])
    // ...
}
```

**Replace with**:
```go
if buf[0] == 0x1b {
    // Read additional bytes with VTIME timeout (no artificial delay)
    n2, err := b.reader.Read(buf[1:])
    if err != nil && err != io.EOF {
        return b.parser.Parse(buf[:1])  // Treat as bare ESC
    }
    n += n2
}
```

### Step 1.3: Validate

```bash
# Run tests
go test ./input

# Run benchmark (compare with baseline)
go test -bench=BenchmarkEscapeKeyLatency ./input

# Lint
golangci-lint run ./input
```

**Success Criteria**: Latency reduced from ~5ms to <1ms

---

## Phase 2: Add Buffer Pooling (P2)

**Goal**: Eliminate 256-byte allocation per keypress

**Estimated Time**: 3-4 hours

### Step 2.1: Add Allocation Benchmark

**File**: `input/allocation_bench_test.go` (new)

```go
package input

import "testing"

func BenchmarkReadEventAllocations(b *testing.B) {
    backend := newBackend().(*unixBackend)
    backend.Init()
    defer backend.Restore()

    b.ReportAllocs()  // Critical: enable allocation tracking
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, _ = backend.ReadEvent()
    }
}
```

**Run**: `go test -bench=BenchmarkReadEventAllocations -benchmem ./input`

**Baseline**: Should show `256 B/op, 1 allocs/op`

### Step 2.2: Add sync.Pool

**File**: `input/backend_unix.go`

**Add at package level** (after imports):
```go
var readBufferPool = sync.Pool{
    New: func() interface{} {
        b := make([]byte, 256)
        return &b
    },
}
```

**Add to imports**:
```go
import (
    "sync"  // ← Add this
    // ... existing imports
)
```

### Step 2.3: Add pendingBuf Field

**File**: `input/backend_unix.go`

**Modify struct** (around line 17-23):
```go
type unixBackend struct {
    fd            int
    originalState *unix.Termios
    parser        *SequenceParser
    reader        io.Reader
    initialized   bool

    // Accumulator for partial UTF-8 and escape sequences
    pendingBuf    []byte
}
```

### Step 2.4: Update ReadEvent()

**File**: `input/backend_unix.go`

**Replace entire function** (around line 113-142):
```go
func (b *unixBackend) ReadEvent() (Event, error) {
    // Get buffer from pool
    bufPtr := readBufferPool.Get().(*[]byte)
    defer readBufferPool.Put(bufPtr)
    buf := *bufPtr

    // Read chunk
    n, err := b.reader.Read(buf)
    if err != nil {
        return Event{}, err
    }
    if n == 0 {
        return Event{}, io.EOF
    }

    // Append to pending buffer (CRITICAL: copy before pool return)
    b.pendingBuf = append(b.pendingBuf, buf[:n]...)

    // For escape sequences, continue reading until VTIME timeout
    if len(b.pendingBuf) > 0 && b.pendingBuf[0] == 0x1b {
        for len(b.pendingBuf) < 16 {  // Max escape sequence length
            n, err := b.reader.Read(buf)
            if err != nil || n == 0 {
                break  // VTIME timeout or error
            }
            b.pendingBuf = append(b.pendingBuf, buf[:n]...)
        }
    }

    // Parse and clear buffer
    event, err := b.parser.Parse(b.pendingBuf)
    b.pendingBuf = b.pendingBuf[:0]  // Clear for next read

    return event, err
}
```

### Step 2.5: Validate

```bash
# Run allocation benchmark
go test -bench=BenchmarkReadEventAllocations -benchmem ./input

# Should show: 0 B/op, 0 allocs/op ✅

# Run all tests
go test ./...

# Lint
golangci-lint run ./...
```

**Success Criteria**: 0 allocations/op, all tests pass

---

## Phase 3: Add UTF-8 Support (P3)

**Goal**: Decode multi-byte UTF-8 characters correctly

**Estimated Time**: 4-5 hours

### Step 3.1: Add UTF-8 Test Cases

**File**: `tests/contract/utf8_test.go` (new)

```go
package contract_test

import (
    "testing"
    "github.com/dshills/gokeys/input"
)

func TestUTF8TwoByte(t *testing.T) {
    parser := input.NewSequenceParser()

    tests := []struct {
        name string
        seq  []byte
        want rune
    }{
        {"e-acute", []byte{0xc3, 0xa9}, 'é'},        // U+00E9
        {"n-tilde", []byte{0xc3, 0xb1}, 'ñ'},        // U+00F1
        {"a-umlaut", []byte{0xc3, 0xa4}, 'ä'},       // U+00E4
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            event, err := parser.Parse(tt.seq)
            if err != nil {
                t.Fatalf("Parse() error = %v", err)
            }
            if event.Rune != tt.want {
                t.Errorf("Rune = %c (U+%04X), want %c (U+%04X)",
                    event.Rune, event.Rune, tt.want, tt.want)
            }
        })
    }
}

func TestUTF8ThreeByte(t *testing.T) {
    parser := input.NewSequenceParser()

    tests := []struct {
        name string
        seq  []byte
        want rune
    }{
        {"euro", []byte{0xe2, 0x82, 0xac}, '€'},      // U+20AC
        {"hiragana-a", []byte{0xe3, 0x81, 0x82}, 'あ'}, // U+3042
        {"kanji-day", []byte{0xe6, 0x97, 0xa5}, '日'},  // U+65E5
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            event, err := parser.Parse(tt.seq)
            if err != nil {
                t.Fatalf("Parse() error = %v", err)
            }
            if event.Rune != tt.want {
                t.Errorf("Rune = %c (U+%04X), want %c (U+%04X)",
                    event.Rune, event.Rune, tt.want, tt.want)
            }
        })
    }
}
```

**Run**: `go test ./tests/contract -v`

**Expected**: Tests FAIL (UTF-8 not implemented yet)

### Step 3.2: Add UTF-8 Decoding to Parser

**File**: `input/parser.go`

**Add import**:
```go
import (
    "unicode/utf8"  // ← Add this
    // ... existing imports
)
```

**Modify Parse() function** (around line 38-100):

**Find** this section:
```go
// Printable ASCII
if b >= 0x20 && b <= 0x7e {
    event.Rune = rune(b)
    event.Key = p.runeToKey(event.Rune)
    return event, nil
}
```

**Replace with**:
```go
// UTF-8 or printable ASCII
if b >= 0x20 {
    // Check for complete UTF-8 sequence
    if !utf8.FullRune(seq) {
        // Incomplete - caller should accumulate more bytes
        return Event{}, fmt.Errorf("incomplete UTF-8 sequence")
    }

    r, size := utf8.DecodeRune(seq)
    event.Rune = r

    // Only ASCII letters/numbers get specific Key codes
    if r >= 0x20 && r <= 0x7e {
        event.Key = p.runeToKey(r)
    } else {
        // Non-ASCII characters map to KeyUnknown
        event.Key = KeyUnknown
    }

    return event, nil
}
```

### Step 3.3: Handle Incomplete UTF-8 in ReadEvent()

**File**: `input/backend_unix.go`

**Update ReadEvent()** to handle incomplete UTF-8:

```go
func (b *unixBackend) ReadEvent() (Event, error) {
    // ... buffer pool code ...

    b.pendingBuf = append(b.pendingBuf, buf[:n]...)

    // For UTF-8, check if we have a complete character
    if len(b.pendingBuf) > 0 && b.pendingBuf[0] >= 0x80 {  // Non-ASCII
        if !utf8.FullRune(b.pendingBuf) {
            // Incomplete UTF-8 - read more bytes
            if len(b.pendingBuf) < utf8.UTFMax {
                n, err := b.reader.Read(buf)
                if err == nil && n > 0 {
                    b.pendingBuf = append(b.pendingBuf, buf[:n]...)
                }
            }
        }
    }

    // ... escape sequence code ...

    event, err := b.parser.Parse(b.pendingBuf)
    if err != nil {
        // If incomplete, try reading more
        if err.Error() == "incomplete UTF-8 sequence" {
            return b.ReadEvent()  // Recurse
        }
    }

    b.pendingBuf = b.pendingBuf[:0]
    return event, err
}
```

**Add import**:
```go
import (
    "unicode/utf8"  // ← Add this
    // ... existing imports
)
```

### Step 3.4: Validate

```bash
# Run UTF-8 tests
go test ./tests/contract -v -run TestUTF8

# Run all tests
go test ./...

# Benchmark UTF-8 performance
go test -bench=. ./input

# Lint
golangci-lint run ./...
```

**Success Criteria**:
- All UTF-8 tests pass
- Existing ASCII tests still pass
- 0 lint issues

---

## Final Validation

### Run Complete Test Suite

```bash
# All tests
go test ./... -v

# With race detector
go test ./... -race

# Benchmarks
go test -bench=. -benchmem ./input

# Lint
golangci-lint run ./...
```

### Verify Success Criteria

- [ ] **SC-001**: Escape latency <1ms (check benchmark output)
- [ ] **SC-002**: 50%+ allocation reduction (0 B/op vs 256 B/op baseline)
- [ ] **SC-003**: GC pressure reduced (run with `-benchmem`)
- [ ] **SC-004**: UTF-8 100% accurate (all TestUTF8* pass)
- [ ] **SC-005**: Mixed ASCII+UTF-8 works (integration test)
- [ ] **SC-007**: Existing tests pass (backward compatibility)

### Performance Report

```bash
# Generate before/after comparison
go test -bench=. -benchmem ./input > after.txt

# Compare with baseline (from git stash or separate branch)
git stash
go test -bench=. -benchmem ./input > before.txt
git stash pop

# Use benchcmp or manual comparison
# Expected improvements:
#   - Latency: 5x-10x faster for Escape
#   - Allocations: 256 B/op → 0 B/op
#   - Throughput: Higher ops/sec
```

---

## Troubleshooting

### Issue: Tests fail with "incomplete UTF-8 sequence"

**Cause**: Parser called before complete character accumulated

**Fix**: Ensure ReadEvent() checks `utf8.FullRune()` before parsing

### Issue: Race detector reports data race on pendingBuf

**Cause**: Concurrent access to backend (violates contract)

**Fix**: Verify only single goroutine calls ReadEvent()

### Issue: Benchmark shows allocations > 0

**Cause**: Pooled buffer not properly returned or slice escape

**Fix**: Check defer placement and ensure data copied before pool return

### Issue: Escape sequences broken

**Cause**: Removed sleep but didn't implement VTIME loop

**Fix**: Add loop to read until n=0 (VTIME timeout) for escape sequences

---

## Next Steps

After completing this quickstart:

1. Run `/speckit.tasks` to generate detailed task list
2. Implement tasks in priority order (P1 → P2 → P3)
3. Create PR with benchmarks showing improvements
4. Update documentation with UTF-8 support announcement

---

**Quickstart Complete**: Ready for implementation!
