# Network Protocol for Matrix CUI - Remote Access Plan

## Overview

Extend Matrix CUI to support remote access via socket-based communication, enabling:
- Local and remote Go applications to share the same API
- Browser-based clients using WebSocket
- Multiple simultaneous clients (foundation for terminal multiplexer)
- Server-side rendering with command-based protocol

## Architecture

### Chosen Model: Terminal Server with Multi-Transport Support

```
[Terminal] <-> [Server] <-> [Unix Socket] <-> [Local Go App]
                        \-> [TCP Socket]  <-> [Remote Go App]
                        \-> [WebSocket]   <-> [Browser Client]
```

**Server-side rendering**: The server owns the terminal and Matrix instance. Clients send drawing commands and receive events.

### Why This Approach?

1. **Natural API mapping**: Client API can mirror local Matrix API exactly
2. **Multiple clients**: Foundation for terminal multiplexer functionality
3. **Event routing**: Server can broadcast events or target specific clients
4. **Efficient for small updates**: Only commands sent, not full matrix state
5. **Browser compatibility**: WebSocket is the standard for browser real-time communication
6. **Aligns with project goals**: "Replace classic TUIs" - enables sharing one terminal

### Alternative Considered and Rejected

**Client-side rendering (VNC-style)**: Server sends matrix state, client renders locally.
- ❌ Less efficient for small updates (full state vs individual commands)
- ❌ Doesn't support multiple apps sharing a terminal elegantly
- ❌ Requires terminal emulation in browser

## Communication Protocol Specification

### Design Principles

1. **Transport-agnostic**: Same protocol works over Unix socket, TCP, and WebSocket
2. **JSON-based initially**: Human-readable, debuggable, widely supported
3. **Request-response pattern**: Commands get responses, events are broadcast
4. **Extensible**: Easy to add new commands and event types
5. **Future-proof**: Can optimize to binary (MessagePack) later

### Message Envelope Format

All messages follow this JSON envelope structure:

```json
{
  "type": "command|response|event",
  "id": "uuid-for-request-response-matching",
  "payload": { ... }
}
```

- **type**: Discriminator for message category
- **id**: UUID for matching requests to responses (omitted for events)
- **payload**: Message-specific data

### Commands (Client → Server)

Commands are requests from client to server. Each command receives a response.

#### Matrix Operations

**Put** - Set a cell at coordinates:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
    "cmd": "put",
    "x": 10,
    "y": 5,
    "cell": {
      "char": "A",
      "fg": "#00FF00",
      "bg": "#000000",
      "style": 1
    }
  }
}
```

**Get** - Query a cell at coordinates:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "payload": {
    "cmd": "get",
    "x": 10,
    "y": 5
  }
}
```

**Clear** - Reset all cells to empty:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "payload": {
    "cmd": "clear"
  }
}
```

**Fill** - Fill entire matrix with a cell:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "payload": {
    "cmd": "fill",
    "cell": {
      "char": " ",
      "fg": "",
      "bg": "#000000",
      "style": 0
    }
  }
}
```

**PutString** - Write a string at coordinates:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440004",
  "payload": {
    "cmd": "putString",
    "x": 0,
    "y": 0,
    "text": "Hello, World!",
    "cell": {
      "char": "",
      "fg": "#FFFFFF",
      "bg": "",
      "style": 0
    }
  }
}
```
Note: The `cell` parameter provides styling (fg, bg, style) for all characters. The `char` field is ignored.

**Resize** - Change matrix dimensions:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "payload": {
    "cmd": "resize",
    "width": 100,
    "height": 30
  }
}
```

#### Event Subscription

**Subscribe** - Start receiving events:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440006",
  "payload": {
    "cmd": "subscribe",
    "events": ["key", "mouse", "resize"]
  }
}
```

**Unsubscribe** - Stop receiving events:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440007",
  "payload": {
    "cmd": "unsubscribe"
  }
}
```

#### Session Management

**GetSize** - Query current matrix dimensions:
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440008",
  "payload": {
    "cmd": "getSize"
  }
}
```

**GetSnapshot** - Retrieve full matrix state (for initial render or reconnect):
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440009",
  "payload": {
    "cmd": "getSnapshot"
  }
}
```

### Cell Format

Cells are represented as JSON objects:

```json
{
  "char": "A",      // Single character (rune as string)
  "fg": "#00FF00",  // Foreground color (hex, 256-color, or named)
  "bg": "#000000",  // Background color (hex, 256-color, or named)
  "style": 1        // Bitfield: Bold=1, Italic=2, Underline=4, Blink=8, Reverse=16
}
```

**Empty cell**:
```json
{
  "char": " ",
  "fg": "",
  "bg": "",
  "style": 0
}
```

### Responses (Server → Client)

Responses match the `id` of the command they're responding to.

#### Success Response

Simple acknowledgment:
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
    "ok": true
  }
}
```

#### Data Responses

**Get** response:
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "payload": {
    "ok": true,
    "cell": {
      "char": "A",
      "fg": "#00FF00",
      "bg": "#000000",
      "style": 1
    }
  }
}
```

**GetSize** response:
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440008",
  "payload": {
    "ok": true,
    "width": 80,
    "height": 24
  }
}
```

**GetSnapshot** response:
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440009",
  "payload": {
    "ok": true,
    "width": 80,
    "height": 24,
    "cells": [
      [ {"char": "A", "fg": "#FFF", ...}, ... ],  // Row 0
      [ {"char": "B", "fg": "#FFF", ...}, ... ],  // Row 1
      ...
    ]
  }
}
```

#### Error Response

```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
    "ok": false,
    "error": "out of bounds: (100, 50) exceeds matrix size (80, 24)"
  }
}
```

### Events (Server → Client)

Events are broadcast to subscribed clients. No `id` field since they're not request-response.

#### KeyEvent

```json
{
  "type": "event",
  "payload": {
    "event": "key",
    "key": "Enter",      // Special key name (or "" if printable)
    "rune": 0,           // Rune code (or 0 if special key)
    "alt": false,
    "ctrl": false,
    "shift": false
  }
}
```

Examples:
```json
// Letter 'a'
{"type": "event", "payload": {"event": "key", "key": "", "rune": 97, "alt": false, "ctrl": false, "shift": false}}

// Escape key
{"type": "event", "payload": {"event": "key", "key": "Escape", "rune": 0, "alt": false, "ctrl": false, "shift": false}}

// Ctrl+C
{"type": "event", "payload": {"event": "key", "key": "Ctrl+C", "rune": 3, "alt": false, "ctrl": true, "shift": false}}
```

#### MouseEvent

```json
{
  "type": "event",
  "payload": {
    "event": "mouse",
    "x": 10,
    "y": 5,
    "button": "Left",      // "None", "Left", "Middle", "Right", "WheelUp", "WheelDown"
    "action": "Press",     // "Press", "Release", "Move"
    "alt": false,
    "ctrl": false,
    "shift": false
  }
}
```

#### ResizeEvent

```json
{
  "type": "event",
  "payload": {
    "event": "resize",
    "width": 100,
    "height": 30
  }
}
```

## Implementation Plan

### Phase 1: Core Protocol + Unix Socket

**Goals**: Establish protocol, basic server/client, Unix socket transport

**Files to create**:
- `protocol/messages.go` - Message type definitions
- `protocol/cell.go` - Cell serialization helpers
- `protocol/protocol_test.go` - Protocol serialization tests
- `server/server.go` - Server type, connection handling
- `server/session.go` - Per-client session management
- `server/transport.go` - Transport abstraction
- `server/server_test.go` - Server tests
- `client/remote_matrix.go` - RemoteMatrix type implementing Matrix interface
- `client/connection.go` - Connection management, request/response handling
- `client/client_test.go` - Client tests
- `examples/matrix-server/main.go` - Standalone server binary
- `examples/remote-paint/main.go` - Remote version of interactive-paint demo

**API Design**:

```go
// Server
server := matrixcui.NewServer(matrix)
server.ListenUnix("/tmp/matrix.sock")

// Client
m := matrixcui.NewMatrixRemote("unix:///tmp/matrix.sock")
m.Put(10, 10, cell) // Same API as local Matrix
```

**Testing strategy**:
1. Unit tests for message serialization/deserialization
2. Server tests using in-memory connections
3. Integration test: server + client over real Unix socket

### Phase 2: Event System

**Goals**: Bidirectional communication, event subscription, event routing

**Additions**:
- Subscribe/unsubscribe command handling
- Event broadcaster in server
- Event receiver in client (channel-based, matching local API)

**API Design**:

```go
// Client receives events just like local usage
reader := m.EventReader() // Returns same EventReader interface
eventChan, cleanup := matrixcui.StartEventChannel(reader)
defer cleanup()

for event := range eventChan {
    // Handle events same as local
}
```

### Phase 3: TCP + WebSocket + Browser Client

**Goals**: Network access, browser support, example web client

**Files to create**:
- `server/websocket.go` - WebSocket transport handler
- `examples/web-client/index.html` - Browser client UI
- `examples/web-client/matrix-renderer.js` - Canvas-based renderer
- `examples/web-client/protocol.js` - Protocol implementation in JS

**Server additions**:
```go
server.ListenTCP(":8080", tlsConfig) // Optional TLS
server.ListenWebSocket(":8080", "/matrix", corsConfig)
```

**Browser client**:
- WebSocket connection to server
- Canvas 2D rendering with monospace font
- Mouse event capture and forwarding
- Keyboard event capture and forwarding

**Security considerations**:
- CORS configuration for WebSocket
- Optional authentication token
- Rate limiting

### Phase 4: Advanced Features (Future)

**Goals**: Optimization, multi-session, advanced features

**Features**:
1. **Authentication**: Token-based auth, per-client permissions
2. **Multi-session**: Isolated matrices per client vs shared matrix
3. **Delta encoding**: Only send changed cells (efficiency)
4. **Compression**: gzip for WebSocket, msgpack for binary efficiency
5. **Reconnection**: Client auto-reconnect with state recovery
6. **Multiplexer features**: Windows/panes, client switching

## File Structure

```
matrix-cui/
├── protocol/
│   ├── messages.go           # Message envelope, command/response/event types
│   ├── cell.go               # Cell JSON serialization
│   └── protocol_test.go      # Serialization tests
├── server/
│   ├── server.go             # Server type, listener management
│   ├── session.go            # Client session, event subscription
│   ├── transport.go          # Transport abstraction (Unix/TCP/WebSocket)
│   ├── websocket.go          # WebSocket-specific handler (Phase 3)
│   └── server_test.go        # Server unit tests
├── client/
│   ├── remote_matrix.go      # RemoteMatrix implementing Matrix interface
│   ├── connection.go         # Connection, request/response, reconnect
│   └── client_test.go        # Client unit tests
├── examples/
│   ├── matrix-server/        # Standalone server binary
│   │   └── main.go
│   ├── remote-paint/         # Remote version of interactive-paint
│   │   └── main.go
│   └── web-client/           # Browser-based client (Phase 3)
│       ├── index.html
│       ├── matrix-renderer.js
│       └── protocol.js
├── PROTOCOL.md               # This document (protocol reference)
└── README.md                 # Updated with remote access examples
```

## Browser Client Design (Phase 3)

### Rendering Strategy

**Chosen: Canvas 2D with monospace font**

Advantages:
- Best performance for full redraws
- Pixel-perfect control over rendering
- Easy to implement cell-based rendering
- Good browser support

Implementation:
```javascript
// Canvas setup
const canvas = document.getElementById('matrix');
const ctx = canvas.getContext('2d');
const cellWidth = 8;   // Measured from font
const cellHeight = 16; // Measured from font
ctx.font = '16px "Courier New", monospace';

// Render cell
function renderCell(x, y, cell) {
    // Draw background
    ctx.fillStyle = cell.bg;
    ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);

    // Draw character
    ctx.fillStyle = cell.fg;
    ctx.fillText(cell.char, x * cellWidth, (y + 1) * cellHeight - 2);

    // Apply styles (bold, underline, etc.)
    // ...
}
```

### Event Capture

**Keyboard**:
```javascript
document.addEventListener('keydown', (e) => {
    e.preventDefault();
    sendEvent({
        event: 'key',
        key: mapKey(e.key),
        rune: e.key.length === 1 ? e.key.charCodeAt(0) : 0,
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    });
});
```

**Mouse**:
```javascript
canvas.addEventListener('mousedown', (e) => {
    const rect = canvas.getBoundingClientRect();
    const x = Math.floor((e.clientX - rect.left) / cellWidth);
    const y = Math.floor((e.clientY - rect.top) / cellHeight);

    sendEvent({
        event: 'mouse',
        x: x,
        y: y,
        button: mapButton(e.button),
        action: 'Press',
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    });
});
```

## Performance Considerations

### Initial Implementation (Phase 1-2)

- Full matrix snapshots on connect
- Individual cell updates sent as commands
- No optimization - focus on correctness and API design

### Future Optimizations (Phase 4)

1. **Delta encoding**: Track dirty cells, only send changes
2. **Batching**: Accumulate commands, send in batches
3. **Compression**: gzip WebSocket frames, msgpack for binary
4. **Caching**: Client-side cell cache, only update on change
5. **Diff rendering**: Server tracks what clients have seen

## Security Considerations

### Phase 1 (Unix Socket)

- File permissions on socket file
- Local-only access

### Phase 3 (Network/WebSocket)

- **Authentication**: Optional token-based auth
- **CORS**: Configurable allowed origins
- **Rate limiting**: Prevent DoS
- **TLS**: Optional HTTPS/WSS support
- **Input validation**: Bounds checking, sanitization

## Testing Strategy

### Unit Tests

- Protocol serialization/deserialization (protocol_test.go)
- Server command handling (server_test.go)
- Client API matching local Matrix (client_test.go)

### Integration Tests

- Server + client over Unix socket
- Multiple simultaneous clients
- Event broadcasting
- Connection lifecycle (connect, disconnect, reconnect)

### Manual Testing

- Run matrix-server example
- Connect with remote-paint example
- Verify event handling, rendering

## Migration Path for Existing Code

Existing code using local Matrix should work unchanged:

```go
// Local (existing)
m := matrixcui.NewMatrix(80, 24)
m.Put(10, 10, cell)

// Remote (new, same API)
m := matrixcui.NewMatrixRemote("unix:///tmp/matrix.sock")
m.Put(10, 10, cell) // Identical usage
```

Only initialization changes. Rest of code is identical.

## Success Criteria

### Phase 1

- [ ] RemoteMatrix implements all Matrix interface methods
- [ ] Server accepts Unix socket connections
- [ ] Commands execute and return responses
- [ ] remote-paint example works identically to local paint

### Phase 2

- [ ] Event subscription/unsubscribe works
- [ ] KeyEvents, MouseEvents, ResizeEvents forwarded to clients
- [ ] Multiple clients can subscribe independently

### Phase 3

- [ ] WebSocket transport works
- [ ] Browser client renders matrix correctly
- [ ] Browser client sends keyboard/mouse events
- [ ] Browser client is usable for interactive demos

## Future Ideas

- Terminal multiplexer (tmux-like) using this protocol
- Web-based terminal dashboard
- Record/replay sessions (protocol is already JSON)
- Remote debugging: inspect matrix state from browser
- Collaborative editing: multiple clients, one terminal
- Terminal sharing: obs-style streaming to viewers

## Open Questions

1. **Shared vs isolated matrices**: Should multiple clients share one matrix or get isolated views?
   - Initial: Shared (simpler, enables collaboration)
   - Future: Add session isolation option

2. **Synchronous vs async API**: Should `Put()` wait for server response?
   - Initial: Synchronous (simpler, matches local API)
   - Future: Add async mode with batching

3. **Error handling**: How should network errors surface in Matrix API?
   - Initial: Return false (matches bounds errors)
   - Future: Add `LastError() error` method

4. **Session persistence**: Should server persist matrix state?
   - Phase 1: No (stateful server, clients reconnect loses state)
   - Future: Optional persistence or snapshot API

5. **Binary protocol**: When to switch from JSON to MessagePack/protobuf?
   - Initial: JSON (debuggable, web-friendly)
   - Future: Negotiate protocol version, binary for Go-to-Go

## References

- [WebSocket Protocol (RFC 6455)](https://datatracker.ietf.org/doc/html/rfc6455)
- [JSON-RPC 2.0](https://www.jsonrpc.org/specification) - Similar request/response pattern
- [gorilla/websocket](https://github.com/gorilla/websocket) - Go WebSocket library
- [MessagePack](https://msgpack.org/) - Binary serialization alternative
