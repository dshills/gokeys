# Feature Specification: GameInput Action Mapping API

**Feature Branch**: `004-game-input-api`
**Created**: 2025-10-18
**Status**: Draft
**Input**: User description: "1. GameInput"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Action Binding (Priority: P1)

A game developer wants to map logical action names (like "jump", "fire", "move-left") to physical keys. They need a simple API to bind actions and query whether those actions are currently active, without having to check individual key states.

**Why this priority**: This is the foundational use case for action mapping - it provides the core value of decoupling game logic from physical keys. This enables all other action-mapping features and represents the minimum viable product.

**Independent Test**: Can be fully tested by creating a GameInput instance, binding single actions to single keys (e.g., "jump" → Space), pressing those keys, and verifying that IsActionPressed("jump") returns true when Space is held and false when released. Delivers immediate value by simplifying game code.

**Acceptance Scenarios**:

1. **Given** a GameInput instance is started, **When** Bind("jump", KeySpace) is called, **Then** pressing Space makes IsActionPressed("jump") return true
2. **Given** an action is bound to a key, **When** that key is released, **Then** IsActionPressed returns false for that action
3. **Given** multiple actions are bound to different keys, **When** each key is pressed, **Then** only the corresponding action returns true from IsActionPressed
4. **Given** an action has no binding, **When** IsActionPressed is called for that action, **Then** it returns false (not an error)
5. **Given** a GameInput instance, **When** Start() and Stop() are called, **Then** the underlying input system lifecycle is managed correctly

---

### User Story 2 - Multiple Keys Per Action (Priority: P2)

A game developer wants to support alternative control schemes by allowing multiple keys to trigger the same action. For example, both arrow keys and WASD should work for movement, or both Enter and Space should confirm dialogs.

**Why this priority**: Extends P1 to support real-world game requirements like accessibility, alternative control schemes (WASD vs arrows), and multiple confirm/cancel keys. Builds directly on single-key binding.

**Independent Test**: Can be tested by binding multiple keys to one action (e.g., "confirm" → Enter, Space, Y), pressing any of those keys, and verifying IsActionPressed("confirm") returns true. Delivers value by supporting flexible control schemes.

**Acceptance Scenarios**:

1. **Given** an action "fire" is bound to both Space and Enter, **When** either key is pressed, **Then** IsActionPressed("fire") returns true
2. **Given** an action is bound to 3 keys, **When** all keys are released, **Then** IsActionPressed returns false
3. **Given** movement actions bound to both arrows and WASD, **When** either set is used, **Then** the corresponding action is detected
4. **Given** multiple keys bound to an action, **When** one key is held and another is pressed, **Then** IsActionPressed remains true until all are released

---

### User Story 3 - Dynamic Action Rebinding (Priority: P3)

A game developer wants to allow players to rebind controls at runtime through a settings menu. They need to replace existing key bindings for actions and have the changes take effect immediately without restarting the game.

**Why this priority**: This is a convenience feature for advanced games that want customizable controls. It builds on P1 and P2 but isn't essential for basic action mapping functionality.

**Independent Test**: Can be tested by binding an action to one key, then calling Bind again with a different key, and verifying that only the new binding works. Delivers value by enabling player control customization.

**Acceptance Scenarios**:

1. **Given** "jump" is initially bound to Space, **When** Bind("jump", KeyJ) is called, **Then** only J triggers the jump action (Space no longer works)
2. **Given** an action has existing bindings, **When** Bind is called with new keys, **Then** the old bindings are completely replaced
3. **Given** rebinding happens during gameplay, **When** the new bindings are applied, **Then** they take effect immediately on the next key press
4. **Given** an action is bound, **When** Bind is called with no keys (empty list), **Then** the action is unbound and IsActionPressed always returns false

---

### Edge Cases

- What happens when the same key is bound to multiple actions? (Both actions should return true when that key is pressed)
- How does the system handle calling IsActionPressed with an empty string action name? (Should treat it as a valid action name, returning false if not bound)
- What happens if Bind is called before Start()? (Bindings should be stored and take effect when Start() is called)
- How does the system handle concurrent Bind and IsActionPressed calls from different goroutines? (Must be thread-safe with proper synchronization)
- What happens when Stop() is called while actions are pressed? (IsActionPressed should still work based on last known state until system is fully stopped)
- How many keys can be bound to a single action? (Document any practical limits - suggest supporting at least 10 keys per action)
- What happens with case sensitivity in action names? (Action names should be case-sensitive: "Jump" ≠ "jump")

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a GameInput interface that wraps the lower-level Input interface
- **FR-002**: System MUST provide a Bind method that associates one or more keys with a named action
- **FR-003**: System MUST provide an IsActionPressed method that returns true if any key bound to that action is currently pressed
- **FR-004**: System MUST allow binding multiple keys to the same action (OR relationship - any key triggers the action)
- **FR-005**: System MUST treat action names as case-sensitive strings
- **FR-006**: System MUST replace existing bindings when Bind is called for an action that already has bindings
- **FR-007**: System MUST return false (not error) from IsActionPressed when querying an unbound action
- **FR-008**: System MUST allow the same physical key to be bound to multiple different actions simultaneously
- **FR-009**: System MUST provide Start and Stop methods that delegate to the underlying Input system
- **FR-010**: System MUST support binding changes at runtime that take effect immediately
- **FR-011**: System MUST be thread-safe for concurrent Bind and IsActionPressed calls
- **FR-012**: System MUST support unbinding an action by calling Bind with an empty key list
- **FR-013**: System MUST query the underlying Input.IsPressed for each bound key to determine action state
- **FR-014**: System MUST support at least 100 unique action names simultaneously
- **FR-015**: System MUST support at least 10 keys bound to a single action

### Key Entities

- **GameInput Interface**: Higher-level abstraction providing action mapping on top of the Input interface. Contains methods: Start, Stop, Bind, IsActionPressed
- **Action Binding**: Association between a string action name and a list of Key values that trigger that action
- **Action Name**: Case-sensitive string identifier for a logical game action (e.g., "jump", "fire", "move-left")

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Game developers can bind an action and query its state in under 5 lines of code
- **SC-002**: IsActionPressed queries respond in under 1 millisecond (faster than 60fps frame time of 16ms)
- **SC-003**: System correctly handles 100 simultaneous action bindings without performance degradation
- **SC-004**: Action rebinding takes effect within one input poll cycle (under 20 milliseconds)
- **SC-005**: Game logic using action names has zero hardcoded key references, enabling 100% rebindable controls
- **SC-006**: Thread-safe operations allow concurrent Bind calls from UI thread and IsActionPressed from game loop without race conditions
- **SC-007**: Games using action mapping can run at 60fps with 10+ actions being queried per frame without input latency

### Assumptions

- The underlying Input interface is already implemented and provides IsPressed(Key) functionality
- Game developers understand the concept of logical actions vs physical keys
- Most games will have between 5-20 distinct actions
- Most actions will have 1-3 keys bound, rarely exceeding 10 keys per action
- Action names will be chosen by developers and stored as constants in game code
- Concurrent access patterns will be: frequent reads (IsActionPressed) and rare writes (Bind)
- The GameInput API is optional - developers can still use the lower-level Input API directly if preferred
