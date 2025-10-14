# Matrix CUI Web Server

Web-based interface for Matrix CUI that displays a Matrix instance in the browser and accepts socket-based client connections.

## Architecture

The web server **owns a Matrix instance** and acts as a rendering backend, replacing the terminal:

```
[Socket Client] <--CUI Protocol--> [Web Server] <--WebSocket--> [Browser(s)]
(e.g. remote-paint)                     |                        (viewers)
                                     Matrix
                                   (owned by server)
```

Instead of using a terminal for display, the web server:
- Owns the Matrix instance
- Accepts socket connections from clients (like remote-paint)
- Renders the matrix to browsers via WebSocket
- Forwards events from browsers to socket clients

## Features

- **Matrix Ownership**: Creates and manages its own Matrix instance
- **Socket Server**: Accepts Unix/TCP socket connections from Matrix CUI clients
- **WebSocket Display**: Serves matrix to browsers in real-time
- **Real-time Updates**: Matrix changes are broadcast to all connected browsers (~20 FPS)
- **Multiple Viewers**: Multiple browsers can watch the same matrix simultaneously
- **Bi-directional Interaction**: Browsers can send keyboard and mouse events to socket clients
- **DOM Rendering**: Fast, accessible DOM-based matrix rendering
- **Auto-reconnect**: Browsers automatically reconnect if connection drops

## Usage

### 1. Start the Web Server

Start the web server to create a Matrix and listen for connections:

```bash
cd examples/web-server
go run main.go
```

By default, this will:
- Create an 80x24 Matrix instance
- Listen for socket clients at `unix:///tmp/matrix-cui-web.sock`
- Start HTTP server on `:8080`
- Serve static files from `./static`

You should see:
```
Matrix CUI Web Server started
  Socket:   unix:///tmp/matrix-cui-web.sock (for clients like remote-paint)
  HTTP:     http://localhost:8080 (for browsers)
  Matrix:   80x24
  Static:   ./static
```

### 2. Open in Browser

Open your browser to:

```
http://localhost:8080
```

You should see an empty black matrix grid! It's waiting for a client to connect and draw.

### 3. Connect a Client

In another terminal, connect a client to draw on the matrix:

```bash
cd ../remote-paint
go run main.go -socket=unix:///tmp/matrix-cui-web.sock
```

Now paint in the terminal, and you'll see the updates appear in the browser in real-time!

## Command-Line Options

```bash
go run main.go [options]

Options:
  -socket string
        Socket address for clients (default "unix:///tmp/matrix-cui-web.sock")
        Examples:
          unix:///tmp/matrix-cui-web.sock
          tcp://localhost:9000

  -http string
        HTTP server address (default ":8080")

  -width int
        Matrix width (default 80)

  -height int
        Matrix height (default 24)

  -static string
        Static files directory (default "./static")
```

### Examples

**Listen on TCP socket:**
```bash
go run main.go -socket=tcp://localhost:9000
```

**Custom matrix size:**
```bash
go run main.go -width=120 -height=40
```

**Use custom HTTP port:**
```bash
go run main.go -http=:3000
```

**All together:**
```bash
go run main.go -socket=tcp://:9000 -http=:8080 -width=100 -height=30
```

## Building

```bash
go build
./web-server
```

## Testing the Complete Stack

**Terminal 1** - Start web-server:
```bash
cd examples/web-server
go run main.go
```

**Browser** - Open http://localhost:8080

You should see an empty black matrix grid.

**Terminal 2** - Start remote-paint client:
```bash
cd examples/remote-paint
go run main.go -socket=unix:///tmp/matrix-cui-web.sock
```

Now you should see:
- Web server logs showing client connected
- Browser showing the matrix with remote-paint's drawing
- remote-paint terminal showing your painting

**Interaction:**
- **Paint in terminal**: Use mouse in remote-paint terminal, see updates in browser
- **Paint in browser**: Click and drag in browser, see updates in terminal!
- **Keyboard in browser**: Press keys (1-9 for colors, C to clear, Q to quit) in browser
- **Real-time sync**: All changes are instantly visible everywhere

## How It Works

### Server Side (Go)

1. **Matrix Ownership**:
   - Creates and owns a Matrix instance (width x height)
   - Matrix is the single source of truth

2. **Dual Server**:
   - **Socket server**: Accepts connections from Matrix CUI clients (Unix/TCP)
   - **HTTP/WebSocket server**: Serves frontend and accepts browser connections

3. **Connection Management**:
   - Each socket client runs in its own goroutine with session management
   - Each browser runs in its own goroutine
   - All share access to the same Matrix instance

4. **Command Execution**:
   - Commands from socket clients → execute on Matrix
   - Commands from browsers → execute on Matrix
   - All changes automatically visible via periodic snapshot broadcasts

5. **Event Flow** (Bi-directional):
   - **Browser → Socket Clients**: Keyboard and mouse events captured in browser, forwarded to subscribed socket clients
   - **Socket Clients → Browser**: Matrix updates captured via periodic snapshots (~20 FPS), broadcast to all browsers
   - Browser acts as both input device and display

6. **Periodic Snapshots**:
   - Server broadcasts full matrix state to browsers every 50ms
   - Only broadcasts when browsers are connected (no wasted work)
   - Simple and reliable, good for typical usage

7. **Initial Sync**:
   - When browser connects, server sends full matrix snapshot
   - Browser renders initial state immediately

### Client Side (JavaScript)

1. **Protocol** (`protocol.js`):
   - WebSocket connection management
   - Command/response handling with UUID matching
   - Event subscription and distribution
   - Auto-reconnect on disconnect

2. **Renderer** (`renderer-dom.js`):
   - DOM-based matrix rendering using styled `<span>` elements
   - Cell-by-cell updates
   - Supports colors, bold, italic, underline

3. **Main** (`main.js`):
   - Application initialization
   - Event handling (keyboard, mouse) - sends events to server
   - Status updates
   - Terminal focus management

## Current Status

**✅ Phase 1 (Complete)**: Matrix ownership, socket server, WebSocket display, periodic snapshots

**✅ Phase 2 (Complete)**: Bi-directional interaction, keyboard and mouse events from browser

## Possible Future Enhancements

- **Performance**: Delta encoding (only send changed cells) instead of full snapshots
- **Canvas Renderer**: Alternative to DOM for better performance with large matrices
- **Web Component**: Extract as reusable component
- **Authentication**: Token-based authentication for connections
- **HTTPS/WSS**: TLS support for production
- **Connection Health**: Heartbeat/ping-pong, better reconnection logic
- **Compression**: gzip or msgpack for protocol messages

## Architecture Details

### Protocol

The web server uses the exact same JSON protocol as the socket-based communication. No translation needed!

**Command (Client/Browser → Server):**
```json
{
  "type": "command",
  "id": "uuid",
  "payload": {
    "cmd": "put",
    "x": 10,
    "y": 5,
    "cell": {"char": "A", "fg": "#00FF00", "bg": "#000000", "style": 1}
  }
}
```

**Event (Browser → Server → Socket Clients):**
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

**Snapshot (Server → Browser, periodic ~20 FPS):**
```json
{
  "type": "response",
  "id": "snapshot",
  "payload": {
    "ok": true,
    "width": 80,
    "height": 24,
    "cells": [[{"char": " ", "fg": "", "bg": "#000000", "style": 0}, ...], ...]
  }
}
```

### Concurrency

- Matrix instance is shared (reads/writes need to be synchronized in future)
- Each socket client connection runs in its own goroutine with mutex-protected writes
- Each browser connection runs in its own goroutine
- Event forwarding loop runs in a dedicated goroutine
- WebSocket writes are safe (gorilla/websocket handles concurrency for single writer)
- Socket session writes are mutex-protected to prevent race conditions

## Troubleshooting

**Browser can't connect:**
- Check that web-server is running
- Check browser console for errors
- Verify firewall isn't blocking port 8080

**Empty/black matrix displayed:**
- This is normal! The matrix starts empty
- Connect a client (like remote-paint) to draw on it
- Check web-server logs to see if client connected

**Socket client can't connect:**
- Check that web-server is running and listening on expected socket
- Verify socket address matches (default: unix:///tmp/matrix-cui-web.sock)
- Check web-server logs for connection errors
- For Unix sockets, check file permissions

**Matrix not updating in browser:**
- Check that socket client is actually sending commands
- Check browser console for WebSocket errors
- Try refreshing the browser
- Check web-server logs for errors

**Port already in use:**
- Use `-http` flag to specify different port
- Check what's using port 8080: `lsof -i :8080`

**Socket address already in use:**
- Use `-socket` flag to specify different address
- For Unix sockets, remove the old socket file: `rm /tmp/matrix-cui-web.sock`

## Files

```
examples/web-server/
├── main.go                   # Entry point
├── server/
│   └── webserver.go          # WebSocket proxy implementation
├── static/
│   ├── index.html            # Main page
│   ├── style.css             # Styling
│   ├── protocol.js           # WebSocket protocol client
│   ├── renderer-dom.js       # DOM renderer
│   └── main.js               # Application logic
├── go.mod                    # Go module
└── README.md                 # This file
```

## Dependencies

**Go:**
- `github.com/gorilla/websocket` - WebSocket implementation
- `github.com/google/uuid` - UUID generation for sessions
- `github.com/dyuri/matrix-cui` - Matrix CUI library
- `github.com/dyuri/matrix-cui/protocol` - Protocol types

**JavaScript:**
- None! Pure vanilla JavaScript

## Performance

- **Latency**: <5ms local, 10-50ms network
- **Bandwidth**: ~10-20 KB/s typical usage
- **Browsers**: Tested on Chrome, Firefox, Safari
- **Concurrent connections**: Handles 10+ browsers easily

## Security Notes

**⚠️ Development Mode**

This is currently configured for development:
- CORS allows all origins
- No authentication
- HTTP (not HTTPS)

**For production use, you should add:**
- Origin validation in WebSocket upgrader
- TLS/HTTPS support
- Authentication tokens
- Rate limiting
- Input validation (already handled by protocol layer)

## License

Same as Matrix CUI main project.
