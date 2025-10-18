package input

import (
	"fmt"
	"time"
	"unicode/utf8"
)

// SequenceNode represents a node in the escape sequence trie.
// Each node can either be a leaf (with a Key value) or an internal
// node with children for continued parsing.
type SequenceNode struct {
	key      Key
	modifier Modifier
	children map[byte]*SequenceNode
}

// SequenceParser parses terminal escape sequences into normalized Events.
// It uses a trie structure for efficient multi-byte sequence recognition.
type SequenceParser struct {
	root *SequenceNode
}

// NewSequenceParser creates a new parser initialized with common
// terminal escape sequences.
func NewSequenceParser() *SequenceParser {
	p := &SequenceParser{
		root: &SequenceNode{
			children: make(map[byte]*SequenceNode),
		},
	}
	p.buildTrie()
	return p
}

// Parse converts a byte sequence into an Event.
// It recognizes escape sequences, control characters, and printable characters.
func (p *SequenceParser) Parse(seq []byte) (Event, error) {
	if len(seq) == 0 {
		return Event{}, fmt.Errorf("empty sequence")
	}

	event := Event{
		Timestamp: time.Now(),
		Pressed:   true,
		Repeat:    false,
	}

	// Handle single-byte sequences
	if len(seq) == 1 {
		b := seq[0]

		// Escape key (standalone ESC)
		if b == 0x1b {
			event.Key = KeyEscape
			return event, nil
		}

		// Control characters (Ctrl+A through Ctrl+Z)
		if b >= 0x01 && b <= 0x1a {
			event.Key = p.ctrlCharToKey(b)
			event.Modifiers = ModCtrl
			return event, nil
		}

		// Tab
		if b == 0x09 {
			event.Key = KeyTab
			event.Rune = '\t'
			return event, nil
		}

		// Enter/Return
		if b == 0x0d {
			event.Key = KeyEnter
			event.Rune = '\r'
			return event, nil
		}

		// Backspace
		if b == 0x7f || b == 0x08 {
			event.Key = KeyBackspace
			return event, nil
		}

		// Space
		if b == 0x20 {
			event.Key = KeySpace
			event.Rune = ' '
			return event, nil
		}

		// Printable ASCII (single-byte UTF-8)
		if b >= 0x20 && b <= 0x7e {
			event.Rune = rune(b)
			event.Key = p.runeToKey(event.Rune)
			return event, nil
		}

		// Check if this is the start of a multi-byte UTF-8 character
		// UTF-8 lead bytes: 0x80-0xFF
		if b >= 0x80 {
			// Multi-byte UTF-8 - fallthrough to UTF-8 handling below
			// (this will be caught by the multi-byte sequence handler)
		} else {
			// Unknown single byte
			event.Key = KeyUnknown
			return event, nil
		}
	}

	// UTF-8 multi-byte character handling
	// Check if this is a valid UTF-8 sequence (not an escape sequence)
	if len(seq) > 0 && seq[0] >= 0x80 && seq[0] != 0x1b {
		// Check for complete UTF-8 sequence
		if !utf8.FullRune(seq) {
			return Event{}, fmt.Errorf("incomplete UTF-8 sequence")
		}

		r, size := utf8.DecodeRune(seq)
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 encoding
			event.Key = KeyUnknown
			event.Rune = utf8.RuneError
			return event, nil
		}

		event.Rune = r
		// Non-ASCII characters map to KeyUnknown
		event.Key = KeyUnknown
		return event, nil
	}

	// Multi-byte sequences - check trie
	node := p.root
	for _, b := range seq {
		if node.children == nil {
			event.Key = KeyUnknown
			return event, nil
		}

		next, ok := node.children[b]
		if !ok {
			event.Key = KeyUnknown
			return event, nil
		}

		node = next
	}

	// Check if we landed on a leaf node
	if node.key != KeyUnknown {
		event.Key = node.key
		event.Modifiers = node.modifier
		return event, nil
	}

	// Unknown sequence
	event.Key = KeyUnknown
	return event, nil
}

// buildTrie constructs the escape sequence trie with common terminal sequences.
func (p *SequenceParser) buildTrie() {
	// CSI sequences (ESC [)
	p.addSequence([]byte{0x1b, '[', 'A'}, KeyUp, ModNone)
	p.addSequence([]byte{0x1b, '[', 'B'}, KeyDown, ModNone)
	p.addSequence([]byte{0x1b, '[', 'C'}, KeyRight, ModNone)
	p.addSequence([]byte{0x1b, '[', 'D'}, KeyLeft, ModNone)

	p.addSequence([]byte{0x1b, '[', 'H'}, KeyHome, ModNone)
	p.addSequence([]byte{0x1b, '[', 'F'}, KeyEnd, ModNone)

	p.addSequence([]byte{0x1b, '[', '2', '~'}, KeyInsert, ModNone)
	p.addSequence([]byte{0x1b, '[', '3', '~'}, KeyDelete, ModNone)
	p.addSequence([]byte{0x1b, '[', '5', '~'}, KeyPageUp, ModNone)
	p.addSequence([]byte{0x1b, '[', '6', '~'}, KeyPageDown, ModNone)

	// Function keys (SS3 sequences: ESC O)
	p.addSequence([]byte{0x1b, 'O', 'P'}, KeyF1, ModNone)
	p.addSequence([]byte{0x1b, 'O', 'Q'}, KeyF2, ModNone)
	p.addSequence([]byte{0x1b, 'O', 'R'}, KeyF3, ModNone)
	p.addSequence([]byte{0x1b, 'O', 'S'}, KeyF4, ModNone)

	// Function keys (CSI sequences: ESC [)
	p.addSequence([]byte{0x1b, '[', '1', '5', '~'}, KeyF5, ModNone)
	p.addSequence([]byte{0x1b, '[', '1', '7', '~'}, KeyF6, ModNone)
	p.addSequence([]byte{0x1b, '[', '1', '8', '~'}, KeyF7, ModNone)
	p.addSequence([]byte{0x1b, '[', '1', '9', '~'}, KeyF8, ModNone)
	p.addSequence([]byte{0x1b, '[', '2', '0', '~'}, KeyF9, ModNone)
	p.addSequence([]byte{0x1b, '[', '2', '1', '~'}, KeyF10, ModNone)
	p.addSequence([]byte{0x1b, '[', '2', '3', '~'}, KeyF11, ModNone)
	p.addSequence([]byte{0x1b, '[', '2', '4', '~'}, KeyF12, ModNone)
}

// addSequence adds a byte sequence to the trie with the given key and modifier.
func (p *SequenceParser) addSequence(seq []byte, key Key, mod Modifier) {
	node := p.root
	for _, b := range seq {
		if node.children == nil {
			node.children = make(map[byte]*SequenceNode)
		}

		if node.children[b] == nil {
			node.children[b] = &SequenceNode{
				key:      KeyUnknown,
				children: make(map[byte]*SequenceNode),
			}
		}

		node = node.children[b]
	}

	// Set the key at the leaf node
	node.key = key
	node.modifier = mod
}

// ctrlCharToKey converts a control character byte to its corresponding Key.
func (p *SequenceParser) ctrlCharToKey(b byte) Key {
	switch b {
	case 0x01:
		return KeyCtrlA
	case 0x02:
		return KeyCtrlB
	case 0x03:
		return KeyCtrlC
	case 0x04:
		return KeyCtrlD
	case 0x05:
		return KeyCtrlE
	case 0x06:
		return KeyCtrlF
	case 0x07:
		return KeyCtrlG
	case 0x08:
		return KeyCtrlH
	// 0x09 is Tab, handled separately
	case 0x0a:
		return KeyCtrlJ
	case 0x0b:
		return KeyCtrlK
	case 0x0c:
		return KeyCtrlL
	// 0x0d is Enter, handled separately
	case 0x0e:
		return KeyCtrlN
	case 0x0f:
		return KeyCtrlO
	case 0x10:
		return KeyCtrlP
	case 0x11:
		return KeyCtrlQ
	case 0x12:
		return KeyCtrlR
	case 0x13:
		return KeyCtrlS
	case 0x14:
		return KeyCtrlT
	case 0x15:
		return KeyCtrlU
	case 0x16:
		return KeyCtrlV
	case 0x17:
		return KeyCtrlW
	case 0x18:
		return KeyCtrlX
	case 0x19:
		return KeyCtrlY
	case 0x1a:
		return KeyCtrlZ
	default:
		return KeyUnknown
	}
}

// runeToKey converts a printable rune to its corresponding Key.
// For letters, it returns the normalized uppercase Key (KeyA-KeyZ).
// For numbers, it returns Key0-Key9.
// For other printable characters, it returns the specific Key or KeyUnknown.
func (p *SequenceParser) runeToKey(r rune) Key {
	switch {
	case r >= 'a' && r <= 'z':
		return Key(int(KeyA) + int(r-'a'))
	case r >= 'A' && r <= 'Z':
		return Key(int(KeyA) + int(r-'A'))
	case r >= '0' && r <= '9':
		return Key(int(Key0) + int(r-'0'))
	case r == ' ':
		return KeySpace
	case r == '\t':
		return KeyTab
	case r == '\r' || r == '\n':
		return KeyEnter
	default:
		return KeyUnknown
	}
}
