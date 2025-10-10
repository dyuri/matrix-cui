# Matrix CUI - Project Context for Claude

## Project Overview

Matrix CUI is a Go library that provides a simplified API for creating character-based terminal user interfaces. Instead of dealing with ANSI escape sequences directly, developers work with an NxM matrix of cells.

### Goals

1. **Experiment with API design**: This is the first iteration to test the matrix-based API concept
2. **Emulation layer**: Currently emulates the matrix API using ANSI terminal output via Lipgloss
3. **Replace classic TUIs**: Eventually replace traditional TUI implementations that require manual ANSI sequence management
4. **Future-proof**: Design an API that can later support different rendering backends

### Design Philosophy

- **Simplicity first**: API should be intuitive - `Put(x, y, content)` instead of cursor movement
- **Cell-centric**: Each cell is independent with its own character, colors, and styles
- **Immutable patterns**: Cell methods return new cells rather than modifying in place
- **Bounds-safe**: Operations on out-of-bounds coordinates fail gracefully

## Architecture

### Core Components

#### Rendering Layer

1. **cell.go**: Defines `Cell` type and styling
   - `Cell` struct: character + foreground + background + style flags
   - Style constants: Bold, Italic, Underline, Blink, Reverse
   - Helper constructors and immutable modifier methods

2. **matrix.go**: The main `Matrix` type
   - 2D array of cells with width/height
   - CRUD operations: Put, Get, Clear, Fill, Resize
   - Utility: InBounds checking, PutString for text
   - Auto-sizing from terminal dimensions

3. **render.go**: ANSI rendering/emulation layer
   - Converts matrix to ANSI string using Lipgloss
   - Supports partial rendering (regions, rows, columns)
   - Applies all cell styling through Lipgloss style builder
   - Uses `\r\n` for raw mode terminal compatibility

#### Event Handling Layer

4. **event.go**: Event type system
   - `Event` interface with type checking
   - `KeyEvent`: keyboard input with modifiers (Alt, Ctrl, Shift)
   - `MouseEvent`: mouse input with position, button, action, modifiers
   - `ResizeEvent`: terminal size changes
   - Key/Button/Action enums with String() methods

5. **terminal.go**: Terminal state management
   - `Terminal` type for fullscreen TUI setup
   - Raw mode: disables line buffering, enables character-by-character input
   - Alternate screen buffer: fullscreen without affecting scrollback
   - Mouse tracking: SGR extended mode (1000h, 1002h, 1006h)
   - Cursor hiding/showing
   - Signal handling for clean Ctrl+C exit

6. **input.go**: Event reader and ANSI parser
   - `EventReader`: reads from stdin and parses events
   - ANSI escape sequence parsing for keyboard and mouse
   - CSI sequence handler for arrow keys, function keys, navigation
   - SGR mouse event parser (extended format with modifiers)
   - Terminal resize signal handling (SIGWINCH)
   - Channel-based event loop via `StartEventChannel()`
   - ESC key timeout (50ms) to distinguish standalone ESC from escape sequences

#### Network Protocol Layer

7. **protocol/messages.go**: Protocol message types
   - JSON-based message envelope system
   - Command, Response, and Event payload types
   - Helper functions for wrapping/unwrapping messages
   - CellProto for JSON-serializable cell representation
   - Request-response matching via UUIDs

8. **protocol/cell.go**: Cell serialization
   - Conversion between library Cell and protocol CellProto
   - MatrixToProto for full matrix snapshots
   - Handles color and style serialization

9. **server/server.go**: Server implementation
   - Manages Matrix instance and Terminal
   - Accepts client connections (Unix socket, TCP)
   - Command execution (Put, Get, Clear, Fill, etc.)
   - Event broadcasting to subscribed clients
   - Concurrent client handling with goroutines

10. **server/session.go**: Client session management
    - Per-client connection state
    - Subscribe/unsubscribe to events
    - Event filtering by type (key, mouse, resize)
    - Request/response handling

11. **client/remote_matrix.go**: Remote matrix client
    - RemoteMatrix type implementing Matrix interface
    - Connection management (Unix socket, TCP, WebSocket)
    - Synchronous command execution with response waiting
    - Event subscription and channel-based event reception
    - Transparent API matching local Matrix

#### Examples

12. **examples/simple-demo/**: Basic matrix API demo
    - Shows basic text, boxes, colors, styling
    - Includes simple animation example

13. **examples/matrix-rain/**: Interactive animation with keyboard input
    - Matrix digital rain effect
    - ESC/Q to quit, +/- for speed, Space to pause
    - Terminal resize handling, FPS counter

14. **examples/interactive-paint/**: Mouse-driven drawing
    - Click and drag painting
    - Color palette (9 colors via number keys)
    - Demonstrates mouse tracking and drag events

15. **examples/key-test/**: Event debugging tool
    - Shows all keyboard and mouse events
    - Useful for testing input parsing

16. **examples/matrix-server/**: Standalone network server
    - Fullscreen TUI server listening on Unix socket
    - Forwards keyboard/mouse events to clients
    - Renders matrix to terminal at ~60 FPS
    - Press 'q' or Ctrl+C to quit

17. **examples/remote-paint/**: Remote client demo
    - Connects to matrix-server via Unix socket
    - Same functionality as interactive-paint
    - Demonstrates transparent local vs remote API

### Dependencies

- **Lipgloss** (`github.com/charmbracelet/lipgloss`): Color and ANSI styling
  - Chosen for: Lightweight, well-maintained, great API, Charm ecosystem
  - Used in: Cell rendering, color definitions

- **x/term** (`github.com/charmbracelet/x/term`): Terminal size detection
  - Used in: `NewMatrixAuto()` for auto-sizing

- **UUID** (`github.com/google/uuid`): UUID generation
  - Used in: Protocol request-response matching, session IDs

### Key Design Decisions

1. **Lipgloss over Bubbletea**: Bubbletea is a full framework with event loops - too heavyweight for just rendering
2. **Color type**: Uses `lipgloss.Color` directly for maximum flexibility (256-color, hex, etc.)
3. **Coordinate system**: (0,0) is top-left, x is horizontal, y is vertical
4. **Error handling**: Out-of-bounds operations return false/empty rather than panicking (except constructors)
5. **Immutable cells**: Cell modification methods return new cells to avoid shared state bugs
6. **Channel-based events**: Go-idiomatic event loop, not Elm architecture like Bubbletea
7. **SGR mouse mode**: Extended format for better coordinate support and modifiers
8. **Raw mode rendering**: Uses `\r\n` instead of `\n` for proper line wrapping
9. **Server-side rendering**: Network protocol uses server-side rendering, not VNC-style state sync
10. **JSON protocol**: Human-readable JSON initially, can optimize to binary (MessagePack) later
11. **Transport-agnostic**: Same protocol works over Unix socket, TCP, and WebSocket

## Common Tasks

### Adding New Cell Styles

1. Add constant to `CellStyle` type in cell.go
2. Update `HasStyle()` and `AddStyle()` methods if needed
3. Add rendering logic in `renderCell()` in render.go
4. Update documentation in README.md

### Adding New Matrix Operations

1. Add method to Matrix type in matrix.go
2. Ensure bounds checking for safety
3. Add corresponding test in matrix_test.go
4. Document in README.md

### Adding New Rendering Modes

1. Add method to render.go
2. Use `renderCell()` for consistent styling
3. Consider performance for large matrices

### Adding New Event Types

1. Add event struct to event.go implementing `Event` interface
2. Add event type constant to `EventType` enum
3. Add parsing logic to input.go's `readKeyEvent()` or `parseCSICommand()`
4. Add unit tests to event_test.go
5. Update README.md with usage examples

### Working with Terminal Modes

- **Raw mode** is required for character-by-character input (no line buffering)
- **Alternate screen** prevents polluting terminal scrollback
- **Mouse tracking** requires both terminal setup and escape sequence parsing
- Always use `defer term.Close()` to ensure cleanup
- `SetupCleanupOnSignal()` handles Ctrl+C gracefully

## Testing

- Unit tests in matrix_test.go and event_test.go
- Test coverage priorities:
  1. Bounds checking (most critical for safety)
  2. Cell operations (immutability, style flags)
  3. Matrix operations (Put, Get, Resize)
  4. Rendering (correctness of ANSI output)
  5. Event parsing (keyboard, mouse, escape sequences)

### Testing Event Parsing

- Use pipe-based test helpers: `newTestEventReader(input string)`
- Test both basic input (characters) and escape sequences
- Mouse events use SGR format: `\x1b[<Cb;Cx;CyM` (press) or `m` (release)
- Arrow keys: `\x1b[A` (up), `\x1b[B` (down), etc.
- Function keys: `\x1bOP` (F1 SS3 format) or `\x1b[15~` (F5 CSI format)

## Future Considerations

### Completed Features

1. ✅ **Matrix API**: Simple 2D cell-based interface
2. ✅ **ANSI Rendering**: Lipgloss-based with raw mode support
3. ✅ **Event Handling**: Keyboard and mouse via channels
4. ✅ **Terminal Management**: Raw mode, alternate screen, mouse tracking
5. ✅ **Network Protocol**: Remote access via Unix sockets and TCP (Phase 1)

### Planned Features

1. **Native backends**: Skip ANSI emulation entirely
2. **Layout engine**: Containers, flexbox-like layouts
3. **Diff rendering**: Only update changed cells
4. **Double buffering**: Smooth animations
5. **Wide char support**: Better Unicode/emoji handling
6. **Bracketed paste**: Large text paste handling
7. **Focus events**: Terminal focus in/out detection

### API Stability

This is v0 - API may change based on usage feedback. Once we have real-world usage in the dependent projects, we'll stabilize to v1.

### Performance Notes

- Current implementation is not optimized - acceptable for initial experimentation
- For large matrices, consider lazy rendering or dirty tracking
- String builder pre-allocation helps but could be tuned

## Code Style

- Standard Go formatting (gofmt)
- Exported types/functions have doc comments
- Error returns for fallible operations (except constructors which panic)
- Prefer composition over inheritance
- Keep dependencies minimal

## Related Projects

Projects that will eventually use Matrix CUI:
- (List user's projects here as they adopt the library)

## Development Workflow

1. Make changes to library code
2. Run `go test ./...` to verify all tests pass
3. Test with examples:
   - `cd examples/simple-demo && go run demo.go`
   - `cd examples/matrix-rain && go run main.go`
   - `cd examples/interactive-paint && go run main.go`
   - `cd examples/key-test && go run main.go` (for event debugging)
   - `cd examples/matrix-server && go run main.go` (for network server)
   - `cd examples/remote-paint && go run main.go` (for remote client, requires matrix-server running)
4. Update README.md if API changed
5. Update this file if architecture changed
6. Update PROTOCOL.md if network protocol changed

## Event Handling Architecture

### Event Flow

1. **Terminal Setup**: `Terminal.SetupFullscreen()` enables raw mode + alt screen
2. **Event Reader**: `EventReader` reads from stdin in raw mode
3. **ANSI Parsing**: Escape sequences converted to Event structs
4. **Channel Distribution**: `StartEventChannel()` runs goroutine sending events to channel
5. **User Loop**: Application reads from event channel and handles events

### ANSI Escape Sequences

**Keyboard:**
- Regular chars: sent as-is (1 byte)
- Control chars: ASCII codes 0-31 (e.g., Ctrl+C = 3)
- Arrow keys: `ESC [ A/B/C/D` (CSI format)
- Function keys: `ESC O P/Q/R/S` (SS3) or `ESC [ 15~` (CSI)
- ESC key: Single `\x1b` with 50ms timeout to detect standalone press

**Mouse (SGR Extended Format):**
- Press: `ESC [ < Cb ; Cx ; Cy M`
- Release: `ESC [ < Cb ; Cx ; Cy m`
- Cb = button code (bits: 0-1=button, 2=shift, 3=alt, 4=ctrl, 5=motion, 6-7=wheel)
- Cx, Cy = 1-based coordinates (converted to 0-based)

**Terminal Resize:**
- SIGWINCH signal caught by goroutine
- Queries terminal size via `term.GetSize()`
- Sends `ResizeEvent` to channel

### Critical Implementation Details

1. **Raw mode newlines**: Must use `\r\n` not `\n` in Render()
2. **ESC timeout**: Standalone ESC needs 50ms timeout to distinguish from sequences
3. **Mouse tracking codes**: Use SGR extended (1006h) for better coordinate range
4. **Coordinate systems**: Mouse events use 0-based, terminal uses 1-based
5. **Event channel buffering**: 10-event buffer prevents blocking on slow handlers
6. **Cleanup order**: Mouse → Alt screen → Raw mode → Show cursor

## Network Protocol Architecture

### Protocol Design

**Server-Side Rendering Model:**
- Server owns the Terminal and Matrix instance
- Clients send commands to modify matrix state
- Server broadcasts events to subscribed clients
- Rendering happens only on the server

**Message Types:**
1. **Command** (Client → Server): Matrix operations (put, get, clear, etc.)
2. **Response** (Server → Client): Success/failure with optional data
3. **Event** (Server → Clients): Terminal events (key, mouse, resize)

**Message Envelope:**
```json
{
  "type": "command|response|event",
  "id": "uuid-for-request-response",
  "payload": { ... }
}
```

### Transport Layer

**Wire Format:**
- JSON objects, one per line, separated by `\n`
- Newline-delimited JSON (NDJSON)
- Human-readable for debugging
- Easy to parse and stream

**Supported Transports:**
- Unix sockets: `unix:///path/to/socket` (best for local IPC)
- TCP: `tcp://host:port` (network access)
- WebSocket: `ws://host:port/path` (future, for browsers)

### Client-Server Interaction

**Connection Lifecycle:**
1. Client connects to server
2. Client sends commands (Put, Get, Clear, etc.)
3. Server executes commands, sends responses
4. Client optionally subscribes to events
5. Server broadcasts events to subscribed clients
6. Client disconnects when done

**Request-Response Pattern:**
- Client generates UUID for each command
- Server matches response to request via UUID
- Synchronous from client perspective
- Async handling on server

**Event Broadcasting:**
- Server maintains list of subscribed sessions
- Events forwarded to all subscribed clients
- Client-specific event filtering (key, mouse, resize)
- Non-blocking sends (full channel = dropped event)

### Concurrency Model

**Server:**
- Main goroutine: Accept loop for new connections
- Per-session goroutine: Handle client messages
- Event forwarding goroutine: Broadcast terminal events
- Mutex-protected session registry

**Client:**
- Main goroutine: Application logic
- Receive loop goroutine: Read messages from server
- Pending requests map: Match responses to commands

### Shared State

**Matrix is shared:**
- All clients see the same matrix
- Changes by one client visible to all
- No client isolation (Phase 1)
- Commands executed sequentially

**Future considerations:**
- Session-specific matrices
- Access control and permissions
- Command batching for efficiency

### Protocol Extensions

**Phase 1 (Current):**
- Basic commands (put, get, clear, fill, etc.)
- Event subscription (key, mouse, resize)
- Unix socket and TCP transports

**Phase 2 (Planned):**
- Authentication tokens
- Session management
- Connection keep-alive

**Phase 3 (Future):**
- WebSocket transport
- Browser client with Canvas rendering
- Delta encoding (only send changed cells)
- Compression (gzip, msgpack)

### Testing Network Components

**Unit Tests:**
- Protocol serialization (protocol_test.go)
- Message envelope wrapping/unwrapping
- Cell conversion (Cell ↔ CellProto)

**Integration Tests:**
- Server + client over real Unix socket
- Multiple simultaneous clients
- Event broadcasting
- Connection lifecycle

**Manual Testing:**
- Run matrix-server example
- Connect with remote-paint client
- Verify commands execute correctly
- Verify events forwarded properly

## Questions for Future

- Should we support z-index/layers?
- Do we need a cursor abstraction?
- Should cells support multiple characters (grapheme clusters)?
- Animation API - should it be built-in or separate?
- Terminal capability detection - how much to support?
- Should we add a Component abstraction for reusable UI elements?
- How to handle Unicode width (wide characters, emoji)?
- Should we support sixel graphics for images?

## Troubleshooting

### Common Issues

**Mouse events not working:**
- Ensure `EnableMouseTracking()` is called after `SetupFullscreen()`
- Terminal must support mouse tracking (most modern terminals do)
- Some terminals (tmux) may need specific configuration

**ESC key not detected:**
- 50ms timeout distinguishes ESC from escape sequences
- Fast typists might trigger sequences - adjust timeout if needed

**Rendering artifacts:**
- Ensure using `\r\n` for newlines in raw mode
- Check terminal size matches matrix dimensions
- Try clearing screen before first render

**Terminal not restored on exit:**
- Always use `defer term.Close()`
- Use `SetupCleanupOnSignal()` for Ctrl+C handling
- Consider panic recovery to ensure cleanup

### Performance

- Current implementation renders entire matrix each frame
- For 80x24 terminal ~= 2K cells, easily handles 60+ FPS
- Larger terminals or slower systems may need optimization:
  - Implement dirty cell tracking
  - Only render changed regions
  - Use double buffering to reduce flicker
