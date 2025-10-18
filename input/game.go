package input

// GameInput provides a higher-level abstraction over Input for game development.
// It allows mapping logical actions (e.g., "jump", "fire") to physical keys,
// enabling key rebinding and simplifying game input logic.
type GameInput interface {
	// Start initializes the underlying Input system.
	// Delegates to the wrapped Input.Start().
	//
	// Returns error if underlying Input fails to start.
	Start() error

	// Stop cleans up and restores terminal state.
	// Delegates to the wrapped Input.Stop().
	Stop()

	// IsActionPressed returns true if any key bound to the action is currently pressed.
	// Returns false if action has no bound keys or none are pressed.
	//
	// Thread-safe: Safe for concurrent calls.
	IsActionPressed(action string) bool

	// Bind associates one or more keys with a logical action name.
	// If the action already has bindings, they are replaced.
	// Passing no keys unbinds the action.
	//
	// Thread-safe: Safe for concurrent calls.
	Bind(action string, keys ...Key)
}

// NewGameInput creates a new GameInput instance.
// If input is nil, creates a default Input via input.New().
func NewGameInput(input Input) GameInput {
	if input == nil {
		input = New()
	}
	return &gameInputImpl{
		input:    input,
		bindings: make(map[string][]Key),
	}
}
