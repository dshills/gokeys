# Data Model: Performance and Efficiency Improvements

**Feature**: 003-fix-performance-issues
**Date**: 2025-10-18

## Overview

This feature modifies existing data structures to support UTF-8 decoding and buffer reuse. No new public types are added, maintaining backward compatibility.

## Modified Entities

### 1. unixBackend (Internal)

**Location**: `input/backend_unix.go`

**Current Fields**:
```go
type unixBackend struct {
    fd            int
    originalState *unix.Termios
    parser        *SequenceParser
    reader        io.Reader
    initialized   bool
}
```

**New Fields Added**:
```go
type unixBackend struct {
    fd            int
    originalState *unix.Termios
    parser        *SequenceParser
    reader        io.Reader
    initialized   bool

    // NEW: Accumulator for partial sequences (UTF-8, escape sequences)
    pendingBuf    []byte

    // NEW: Reusable buffer for Read() operations (reduces to field from local var)
    // Note: Alternative design uses sync.Pool instead of field
}
```

**Field Descriptions**:

- **pendingBuf**: Persistent byte buffer that accumulates partial sequences across multiple `Read()` calls
  - **Purpose**: Handle UTF-8 characters and escape sequences split across I/O boundaries
  - **Lifecycle**: Grows as bytes arrive, shrinks as complete sequences are parsed
  - **Max Size**: 16 bytes (longest escape sequence is F12 = 6 bytes, UTF-8 max = 4 bytes)
  - **Invariant**: Never contains more than one complete sequence at the start

**State Transitions**:
```
Empty → [Read chunk] → Partial → [Read more] → Complete → [Parse] → Empty
                         ↓
                    [Timeout/Error] → Parse anyway
```

---

### 2. Buffer Pool (New Global)

**Location**: `input/backend_unix.go`

**Definition**:
```go
var readBufferPool = sync.Pool{
    New: func() interface{} {
        b := make([]byte, 256)
        return &b
    },
}
```

**Purpose**: Reuse 256-byte read buffers across `ReadEvent()` calls to eliminate allocations

**Usage Pattern**:
```go
bufPtr := readBufferPool.Get().(*[]byte)
defer readBufferPool.Put(bufPtr)

buf := *bufPtr
n, _ := b.reader.Read(buf)

// CRITICAL: Copy data out before returning to pool
b.pendingBuf = append(b.pendingBuf, buf[:n]...)
```

**Safety Invariants**:
- Buffer MUST be returned to pool after use (via defer)
- Data MUST be copied to `pendingBuf` before buffer returns to pool
- Pooled buffers MUST NOT be referenced after `Put()`

---

### 3. Event (No Changes)

**Location**: `input/event.go`

**Current Definition**:
```go
type Event struct {
    Key        Key
    Rune       rune         // Already supports Unicode
    Modifiers  Modifier
    Timestamp  time.Time
    Pressed    bool
    Repeat     bool
}
```

**Impact**: No structural changes. The existing `Rune` field already supports full Unicode range (int32). UTF-8 decoding will populate this field with non-ASCII runes.

---

## Data Flow

### Current Flow (ASCII only)
```
Terminal → Read(buf[256]) → Parse(buf[:n]) → Event{Rune: 'a'}
           ↑ allocates      ↑ immediate
```

### New Flow (UTF-8 + Pooling)
```
Terminal → Read(pooled[256]) → Copy to pendingBuf → Parse when complete → Event{Rune: '日'}
           ↑ from pool          ↑ persistent         ↑ utf8.FullRune check
```

### Edge Case: Split UTF-8 Character
```
Read 1: [0xe6]           → pendingBuf = [0xe6]           → !utf8.FullRune → wait
Read 2: [0x97, 0xa5]     → pendingBuf = [0xe6,0x97,0xa5] → utf8.FullRune → parse '日'
```

### Edge Case: Escape Sequence
```
Read 1: [0x1b]           → pendingBuf = [0x1b]           → incomplete → wait
Read 2: [0x5b, 0x41]     → pendingBuf = [0x1b,0x5b,0x41] → complete (ArrowUp) → parse
```

### Edge Case: VTIME Timeout (Bare ESC)
```
Read 1: [0x1b]           → pendingBuf = [0x1b]           → incomplete → wait
Read 2: timeout (n=0)    → pendingBuf = [0x1b]           → parse as KeyEscape
```

---

## Validation Rules

### pendingBuf Constraints

1. **Max Size**: 16 bytes
   - Longest escape sequence: `ESC [ 2 4 ~` (F12) = 6 bytes
   - UTF-8 max: 4 bytes
   - Safety margin: 6 bytes
   - **Enforcement**: If `len(pendingBuf) > 16`, parse first byte as invalid and continue

2. **No Leading Complete Sequences**
   - After `Parse()`, `pendingBuf` MUST NOT start with a complete sequence
   - **Enforcement**: Parser consumes and removes exactly one complete sequence

3. **Thread Safety**
   - `pendingBuf` owned by `unixBackend` instance
   - `ReadEvent()` called from single goroutine (per constitution)
   - **No locking required**

### Buffer Pool Safety

1. **No Slice Escapes**
   - Pooled buffer data MUST be copied before return
   - **Violation Detection**: `go test -race` will catch slice escapes

2. **Deferred Return**
   - `defer readBufferPool.Put(bufPtr)` MUST be called immediately after `Get()`
   - **Enforcement**: Code review, no runtime check possible

---

## Migration Notes

### Backward Compatibility

**Public API**: No changes
- `Event` struct unchanged
- `Input` interface unchanged
- Existing tests continue to pass

**Internal Changes**:
- `unixBackend` adds fields (private struct)
- `ReadEvent()` signature unchanged
- Behavior changes:
  - Now supports UTF-8 (was ASCII-only)
  - Faster (no sleep)
  - Lower allocations

### Testing Strategy

**New Tests Required**:
```
tests/contract/utf8_test.go
├── TestUTF8TwoByteChars     (é, ñ, etc.)
├── TestUTF8ThreeByteChars   (€, 日, etc.)
├── TestUTF8FourByteChars    (𝄞, emoji)
└── TestUTF8SplitAcrossReads (partial sequence handling)

tests/benchmarks/input_bench_test.go
├── BenchmarkEscapeKeyLatency      (before/after comparison)
├── BenchmarkReadEventAllocations  (verify 0 allocs/op)
└── BenchmarkUTF8Decoding          (performance of multi-byte chars)
```

**Modified Tests**:
- `tests/contract/normalization_test.go`: Add UTF-8 characters to existing cases
- `tests/integration/unix_test.go`: Add timing assertions (verify no 5ms delay)

---

## Performance Characteristics

### Memory

**Before**:
- Allocation per event: 256 bytes (buffer)
- At 60 FPS: 15,360 bytes/sec = 15 KB/s garbage

**After**:
- Allocation per event: 0 bytes (pooled buffer)
- Persistent memory: ~16 bytes (pendingBuf, worst case)
- Net improvement: ~99.9% reduction in allocations

### Latency

**Before**:
- Escape key: 5ms minimum (time.Sleep)
- Other keys: <1ms

**After**:
- All keys: <1ms (no artificial delay)
- Escape sequences: Bounded by VTIME (100ms max)

### CPU

**Before**:
- UTF-8 lookup: O(1) for ASCII, KeyUnknown for UTF-8
- Sleep overhead: ~5ms idle per escape

**After**:
- UTF-8 decode: O(1) via `utf8.DecodeRune` (optimized assembly for common cases)
- No sleep overhead
- Net: Slightly higher CPU for UTF-8 decoding, but negligible (<1μs)

---

## Open Questions

**Q: Should pendingBuf be pre-allocated?**
**A**: No. Start empty (`nil` slice), grow as needed. Max 16 bytes is tiny, and most events don't need accumulation.

**Q: What happens if pendingBuf grows beyond 16 bytes?**
**A**: Safety check: parse first byte as invalid (`RuneError`), shift buffer, continue. This prevents memory leaks on malformed input.

**Q: Should we add metrics for pendingBuf usage?**
**A**: Out of scope. Could add in future (average size, max size, flush rate) but not needed for MVP.

---

**Data Model Complete**: All entities defined, validation rules specified, backward compatibility ensured.
