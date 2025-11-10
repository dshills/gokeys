# Advanced Demo - Comprehensive gokeys Functionality Test

This advanced demo showcases the full capabilities of the gokeys library through an interactive menu-driven interface.

## Features

The demo includes 6 different modes, each testing specific aspects of the library:

### 1. Event Inspector
**Tests:** Detailed event analysis
- Displays all event fields (Key, Rune, Modifiers, Timestamp, Pressed, Repeat)
- Shows event latency (processing delay from capture to display)
- Demonstrates blocking `Poll()` API
- Tracks event count and timing information

**What to try:**
- Press different keys to see their normalized Key codes
- Try special keys (arrows, function keys, escape sequences)
- Hold keys to see repeat events
- Type characters to see Rune values and Unicode codes

### 2. State Tracker
**Tests:** Real-time key state monitoring
- Tracks multiple key states simultaneously using `IsPressed()`
- Demonstrates non-blocking `Next()` API
- Shows live state updates at 10 FPS
- Tests concurrent key presses

**What to try:**
- Hold multiple keys simultaneously (e.g., WASD)
- Press and release keys rapidly
- Hold keys for extended periods
- Try combinations like arrow keys + space

**Tracked keys:** A, S, D, W, Arrow keys, Space, Enter, Ctrl+C

### 3. Game Input Demo
**Tests:** Action mapping and game loop patterns
- Demonstrates `GameInput` interface
- Tests action binding with multiple keys per action (P2 feature)
- Shows runtime rebinding (P3 feature)
- Implements 60 FPS game loop
- Visual feedback with character movement

**Controls:**
- **Movement:** WASD or Arrow Keys (demonstrates multi-key bindings)
- **Jump:** SPACE or J (can be rebound)
- **Fire:** F or ENTER
- **Special:** E or Z
- **Pause:** P (demonstrates pause handling)
- **Rebind:** R (rebinds Jump action to X key only)
- **Quit:** ESC or Q

**What to try:**
- Move the @ character around the screen
- Press action keys to see visual feedback
- Press R to rebind the jump action, then try different keys
- Test that SPACE/J stop working after rebind, only X works
- Hold multiple movement keys simultaneously

### 4. Performance Monitor
**Tests:** Latency and throughput measurement
- Measures event processing latency (time from event capture to processing)
- Calculates events per second (throughput)
- Tracks min/max/average latency
- Updates metrics in real-time

**Metrics displayed:**
- Total event count
- Events per second (throughput)
- Minimum latency (best case)
- Maximum latency (worst case)
- Average latency (overall performance)
- System uptime

**What to try:**
- Type rapidly to increase events/sec
- Hold keys to see autorepeat throughput
- Watch latency metrics (should be sub-millisecond)
- Compare performance with different typing patterns

### 5. UTF-8 Test
**Tests:** Multi-byte character support
- Demonstrates UTF-8 character decoding
- Shows character buffer editing (backspace support)
- Displays Unicode code points
- Measures byte size of characters

**What to try:**
- Type ASCII characters: a, b, c
- Type extended ASCII: é, ñ, ü (if your keyboard supports it)
- Copy/paste emoji: 😀, 🎮, 🚀 (may not work in all terminals)
- Copy/paste CJK characters: 中, 日, 한
- Use backspace to delete characters
- Press Enter to add newlines

**Displays:**
- Input buffer (editable text)
- Last character with Unicode code point (U+XXXX)
- Character byte size (1-4 bytes for UTF-8)
- Buffer length in characters vs bytes

### 6. Modifier Test
**Tests:** Modifier key combinations
- Demonstrates Shift, Alt, Ctrl detection
- Shows modifier combinations (Ctrl+Shift, etc.)
- Displays history of recent key combinations
- Tests modifier bitflag system

**What to try:**
- Press Shift + letter keys
- Press Ctrl + letter keys
- Press Alt + letter keys (may be captured by terminal/OS)
- Try combinations: Ctrl+Shift+A
- Try Ctrl+C (shows in history but doesn't terminate)
- Press special keys with modifiers

## Building and Running

```bash
# Build
cd examples/advanced
go build

# Run
./advanced
```

Or run directly with:
```bash
go run examples/advanced/main.go
```

## Architecture

The demo uses a mode-based architecture:

```go
type DemoMode int

const (
    ModeMenu           // Main menu
    ModeEventInspector // Event details
    ModeStateTracker   // IsPressed() tracking
    ModeGameInput      // Action mapping
    ModePerformance    // Latency/throughput
    ModeUTF8Test       // UTF-8 characters
    ModeModifierTest   // Modifier combos
)
```

Each mode:
1. Starts the input system
2. Runs its specific test/demo loop
3. Stops the input system
4. Returns to the menu

## Statistics Tracking

The demo includes comprehensive statistics tracking:

```go
type Stats struct {
    eventCount    int           // Total events processed
    startTime     time.Time     // Demo start time
    lastEventTime time.Time     // Most recent event
    minLatency    time.Duration // Best latency
    maxLatency    time.Duration // Worst latency
    totalLatency  time.Duration // Sum for average
}
```

Statistics are reset when switching modes.

## API Coverage

This demo exercises all major gokeys APIs:

### Input Interface
- ✅ `Start()` - Initialize input system
- ✅ `Stop()` - Cleanup and restore terminal
- ✅ `Poll()` - Blocking event retrieval
- ✅ `Next()` - Non-blocking event retrieval
- ✅ `IsPressed(Key)` - Key state queries

### Event Fields
- ✅ `Key` - Normalized key code
- ✅ `Rune` - Unicode character
- ✅ `Modifiers` - Shift/Alt/Ctrl flags
- ✅ `Timestamp` - Event capture time
- ✅ `Pressed` - Key press/release state
- ✅ `Repeat` - OS autorepeat indicator

### GameInput Interface
- ✅ `Start()` - Initialize game input
- ✅ `Stop()` - Cleanup
- ✅ `Bind(action, keys...)` - Bind actions to keys
- ✅ `IsActionPressed(action)` - Query action state
- ✅ Multiple keys per action (P2)
- ✅ Runtime rebinding (P3)

### Key Types
- ✅ Letter keys (A-Z)
- ✅ Number keys (0-9)
- ✅ Arrow keys (Up, Down, Left, Right)
- ✅ Special keys (Escape, Enter, Space, Backspace)
- ✅ Function keys (F1-F12)
- ✅ Control combinations (Ctrl+A through Ctrl+Z)

### Modifiers
- ✅ `ModShift` - Shift key
- ✅ `ModAlt` - Alt key
- ✅ `ModCtrl` - Ctrl key
- ✅ Modifier combinations (bitflags)

## Performance Expectations

Based on gokeys performance characteristics:

- **Latency:** Sub-millisecond (<1ms) for all events
- **Throughput:** Hundreds of events per second
- **Allocations:** Zero allocations per event (buffer pooling)
- **UTF-8 Parsing:** ~30ns/op for multi-byte characters

The Performance Monitor mode lets you verify these characteristics on your system.

## Terminal Compatibility

This demo works with any terminal that gokeys supports:
- Unix/Linux terminals (xterm, gnome-terminal, etc.)
- macOS Terminal.app, iTerm2
- Windows Command Prompt, PowerShell, Windows Terminal
- SSH clients
- tmux, screen (terminal multiplexers)

Some features may vary by terminal:
- Alt key combinations (often captured by terminal/OS)
- Emoji support (terminal font dependent)
- ANSI escape sequence support (clear screen, cursor positioning)

## Exit

Press **Q** in the main menu to exit the demo.

Each mode returns to the menu with **ESC** (except Game Input mode, which uses ESC or Q to quit).

## Troubleshooting

**Problem:** Keys not responding
- **Solution:** Make sure the terminal window has focus

**Problem:** Display looks corrupted
- **Solution:** Some terminals may not support ANSI escape sequences properly. Try a different terminal.

**Problem:** Alt combinations don't work
- **Solution:** Many terminals capture Alt+key for menus. This is expected behavior.

**Problem:** Emoji don't display
- **Solution:** Your terminal font may not support emoji. Use a modern font like "Noto Color Emoji" or "Apple Color Emoji".

**Problem:** Performance metrics show high latency
- **Solution:** Latency measures processing delay, not input lag. Some overhead is expected for metrics collection.

## Code Structure

- `AdvancedDemo` - Main demo coordinator
- `Stats` - Performance statistics tracker
- `DemoMode` - Mode enumeration
- 6 mode-specific methods (`runEventInspector`, `runStateTracker`, etc.)
- Helper functions (`clearScreen`, `printHeader`, `formatModifiers`)

Total: ~500 lines of comprehensive testing code.
