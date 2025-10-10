package matrixcui

// Event is the interface for all terminal events.
type Event interface {
	Type() EventType
}

// EventType represents the type of event.
type EventType int

const (
	EventTypeKey EventType = iota
	EventTypeResize
)

// KeyEvent represents a keyboard event.
type KeyEvent struct {
	Key  Key    // Special key (if any)
	Rune rune   // Character typed (if printable)
	Alt  bool   // Alt modifier
	Ctrl bool   // Ctrl modifier
	Shift bool  // Shift modifier (mainly for special keys)
}

// Type returns the event type.
func (k KeyEvent) Type() EventType {
	return EventTypeKey
}

// ResizeEvent represents a terminal resize event.
type ResizeEvent struct {
	Width  int
	Height int
}

// Type returns the event type.
func (r ResizeEvent) Type() EventType {
	return EventTypeResize
}

// Key represents special keys (non-printable characters).
type Key int

const (
	KeyNone Key = iota // No special key (use Rune instead)

	// Control keys
	KeyEnter
	KeyBackspace
	KeyTab
	KeyEscape
	KeySpace

	// Arrow keys
	KeyUp
	KeyDown
	KeyLeft
	KeyRight

	// Navigation keys
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	// Editing keys
	KeyInsert
	KeyDelete

	// Function keys
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// Ctrl+letter combinations (for common ones)
	KeyCtrlC
	KeyCtrlD
	KeyCtrlZ
)

// String returns a string representation of the key.
func (k Key) String() string {
	switch k {
	case KeyNone:
		return "None"
	case KeyEnter:
		return "Enter"
	case KeyBackspace:
		return "Backspace"
	case KeyTab:
		return "Tab"
	case KeyEscape:
		return "Escape"
	case KeySpace:
		return "Space"
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
	case KeyInsert:
		return "Insert"
	case KeyDelete:
		return "Delete"
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
	case KeyCtrlC:
		return "Ctrl+C"
	case KeyCtrlD:
		return "Ctrl+D"
	case KeyCtrlZ:
		return "Ctrl+Z"
	default:
		return "Unknown"
	}
}
