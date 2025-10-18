package input

import "sync"

// gameInputImpl is the concrete implementation of GameInput.
type gameInputImpl struct {
	input    Input
	bindings map[string][]Key
	mu       sync.RWMutex
}

// Start delegates to the underlying Input.
func (g *gameInputImpl) Start() error {
	return g.input.Start()
}

// Stop delegates to the underlying Input.
func (g *gameInputImpl) Stop() {
	g.input.Stop()
}

// IsActionPressed returns true if any key bound to the action is pressed.
func (g *gameInputImpl) IsActionPressed(action string) bool {
	g.mu.RLock()
	keys, ok := g.bindings[action]
	if !ok {
		g.mu.RUnlock()
		return false // Unbound action
	}

	// Copy keys slice to avoid race condition after unlock
	keysCopy := make([]Key, len(keys))
	copy(keysCopy, keys)
	g.mu.RUnlock()

	// OR logic: any key pressed → action pressed
	for _, key := range keysCopy {
		if g.input.IsPressed(key) {
			return true
		}
	}
	return false
}

// Bind associates keys with an action. Empty keys unbinds the action.
func (g *gameInputImpl) Bind(action string, keys ...Key) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(keys) == 0 {
		delete(g.bindings, action) // Unbind
	} else {
		// Defensive copy to prevent aliasing issues
		keyCopy := make([]Key, len(keys))
		copy(keyCopy, keys)
		g.bindings[action] = keyCopy // Bind/rebind
	}
}
