package matrixcui

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/term"
)

// Terminal manages terminal state for fullscreen TUI applications.
type Terminal struct {
	input          *os.File
	output         *os.File
	origState      *term.State
	inRawMode      bool
	inAltScreen    bool
	mouseTracking  bool
}

// NewTerminal creates a new terminal manager.
// By default uses stdin for input and stdout for output.
func NewTerminal() (*Terminal, error) {
	return NewTerminalWithIO(os.Stdin, os.Stdout)
}

// NewTerminalWithIO creates a new terminal manager with custom I/O.
func NewTerminalWithIO(input *os.File, output *os.File) (*Terminal, error) {
	return &Terminal{
		input:  input,
		output: output,
	}, nil
}

// EnterRawMode puts the terminal into raw mode.
// Returns an error if the terminal is already in raw mode or if setting raw mode fails.
func (t *Terminal) EnterRawMode() error {
	if t.inRawMode {
		return fmt.Errorf("terminal already in raw mode")
	}

	// Save the current terminal state
	state, err := term.MakeRaw(t.input.Fd())
	if err != nil {
		return fmt.Errorf("failed to enter raw mode: %w", err)
	}

	t.origState = state
	t.inRawMode = true
	return nil
}

// ExitRawMode restores the terminal to its original state.
func (t *Terminal) ExitRawMode() error {
	if !t.inRawMode {
		return nil
	}

	if t.origState != nil {
		if err := term.Restore(t.input.Fd(), t.origState); err != nil {
			return fmt.Errorf("failed to restore terminal state: %w", err)
		}
	}

	t.inRawMode = false
	return nil
}

// EnterAltScreen switches to the alternate screen buffer.
// This allows fullscreen TUI applications without affecting the terminal scrollback.
func (t *Terminal) EnterAltScreen() error {
	if t.inAltScreen {
		return fmt.Errorf("already in alternate screen")
	}

	// ANSI escape sequence to enter alternate screen
	// \033[?1049h - Save cursor position and switch to alternate screen
	if _, err := io.WriteString(t.output, "\033[?1049h"); err != nil {
		return fmt.Errorf("failed to enter alternate screen: %w", err)
	}

	// Flush to ensure terminal processes the sequence
	t.output.Sync()

	t.inAltScreen = true
	return nil
}

// ExitAltScreen returns to the normal screen buffer.
func (t *Terminal) ExitAltScreen() error {
	if !t.inAltScreen {
		return nil
	}

	// ANSI escape sequence to exit alternate screen
	// \033[?1049l - Restore cursor position and return to normal screen
	if _, err := io.WriteString(t.output, "\033[?1049l"); err != nil {
		return fmt.Errorf("failed to exit alternate screen: %w", err)
	}

	t.inAltScreen = false
	return nil
}

// HideCursor hides the terminal cursor.
func (t *Terminal) HideCursor() error {
	_, err := io.WriteString(t.output, "\033[?25l")
	return err
}

// ShowCursor shows the terminal cursor.
func (t *Terminal) ShowCursor() error {
	_, err := io.WriteString(t.output, "\033[?25h")
	return err
}

// EnableMouseTracking enables mouse event tracking.
// This enables button press/release and motion tracking.
func (t *Terminal) EnableMouseTracking() error {
	if t.mouseTracking {
		return fmt.Errorf("mouse tracking already enabled")
	}

	// Enable mouse tracking:
	// \033[?1000h - Enable basic mouse tracking (press/release)
	// \033[?1002h - Enable button motion tracking (drag)
	// \033[?1003h - Enable all motion tracking (hover)
	// \033[?1006h - Enable SGR extended mouse mode (better coordinate support)
	sequences := "\033[?1000h\033[?1002h\033[?1006h"

	if _, err := io.WriteString(t.output, sequences); err != nil {
		return fmt.Errorf("failed to enable mouse tracking: %w", err)
	}

	t.output.Sync()
	t.mouseTracking = true
	return nil
}

// DisableMouseTracking disables mouse event tracking.
func (t *Terminal) DisableMouseTracking() error {
	if !t.mouseTracking {
		return nil
	}

	// Disable mouse tracking (reverse order of enable)
	sequences := "\033[?1006l\033[?1002l\033[?1000l"

	if _, err := io.WriteString(t.output, sequences); err != nil {
		return fmt.Errorf("failed to disable mouse tracking: %w", err)
	}

	t.mouseTracking = false
	return nil
}

// ClearScreen clears the entire screen.
func (t *Terminal) ClearScreen() error {
	_, err := io.WriteString(t.output, "\033[2J\033[H")
	return err
}

// Close restores the terminal to its original state.
// This should be called when done with the terminal (typically via defer).
func (t *Terminal) Close() error {
	var firstErr error

	// Disable mouse tracking first
	if err := t.DisableMouseTracking(); err != nil && firstErr == nil {
		firstErr = err
	}

	// Exit alternate screen
	if err := t.ExitAltScreen(); err != nil && firstErr == nil {
		firstErr = err
	}

	// Then restore raw mode
	if err := t.ExitRawMode(); err != nil && firstErr == nil {
		firstErr = err
	}

	// Show cursor
	if err := t.ShowCursor(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

// SetupFullscreen prepares the terminal for fullscreen TUI usage.
// This enters raw mode, alternate screen, hides cursor, and clears the screen.
func (t *Terminal) SetupFullscreen() error {
	if err := t.EnterRawMode(); err != nil {
		return err
	}
	if err := t.EnterAltScreen(); err != nil {
		t.ExitRawMode() // Cleanup
		return err
	}
	if err := t.HideCursor(); err != nil {
		t.Close() // Cleanup
		return err
	}
	if err := t.ClearScreen(); err != nil {
		t.Close() // Cleanup
		return err
	}
	return nil
}

// SetupCleanupOnSignal sets up automatic cleanup on interrupt signals (Ctrl+C, etc.).
// This ensures the terminal is properly restored even if the program is interrupted.
func (t *Terminal) SetupCleanupOnSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigChan
		t.Close()
		os.Exit(0)
	}()
}
