# Advanced Demo - Quick Start Guide

## Running the Demo

```bash
# Option 1: Run directly
go run examples/advanced/main.go

# Option 2: Build and run
cd examples/advanced
go build
./advanced
```

## Navigation

The demo has a main menu with 6 modes. Use number keys 1-6 to select a mode, Q to quit.

## Quick Tour (5 minutes)

Follow this sequence for a comprehensive test:

### 1. Start with Event Inspector (Press 1)
- Type a few letters: `hello`
- Try arrow keys: ← → ↑ ↓
- Press function keys: F1, F2, F3
- Try special keys: Space, Enter, Backspace
- Hold a key to see Repeat=true
- **Press ESC to return to menu**

### 2. State Tracker (Press 2)
- Hold W, A, S, D keys (one at a time)
- Hold multiple keys simultaneously (W+D)
- Try arrow keys and Space
- Watch the live state display update
- **Press ESC to return to menu**

### 3. Game Input (Press 3)
- Move the @ character with WASD or arrow keys
- Press SPACE or J to see "JUMP" message
- Press F or ENTER to see "FIRE" message
- Press E or Z to see "SPECIAL" message
- **Press R** to rebind jump action
- Try SPACE/J (won't work anymore)
- Try X key (now triggers jump)
- Press P to pause
- **Press ESC or Q to return to menu**

### 4. Performance Monitor (Press 4)
- Type rapidly: `asdfasdfasdf`
- Hold a key for autorepeat
- Watch metrics update:
  - Events/sec should be 50-200+ when typing
  - Latency should be <1ms
- **Press ESC to return to menu**

### 5. UTF-8 Test (Press 5)
- Type regular text: `Hello World`
- Try backspace to delete characters
- Press Enter to add newlines
- If possible, copy/paste special characters:
  - Extended: `café`, `niño`, `Zürich`
  - Emoji: `😀 🎮 🚀 ⚡ 💻`
  - CJK: `你好 こんにちは 안녕하세요`
- Watch the Unicode code points (U+XXXX)
- Notice byte size differences (1-4 bytes)
- **Press ESC to return to menu**

### 6. Modifier Test (Press 6)
- Press Shift+A (should show "Shift + A")
- Press Ctrl+C (shows in history, doesn't quit)
- Press Ctrl+A, Ctrl+Z
- Try Alt+A (may be captured by terminal)
- Try Shift+Ctrl+A (multiple modifiers)
- Watch the history list grow
- **Press ESC to return to menu**

### 7. Exit (Press Q from menu)

## What Each Mode Tests

| Mode | Tests | Key APIs |
|------|-------|----------|
| **Event Inspector** | Full event details | `Poll()`, Event fields |
| **State Tracker** | Multi-key states | `Next()`, `IsPressed()` |
| **Game Input** | Action mapping | `GameInput`, `Bind()`, `IsActionPressed()` |
| **Performance** | Latency/throughput | Timestamp, metrics |
| **UTF-8** | Multi-byte chars | Rune field, Unicode |
| **Modifier** | Key combinations | Modifiers bitflags |

## Expected Results

### Event Inspector
- All key presses should show:
  - Key: Normalized key name
  - Rune: Character or empty
  - Modifiers: Shift/Alt/Ctrl if pressed
  - Pressed: true
  - Repeat: false first time, true when held
  - Latency: <1ms

### State Tracker
- Keys show "PRESSED" while held
- Multiple keys can be pressed simultaneously
- State updates in real-time (10 FPS)

### Game Input
- @ character moves smoothly in all directions
- Both WASD and arrows work for movement
- Multiple action bindings work (Space and J both jump)
- After rebind (R key), only X triggers jump
- Action messages appear instantly when keys pressed

### Performance Monitor
- Events/sec: 50-200+ when typing rapidly
- Min latency: <1ms (often <0.1ms)
- Max latency: <2ms (occasional spikes OK)
- Avg latency: <1ms

### UTF-8 Test
- ASCII chars: 1 byte each
- Extended ASCII (é, ñ): 2 bytes
- Most emoji: 4 bytes
- Buffer edits work correctly with multi-byte chars

### Modifier Test
- Shift combinations always work
- Ctrl combinations always work
- Alt combinations may not work (OS/terminal captures them)
- Multiple modifier combinations work (Ctrl+Shift+key)

## Common Issues

**Q: Alt+key doesn't work**
A: Most terminals/OS capture Alt for menus. This is normal.

**Q: Display is corrupted**
A: Your terminal may not support ANSI codes. Try a modern terminal.

**Q: Can't type emoji in UTF-8 test**
A: Copy/paste them from elsewhere. Most terminals don't have emoji input methods.

**Q: Keys seem delayed**
A: Check Performance Monitor. Latency should be <1ms. Higher values indicate system load.

**Q: Game Input mode flickers**
A: Expected - it renders at 10 FPS to balance smoothness and CPU usage.

## Tips

- **Terminal focus**: Ensure terminal window has focus
- **Full screen**: Maximize terminal for best experience
- **Font**: Use a monospace font for proper alignment
- **Colors**: Demo uses basic ANSI - works on all terminals
- **Escape key**: Always returns to menu (except in Game mode, also accepts Q)

## Development Notes

This demo was designed to:
1. Test every major gokeys API
2. Demonstrate real-world usage patterns
3. Verify performance characteristics
4. Showcase cross-platform compatibility
5. Provide example code for developers

Total code: ~500 lines covering:
- Both blocking and non-blocking APIs
- Event inspection and state tracking
- Game loop patterns at 60 FPS
- Performance measurement
- UTF-8 multi-byte handling
- Modifier key combinations
- Runtime action rebinding

## Questions?

See the main README.md for detailed explanations of each mode and the APIs they test.
