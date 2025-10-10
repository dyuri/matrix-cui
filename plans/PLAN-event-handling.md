# Event Handling Implementation Plan

## 1. Core Event System (`event.go`)

**Event Type Definitions:**
- `Event` interface with `Type() EventType` method
- `EventType` enum (KeyPress, MousePress, MouseRelease, MouseMove, WindowResize, etc.)
- `KeyEvent` struct: Key, Rune, Modifiers (Alt, Ctrl, Shift, Meta)
- `MouseEvent` struct: X, Y, Button, Action, Modifiers
- `ResizeEvent` struct: Width, Height
- Common key constants (Enter, Escape, Arrow keys, F1-F12, etc.)

## 2. Event Reader (`input.go`)

**Terminal Input Handler:**
- `EventReader` struct managing raw terminal mode
- `NewEventReader()` - initializes stdin in raw mode, stores original state
- `ReadEvent() (Event, error)` - blocking read of next event
- `Close()` - restores terminal to original state
- Parse ANSI escape sequences for special keys and mouse events
- Support standard input sequences: arrows, function keys, home/end, etc.

**Using Charm's existing libraries:**
- Leverage `github.com/charmbracelet/x/term` for raw mode
- Leverage `github.com/charmbracelet/x/ansi` for escape sequence parsing
- Mouse tracking via CSI sequences (if terminal supports it)

## 3. Event Loop Patterns (`loop.go`)

**Channel-based approach (Go idiomatic):**
```go
func (m *Matrix) StartEventLoop() (<-chan Event, func())
```
- Returns read-only channel and cleanup function
- User reads events in their own goroutine
- Gives full control to the user

## 4. Terminal Management (`terminal.go`)

**Terminal State Handler:**
- `Terminal` struct tracking raw mode state
- Enable/disable raw mode with proper restoration
- Enable/disable mouse tracking (optional)
- Handle alternate screen buffer (optional)
- Proper cleanup on panic/signals (defer and signal handling)

## 5. File Structure

```
event.go        - Event types and interfaces
input.go        - EventReader and ANSI parsing
loop.go         - Event loop helpers
terminal.go     - Terminal mode management
event_test.go   - Event parsing tests
```

## 6. Examples

### `examples/input-demo/main.go`
- Show channel-based pattern
- Display keypresses on screen at cursor position
- Show mouse clicks as colored cells
- Display terminal resize events
- Clean exit on Ctrl+C or 'q'

### `examples/interactive-paint/main.go`
- Mouse-based drawing tool
- Click to paint cells
- Keyboard shortcuts for colors
- Demonstrates practical event usage with Matrix API

## 7. API Design Principles

**Keep it simple and unopinionated:**
- Don't force Elm architecture (unlike Bubbletea)
- Provide primitives, let users build patterns
- Zero breaking changes to existing Matrix API
- Event system is opt-in via separate imports

**Thread safety:**
- EventReader is NOT thread-safe (document this)
- Users manage concurrency with channels if needed
- Matrix operations remain non-concurrent

## 8. Testing Strategy

- Unit tests for ANSI escape sequence parsing
- Mock input streams for event reader tests
- Integration tests for common key sequences
- Document terminal compatibility (xterm, vt100, etc.)
- Test mouse event encoding/decoding

## 9. Documentation Updates

- Add event handling section to README.md
- Update CLAUDE.md with event architecture
- Document terminal compatibility matrix
- Add migration guide for Bubbletea users
- Explain channel-based pattern usage

## 10. Future Enhancements (out of scope)

- Bracketed paste support
- IME/Unicode input handling
- Terminal capability detection
- Focus events
- Clipboard integration (via OSC 52)

## Implementation Approach

This approach provides a clean, minimal event system that fits matrix-cui's philosophy while leveraging Charm's proven low-level libraries. The channel-based pattern is idiomatic Go and gives users maximum flexibility in how they structure their event loops.
