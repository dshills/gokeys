package contract_test

import (
	"testing"

	"github.com/dshills/gokeys/input"
)

// TestEscapeSequenceNormalization validates that arrow key escape sequences
// are correctly normalized to their corresponding Key constants.
// This ensures consistent behavior across different terminal emulators.
func TestEscapeSequenceNormalization(t *testing.T) {
	tests := []struct {
		name     string
		sequence []byte
		wantKey  input.Key
		wantMods input.Modifier
	}{
		{
			name:     "Up Arrow",
			sequence: []byte{0x1b, '[', 'A'},
			wantKey:  input.KeyUp,
			wantMods: input.ModNone,
		},
		{
			name:     "Down Arrow",
			sequence: []byte{0x1b, '[', 'B'},
			wantKey:  input.KeyDown,
			wantMods: input.ModNone,
		},
		{
			name:     "Right Arrow",
			sequence: []byte{0x1b, '[', 'C'},
			wantKey:  input.KeyRight,
			wantMods: input.ModNone,
		},
		{
			name:     "Left Arrow",
			sequence: []byte{0x1b, '[', 'D'},
			wantKey:  input.KeyLeft,
			wantMods: input.ModNone,
		},
		{
			name:     "F1 Key",
			sequence: []byte{0x1b, 'O', 'P'},
			wantKey:  input.KeyF1,
			wantMods: input.ModNone,
		},
		{
			name:     "Home Key",
			sequence: []byte{0x1b, '[', 'H'},
			wantKey:  input.KeyHome,
			wantMods: input.ModNone,
		},
		{
			name:     "End Key",
			sequence: []byte{0x1b, '[', 'F'},
			wantKey:  input.KeyEnd,
			wantMods: input.ModNone,
		},
	}

	parser := input.NewSequenceParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.Parse(tt.sequence)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if event.Key != tt.wantKey {
				t.Errorf("Key = %v, want %v", event.Key, tt.wantKey)
			}

			if event.Modifiers != tt.wantMods {
				t.Errorf("Modifiers = %v, want %v", event.Modifiers, tt.wantMods)
			}
		})
	}
}

// TestCtrlKeyNormalization validates that Ctrl+Key combinations are
// correctly normalized with appropriate modifier flags.
func TestCtrlKeyNormalization(t *testing.T) {
	tests := []struct {
		name     string
		sequence []byte
		wantKey  input.Key
		wantMods input.Modifier
	}{
		{
			name:     "Ctrl+C",
			sequence: []byte{0x03},
			wantKey:  input.KeyCtrlC,
			wantMods: input.ModCtrl,
		},
		{
			name:     "Ctrl+A",
			sequence: []byte{0x01},
			wantKey:  input.KeyCtrlA,
			wantMods: input.ModCtrl,
		},
		{
			name:     "Ctrl+Z",
			sequence: []byte{0x1a},
			wantKey:  input.KeyCtrlZ,
			wantMods: input.ModCtrl,
		},
	}

	parser := input.NewSequenceParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.Parse(tt.sequence)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if event.Key != tt.wantKey {
				t.Errorf("Key = %v, want %v", event.Key, tt.wantKey)
			}

			if event.Modifiers != tt.wantMods {
				t.Errorf("Modifiers = %v, want %v", event.Modifiers, tt.wantMods)
			}
		})
	}
}

// TestUnknownSequenceHandling validates that unparsable sequences
// are gracefully handled by returning KeyUnknown rather than panicking.
func TestUnknownSequenceHandling(t *testing.T) {
	tests := []struct {
		name     string
		sequence []byte
		wantKey  input.Key
	}{
		{
			name:     "Unknown CSI sequence",
			sequence: []byte{0x1b, '[', '9', '9', '9', '~'},
			wantKey:  input.KeyUnknown,
		},
		{
			name:     "Incomplete escape sequence",
			sequence: []byte{0x1b},
			wantKey:  input.KeyEscape,
		},
		{
			name:     "Invalid sequence",
			sequence: []byte{0x1b, '[', 'Z', 'Z'},
			wantKey:  input.KeyUnknown,
		},
	}

	parser := input.NewSequenceParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.Parse(tt.sequence)
			if err != nil {
				t.Fatalf("Parse() should not error on unknown sequences, got: %v", err)
			}

			if event.Key != tt.wantKey {
				t.Errorf("Key = %v, want %v", event.Key, tt.wantKey)
			}
		})
	}
}
