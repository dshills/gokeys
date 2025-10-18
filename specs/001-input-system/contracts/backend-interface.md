# Contract: Backend Interface (Internal)

**Package**: `internal/backend`
**Type**: Interface (Internal - Not Exported)
**Purpose**: Platform-specific terminal I/O abstraction

## Interface Definition

```go
package backend

import "github.com/dshills/gokeys/input"

// Backend defines the internal contract for platform-specific terminal I/O.
// This interface is internal and not exported from the input package.
// It abstracts the differences between Unix termios and Windows Console API.
type Backend interface {
    // Init enters raw mode and saves the current terminal state.
    // Must be idempotent - calling multiple times is a no-op.
    //
    // Returns error if:
    //   - Platform initialization fails
    //   - Insufficient permissions
    //   - Terminal is unsupported
    Init() error

    // Restore exits raw mode and restores the terminal to its original state.
    // Must be idempotent and safe to call even if Init() failed.
    // Guaranteed to be called during cleanup.
    Restore() error

    // ReadEvent blocks until a keyboard event is available, then parses
    // and returns a normalized Event.
    //
    // Returns error if:
    //   - Read operation fails (terminal disconnected, etc.)
    //   - Should not error on unparsable sequences (return KeyUnknown instead)
    //
    // Thread-safety: Only called from single capture goroutine.
    ReadEvent() (input.Event, error)
}
```

## Method Contracts

### Init()

**Platform-Specific Behavior**:

**Unix/Linux/macOS**:
```go
1. Get current terminal attributes via tcgetattr
2. Save to terminalState struct
3. Create new termios with raw mode flags:
   - c_lflag &= ~(ICANON | ECHO | ISIG)
   - c_cc[VMIN] = 1
   - c_cc[VTIME] = 0
4. Apply via tcsetattr(TCSANOW)
```

**Windows**:
```go
1. Get stdin handle via GetStdHandle(STD_INPUT_HANDLE)
2. Get current console mode via GetConsoleMode
3. Save to consoleMode field
4. Set new mode:
   - Enable ENABLE_VIRTUAL_TERMINAL_INPUT
   - Disable ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT
   - Enable ENABLE_WINDOW_INPUT
5. Apply via SetConsoleMode
```

**Error Handling**:
```go
// Unix
if syscall.Tcgetattr(fd, &termios) != nil {
    return ErrTerminalInit
}

// Windows
if GetConsoleMode(handle, &mode) == 0 {
    return ErrConsoleInit
}
```

**Idempotency**:
```go
type UnixBackend struct {
    initialized bool
    savedState  syscall.Termios
}

func (b *UnixBackend) Init() error {
    if b.initialized {
        return nil // Already initialized
    }
    // ... perform initialization
    b.initialized = true
    return nil
}
```

### Restore()

**Platform-Specific Behavior**:

**Unix**:
```go
1. Apply saved termios via tcsetattr(TCSANOW)
2. Flush input/output buffers
3. Handle error gracefully (log but don't fail)
```

**Windows**:
```go
1. Restore saved console mode via SetConsoleMode
2. Flush console input buffer
3. Handle error gracefully
```

**Cleanup Guarantees**:
- Always attempt restoration even if Init() partially failed
- Never panic
- Log errors but return nil (best-effort)

**Example Implementation**:
```go
func (b *UnixBackend) Restore() error {
    if !b.initialized {
        return nil // Nothing to restore
    }

    fd := int(os.Stdin.Fd())
    if err := syscall.Tcsetattr(fd, syscall.TCSANOW, &b.savedState); err != nil {
        // Log but don't fail
        log.Printf("Warning: Failed to restore terminal: %v", err)
    }

    b.initialized = false
    return nil
}
```

### ReadEvent()

**Blocking Behavior**:
- Blocks until at least one byte available
- May perform multiple reads to complete escape sequence
- Uses timeout for multi-byte sequences (50ms)

**Event Construction**:
```go
1. Read byte(s) from terminal
2. Parse escape sequence or single key
3. Normalize to Key constant
4. Detect modifiers from sequence
5. Set Timestamp to time.Now()
6. Detect Repeat via timing heuristic (optional)
7. Set Pressed field (platform-dependent)
8. Populate Rune for printable keys
9. Return Event
```

**Platform-Specific Parsing**:

**Unix** (escape sequences):
```go
// Example: Arrow Up = \x1b[A
buf := make([]byte, 32)
n, err := os.Stdin.Read(buf)
if err != nil {
    return Event{}, err
}

if buf[0] == 0x1b {
    // Escape sequence
    seq := readSequence(buf[:n]) // May read more bytes with timeout
    key, mods := parseSequence(seq)
    return Event{
        Key:       key,
        Modifiers: mods,
        Timestamp: time.Now(),
        Pressed:   true,
        Repeat:    detectRepeat(key), // Timing-based
        Rune:      0,
    }, nil
}

// Single byte key
return parseSingleByte(buf[0])
```

**Windows** (Console API):
```go
// Example: Using INPUT_RECORD
var inputRecord INPUT_RECORD
var numRead uint32

if !ReadConsoleInput(handle, &inputRecord, 1, &numRead) {
    return Event{}, ErrConsoleRead
}

if inputRecord.EventType == KEY_EVENT {
    keyEvent := inputRecord.KeyEvent
    key := normalizeVirtualKeyCode(keyEvent.wVirtualKeyCode)

    return Event{
        Key:       key,
        Modifiers: parseModifiers(keyEvent.dwControlKeyState),
        Timestamp: time.Now(),
        Pressed:   keyEvent.bKeyDown != 0,
        Repeat:    keyEvent.wRepeatCount > 1,
        Rune:      rune(keyEvent.UnicodeChar),
    }, nil
}
```

**Sequence Timeout Strategy**:
```go
// Unix: Distinguish Escape alone vs Escape prefix
func readWithTimeout(fd int, timeout time.Duration) ([]byte, error) {
    syscall.SetNonblock(fd, true)
    defer syscall.SetNonblock(fd, false)

    buf := make([]byte, 32)
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        n, err := syscall.Read(fd, buf)
        if n > 0 {
            return buf[:n], nil
        }
        if err == syscall.EAGAIN {
            time.Sleep(1 * time.Millisecond)
            continue
        }
        return nil, err
    }
    return nil, ErrTimeout
}
```

**Normalization Examples**:

```go
// Unix escape sequences
"\x1b[A"     → KeyUp
"\x1b[B"     → KeyDown
"\x1b[1;2A"  → KeyUp with ModShift
"\x1b[1;5A"  → KeyUp with ModCtrl
"\x1bx"      → KeyX with ModAlt
"\x01"       → KeyCtrlA (Ctrl+A)

// Windows virtual key codes
VK_UP (0x26)        → KeyUp
VK_SPACE (0x20)     → KeySpace
VK_RETURN (0x0D)    → KeyEnter
'A' + Shift         → KeyA with ModShift, Rune='A'
'a' + Ctrl          → KeyCtrlA with ModCtrl
```

**Unknown Sequence Handling**:
```go
// Always return Event, never error on parse failure
if key, ok := sequenceMap[seq]; ok {
    return Event{Key: key, ...}, nil
}

// Unknown sequence
return Event{
    Key:       input.KeyUnknown,
    Timestamp: time.Now(),
    Pressed:   true,
}, nil // No error!
```

## Implementation Examples

### UnixBackend Structure

```go
package backend

import (
    "os"
    "syscall"
    "time"
    "github.com/dshills/gokeys/input"
)

type UnixBackend struct {
    initialized bool
    savedState  syscall.Termios
    fd          int
    parser      *SequenceParser
}

func NewUnixBackend() *UnixBackend {
    return &UnixBackend{
        fd:     int(os.Stdin.Fd()),
        parser: NewSequenceParser(),
    }
}

func (b *UnixBackend) Init() error {
    if b.initialized {
        return nil
    }

    // Get current state
    if err := syscall.Tcgetattr(b.fd, &b.savedState); err != nil {
        return err
    }

    // Create raw mode termios
    raw := b.savedState
    raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ISIG
    raw.Cc[syscall.VMIN] = 1
    raw.Cc[syscall.VTIME] = 0

    // Apply
    if err := syscall.Tcsetattr(b.fd, syscall.TCSANOW, &raw); err != nil {
        return err
    }

    b.initialized = true
    return nil
}

func (b *UnixBackend) Restore() error {
    if !b.initialized {
        return nil
    }
    syscall.Tcsetattr(b.fd, syscall.TCSANOW, &b.savedState)
    b.initialized = false
    return nil
}

func (b *UnixBackend) ReadEvent() (input.Event, error) {
    buf := make([]byte, 32)
    n, err := syscall.Read(b.fd, buf)
    if err != nil {
        return input.Event{}, err
    }

    return b.parser.Parse(buf[:n])
}
```

### WindowsBackend Structure

```go
package backend

import (
    "syscall"
    "unsafe"
    "github.com/dshills/gokeys/input"
)

type WindowsBackend struct {
    initialized bool
    handle      syscall.Handle
    savedMode   uint32
}

func NewWindowsBackend() *WindowsBackend {
    return &WindowsBackend{}
}

func (b *WindowsBackend) Init() error {
    if b.initialized {
        return nil
    }

    // Get stdin handle
    b.handle, _ = syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)

    // Get and save current mode
    var mode uint32
    syscall.GetConsoleMode(b.handle, &b.savedMode)

    // Set raw mode
    mode = b.savedMode
    mode &^= ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT
    mode |= ENABLE_VIRTUAL_TERMINAL_INPUT | ENABLE_WINDOW_INPUT

    if err := syscall.SetConsoleMode(b.handle, mode); err != nil {
        return err
    }

    b.initialized = true
    return nil
}

func (b *WindowsBackend) Restore() error {
    if !b.initialized {
        return nil
    }
    syscall.SetConsoleMode(b.handle, b.savedMode)
    b.initialized = false
    return nil
}

func (b *WindowsBackend) ReadEvent() (input.Event, error) {
    var record INPUT_RECORD
    var numRead uint32

    for {
        if !ReadConsoleInput(b.handle, &record, 1, &numRead) {
            return input.Event{}, syscall.GetLastError()
        }

        if record.EventType == KEY_EVENT && record.KeyEvent.bKeyDown != 0 {
            return b.parseKeyEvent(record.KeyEvent)
        }
    }
}
```

## Testing Strategy

### Contract Tests (Cross-Backend)

```go
// tests/contract/backend_test.go
package contract_test

func TestBackendNormalization(t *testing.T) {
    backends := []struct {
        name    string
        backend backend.Backend
    }{
        {"Unix", backend.NewUnixBackend()},
        {"Windows", backend.NewWindowsBackend()},
    }

    for _, tt := range backends {
        t.Run(tt.name, func(t *testing.T) {
            // Inject same escape sequence
            // Verify same Key produced
        })
    }
}
```

### Platform-Specific Tests

```go
//go:build !windows
// +build !windows

package backend_test

func TestUnixTermios(t *testing.T) {
    backend := backend.NewUnixBackend()
    if err := backend.Init(); err != nil {
        t.Fatal(err)
    }
    defer backend.Restore()

    // Verify raw mode flags set correctly
}
```

```go
//go:build windows
// +build windows

package backend_test

func TestWindowsConsoleMode(t *testing.T) {
    backend := backend.NewWindowsBackend()
    if err := backend.Init(); err != nil {
        t.Fatal(err)
    }
    defer backend.Restore()

    // Verify console mode flags set correctly
}
```

## Performance Requirements

| Operation | Latency Target | Notes |
|-----------|----------------|-------|
| Init() | <100ms | One-time setup cost |
| Restore() | <50ms | Cleanup on exit |
| ReadEvent() | <10ms | From physical key to Event |

**Parsing Performance**:
- Single-byte keys: <1ms
- Escape sequences: <5ms (including timeout)
- Trie lookup: O(sequence length), typically <10 bytes

## Version Compatibility

**Interface Stability**: Internal interface, may change between minor versions

**Platform Support**:
- Unix: Linux, macOS, BSD (any POSIX-compliant system)
- Windows: Windows 10+ (Console API with VT support)

**Build Tags**:
```go
// unix.go
//go:build !windows
// +build !windows

// windows.go
//go:build windows
// +build windows
```
