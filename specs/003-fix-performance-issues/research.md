# Research: Performance and Efficiency Improvements

**Feature**: 003-fix-performance-issues
**Date**: 2025-10-18
**Status**: Complete

## Overview

This document consolidates research findings for optimizing the gokeys input system to address three performance issues: escape key latency, memory allocations, and UTF-8 support.

## Key Technical Decisions

### Decision 1: UTF-8 Decoding Strategy

**Decision**: Use manual buffer accumulation with `unicode/utf8.FullRune` instead of `bufio.Reader.ReadRune()`

**Rationale**:
- gokeys parser requires byte-level access to detect escape sequences (`0x1b`)
- `bufio.Reader` adds unnecessary buffering layer on top of termios buffering
- Manual approach provides precise control over partial sequence handling
- Performance: avoids double buffering overhead

**Implementation Pattern**:
```go
// Accumulate partial UTF-8 sequences
if utf8.FullRune(buf) {
    r, size := utf8.DecodeRune(buf)
    // Process complete rune
} else {
    // Wait for more bytes
}
```

**Alternatives Considered**:
- `bufio.Reader.ReadRune()`: Rejected due to abstraction overhead and loss of byte-level control
- `io.RuneReader` interface: Rejected for same reasons

**References**:
- Go unicode/utf8 package: https://pkg.go.dev/unicode/utf8
- Go Issue #45898 (UTF-8 incomplete detection): https://github.com/golang/go/issues/45898

---

### Decision 2: Buffer Reuse with sync.Pool

**Decision**: Use `sync.Pool` for 256-byte read buffers to eliminate per-keypress allocations

**Rationale**:
- Current implementation allocates 256 bytes per `ReadEvent()` call
- At 60 FPS input (gaming scenario), this creates 15KB/sec garbage
- `sync.Pool` provides zero-allocation buffer reuse after warmup
- Cloudflare reports 4x faster response time and 50-90% allocation reduction

**Implementation Pattern**:
```go
var readBufferPool = sync.Pool{
    New: func() interface{} {
        b := make([]byte, 256)
        return &b
    },
}

func (b *unixBackend) ReadEvent() (Event, error) {
    bufPtr := readBufferPool.Get().(*[]byte)
    defer readBufferPool.Put(bufPtr)

    buf := *bufPtr
    n, _ := b.reader.Read(buf)

    // CRITICAL: Copy data out before returning buffer to pool
    b.pendingBuf = append(b.pendingBuf, buf[:n]...)

    return b.parser.Parse(b.pendingBuf)
}
```

**Safety Considerations**:
- NEVER return slices that reference pooled memory
- ALWAYS copy data to persistent buffer before pool return
- Reset pooled buffers before use (not needed for []byte, needed for bytes.Buffer)

**Alternatives Considered**:
- Pre-allocated struct field buffer: Simpler but prevents concurrent backend instances
- Object pooling library: Adds external dependency (violates constitution)

**References**:
- Cloudflare buffer pooling guide: https://blog.cloudflare.com/recycling-memory-buffers-in-go/
- Go sync.Pool documentation: https://pkg.go.dev/sync#Pool

---

### Decision 3: Escape Sequence Timeout Without Artificial Delay

**Decision**: Remove `time.Sleep(5 * time.Millisecond)` and rely on termios VTIME configuration

**Rationale**:
- Current 5ms sleep adds latency to EVERY escape keypress
- Existing VMIN=0, VTIME=1 settings provide 100ms inter-character timeout
- VTIME is hardware-level timeout (no polling overhead)
- This is the standard approach used by tcell, termbox-go, and golang.org/x/term

**Current Problem**:
```go
if buf[0] == 0x1b {
    time.Sleep(5 * time.Millisecond)  // ❌ Artificial delay
    n2, _ := b.reader.Read(buf[1:])
}
```

**Improved Approach**:
```go
// Read until VTIME timeout (100ms inter-char)
for b.pendingBuf[0] == 0x1b && len(b.pendingBuf) < 16 {
    n, err := b.reader.Read(buf)
    if err != nil || n == 0 {
        break  // VTIME timeout - no more data
    }
    b.pendingBuf = append(b.pendingBuf, buf[:n]...)
}
```

**How VTIME Works**:
- After first byte read, timer starts
- If no byte arrives within VTIME*100ms, Read() returns with 0 bytes
- This naturally distinguishes bare ESC from escape sequences

**Alternatives Considered**:
- Application-level timer (like tcell's 50ms timer): More complex, unnecessary given VTIME
- Shorter timeout (VTIME=0.5): Risks cutting off legitimate sequences on slow SSH

**References**:
- VMIN/VTIME guide: http://www.unixwiz.net/techtips/termios-vmin-vtime.html
- tcell implementation: https://github.com/gdamore/tcell/blob/main/tscreen.go

---

### Decision 4: Partial Sequence Accumulation

**Decision**: Add persistent `pendingBuf []byte` field to `unixBackend` for accumulating partial sequences

**Rationale**:
- UTF-8 characters can arrive across multiple Read() calls
- Escape sequences may be split by network latency (SSH, remote terminals)
- Need to accumulate bytes until a complete sequence is detected

**State Machine**:
```
1. Read chunk → append to pendingBuf
2. Check for complete sequence:
   - UTF-8: use utf8.FullRune()
   - Escape: check terminator (letter, ~, timeout)
3. Parse complete sequence
4. Remove processed bytes from pendingBuf
5. Repeat
```

**Edge Cases**:
- Partial UTF-8 at buffer end: Carried to next read
- Incomplete escape sequence: VTIME timeout triggers parse
- Invalid UTF-8 after 4 bytes: Treat first byte as complete

**Alternatives Considered**:
- Stateless parsing: Would drop partial sequences on buffer boundary
- Fixed-size ring buffer: More complex, no clear benefit over dynamic slice

---

### Decision 5: Performance Benchmarking Strategy

**Decision**: Add comprehensive benchmarks using `testing.B` with sub-benchmarks for different scenarios

**Benchmark Coverage**:
```go
// Latency measurement
BenchmarkEscapeKeyLatency     // Before/after sleep removal
BenchmarkReadEventLatency     // End-to-end event processing

// Allocation measurement
BenchmarkReadEventNoPool      // Current: 256 B/op, 1 allocs/op
BenchmarkReadEventWithPool    // Target:    0 B/op, 0 allocs/op

// UTF-8 performance
BenchmarkParseASCII           // Single byte
BenchmarkParseUTF8_2byte      // é (U+00E9)
BenchmarkParseUTF8_3byte      // € (U+20AC)
BenchmarkParseUTF8_4byte      // 𝄞 (U+1D11E)
```

**Run Commands**:
```bash
# Allocation benchmarks
go test -bench=. -benchmem ./input

# CPU profiling
go test -bench=BenchmarkReadEvent -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -bench=BenchmarkReadEvent -memprofile=mem.prof
go tool pprof mem.prof
```

**Success Criteria Validation**:
- SC-001: Escape latency <1ms → Benchmark before/after comparison
- SC-002: 50% allocation reduction → Compare B/op and allocs/op
- SC-003: GC pressure reduction → Memory profile analysis

**Alternatives Considered**:
- Manual timing with time.Now(): Less accurate, no integration with go test tooling
- External benchmarking tools: Adds complexity, go test -bench is sufficient

---

## Implementation Priorities

### Phase 1: Remove Latency (P1)
**Files**: `input/backend_unix.go`
- Remove `time.Sleep(5 * time.Millisecond)`
- Add loop for VTIME-based reads
- Add benchmarks for latency measurement

**Validation**: Run benchmarks, verify <1ms escape latency

---

### Phase 2: Add Buffer Pooling (P2)
**Files**: `input/backend_unix.go`
- Add `sync.Pool` for read buffers
- Add `pendingBuf []byte` field to `unixBackend`
- Implement safe copy pattern

**Validation**: Run allocation benchmarks, verify 0 allocs/op

---

### Phase 3: UTF-8 Support (P3)
**Files**: `input/parser.go`, `input/backend_unix.go`
- Import `unicode/utf8` package
- Add `utf8.FullRune` checks
- Update parser to handle multi-byte runes
- Add UTF-8 test cases (Japanese, emoji, etc.)

**Validation**: Test with actual UTF-8 input, verify 100% accuracy

---

## Risk Assessment

### Low Risk
- **Buffer pooling**: Well-established pattern, extensive test coverage in Go stdlib
- **VTIME usage**: Already configured, just leveraging it properly

### Medium Risk
- **UTF-8 partial sequences**: Edge case testing critical (network splits, terminal disconnects)
- **Backward compatibility**: Must not break existing ASCII-only applications

### Mitigation Strategies
- Comprehensive test suite with edge cases
- Benchmark suite to catch performance regressions
- Code review focused on buffer safety (no pooled slice returns)

---

## Open Questions

**Q: Should we support non-UTF-8 encodings (ISO-8859-1, Windows-1252)?**
**A**: No. Scope limited to UTF-8 (spec assumption: "Terminal is configured with UTF-8 encoding"). Other encodings would require iconv dependency (violates zero-dependency constraint).

**Q: What's the maximum pending buffer size before we should flush/error?**
**A**: 16 bytes is sufficient for longest escape sequence (F12 = ESC [ 2 4 ~). Add safety check to prevent memory leak on malformed input.

**Q: Should we add context.Context support for cancellation?**
**A**: Out of scope for this feature (explicitly listed in spec). Will be separate feature later.

---

## References

### Go Standard Library
- `unicode/utf8`: https://pkg.go.dev/unicode/utf8
- `sync.Pool`: https://pkg.go.dev/sync#Pool
- `testing` (benchmarks): https://pkg.go.dev/testing

### External Libraries (for reference only)
- tcell: https://github.com/gdamore/tcell
- termbox-go: https://github.com/nsf/termbox-go
- golang.org/x/term: https://pkg.go.dev/golang.org/x/term

### Best Practices
- Cloudflare buffer pooling: https://blog.cloudflare.com/recycling-memory-buffers-in-go/
- VMIN/VTIME: http://www.unixwiz.net/techtips/termios-vmin-vtime.html
- Go benchmarking: https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go

---

**Research Complete**: All technical unknowns resolved. Ready for Phase 1 design artifacts.
