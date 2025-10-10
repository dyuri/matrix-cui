# Matrix CUI Network Protocol

This document describes the network protocol used for remote access to Matrix CUI servers.

## Overview

The Matrix CUI protocol enables remote clients to control a terminal-based matrix display over Unix sockets, TCP, or WebSocket connections. The protocol uses JSON-based message envelopes with three message types:

- **Commands**: Requests from client to server
- **Responses**: Replies from server to client
- **Events**: Asynchronous notifications from server to client

## Transport

### Supported Transports

1. **Unix Sockets** - Local IPC, best performance
   - Address format: `unix:///path/to/socket`
   - Example: `unix:///tmp/matrix-cui.sock`

2. **TCP** - Network connections
   - Address format: `tcp://host:port`
   - Example: `tcp://localhost:8080`

3. **WebSocket** - Browser-compatible (future)
   - Address format: `ws://host:port/path`
   - Example: `ws://localhost:8080/matrix`

### Wire Format

Messages are JSON objects, one per line, separated by `\n`:

```
{"type":"command","id":"uuid","payload":{...}}\n
{"type":"response","id":"uuid","payload":{...}}\n
{"type":"event","payload":{...}}\n
```

## Message Envelope

All messages follow this envelope structure:

```json
{
  "type": "command|response|event",
  "id": "uuid-for-request-response-matching",
  "payload": { ... }
}
```

**Fields:**
- `type` (string): Message category - `"command"`, `"response"`, or `"event"`
- `id` (string, optional): UUID for matching requests/responses. Present for commands and responses, omitted for events.
- `payload` (object): Message-specific data

## Cell Format

Cells are represented as JSON objects:

```json
{
  "char": "A",
  "fg": "#00FF00",
  "bg": "#000000",
  "style": 1
}
```

**Fields:**
- `char` (string): Single character (rune as string)
- `fg` (string): Foreground color - hex (`"#RRGGBB"`), 256-color number, or named color
- `bg` (string): Background color - same format as `fg`
- `style` (integer): Bitfield for text styles:
  - `1` = Bold
  - `2` = Italic
  - `4` = Underline
  - `8` = Blink
  - `16` = Reverse

**Example - Bold underlined green text on black:**
```json
{
  "char": "A",
  "fg": "#00FF00",
  "bg": "#000000",
  "style": 5
}
```

## Commands

Commands are requests from client to server. Each command receives a response with the same `id`.

### Matrix Operations

#### Put

Set a cell at coordinates.

**Request:**
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

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
    "ok": true
  }
}
```

#### Get

Query a cell at coordinates.

**Request:**
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

**Response:**
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

#### Clear

Reset all cells to empty.

**Request:**
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "payload": {
    "cmd": "clear"
  }
}
```

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "payload": {
    "ok": true
  }
}
```

#### Fill

Fill entire matrix with a cell.

**Request:**
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

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440003",
  "payload": {
    "ok": true
  }
}
```

#### PutString

Write a string at coordinates with styling.

**Request:**
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

Note: The `cell.char` field is ignored for `putString`. The `cell` provides styling (fg, bg, style) for all characters in the string.

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440004",
  "payload": {
    "ok": true
  }
}
```

#### Resize

Change matrix dimensions.

**Request:**
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

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "payload": {
    "ok": true
  }
}
```

### Session Management

#### GetSize

Query current matrix dimensions.

**Request:**
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440008",
  "payload": {
    "cmd": "getSize"
  }
}
```

**Response:**
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

#### GetSnapshot

Retrieve full matrix state (useful for initial render or reconnection).

**Request:**
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440009",
  "payload": {
    "cmd": "getSnapshot"
  }
}
```

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440009",
  "payload": {
    "ok": true,
    "width": 80,
    "height": 24,
    "cells": [
      [ {"char": "A", "fg": "#FFF", "bg": "", "style": 0}, ... ],
      [ {"char": "B", "fg": "#FFF", "bg": "", "style": 0}, ... ],
      ...
    ]
  }
}
```

The `cells` array is indexed as `cells[y][x]` (row-major order).

### Event Subscription

#### Subscribe

Start receiving events from the server.

**Request:**
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

The `events` array specifies which event types to receive:
- `"key"` - Keyboard events
- `"mouse"` - Mouse events
- `"resize"` - Terminal resize events

If `events` is empty or omitted, all event types are subscribed.

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440006",
  "payload": {
    "ok": true
  }
}
```

#### Unsubscribe

Stop receiving events from the server.

**Request:**
```json
{
  "type": "command",
  "id": "550e8400-e29b-41d4-a716-446655440007",
  "payload": {
    "cmd": "unsubscribe"
  }
}
```

**Response:**
```json
{
  "type": "response",
  "id": "550e8400-e29b-41d4-a716-446655440007",
  "payload": {
    "ok": true
  }
}
```

## Events

Events are broadcast from server to subscribed clients. They have no `id` field and expect no response.

### KeyEvent

Keyboard input event.

```json
{
  "type": "event",
  "payload": {
    "event": "key",
    "key": "Enter",
    "rune": 13,
    "alt": false,
    "ctrl": false,
    "shift": false
  }
}
```

**Fields:**
- `key` (string): Special key name (see Key Names below), or empty string for printable characters
- `rune` (integer): Character code point, or 0 for special keys
- `alt` (boolean): Alt modifier pressed
- `ctrl` (boolean): Ctrl modifier pressed
- `shift` (boolean): Shift modifier pressed

**Key Names:**
- `"Enter"`, `"Backspace"`, `"Tab"`, `"Escape"`, `"Space"`
- `"Up"`, `"Down"`, `"Left"`, `"Right"`
- `"Home"`, `"End"`, `"PageUp"`, `"PageDown"`
- `"Insert"`, `"Delete"`
- `"F1"` through `"F12"`
- `"Ctrl+C"`, `"Ctrl+D"`, `"Ctrl+Z"`

**Examples:**

Letter 'a':
```json
{
  "type": "event",
  "payload": {
    "event": "key",
    "key": "",
    "rune": 97,
    "alt": false,
    "ctrl": false,
    "shift": false
  }
}
```

Escape key:
```json
{
  "type": "event",
  "payload": {
    "event": "key",
    "key": "Escape",
    "rune": 0,
    "alt": false,
    "ctrl": false,
    "shift": false
  }
}
```

Ctrl+C:
```json
{
  "type": "event",
  "payload": {
    "event": "key",
    "key": "Ctrl+C",
    "rune": 3,
    "alt": false,
    "ctrl": true,
    "shift": false
  }
}
```

### MouseEvent

Mouse input event.

```json
{
  "type": "event",
  "payload": {
    "event": "mouse",
    "x": 10,
    "y": 5,
    "button": "Left",
    "action": "Press",
    "alt": false,
    "ctrl": false,
    "shift": false
  }
}
```

**Fields:**
- `x` (integer): Column position (0-based)
- `y` (integer): Row position (0-based)
- `button` (string): Button identifier (see Button Names below)
- `action` (string): Mouse action (see Action Names below)
- `alt` (boolean): Alt modifier pressed
- `ctrl` (boolean): Ctrl modifier pressed
- `shift` (boolean): Shift modifier pressed

**Button Names:**
- `"None"` - No button (for move events)
- `"Left"` - Left mouse button
- `"Middle"` - Middle mouse button
- `"Right"` - Right mouse button
- `"WheelUp"` - Mouse wheel scroll up
- `"WheelDown"` - Mouse wheel scroll down

**Action Names:**
- `"Press"` - Button pressed
- `"Release"` - Button released
- `"Move"` - Mouse moved (with or without button held)

### ResizeEvent

Terminal resize event.

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

**Fields:**
- `width` (integer): New terminal width in columns
- `height` (integer): New terminal height in rows

## Error Responses

When a command fails, the response has `ok: false` and an `error` field:

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

## Client Implementation Guide

### Connection Lifecycle

1. **Connect**: Establish socket connection
2. **GetSize**: Query initial dimensions (optional)
3. **GetSnapshot**: Get initial state (optional)
4. **Subscribe**: Start receiving events (optional)
5. **Commands**: Send Put/Get/Clear/etc. commands
6. **Events**: Receive and handle events
7. **Unsubscribe**: Stop events (optional)
8. **Disconnect**: Close connection

### Request-Response Pattern

1. Generate a unique UUID for the command
2. Wrap command in envelope with UUID
3. Send envelope as JSON + newline
4. Wait for response with matching UUID

### Event Handling

1. Subscribe to desired event types
2. Read messages continuously
3. Filter by `type: "event"`
4. Handle events asynchronously

### Example Session

```
Client: {"type":"command","id":"1","payload":{"cmd":"getSize"}}\n
Server: {"type":"response","id":"1","payload":{"ok":true,"width":80,"height":24}}\n

Client: {"type":"command","id":"2","payload":{"cmd":"subscribe","events":["key"]}}\n
Server: {"type":"response","id":"2","payload":{"ok":true}}\n

Client: {"type":"command","id":"3","payload":{"cmd":"put","x":10,"y":5,"cell":{"char":"A","fg":"#FF0000","bg":"","style":0}}}\n
Server: {"type":"response","id":"3","payload":{"ok":true}}\n

[User presses Enter key on server terminal]
Server: {"type":"event","payload":{"event":"key","key":"Enter","rune":13,"alt":false,"ctrl":false,"shift":false}}\n

Client: {"type":"command","id":"4","payload":{"cmd":"unsubscribe"}}\n
Server: {"type":"response","id":"4","payload":{"ok":true}}\n
```

## Server Implementation Notes

### Rendering

The server renders the matrix to its local terminal. Clients don't receive rendered output - they only send commands to modify the matrix state.

### Shared State

All clients share the same matrix. Changes made by one client are visible to all other clients.

### Event Broadcasting

Events are broadcast to all subscribed clients. The server doesn't track which client caused a change.

### Concurrency

The server handles multiple concurrent clients. Commands are processed sequentially to maintain consistency.

## Future Extensions

### Planned Features (Phase 3+)

- **Authentication**: Token-based client authentication
- **Session Isolation**: Separate matrices per client
- **Delta Updates**: Only send changed cells
- **Compression**: gzip for WebSocket, msgpack for binary protocol
- **Reconnection**: State recovery after disconnect
- **Multiplexer**: Window/pane management

### Protocol Versioning

The protocol may evolve. Future versions will include a version negotiation step:

```json
{
  "type": "command",
  "id": "...",
  "payload": {
    "cmd": "hello",
    "version": "1.0",
    "capabilities": ["events", "compression"]
  }
}
```

## Security Considerations

### Unix Sockets

- File permissions control access
- Local-only, no network exposure
- Consider socket file location carefully

### TCP/WebSocket

- No authentication in Phase 1 (local testing only)
- Phase 3 will add:
  - Token-based authentication
  - TLS/WSS encryption
  - CORS configuration
  - Rate limiting

## Performance

### Current Implementation

- Full JSON encoding/decoding per message
- No compression
- Synchronous command processing
- Acceptable for local usage and prototyping

### Future Optimizations

- Binary protocol (MessagePack)
- Command batching
- Delta encoding for large matrices
- Client-side caching

## References

- [JSON Specification](https://www.json.org/)
- [WebSocket Protocol (RFC 6455)](https://datatracker.ietf.org/doc/html/rfc6455)
- [UUID (RFC 4122)](https://datatracker.ietf.org/doc/html/rfc4122)
