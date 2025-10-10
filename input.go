package matrixcui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/x/term"
)

// EventReader reads events from terminal input.
// It is NOT thread-safe - only one goroutine should read events.
type EventReader struct {
	input      *os.File
	reader     *bufio.Reader
	resizeChan chan ResizeEvent
	stopResize chan struct{}
}

// NewEventReader creates a new event reader using stdin.
func NewEventReader() *EventReader {
	return NewEventReaderWithInput(os.Stdin)
}

// NewEventReaderWithInput creates a new event reader with custom input.
func NewEventReaderWithInput(input *os.File) *EventReader {
	er := &EventReader{
		input:      input,
		reader:     bufio.NewReader(input),
		resizeChan: make(chan ResizeEvent, 10),
		stopResize: make(chan struct{}),
	}

	// Start resize event listener
	er.startResizeListener()

	return er
}

// ReadEvent reads the next event from the terminal.
// This is a blocking call. Returns io.EOF when input is closed.
func (er *EventReader) ReadEvent() (Event, error) {
	// Check for resize event first (non-blocking)
	select {
	case resize := <-er.resizeChan:
		return resize, nil
	default:
		// No resize event, continue to read keyboard input
	}

	// Read keyboard input
	return er.readKeyEvent()
}

// readKeyEvent reads a keyboard event from input.
func (er *EventReader) readKeyEvent() (Event, error) {
	b, err := er.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Handle escape sequences
	if b == '\033' { // ESC
		return er.readEscapeSequence()
	}

	// Handle control characters
	if b < 32 {
		return er.handleControlChar(b), nil
	}

	// Handle DEL (127)
	if b == 127 {
		return KeyEvent{Key: KeyBackspace}, nil
	}

	// Regular printable character
	return KeyEvent{
		Key:  KeyNone,
		Rune: rune(b),
	}, nil
}

// readEscapeSequence reads an ANSI escape sequence.
func (er *EventReader) readEscapeSequence() (Event, error) {
	// Use a channel to read with timeout
	type peekResult struct {
		data []byte
		err  error
	}

	peekChan := make(chan peekResult, 1)
	go func() {
		data, err := er.reader.Peek(1)
		peekChan <- peekResult{data, err}
	}()

	// Wait for peek with timeout
	select {
	case result := <-peekChan:
		if result.err != nil {
			// Error peeking, treat as standalone ESC
			return KeyEvent{Key: KeyEscape}, nil
		}

		// Check for CSI sequence (ESC [)
		if result.data[0] == '[' {
			er.reader.ReadByte() // consume '['
			return er.readCSISequence()
		}

		// Check for SS3 sequence (ESC O)
		if result.data[0] == 'O' {
			er.reader.ReadByte() // consume 'O'
			return er.readSS3Sequence()
		}

		// Alt + key combination
		altKey, err := er.reader.ReadByte()
		if err != nil {
			return KeyEvent{Key: KeyEscape}, nil
		}

		return KeyEvent{
			Key:  KeyNone,
			Rune: rune(altKey),
			Alt:  true,
		}, nil

	case <-time.After(50 * time.Millisecond):
		// Timeout - standalone ESC key
		return KeyEvent{Key: KeyEscape}, nil
	}
}

// readCSISequence reads a CSI (Control Sequence Introducer) sequence.
// Format: ESC [ params letter
func (er *EventReader) readCSISequence() (Event, error) {
	// Read until we hit a letter or tilde (the command)
	var params []byte
	for {
		b, err := er.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		// Check if it's a letter or tilde (end of sequence)
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '~' {
			return er.parseCSICommand(b, params), nil
		}

		// Accumulate parameters (digits and semicolons)
		params = append(params, b)
	}
}

// parseCSICommand parses a CSI command byte with parameters.
func (er *EventReader) parseCSICommand(cmd byte, params []byte) Event {
	paramsStr := string(params)

	switch cmd {
	case 'A':
		return KeyEvent{Key: KeyUp}
	case 'B':
		return KeyEvent{Key: KeyDown}
	case 'C':
		return KeyEvent{Key: KeyRight}
	case 'D':
		return KeyEvent{Key: KeyLeft}
	case 'H':
		return KeyEvent{Key: KeyHome}
	case 'F':
		return KeyEvent{Key: KeyEnd}
	case '~':
		// Extended sequences like Delete (3~), PageUp (5~), etc.
		switch paramsStr {
		case "1":
			return KeyEvent{Key: KeyHome}
		case "2":
			return KeyEvent{Key: KeyInsert}
		case "3":
			return KeyEvent{Key: KeyDelete}
		case "4":
			return KeyEvent{Key: KeyEnd}
		case "5":
			return KeyEvent{Key: KeyPageUp}
		case "6":
			return KeyEvent{Key: KeyPageDown}
		case "11":
			return KeyEvent{Key: KeyF1}
		case "12":
			return KeyEvent{Key: KeyF2}
		case "13":
			return KeyEvent{Key: KeyF3}
		case "14":
			return KeyEvent{Key: KeyF4}
		case "15":
			return KeyEvent{Key: KeyF5}
		case "17":
			return KeyEvent{Key: KeyF6}
		case "18":
			return KeyEvent{Key: KeyF7}
		case "19":
			return KeyEvent{Key: KeyF8}
		case "20":
			return KeyEvent{Key: KeyF9}
		case "21":
			return KeyEvent{Key: KeyF10}
		case "23":
			return KeyEvent{Key: KeyF11}
		case "24":
			return KeyEvent{Key: KeyF12}
		}
	}

	// Unknown sequence - return as escape
	return KeyEvent{Key: KeyEscape}
}

// readSS3Sequence reads an SS3 (Single Shift 3) sequence.
// Format: ESC O letter
func (er *EventReader) readSS3Sequence() (Event, error) {
	b, err := er.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Function keys F1-F4 in some terminals
	switch b {
	case 'P':
		return KeyEvent{Key: KeyF1}, nil
	case 'Q':
		return KeyEvent{Key: KeyF2}, nil
	case 'R':
		return KeyEvent{Key: KeyF3}, nil
	case 'S':
		return KeyEvent{Key: KeyF4}, nil
	case 'H':
		return KeyEvent{Key: KeyHome}, nil
	case 'F':
		return KeyEvent{Key: KeyEnd}, nil
	}

	// Unknown sequence
	return KeyEvent{Key: KeyEscape}, nil
}

// handleControlChar handles control characters (Ctrl+key combinations).
func (er *EventReader) handleControlChar(b byte) Event {
	switch b {
	case 3: // Ctrl+C
		return KeyEvent{Key: KeyCtrlC, Ctrl: true}
	case 4: // Ctrl+D
		return KeyEvent{Key: KeyCtrlD, Ctrl: true}
	case 9: // Tab
		return KeyEvent{Key: KeyTab}
	case 10, 13: // Line feed or carriage return (Enter)
		return KeyEvent{Key: KeyEnter}
	case 26: // Ctrl+Z
		return KeyEvent{Key: KeyCtrlZ, Ctrl: true}
	case 27: // ESC (shouldn't reach here normally)
		return KeyEvent{Key: KeyEscape}
	default:
		// Other Ctrl+letter combinations
		if b >= 1 && b <= 26 {
			return KeyEvent{
				Key:  KeyNone,
				Rune: rune('a' + b - 1), // Map Ctrl+A=1, Ctrl+B=2, etc.
				Ctrl: true,
			}
		}
	}

	// Unknown control char
	return KeyEvent{Key: KeyNone, Rune: rune(b)}
}

// startResizeListener starts listening for terminal resize signals.
func (er *EventReader) startResizeListener() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-sigChan:
				// Get new terminal size
				width, height, err := GetTerminalSize()
				if err == nil {
					select {
					case er.resizeChan <- ResizeEvent{Width: width, Height: height}:
					default:
						// Channel full, drop event
					}
				}
			case <-er.stopResize:
				signal.Stop(sigChan)
				return
			}
		}
	}()
}

// Close stops the event reader and cleans up resources.
func (er *EventReader) Close() error {
	close(er.stopResize)
	return nil
}

// StartEventChannel starts an event loop in a goroutine and returns a channel of events.
// The cleanup function should be called to stop the event loop and close the channel.
// This is the recommended way to integrate event handling into your application.
func StartEventChannel(reader *EventReader) (<-chan Event, func()) {
	eventChan := make(chan Event, 10)
	stopChan := make(chan struct{})

	go func() {
		defer close(eventChan)
		for {
			select {
			case <-stopChan:
				return
			default:
				event, err := reader.ReadEvent()
				if err != nil {
					if err != io.EOF {
						// Send error as a special event? For now, just stop
					}
					return
				}
				select {
				case eventChan <- event:
				case <-stopChan:
					return
				}
			}
		}
	}()

	cleanup := func() {
		close(stopChan)
		reader.Close()
	}

	return eventChan, cleanup
}

// GetTerminalSize returns the current terminal dimensions.
func GetTerminalSize() (width, height int, err error) {
	// Use the existing term package
	w, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get terminal size: %w", err)
	}
	return w, h, nil
}
