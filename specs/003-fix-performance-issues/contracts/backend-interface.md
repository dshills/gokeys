# Backend Interface Contract

**Feature**: 003-fix-performance-issues
**Date**: 2025-10-18

## Overview

This contract defines the expected behavior of the `Backend.ReadEvent()` method after performance optimizations. The interface signature remains unchanged, but internal behavior is enhanced for UTF-8 support and reduced latency.

## Interface Definition

```go
type Backend interface {
    Init() error
    Restore() error
    ReadEvent() (Event, error)  // ← Focus of this contract
}
```

## ReadEvent() Contract

### Signature
```go
func (b *Backend) ReadEvent() (Event, error)
```

### Purpose
Read and parse the next keyboard event from the terminal, blocking until an event is available or an error occurs.

---

## Behavioral Contracts

### 1. UTF-8 Character Handling

**Contract**: ReadEvent() MUST correctly decode multi-byte UTF-8 characters

**Test Cases**:

| Input Bytes | Expected Event.Rune | Expected Event.Key |
|-------------|---------------------|-------------------|
| `[0x61]` | `'a'` (U+0061) | `KeyA` |
| `[0xc3, 0xa9]` | `'é'` (U+00E9) | `KeyUnknown` (non-ASCII) |
| `[0xe2, 0x82, 0xac]` | `'€'` (U+20AC) | `KeyUnknown` |
| `[0xf0, 0x9f, 0x98, 0x80]` | `'😀'` (U+1F600) | `KeyUnknown` |
| `[0xe6, 0x97, 0xa5]` | `'日'` (U+65E5) | `KeyUnknown` |

**Error Handling**:
- **Incomplete UTF-8**: Block until more bytes arrive or VTIME timeout
- **Invalid UTF-8**: Return `Event{Rune: utf8.RuneError, Key: KeyUnknown}`
- **Timeout mid-sequence**: Parse partial bytes as invalid

---

### 2. Escape Sequence Latency

**Contract**: ReadEvent() MUST NOT introduce artificial delays for escape sequences

**Requirements**:
- **Bare ESC press**: Return within 1ms of key press (excluding VTIME)
- **Escape sequences** (arrows, F-keys): Return within VTIME period (100ms max)
- **No sleep calls**: MUST NOT use `time.Sleep` for timing

**Test Validation**:
```go
func BenchmarkEscapeKeyLatency(b *testing.B) {
    backend := NewUnixBackend()
    backend.Init()
    defer backend.Restore()

    // Simulate ESC press
    simulateKeyPress(0x1b)

    start := time.Now()
    event, _ := backend.ReadEvent()
    latency := time.Since(start)

    if latency > 1*time.Millisecond {
        b.Errorf("Escape latency %v exceeds 1ms", latency)
    }
}
```

---

### 3. Memory Allocation

**Contract**: ReadEvent() MUST minimize heap allocations

**Requirements**:
- **Target**: 0 allocations per call (excluding initial warmup)
- **Implementation**: Use `sync.Pool` for read buffers
- **Validation**: `go test -bench=BenchmarkReadEvent -benchmem`

**Expected Benchmark Output**:
```
BenchmarkReadEvent-8    1000000    1200 ns/op    0 B/op    0 allocs/op
                                                  ↑         ↑
                                                  zero      zero
```

---

### 4. Partial Sequence Handling

**Contract**: ReadEvent() MUST accumulate partial sequences across multiple Read() calls

**Scenario 1: UTF-8 Split Across Reads**
```
Call 1: Read() returns [0xe6]              → Block (incomplete)
Call 2: Read() returns [0x97, 0xa5]        → Return Event{Rune: '日'}
```

**Scenario 2: Escape Sequence Split**
```
Call 1: Read() returns [0x1b]              → Block (incomplete)
Call 2: Read() returns [0x5b, 0x41]        → Return Event{Key: KeyUp}
```

**Scenario 3: Network Delay (SSH)**
```
Call 1: Read() returns [0x1b, 0x5b]        → Block (incomplete CSI)
Call 2: Timeout (VTIME expires, n=0)       → Block again
Call 3: Read() returns [0x41]              → Return Event{Key: KeyUp}
```

---

### 5. Backward Compatibility

**Contract**: ReadEvent() MUST maintain identical behavior for ASCII input

**Invariants**:
- ASCII characters (0x20-0x7E): Same Event as before
- Escape sequences: Same Key constants
- Control characters (Ctrl+A, etc.): Same Key constants
- Unknown sequences: Still return KeyUnknown

**Regression Test**:
```go
func TestBackwardCompatibilityASCII(t *testing.T) {
    testCases := []struct {
        input []byte
        want  Key
    }{
        {[]byte{'a'}, KeyA},
        {[]byte{0x1b, '[', 'A'}, KeyUp},
        {[]byte{0x03}, KeyCtrlC},
        {[]byte{' '}, KeySpace},
    }

    backend := NewUnixBackend()
    for _, tc := range testCases {
        event, _ := backend.ReadEvent()
        if event.Key != tc.want {
            t.Errorf("Key = %v, want %v", event.Key, tc.want)
        }
    }
}
```

---

## Performance Contracts

### Latency Guarantees

| Event Type | Max Latency | Validation Method |
|------------|-------------|-------------------|
| ASCII char | <1ms | Benchmark |
| Escape key (bare) | <1ms | Benchmark |
| Escape sequence | <100ms (VTIME) | Integration test |
| UTF-8 char (complete) | <1ms | Benchmark |
| UTF-8 char (split) | <100ms (VTIME) | Integration test |

### Allocation Guarantees

| Operation | Max Allocations | Max Bytes/Op |
|-----------|-----------------|--------------|
| ASCII event | 0 | 0 B |
| UTF-8 event | 0 | 0 B |
| Escape sequence | 0 | 0 B |
| 1000 events | 0 | 0 B |

**Exception**: First call may allocate for pool warmup. Subsequent calls MUST be zero-allocation.

---

## Error Contracts

### Error Conditions

| Condition | Return Value | Behavior |
|-----------|--------------|----------|
| `io.EOF` | `(Event{}, io.EOF)` | Terminal closed |
| Read error | `(Event{}, err)` | Propagate error |
| Invalid UTF-8 (after 4 bytes) | `(Event{Rune: RuneError, Key: KeyUnknown}, nil)` | Graceful degradation |
| Timeout (VTIME) | `(Event{}, nil)` or block again | Depends on partial buffer state |

### No Panics

**Contract**: ReadEvent() MUST NEVER panic

**Prohibited Operations**:
- Slice out-of-bounds access
- Nil pointer dereference
- Unbuffered channel operations

**Validation**: Fuzz testing with invalid input sequences

---

## Thread Safety

**Contract**: ReadEvent() is NOT thread-safe (per constitution)

**Guarantees**:
- Safe to call from single goroutine (input capture loop)
- MUST NOT be called concurrently from multiple goroutines
- No internal locking (performance optimization)

**Caller Responsibility**:
- Ensure serial calls from capture goroutine
- Do NOT share Backend instance across goroutines without external locking

---

## State Invariants

### Pre-conditions
- Backend MUST be initialized (`Init()` called)
- Terminal MUST be in raw mode
- `reader` MUST be readable (stdin or test mock)

### Post-conditions
- Event MUST have valid Key or Rune
- Event.Timestamp MUST be set (monotonic clock)
- Event.Pressed MUST be true (key-up not supported in this feature)
- Partial bytes MUST be retained in `pendingBuf` for next call

---

## Testing Requirements

### Unit Tests
- `TestReadEventUTF8TwoByte`
- `TestReadEventUTF8ThreeByte`
- `TestReadEventUTF8FourByte`
- `TestReadEventSplitSequence`
- `TestReadEventBackwardCompatibility`

### Benchmarks
- `BenchmarkReadEventASCII`
- `BenchmarkReadEventUTF8`
- `BenchmarkReadEventEscapeSequence`
- `BenchmarkReadEventAllocations` (with `-benchmem`)

### Integration Tests
- `TestReadEventRealTerminal` (skipped if not TTY)
- `TestReadEventVTIMETimeout`
- `TestReadEventSSHLatency` (simulated slow terminal)

---

## Migration Path

### Before (Current)
```go
func (b *unixBackend) ReadEvent() (Event, error) {
    buf := make([]byte, 256)  // ❌ Allocates every call
    n, _ := b.reader.Read(buf[:1])

    if buf[0] == 0x1b {
        time.Sleep(5 * time.Millisecond)  // ❌ Artificial delay
        n2, _ := b.reader.Read(buf[1:])
        n += n2
    }

    return b.parser.Parse(buf[:n])  // ❌ No UTF-8 handling
}
```

### After (Target)
```go
func (b *unixBackend) ReadEvent() (Event, error) {
    bufPtr := readBufferPool.Get().(*[]byte)  // ✅ Pooled
    defer readBufferPool.Put(bufPtr)

    buf := *bufPtr
    n, err := b.reader.Read(buf)

    b.pendingBuf = append(b.pendingBuf, buf[:n]...)  // ✅ Accumulate

    // Handle escape sequences with VTIME (no sleep)
    if b.pendingBuf[0] == 0x1b {
        for len(b.pendingBuf) < 16 {
            n, err := b.reader.Read(buf)
            if n == 0 {
                break  // VTIME timeout
            }
            b.pendingBuf = append(b.pendingBuf, buf[:n]...)
        }
    }

    // Check for complete UTF-8 sequence
    if !utf8.FullRune(b.pendingBuf) {
        // Wait for more bytes
        return b.ReadEvent()  // Recurse
    }

    event, err := b.parser.Parse(b.pendingBuf)
    b.pendingBuf = b.pendingBuf[len(b.pendingBuf):]  // Clear

    return event, err
}
```

---

**Contract Status**: DRAFT - Ready for implementation
