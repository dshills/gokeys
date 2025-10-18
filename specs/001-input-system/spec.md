# Feature Specification: Cross-Terminal Input System

**Feature Branch**: `001-input-system`
**Created**: 2025-10-17
**Status**: Draft
**Input**: User description: "review spec/basic_spec.md"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - CLI Tool Developer Captures Keyboard Input (Priority: P1)

A developer building a command-line tool needs to capture keyboard input with normalized key codes across different terminals and operating systems. They need a simple way to handle both blocking and non-blocking input patterns without worrying about platform-specific escape sequences.

**Why this priority**: This is the foundational use case - basic keyboard event capture with cross-platform normalization. Without this, no other functionality is possible. It represents the minimum viable product.

**Independent Test**: Can be fully tested by initializing the input system, capturing arrow keys and standard keys (Enter, Escape, letters), and verifying that the same key produces identical normalized events across different terminals (iTerm2, Windows Terminal, xterm) and operating systems (macOS, Linux, Windows).

**Acceptance Scenarios**:

1. **Given** a CLI tool running on any terminal, **When** the user presses the Up arrow key, **Then** the system produces a normalized KeyUp event
2. **Given** the input system is initialized, **When** the user presses Ctrl+C, **Then** the system produces a normalized KeyCtrlC event with correct modifier flags
3. **Given** a blocking input pattern, **When** Poll() is called, **Then** execution blocks until a key is pressed and returns the event
4. **Given** a non-blocking pattern, **When** Next() is called with no key pressed, **Then** it returns immediately with nil
5. **Given** the input system is active, **When** Stop() is called, **Then** the terminal is restored to its original state
6. **Given** different terminal emulators on the same platform, **When** the same key is pressed, **Then** identical normalized events are produced

---

### User Story 2 - Game Developer Implements Real-Time Input (Priority: P2)

A game developer building a terminal-based game needs to query current key states in real-time (e.g., "is spacebar currently held down?") and distinguish between initial key presses and OS autorepeat events to implement proper game controls.

**Why this priority**: Extends P1 by adding state tracking and repeat detection, which are essential for interactive applications like games but build on the foundational event system.

**Independent Test**: Can be tested independently by capturing key press/release sequences, querying IsPressed() state, and validating that repeat events are correctly flagged. Delivers value by enabling real-time game input without requiring the action mapping system.

**Acceptance Scenarios**:

1. **Given** a key is pressed and held, **When** IsPressed() is called for that key, **Then** it returns true
2. **Given** a key is released, **When** IsPressed() is called for that key, **Then** it returns false
3. **Given** a key is held down triggering OS autorepeat, **When** events are received, **Then** subsequent events have Repeat flag set to true
4. **Given** a key is initially pressed, **When** the first event is received, **Then** the Repeat flag is false and Pressed is true
5. **Given** a platform supporting key-up events, **When** a key is released, **Then** an event with Pressed=false is generated
6. **Given** rapid game loop polling, **When** multiple Next() calls occur between key events, **Then** nil is returned without blocking

---

### User Story 3 - Game Developer Uses Action Mapping (Priority: P3)

A game developer wants to map logical actions (like "jump", "fire", "move-left") to physical keys, allowing players to rebind controls. They need to query actions by name rather than checking individual keys.

**Why this priority**: This is a higher-level convenience API built on P1 and P2. While valuable for game development, it's not essential for the core library functionality and can be implemented as an optional layer.

**Independent Test**: Can be tested by binding actions to keys, triggering those keys, and verifying IsActionPressed() returns correct values. Delivers value by simplifying game input logic even if implemented standalone.

**Acceptance Scenarios**:

1. **Given** an action "jump" is bound to the Spacebar key, **When** Spacebar is pressed, **Then** IsActionPressed("jump") returns true
2. **Given** an action is bound to multiple keys (e.g., "fire" → Spacebar and Enter), **When** either key is pressed, **Then** IsActionPressed("fire") returns true
3. **Given** actions are bound to specific keys, **When** Bind() is called to update bindings, **Then** new bindings take effect immediately
4. **Given** an action has no keys bound, **When** IsActionPressed() is called, **Then** it returns false
5. **Given** multiple actions bound to different keys, **When** queried simultaneously, **Then** each action state is tracked independently

---

### Edge Cases

- What happens when an unrecognized terminal escape sequence is received? (System should produce KeyUnknown rather than panic or drop the event)
- How does the system handle rapid key presses faster than event processing? (Buffered channel should queue events up to capacity; document buffer size and overflow behavior)
- What happens if Start() is called multiple times? (Should return error or no-op after first successful initialization)
- How does the system handle Stop() being called before Start()? (Should no-op gracefully without errors)
- What happens on terminals that don't support key-up events? (Pressed field should use best-effort approximation based on event patterns)
- How are modifier-only key presses handled (pressing Shift/Ctrl/Alt alone)? (Document whether these generate events or require combination with other keys)
- What happens during system clock adjustments? (Timestamps should use monotonic clock to prevent backwards time)
- How does the system behave when terminal is resized during input capture? (Document whether resize events are captured or ignored)
- What happens if Poll() is blocked when Stop() is called from another goroutine? (Poll should unblock and return false to signal shutdown)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide normalized key codes (Key type) that are identical across all supported platforms and terminals for common keys (arrows, Enter, Escape, letters, numbers, function keys)
- **FR-002**: System MUST support modifier key detection (Shift, Alt, Ctrl) using composable bitflags
- **FR-003**: System MUST provide blocking event retrieval (Poll method) that waits until an event is available or shutdown occurs
- **FR-004**: System MUST provide non-blocking event retrieval (Next method) that returns immediately with nil if no event is available
- **FR-005**: System MUST automatically detect and initialize the appropriate platform-specific backend (Unix, Windows, etc.) without requiring user configuration
- **FR-006**: System MUST restore terminal to original state when Stop() is called
- **FR-007**: System MUST include timestamp on all events using monotonic clock source
- **FR-008**: System MUST flag OS autorepeat events with explicit Repeat field
- **FR-009**: System MUST support querying current key state (IsPressed method) for real-time applications
- **FR-010**: System MUST support key-up and key-down event detection where platform allows, indicated by Pressed field
- **FR-011**: System MUST run input capture in separate goroutine to prevent blocking event consumers
- **FR-012**: System MUST use buffered channels for event queue to prevent input loss during consumer processing delays
- **FR-013**: System MUST map unparsable escape sequences to KeyUnknown rather than panicking or dropping events
- **FR-014**: System MUST provide printable character representation (Rune field) for character keys in addition to key code
- **FR-015**: System MUST support action mapping (optional GameInput interface) allowing logical actions to be bound to physical keys
- **FR-016**: System MUST allow multiple keys to be bound to the same action
- **FR-017**: System MUST support dynamic action rebinding during runtime
- **FR-018**: System MUST make Start() and Stop() safe to call from any goroutine
- **FR-019**: Poll() MUST return false when system is shutting down to signal event loop termination
- **FR-020**: System MUST support common terminal emulators: xterm, iTerm2, Windows Terminal, GNOME Terminal, and terminal libraries (tcell, termbox)

### Key Entities

- **Key Event**: Represents a single keyboard event containing normalized key code, optional printable character, modifier flags, timestamp, press/release state, and autorepeat flag
- **Key Code**: Enumerated constant representing a normalized key (e.g., KeyUp, KeyDown, KeyEnter, KeyA) that is platform-independent
- **Modifier**: Bitflag representing key modifiers (Shift, Alt, Ctrl) that can be combined
- **Input Interface**: Core abstraction providing Start/Stop lifecycle, Poll/Next event retrieval, and IsPressed state queries
- **Game Input Interface**: Higher-level abstraction providing action mapping and action state queries
- **Backend**: Platform-specific implementation (unixReader, windowsReader) that handles raw terminal I/O and escape sequence parsing

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can integrate the input system into a basic CLI tool and capture normalized keyboard events in under 10 lines of code
- **SC-002**: The same key press produces byte-identical normalized key codes across at least 4 different terminal emulators (iTerm2, xterm, Windows Terminal, GNOME Terminal)
- **SC-003**: System correctly identifies and normalizes at least 95% of common key sequences (arrows, function keys, modifiers, alphanumeric) across supported platforms
- **SC-004**: Event capture latency is under 10 milliseconds from physical key press to Event availability on standard hardware
- **SC-005**: IsPressed() state queries reflect actual key state with under 16 milliseconds staleness (one frame at 60fps) for game applications
- **SC-006**: System supports at least 100 queued events without loss during consumer processing delays
- **SC-007**: Terminal state is restored correctly 100% of the time after Stop() is called, even after abnormal termination
- **SC-008**: Autorepeat detection accuracy is above 95% (correctly flagging repeat vs initial press events)
- **SC-009**: System works correctly on at least 3 major platforms (Linux, macOS, Windows) without platform-specific code in user applications
- **SC-010**: Developers building terminal games can implement responsive game controls using IsPressed() and action mapping with frame rates above 30fps

### Assumptions

- Target platforms support ANSI/VT100 escape sequences or equivalent terminal control APIs
- Terminal emulators provide some mechanism for raw input mode (disabling line buffering)
- Standard library provides monotonic clock access for event timestamps
- Users will handle graceful shutdown (calling Stop() before program exit) in most cases
- Buffer size of 100 events is sufficient for typical use cases (can be configurable if needed)
- Mouse and resize events are out of scope for initial version (can be added later)
- Most terminals support basic escape sequences; exotic terminals may have reduced key coverage
