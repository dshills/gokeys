package input

import (
	"testing"
	"time"
)

// TestEventZeroValues validates that Event zero values are sensible.
func TestEventZeroValues(t *testing.T) {
	var e Event

	if e.Key != KeyUnknown {
		t.Errorf("Zero Event.Key = %v, want KeyUnknown", e.Key)
	}

	if e.Rune != 0 {
		t.Errorf("Zero Event.Rune = %v, want 0", e.Rune)
	}

	if e.Modifiers != ModNone {
		t.Errorf("Zero Event.Modifiers = %v, want ModNone", e.Modifiers)
	}

	if !e.Timestamp.IsZero() {
		t.Errorf("Zero Event.Timestamp should be zero time")
	}

	if e.Pressed != false {
		t.Errorf("Zero Event.Pressed = %v, want false", e.Pressed)
	}

	if e.Repeat != false {
		t.Errorf("Zero Event.Repeat = %v, want false", e.Repeat)
	}
}

// TestEventFieldValidation validates Event field constraints.
func TestEventFieldValidation(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		valid bool
	}{
		{
			name: "Valid key press",
			event: Event{
				Key:       KeyA,
				Rune:      'a',
				Modifiers: ModNone,
				Timestamp: time.Now(),
				Pressed:   true,
				Repeat:    false,
			},
			valid: true,
		},
		{
			name: "Valid key press with modifiers",
			event: Event{
				Key:       KeyA,
				Rune:      'A',
				Modifiers: ModShift,
				Timestamp: time.Now(),
				Pressed:   true,
				Repeat:    false,
			},
			valid: true,
		},
		{
			name: "Valid autorepeat",
			event: Event{
				Key:       KeySpace,
				Rune:      ' ',
				Modifiers: ModNone,
				Timestamp: time.Now(),
				Pressed:   true,
				Repeat:    true,
			},
			valid: true,
		},
		{
			name: "Non-printable key has zero rune",
			event: Event{
				Key:       KeyUp,
				Rune:      0,
				Modifiers: ModNone,
				Timestamp: time.Now(),
				Pressed:   true,
				Repeat:    false,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate timestamp is set
			if tt.event.Timestamp.IsZero() {
				t.Error("Event.Timestamp must be set")
			}

			// Validate Repeat implies Pressed
			if tt.event.Repeat && !tt.event.Pressed {
				t.Error("Event.Repeat can only be true if Pressed is true")
			}
		})
	}
}

// TestModifierBitflagOperations validates that Modifier bitflags
// can be correctly combined and checked.
func TestModifierBitflagOperations(t *testing.T) {
	tests := []struct {
		name      string
		modifiers Modifier
		hasShift  bool
		hasAlt    bool
		hasCtrl   bool
	}{
		{
			name:      "No modifiers",
			modifiers: ModNone,
			hasShift:  false,
			hasAlt:    false,
			hasCtrl:   false,
		},
		{
			name:      "Shift only",
			modifiers: ModShift,
			hasShift:  true,
			hasAlt:    false,
			hasCtrl:   false,
		},
		{
			name:      "Alt only",
			modifiers: ModAlt,
			hasShift:  false,
			hasAlt:    true,
			hasCtrl:   false,
		},
		{
			name:      "Ctrl only",
			modifiers: ModCtrl,
			hasShift:  false,
			hasAlt:    false,
			hasCtrl:   true,
		},
		{
			name:      "Shift+Ctrl",
			modifiers: ModShift | ModCtrl,
			hasShift:  true,
			hasAlt:    false,
			hasCtrl:   true,
		},
		{
			name:      "Shift+Alt+Ctrl",
			modifiers: ModShift | ModAlt | ModCtrl,
			hasShift:  true,
			hasAlt:    true,
			hasCtrl:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (tt.modifiers & ModShift) != 0; got != tt.hasShift {
				t.Errorf("Has Shift = %v, want %v", got, tt.hasShift)
			}

			if got := (tt.modifiers & ModAlt) != 0; got != tt.hasAlt {
				t.Errorf("Has Alt = %v, want %v", got, tt.hasAlt)
			}

			if got := (tt.modifiers & ModCtrl) != 0; got != tt.hasCtrl {
				t.Errorf("Has Ctrl = %v, want %v", got, tt.hasCtrl)
			}
		})
	}
}

// TestModifierCombinations validates that multiple modifiers can be
// combined using bitwise OR.
func TestModifierCombinations(t *testing.T) {
	// Combine Shift and Ctrl
	combo := ModShift | ModCtrl

	if combo&ModShift == 0 {
		t.Error("Combined modifier should include Shift")
	}

	if combo&ModCtrl == 0 {
		t.Error("Combined modifier should include Ctrl")
	}

	if combo&ModAlt != 0 {
		t.Error("Combined modifier should not include Alt")
	}

	// Combine all three
	allMods := ModShift | ModAlt | ModCtrl

	if allMods&ModShift == 0 {
		t.Error("All modifiers should include Shift")
	}

	if allMods&ModAlt == 0 {
		t.Error("All modifiers should include Alt")
	}

	if allMods&ModCtrl == 0 {
		t.Error("All modifiers should include Ctrl")
	}
}
