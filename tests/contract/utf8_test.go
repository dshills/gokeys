package contract_test

import (
	"testing"

	"github.com/dshills/gokeys/input"
)

// TestUTF8TwoByte verifies correct decoding of 2-byte UTF-8 characters.
// This tests common European characters (accents, umlauts, etc.).
func TestUTF8TwoByte(t *testing.T) {
	parser := input.NewSequenceParser()

	tests := []struct {
		name string
		seq  []byte
		want rune
	}{
		{"e-acute", []byte{0xc3, 0xa9}, 'é'},     // U+00E9
		{"n-tilde", []byte{0xc3, 0xb1}, 'ñ'},     // U+00F1
		{"a-umlaut", []byte{0xc3, 0xa4}, 'ä'},    // U+00E4
		{"o-umlaut", []byte{0xc3, 0xb6}, 'ö'},    // U+00F6
		{"u-umlaut", []byte{0xc3, 0xbc}, 'ü'},    // U+00FC
		{"euro-sign", []byte{0xc2, 0xa3}, '£'},   // U+00A3
		{"cent-sign", []byte{0xc2, 0xa2}, '¢'},   // U+00A2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.Parse(tt.seq)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if event.Rune != tt.want {
				t.Errorf("Rune = %c (U+%04X), want %c (U+%04X)",
					event.Rune, event.Rune, tt.want, tt.want)
			}
			// Non-ASCII characters should map to KeyUnknown
			if event.Key != input.KeyUnknown {
				t.Errorf("Key = %v, want KeyUnknown for non-ASCII", event.Key)
			}
		})
	}
}

// TestUTF8ThreeByte verifies correct decoding of 3-byte UTF-8 characters.
// This tests CJK characters and symbols.
func TestUTF8ThreeByte(t *testing.T) {
	parser := input.NewSequenceParser()

	tests := []struct {
		name string
		seq  []byte
		want rune
	}{
		{"euro", []byte{0xe2, 0x82, 0xac}, '€'},         // U+20AC
		{"hiragana-a", []byte{0xe3, 0x81, 0x82}, 'あ'},    // U+3042
		{"hiragana-i", []byte{0xe3, 0x81, 0x84}, 'い'},    // U+3044
		{"hiragana-u", []byte{0xe3, 0x81, 0x86}, 'う'},    // U+3046
		{"katakana-a", []byte{0xe3, 0x82, 0xa2}, 'ア'},    // U+30A2
		{"kanji-day", []byte{0xe6, 0x97, 0xa5}, '日'},     // U+65E5
		{"kanji-book", []byte{0xe6, 0x9c, 0xac}, '本'},    // U+672C
		{"chinese-good", []byte{0xe5, 0xa5, 0xbd}, '好'},  // U+597D
		{"arrow-right", []byte{0xe2, 0x86, 0x92}, '→'},   // U+2192
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.Parse(tt.seq)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if event.Rune != tt.want {
				t.Errorf("Rune = %c (U+%04X), want %c (U+%04X)",
					event.Rune, event.Rune, tt.want, tt.want)
			}
			// Non-ASCII characters should map to KeyUnknown
			if event.Key != input.KeyUnknown {
				t.Errorf("Key = %v, want KeyUnknown for non-ASCII", event.Key)
			}
		})
	}
}

// TestUTF8FourByte verifies correct decoding of 4-byte UTF-8 characters.
// This tests emoji and other extended Unicode characters.
func TestUTF8FourByte(t *testing.T) {
	parser := input.NewSequenceParser()

	tests := []struct {
		name string
		seq  []byte
		want rune
	}{
		{"grinning-face", []byte{0xf0, 0x9f, 0x98, 0x80}, '😀'},    // U+1F600
		{"thumbs-up", []byte{0xf0, 0x9f, 0x91, 0x8d}, '👍'},        // U+1F44D
		{"heart", []byte{0xf0, 0x9f, 0x92, 0x96}, '💖'},            // U+1F496
		{"musical-note", []byte{0xf0, 0x9d, 0x84, 0x9e}, '𝄞'},     // U+1D11E
		{"rocket", []byte{0xf0, 0x9f, 0x9a, 0x80}, '🚀'},           // U+1F680
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.Parse(tt.seq)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if event.Rune != tt.want {
				t.Errorf("Rune = %c (U+%04X), want %c (U+%04X)",
					event.Rune, event.Rune, tt.want, tt.want)
			}
			// Non-ASCII characters should map to KeyUnknown
			if event.Key != input.KeyUnknown {
				t.Errorf("Key = %v, want KeyUnknown for non-ASCII", event.Key)
			}
		})
	}
}

// TestUTF8ASCIIBackwardCompatibility ensures existing ASCII behavior unchanged.
func TestUTF8ASCIIBackwardCompatibility(t *testing.T) {
	parser := input.NewSequenceParser()

	tests := []struct {
		name string
		seq  []byte
		want rune
		key  input.Key
	}{
		{"lowercase-a", []byte{'a'}, 'a', input.KeyA},
		{"uppercase-A", []byte{'A'}, 'A', input.KeyA},
		{"digit-5", []byte{'5'}, '5', input.Key5},
		{"space", []byte{' '}, ' ', input.KeySpace},
		{"exclamation", []byte{'!'}, '!', input.KeyUnknown}, // Non-alphanumeric ASCII
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parser.Parse(tt.seq)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if event.Rune != tt.want {
				t.Errorf("Rune = %c, want %c", event.Rune, tt.want)
			}
			if event.Key != tt.key {
				t.Errorf("Key = %v, want %v", event.Key, tt.key)
			}
		})
	}
}
