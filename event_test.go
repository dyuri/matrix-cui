package matrixcui

import (
	"os"
	"testing"
)

// Helper to create an EventReader with custom input
func newTestEventReader(input string) *EventReader {
	r, w, _ := os.Pipe()
	go func() {
		w.Write([]byte(input))
		w.Close()
	}()
	return NewEventReaderWithInput(r)
}

func TestKeyEventBasicChars(t *testing.T) {
	tests := []struct {
		input    string
		expected KeyEvent
	}{
		{"a", KeyEvent{Key: KeyNone, Rune: 'a'}},
		{"Z", KeyEvent{Key: KeyNone, Rune: 'Z'}},
		{"5", KeyEvent{Key: KeyNone, Rune: '5'}},
		{" ", KeyEvent{Key: KeyNone, Rune: ' '}},
	}

	for _, tt := range tests {
		reader := newTestEventReader(tt.input)
		event, err := reader.ReadEvent()
		if err != nil {
			t.Errorf("ReadEvent(%q) error: %v", tt.input, err)
			continue
		}

		keyEvent, ok := event.(KeyEvent)
		if !ok {
			t.Errorf("ReadEvent(%q) did not return KeyEvent", tt.input)
			continue
		}

		if keyEvent.Key != tt.expected.Key || keyEvent.Rune != tt.expected.Rune {
			t.Errorf("ReadEvent(%q) = {Key: %v, Rune: %c}, expected {Key: %v, Rune: %c}",
				tt.input, keyEvent.Key, keyEvent.Rune, tt.expected.Key, tt.expected.Rune)
		}
	}
}

func TestKeyEventControlKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected KeyEvent
	}{
		{"\r", KeyEvent{Key: KeyEnter}},        // Carriage return
		{"\n", KeyEvent{Key: KeyEnter}},        // Line feed
		{"\t", KeyEvent{Key: KeyTab}},          // Tab
		{"\x7f", KeyEvent{Key: KeyBackspace}},  // DEL (127)
		{"\x03", KeyEvent{Key: KeyCtrlC, Ctrl: true}}, // Ctrl+C
		{"\x04", KeyEvent{Key: KeyCtrlD, Ctrl: true}}, // Ctrl+D
		{"\x1a", KeyEvent{Key: KeyCtrlZ, Ctrl: true}}, // Ctrl+Z
	}

	for _, tt := range tests {
		reader := newTestEventReader(tt.input)
		event, err := reader.ReadEvent()
		if err != nil {
			t.Errorf("ReadEvent(%q) error: %v", tt.input, err)
			continue
		}

		keyEvent, ok := event.(KeyEvent)
		if !ok {
			t.Errorf("ReadEvent(%q) did not return KeyEvent", tt.input)
			continue
		}

		if keyEvent.Key != tt.expected.Key {
			t.Errorf("ReadEvent(%q) Key = %v, expected %v", tt.input, keyEvent.Key, tt.expected.Key)
		}
		if keyEvent.Ctrl != tt.expected.Ctrl {
			t.Errorf("ReadEvent(%q) Ctrl = %v, expected %v", tt.input, keyEvent.Ctrl, tt.expected.Ctrl)
		}
	}
}

func TestKeyEventArrowKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected Key
	}{
		{"\x1b[A", KeyUp},
		{"\x1b[B", KeyDown},
		{"\x1b[C", KeyRight},
		{"\x1b[D", KeyLeft},
	}

	for _, tt := range tests {
		reader := newTestEventReader(tt.input)
		event, err := reader.ReadEvent()
		if err != nil {
			t.Errorf("ReadEvent(%q) error: %v", tt.input, err)
			continue
		}

		keyEvent, ok := event.(KeyEvent)
		if !ok {
			t.Errorf("ReadEvent(%q) did not return KeyEvent", tt.input)
			continue
		}

		if keyEvent.Key != tt.expected {
			t.Errorf("ReadEvent(%q) Key = %v, expected %v", tt.input, keyEvent.Key, tt.expected)
		}
	}
}

func TestKeyEventNavigationKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected Key
	}{
		{"\x1b[H", KeyHome},
		{"\x1b[F", KeyEnd},
		{"\x1b[1~", KeyHome},
		{"\x1b[4~", KeyEnd},
		{"\x1b[5~", KeyPageUp},
		{"\x1b[6~", KeyPageDown},
		{"\x1b[2~", KeyInsert},
		{"\x1b[3~", KeyDelete},
	}

	for _, tt := range tests {
		reader := newTestEventReader(tt.input)
		event, err := reader.ReadEvent()
		if err != nil {
			t.Errorf("ReadEvent(%q) error: %v", tt.input, err)
			continue
		}

		keyEvent, ok := event.(KeyEvent)
		if !ok {
			t.Errorf("ReadEvent(%q) did not return KeyEvent", tt.input)
			continue
		}

		if keyEvent.Key != tt.expected {
			t.Errorf("ReadEvent(%q) Key = %v, expected %v", tt.input, keyEvent.Key, tt.expected)
		}
	}
}

func TestKeyEventFunctionKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected Key
	}{
		{"\x1bOP", KeyF1},
		{"\x1bOQ", KeyF2},
		{"\x1bOR", KeyF3},
		{"\x1bOS", KeyF4},
		{"\x1b[15~", KeyF5},
		{"\x1b[17~", KeyF6},
		{"\x1b[18~", KeyF7},
		{"\x1b[19~", KeyF8},
		{"\x1b[20~", KeyF9},
		{"\x1b[21~", KeyF10},
		{"\x1b[23~", KeyF11},
		{"\x1b[24~", KeyF12},
	}

	for _, tt := range tests {
		reader := newTestEventReader(tt.input)
		event, err := reader.ReadEvent()
		if err != nil {
			t.Errorf("ReadEvent(%q) error: %v", tt.input, err)
			continue
		}

		keyEvent, ok := event.(KeyEvent)
		if !ok {
			t.Errorf("ReadEvent(%q) did not return KeyEvent", tt.input)
			continue
		}

		if keyEvent.Key != tt.expected {
			t.Errorf("ReadEvent(%q) Key = %v, expected %v", tt.input, keyEvent.Key, tt.expected)
		}
	}
}

func TestKeyEventEscape(t *testing.T) {
	reader := newTestEventReader("\x1b")
	event, err := reader.ReadEvent()
	if err != nil {
		t.Fatalf("ReadEvent error: %v", err)
	}

	keyEvent, ok := event.(KeyEvent)
	if !ok {
		t.Fatal("ReadEvent did not return KeyEvent")
	}

	if keyEvent.Key != KeyEscape {
		t.Errorf("ReadEvent Key = %v, expected %v", keyEvent.Key, KeyEscape)
	}
}

func TestKeyEventAltModifier(t *testing.T) {
	// Alt+a (ESC followed by 'a')
	reader := newTestEventReader("\x1ba")
	event, err := reader.ReadEvent()
	if err != nil {
		t.Fatalf("ReadEvent error: %v", err)
	}

	keyEvent, ok := event.(KeyEvent)
	if !ok {
		t.Fatal("ReadEvent did not return KeyEvent")
	}

	if !keyEvent.Alt {
		t.Error("Expected Alt modifier to be true")
	}
	if keyEvent.Rune != 'a' {
		t.Errorf("Expected rune 'a', got %c", keyEvent.Rune)
	}
}

func TestKeyString(t *testing.T) {
	tests := []struct {
		key      Key
		expected string
	}{
		{KeyNone, "None"},
		{KeyEnter, "Enter"},
		{KeyEscape, "Escape"},
		{KeyUp, "Up"},
		{KeyDown, "Down"},
		{KeyF1, "F1"},
		{KeyCtrlC, "Ctrl+C"},
	}

	for _, tt := range tests {
		result := tt.key.String()
		if result != tt.expected {
			t.Errorf("Key(%v).String() = %q, expected %q", tt.key, result, tt.expected)
		}
	}
}

func TestEventType(t *testing.T) {
	keyEvent := KeyEvent{Key: KeyEnter}
	if keyEvent.Type() != EventTypeKey {
		t.Errorf("KeyEvent.Type() = %v, expected %v", keyEvent.Type(), EventTypeKey)
	}

	resizeEvent := ResizeEvent{Width: 80, Height: 24}
	if resizeEvent.Type() != EventTypeResize {
		t.Errorf("ResizeEvent.Type() = %v, expected %v", resizeEvent.Type(), EventTypeResize)
	}
}

func TestStartEventChannel(t *testing.T) {
	// Create test input with multiple events
	r, w, _ := os.Pipe()
	reader := NewEventReaderWithInput(r)

	eventChan, cleanup := StartEventChannel(reader)
	defer cleanup()

	// Write test events
	go func() {
		w.Write([]byte("a\x1b[A\n"))
		w.Close()
	}()

	// Read events from channel
	var events []Event
	for event := range eventChan {
		events = append(events, event)
		if len(events) >= 3 {
			break
		}
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Verify first event is 'a'
	if ke, ok := events[0].(KeyEvent); ok {
		if ke.Rune != 'a' {
			t.Errorf("First event rune = %c, expected 'a'", ke.Rune)
		}
	} else {
		t.Error("First event is not KeyEvent")
	}

	// Verify second event is Up arrow
	if ke, ok := events[1].(KeyEvent); ok {
		if ke.Key != KeyUp {
			t.Errorf("Second event key = %v, expected KeyUp", ke.Key)
		}
	} else {
		t.Error("Second event is not KeyEvent")
	}

	// Verify third event is Enter
	if ke, ok := events[2].(KeyEvent); ok {
		if ke.Key != KeyEnter {
			t.Errorf("Third event key = %v, expected KeyEnter", ke.Key)
		}
	} else {
		t.Error("Third event is not KeyEvent")
	}
}
