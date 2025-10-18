package input

import "time"

// Key represents a normalized key code that is platform-independent.
// Key codes are identical across all supported terminals and operating systems.
type Key int

const (
	// KeyUnknown represents an unparsable or unrecognized key sequence.
	KeyUnknown Key = iota

	// KeyEscape represents the Escape key.
	KeyEscape
	// KeyEnter represents the Enter/Return key.
	KeyEnter
	// KeyBackspace represents the Backspace key.
	KeyBackspace
	// KeyTab represents the Tab key.
	KeyTab
	// KeyDelete represents the Delete key.
	KeyDelete
	// KeyInsert represents the Insert key.
	KeyInsert

	// KeyUp represents the Up arrow key.
	KeyUp
	// KeyDown represents the Down arrow key.
	KeyDown
	// KeyLeft represents the Left arrow key.
	KeyLeft
	// KeyRight represents the Right arrow key.
	KeyRight

	// KeyHome represents the Home key.
	KeyHome
	// KeyEnd represents the End key.
	KeyEnd
	// KeyPageUp represents the Page Up key.
	KeyPageUp
	// KeyPageDown represents the Page Down key.
	KeyPageDown

	// KeyF1 represents the F1 function key.
	KeyF1
	// KeyF2 represents the F2 function key.
	KeyF2
	// KeyF3 represents the F3 function key.
	KeyF3
	// KeyF4 represents the F4 function key.
	KeyF4
	// KeyF5 represents the F5 function key.
	KeyF5
	// KeyF6 represents the F6 function key.
	KeyF6
	// KeyF7 represents the F7 function key.
	KeyF7
	// KeyF8 represents the F8 function key.
	KeyF8
	// KeyF9 represents the F9 function key.
	KeyF9
	// KeyF10 represents the F10 function key.
	KeyF10
	// KeyF11 represents the F11 function key.
	KeyF11
	// KeyF12 represents the F12 function key.
	KeyF12

	// KeyA represents the A key.
	KeyA
	// KeyB represents the B key.
	KeyB
	// KeyC represents the C key.
	KeyC
	// KeyD represents the D key.
	KeyD
	// KeyE represents the E key.
	KeyE
	// KeyF represents the F key.
	KeyF
	// KeyG represents the G key.
	KeyG
	// KeyH represents the H key.
	KeyH
	// KeyI represents the I key.
	KeyI
	// KeyJ represents the J key.
	KeyJ
	// KeyK represents the K key.
	KeyK
	// KeyL represents the L key.
	KeyL
	// KeyM represents the M key.
	KeyM
	// KeyN represents the N key.
	KeyN
	// KeyO represents the O key.
	KeyO
	// KeyP represents the P key.
	KeyP
	// KeyQ represents the Q key.
	KeyQ
	// KeyR represents the R key.
	KeyR
	// KeyS represents the S key.
	KeyS
	// KeyT represents the T key.
	KeyT
	// KeyU represents the U key.
	KeyU
	// KeyV represents the V key.
	KeyV
	// KeyW represents the W key.
	KeyW
	// KeyX represents the X key.
	KeyX
	// KeyY represents the Y key.
	KeyY
	// KeyZ represents the Z key.
	KeyZ

	// Key0 represents the 0 key.
	Key0
	// Key1 represents the 1 key.
	Key1
	// Key2 represents the 2 key.
	Key2
	// Key3 represents the 3 key.
	Key3
	// Key4 represents the 4 key.
	Key4
	// Key5 represents the 5 key.
	Key5
	// Key6 represents the 6 key.
	Key6
	// Key7 represents the 7 key.
	Key7
	// Key8 represents the 8 key.
	Key8
	// Key9 represents the 9 key.
	Key9

	// KeyCtrlA represents Ctrl+A.
	KeyCtrlA
	// KeyCtrlB represents Ctrl+B.
	KeyCtrlB
	// KeyCtrlC represents Ctrl+C.
	KeyCtrlC
	// KeyCtrlD represents Ctrl+D.
	KeyCtrlD
	// KeyCtrlE represents Ctrl+E.
	KeyCtrlE
	// KeyCtrlF represents Ctrl+F.
	KeyCtrlF
	// KeyCtrlG represents Ctrl+G.
	KeyCtrlG
	// KeyCtrlH represents Ctrl+H.
	KeyCtrlH
	// KeyCtrlI represents Ctrl+I.
	KeyCtrlI
	// KeyCtrlJ represents Ctrl+J.
	KeyCtrlJ
	// KeyCtrlK represents Ctrl+K.
	KeyCtrlK
	// KeyCtrlL represents Ctrl+L.
	KeyCtrlL
	// KeyCtrlM represents Ctrl+M.
	KeyCtrlM
	// KeyCtrlN represents Ctrl+N.
	KeyCtrlN
	// KeyCtrlO represents Ctrl+O.
	KeyCtrlO
	// KeyCtrlP represents Ctrl+P.
	KeyCtrlP
	// KeyCtrlQ represents Ctrl+Q.
	KeyCtrlQ
	// KeyCtrlR represents Ctrl+R.
	KeyCtrlR
	// KeyCtrlS represents Ctrl+S.
	KeyCtrlS
	// KeyCtrlT represents Ctrl+T.
	KeyCtrlT
	// KeyCtrlU represents Ctrl+U.
	KeyCtrlU
	// KeyCtrlV represents Ctrl+V.
	KeyCtrlV
	// KeyCtrlW represents Ctrl+W.
	KeyCtrlW
	// KeyCtrlX represents Ctrl+X.
	KeyCtrlX
	// KeyCtrlY represents Ctrl+Y.
	KeyCtrlY
	// KeyCtrlZ represents Ctrl+Z.
	KeyCtrlZ

	// KeySpace represents the Space key.
	KeySpace
)

// Modifier represents key modifiers that can be combined using bitwise OR.
// Multiple modifiers can be active simultaneously.
type Modifier int

const (
	// ModNone indicates no modifiers are active.
	ModNone Modifier = 0

	// ModShift indicates the Shift key is pressed.
	ModShift Modifier = 1 << iota

	// ModAlt indicates the Alt key is pressed.
	ModAlt

	// ModCtrl indicates the Ctrl key is pressed.
	ModCtrl
)

// Event represents a single keyboard event with all associated metadata.
// Events are produced by the input system and consumed via Poll or Next.
type Event struct {
	// Key is the normalized key code for this event.
	Key Key

	// Rune is the printable character for this key, if applicable.
	// For non-printable keys (arrows, function keys, etc.), Rune is 0.
	Rune rune

	// Modifiers contains the active modifier keys (Shift, Alt, Ctrl).
	// Multiple modifiers can be combined using bitwise OR.
	Modifiers Modifier

	// Timestamp is the monotonic time when this event was captured.
	// Monotonic timestamps are not affected by system clock adjustments.
	Timestamp time.Time

	// Pressed indicates the key state: true for key-down, false for key-up.
	// On platforms without key-up event support, this field uses best-effort
	// approximation based on event patterns.
	Pressed bool

	// Repeat indicates whether this is an OS autorepeat event.
	// The first press has Repeat=false, subsequent repeats have Repeat=true.
	Repeat bool
}

// String returns a human-readable string representation of the Key.
//nolint:cyclop // Unavoidable complexity for 100+ key mappings
func (k Key) String() string {
	switch k {
	case KeyUnknown:
		return "Unknown"
	case KeyEscape:
		return "Escape"
	case KeyEnter:
		return "Enter"
	case KeyBackspace:
		return "Backspace"
	case KeyTab:
		return "Tab"
	case KeyDelete:
		return "Delete"
	case KeyInsert:
		return "Insert"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	case KeySpace:
		return "Space"
	case KeyA:
		return "A"
	case KeyB:
		return "B"
	case KeyC:
		return "C"
	case KeyD:
		return "D"
	case KeyE:
		return "E"
	case KeyF:
		return "F"
	case KeyG:
		return "G"
	case KeyH:
		return "H"
	case KeyI:
		return "I"
	case KeyJ:
		return "J"
	case KeyK:
		return "K"
	case KeyL:
		return "L"
	case KeyM:
		return "M"
	case KeyN:
		return "N"
	case KeyO:
		return "O"
	case KeyP:
		return "P"
	case KeyQ:
		return "Q"
	case KeyR:
		return "R"
	case KeyS:
		return "S"
	case KeyT:
		return "T"
	case KeyU:
		return "U"
	case KeyV:
		return "V"
	case KeyW:
		return "W"
	case KeyX:
		return "X"
	case KeyY:
		return "Y"
	case KeyZ:
		return "Z"
	case Key0:
		return "0"
	case Key1:
		return "1"
	case Key2:
		return "2"
	case Key3:
		return "3"
	case Key4:
		return "4"
	case Key5:
		return "5"
	case Key6:
		return "6"
	case Key7:
		return "7"
	case Key8:
		return "8"
	case Key9:
		return "9"
	case KeyCtrlA:
		return "Ctrl+A"
	case KeyCtrlB:
		return "Ctrl+B"
	case KeyCtrlC:
		return "Ctrl+C"
	case KeyCtrlD:
		return "Ctrl+D"
	case KeyCtrlE:
		return "Ctrl+E"
	case KeyCtrlF:
		return "Ctrl+F"
	case KeyCtrlG:
		return "Ctrl+G"
	case KeyCtrlH:
		return "Ctrl+H"
	case KeyCtrlI:
		return "Ctrl+I"
	case KeyCtrlJ:
		return "Ctrl+J"
	case KeyCtrlK:
		return "Ctrl+K"
	case KeyCtrlL:
		return "Ctrl+L"
	case KeyCtrlM:
		return "Ctrl+M"
	case KeyCtrlN:
		return "Ctrl+N"
	case KeyCtrlO:
		return "Ctrl+O"
	case KeyCtrlP:
		return "Ctrl+P"
	case KeyCtrlQ:
		return "Ctrl+Q"
	case KeyCtrlR:
		return "Ctrl+R"
	case KeyCtrlS:
		return "Ctrl+S"
	case KeyCtrlT:
		return "Ctrl+T"
	case KeyCtrlU:
		return "Ctrl+U"
	case KeyCtrlV:
		return "Ctrl+V"
	case KeyCtrlW:
		return "Ctrl+W"
	case KeyCtrlX:
		return "Ctrl+X"
	case KeyCtrlY:
		return "Ctrl+Y"
	case KeyCtrlZ:
		return "Ctrl+Z"
	default:
		return "Unknown"
	}
}
