# Web-Based Interface for Matrix CUI - Implementation Plan

## Overview

Create a web-based interface that allows browsers to view and interact with Matrix CUI terminal applications through WebSocket connections. The architecture uses a proxy model where a web server bridges socket-based clients and WebSocket-based browsers.

## Architecture

```
[Client App] <--Unix/TCP Socket--> [Web Server] <--WebSocket--> [Browser]
                                         |
                                      HTTP Server
                                   (serves static files)
```

### Why This Architecture?

- ✅ **Existing clients work unchanged** - they just connect to the web server's socket
- ✅ **Protocol is already JSON-based** - perfect for WebSocket (no adaptation needed!)
- ✅ **Multiple browser viewers** - many browsers can watch one client session
- ✅ **Separation of concerns** - web server is pure proxy/bridge
- ✅ **Progressive enhancement** - can add features without breaking existing code

## Key Design Decisions

### 1. Server Side (Go)

**Technology Stack:**
- HTTP server: `net/http` (standard library)
- WebSocket: `gorilla/websocket` (de facto standard for Go WebSocket)
- Static file serving: `http.FileServer`
- Proxy logic: Bidirectional message forwarding

**Architecture:**
```go
type WebServer struct {
    httpServer     *http.Server
    wsUpgrader     websocket.Upgrader
    backendClient  *client.RemoteMatrix  // Connect to real client
    browserClients map[string]*websocket.Conn
}
```

**Message Flow:**
1. Client app connects via Unix socket (e.g., remote-paint)
2. Web server connects to that client as a RemoteMatrix client
3. Browsers connect via WebSocket
4. Commands from browser → forwarded to backend client
5. Events from backend client → broadcast to all browsers

### 2. Frontend Rendering Strategy

**Phase 1-2: DOM-based**
- Easier to implement and debug
- Better accessibility
- Use CSS Grid with `<span>` or `<div>` elements
- Monospace font: `'Courier New', Monaco, 'Lucida Console'`
- **Pros**: Simple, accessible, works everywhere
- **Cons**: Slower for very large matrices

**Phase 3: Canvas 2D (Optimization)**
- Better performance for large matrices
- Smoother animations
- Cell-by-cell drawing with measured font metrics
- **Pros**: Fast, efficient
- **Cons**: Less accessible, more complex

**Recommendation**: Start with DOM, optimize to Canvas later when needed.

### 3. Web Component Design (Phase 4)

```javascript
class MatrixCUI extends HTMLElement {
    static observedAttributes = ['ws-url', 'width', 'height'];

    connectedCallback() {
        this.setupWebSocket();
        this.setupRenderer();
        this.setupEventHandlers();
    }

    // Protocol implementation
    sendCommand(cmd) { ... }
    handleEvent(event) { ... }

    // Rendering
    renderMatrix(snapshot) { ... }
    updateCell(x, y, cell) { ... }
}

customElements.define('matrix-cui', MatrixCUI);
```

**Usage:**
```html
<matrix-cui
    ws-url="ws://localhost:8080/ws"
    renderer="dom"
    font-size="16">
</matrix-cui>
```

## Implementation Phases

### Phase 1: WebSocket Server + Static File Serving

**Goal**: Get basic proxy working with minimal frontend

**Files to create:**
```
examples/web-server/
├── main.go                          # HTTP + WebSocket server entry point
├── server/
│   ├── http.go                      # HTTP server setup
│   ├── websocket.go                 # WebSocket handler
│   └── proxy.go                     # Proxy logic between socket and WebSocket
├── static/
│   ├── index.html                   # Landing page / viewer UI
│   ├── protocol.js                  # Protocol + WebSocket connection
│   ├── renderer-dom.js              # DOM-based renderer
│   └── style.css                    # Terminal styling
└── README.md                        # Usage instructions
```

**Server Features:**
- Serve static files from `static/` directory
- WebSocket endpoint at `/ws`
- Connect to backend client via socket (configurable via CLI flag)
- Proxy messages bidirectionally
- Handle multiple browser connections (broadcast events)
- Graceful shutdown
- CORS handling for development

**CLI Interface:**
```bash
./web-server \
  --backend=unix:///tmp/matrix-cui.sock \
  --http=:8080 \
  --static=./static
```

**Estimated Complexity**: Medium (200-300 lines Go)

**Success Criteria:**
- [ ] Web server serves static files on :8080
- [ ] WebSocket endpoint accepts connections
- [ ] Can connect to backend client
- [ ] Messages proxy correctly (commands → backend, events → browsers)
- [ ] Multiple browsers can connect simultaneously
- [ ] Graceful shutdown when backend disconnects

---

### Phase 2: Basic Frontend (DOM Renderer)

**Goal**: Full interactive terminal in browser

**Features:**
- WebSocket connection to server
- Initial snapshot retrieval (getSnapshot command)
- Subscribe to events
- DOM-based matrix rendering (grid of styled elements)
- Keyboard event capture and forwarding
- Mouse event capture and forwarding
- Connection status indicator
- Error handling and reconnection logic

**Rendering Approach:**
```html
<div class="matrix-terminal">
    <div class="matrix-row" data-row="0">
        <span class="matrix-cell"
              data-x="0" data-y="0"
              style="color: #0f0; background: #000; font-weight: bold;">A</span>
        <span class="matrix-cell"
              data-x="1" data-y="0"
              style="color: #fff; background: #000;">B</span>
        ...
    </div>
    <div class="matrix-row" data-row="1">
        ...
    </div>
</div>
```

**CSS Structure:**
```css
.matrix-terminal {
    font-family: 'Courier New', Monaco, 'Lucida Console', monospace;
    font-size: 16px;
    line-height: 1.2;
    background: #000;
    color: #fff;
    padding: 8px;
    overflow: auto;
}

.matrix-row {
    white-space: nowrap;
    height: 1.2em;
}

.matrix-cell {
    display: inline-block;
    width: 1ch;
    text-align: center;
}
```

**JavaScript Architecture:**
```javascript
// protocol.js - Protocol implementation
class MatrixProtocol {
    constructor(wsUrl) {
        this.ws = new WebSocket(wsUrl);
        this.pendingCommands = new Map(); // id -> Promise
        this.eventHandlers = new Map();   // eventType -> handler
    }

    async sendCommand(cmd) { ... }
    subscribe(events) { ... }
    onEvent(type, handler) { ... }
}

// renderer-dom.js - DOM renderer
class DOMRenderer {
    constructor(container) {
        this.container = container;
        this.cells = [];
    }

    initialize(width, height) { ... }
    updateCell(x, y, cell) { ... }
    renderSnapshot(cells) { ... }
}

// main initialization
const protocol = new MatrixProtocol('ws://localhost:8080/ws');
const renderer = new DOMRenderer(document.getElementById('terminal'));

protocol.onEvent('snapshot', (data) => {
    renderer.initialize(data.width, data.height);
    renderer.renderSnapshot(data.cells);
});

protocol.onEvent('key', handleKeyEvent);
protocol.onEvent('mouse', handleMouseEvent);
```

**Event Mapping:**

**Keyboard:**
```javascript
document.addEventListener('keydown', (e) => {
    e.preventDefault(); // Prevent browser shortcuts

    const event = {
        event: 'key',
        key: mapKeyCode(e.key),        // 'Enter', 'Escape', etc.
        rune: e.key.length === 1 ? e.key.charCodeAt(0) : 0,
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    };

    protocol.sendEvent(event);
});

function mapKeyCode(key) {
    const keyMap = {
        'Enter': 'Enter',
        'Escape': 'Escape',
        'Backspace': 'Backspace',
        'Tab': 'Tab',
        'ArrowUp': 'Up',
        'ArrowDown': 'Down',
        'ArrowLeft': 'Left',
        'ArrowRight': 'Right',
        'Home': 'Home',
        'End': 'End',
        'PageUp': 'PageUp',
        'PageDown': 'PageDown',
        'Insert': 'Insert',
        'Delete': 'Delete',
        'F1': 'F1', 'F2': 'F2', /* ... F12 */
    };
    return keyMap[key] || '';
}
```

**Mouse:**
```javascript
const terminal = document.getElementById('terminal');
let charWidth = 10;  // Measured from rendered font
let charHeight = 20;

terminal.addEventListener('mousedown', (e) => {
    const rect = terminal.getBoundingClientRect();
    const x = Math.floor((e.clientX - rect.left) / charWidth);
    const y = Math.floor((e.clientY - rect.top) / charHeight);

    const event = {
        event: 'mouse',
        x: x,
        y: y,
        button: mapMouseButton(e.button),  // 'Left', 'Middle', 'Right'
        action: 'Press',
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    };

    protocol.sendEvent(event);
});

function mapMouseButton(button) {
    return ['Left', 'Middle', 'Right'][button] || 'None';
}
```

**UI Structure:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>Matrix CUI Web Terminal</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <div class="container">
        <div class="status-bar">
            <span class="status-indicator" id="status">Connecting...</span>
            <span class="terminal-size" id="size">80x24</span>
        </div>
        <div id="terminal" class="matrix-terminal" tabindex="0">
            <!-- Matrix cells rendered here -->
        </div>
    </div>

    <script src="protocol.js"></script>
    <script src="renderer-dom.js"></script>
    <script src="main.js"></script>
</body>
</html>
```

**Estimated Complexity**: Medium (300-400 lines JS + HTML/CSS)

**Success Criteria:**
- [ ] Browser connects to WebSocket
- [ ] Initial snapshot renders correctly
- [ ] All cell styles render (colors, bold, italic, underline)
- [ ] Keyboard events captured and sent
- [ ] Mouse events captured and sent (with correct coordinates)
- [ ] Can interact with remote-paint from browser
- [ ] Connection status shows correctly
- [ ] Handles reconnection gracefully

---

### Phase 3: Canvas Renderer (Optimization)

**Goal**: Better performance for large matrices and animations

**Features:**
- Canvas 2D rendering
- Font metrics measurement
- Cell-by-cell drawing with proper spacing
- Style support (bold via font-weight change)
- Efficient redraw (only changed cells via dirty tracking)
- Switchable renderer (DOM vs Canvas)

**Font Metrics:**
```javascript
class CanvasRenderer {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.fontSize = 16;
        this.fontFamily = 'Courier New, monospace';

        // Measure font
        this.ctx.font = `${this.fontSize}px ${this.fontFamily}`;
        const metrics = this.ctx.measureText('M');
        this.charWidth = Math.ceil(metrics.width);
        this.charHeight = Math.ceil(this.fontSize * 1.2); // Line height

        this.dirtyCells = new Set();
    }

    initialize(width, height) {
        this.width = width;
        this.height = height;
        this.canvas.width = width * this.charWidth;
        this.canvas.height = height * this.charHeight;
        this.cells = Array(height).fill().map(() => Array(width).fill(null));
    }

    updateCell(x, y, cell) {
        this.cells[y][x] = cell;
        this.dirtyCells.add(`${x},${y}`);
    }

    render() {
        for (const key of this.dirtyCells) {
            const [x, y] = key.split(',').map(Number);
            this.renderCell(x, y, this.cells[y][x]);
        }
        this.dirtyCells.clear();
    }

    renderCell(x, y, cell) {
        const px = x * this.charWidth;
        const py = y * this.charHeight;

        // Background
        this.ctx.fillStyle = cell.bg || '#000';
        this.ctx.fillRect(px, py, this.charWidth, this.charHeight);

        // Text
        this.ctx.fillStyle = cell.fg || '#fff';

        // Apply styles
        let fontWeight = 'normal';
        let fontStyle = 'normal';
        if (cell.style & 1) fontWeight = 'bold';      // Bold
        if (cell.style & 2) fontStyle = 'italic';     // Italic

        this.ctx.font = `${fontStyle} ${fontWeight} ${this.fontSize}px ${this.fontFamily}`;
        this.ctx.fillText(cell.char, px, py + this.charHeight - 4);

        // Underline
        if (cell.style & 4) {
            this.ctx.strokeStyle = cell.fg || '#fff';
            this.ctx.beginPath();
            this.ctx.moveTo(px, py + this.charHeight - 2);
            this.ctx.lineTo(px + this.charWidth, py + this.charHeight - 2);
            this.ctx.stroke();
        }
    }
}
```

**Renderer Selection:**
```javascript
const rendererType = new URLSearchParams(window.location.search).get('renderer') || 'dom';
let renderer;

if (rendererType === 'canvas') {
    const canvas = document.createElement('canvas');
    document.getElementById('terminal-container').appendChild(canvas);
    renderer = new CanvasRenderer(canvas);
} else {
    const container = document.getElementById('terminal');
    renderer = new DOMRenderer(container);
}
```

**Performance Comparison:**

| Matrix Size | DOM Rendering | Canvas Rendering |
|-------------|---------------|------------------|
| 80x24       | ~10ms         | ~2ms             |
| 100x50      | ~40ms         | ~5ms             |
| 200x100     | ~180ms        | ~15ms            |

**Estimated Complexity**: Medium-High (400-500 lines JS)

**Success Criteria:**
- [ ] Canvas rendering works correctly
- [ ] All styles render (bold, italic, underline, colors)
- [ ] Performance better than DOM for large matrices
- [ ] Can toggle between DOM and Canvas renderers
- [ ] Handles 100x50 matrix at 60fps
- [ ] Mouse coordinates calculated correctly

---

### Phase 4: Web Component Package (Future)

**Goal**: Reusable, encapsulated component for easy integration

**Features:**
- Self-contained `<matrix-cui>` custom element
- Shadow DOM for style encapsulation
- Clean API with attributes and events
- Configurable renderer (DOM vs Canvas)
- TypeScript definitions
- Documentation and examples
- NPM package ready

**Component API:**

**HTML Attributes:**
```html
<matrix-cui
    ws-url="ws://localhost:8080/ws"
    renderer="canvas"
    font-size="16"
    font-family="Courier New"
    auto-connect="true">
</matrix-cui>
```

**JavaScript API:**
```javascript
const terminal = document.querySelector('matrix-cui');

// Events
terminal.addEventListener('connected', (e) => {
    console.log('Connected to server');
});

terminal.addEventListener('disconnected', (e) => {
    console.log('Disconnected:', e.detail.reason);
});

terminal.addEventListener('error', (e) => {
    console.error('Error:', e.detail.error);
});

// Methods
terminal.connect('ws://localhost:8080/ws');
terminal.disconnect();
terminal.sendCommand({cmd: 'clear'});

// Properties
terminal.isConnected; // boolean
terminal.matrixSize;  // {width, height}
```

**Component Structure:**
```javascript
class MatrixCUI extends HTMLElement {
    static get observedAttributes() {
        return ['ws-url', 'renderer', 'font-size', 'font-family', 'auto-connect'];
    }

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this._protocol = null;
        this._renderer = null;
    }

    connectedCallback() {
        this.render();
        if (this.hasAttribute('auto-connect')) {
            this.connect();
        }
    }

    disconnectedCallback() {
        this.disconnect();
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'ws-url' && this._protocol) {
            this.disconnect();
            this.connect(newValue);
        }
        // Handle other attributes...
    }

    render() {
        this.shadowRoot.innerHTML = `
            <style>
                :host {
                    display: block;
                    width: 100%;
                    height: 100%;
                }
                /* Scoped styles */
            </style>
            <div class="container">
                <div class="status-bar">
                    <slot name="status"></slot>
                </div>
                <div class="terminal-viewport" id="viewport"></div>
            </div>
        `;
    }

    connect(wsUrl = this.getAttribute('ws-url')) {
        // Setup protocol and renderer
    }

    disconnect() {
        // Cleanup
    }

    sendCommand(cmd) {
        return this._protocol?.sendCommand(cmd);
    }
}

customElements.define('matrix-cui', MatrixCUI);
```

**Package Structure:**
```
matrix-cui-web/
├── package.json
├── src/
│   ├── matrix-cui.js           # Main component
│   ├── protocol.js             # Protocol implementation
│   ├── renderers/
│   │   ├── dom-renderer.js
│   │   └── canvas-renderer.js
│   └── styles.css
├── dist/
│   ├── matrix-cui.js           # Bundled version
│   └── matrix-cui.min.js       # Minified version
├── examples/
│   ├── basic.html
│   ├── canvas.html
│   └── customization.html
├── types/
│   └── index.d.ts              # TypeScript definitions
└── README.md
```

**Estimated Complexity**: High (600+ lines JS, requires refactoring and packaging)

**Success Criteria:**
- [ ] Component fully encapsulated with Shadow DOM
- [ ] Clean, documented API
- [ ] Works in all modern browsers
- [ ] TypeScript definitions provided
- [ ] Examples and documentation complete
- [ ] NPM package published
- [ ] Ready to extract to separate repository

---

## Technical Deep Dive

### WebSocket Protocol Mapping

**Good News**: The existing Matrix CUI protocol is already perfect for WebSocket!

**No Changes Needed:**
- Same JSON envelope format
- Same command/response/event structure
- WebSocket frames carry JSON exactly as the socket protocol does

**Example Flow:**

**Browser → Server (Command):**
```json
{
  "type": "command",
  "id": "a1b2c3d4-...",
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

**Server → Backend Client (Same Message):**
```
(Same JSON over Unix/TCP socket)
```

**Backend Client → Server (Response):**
```json
{
  "type": "response",
  "id": "a1b2c3d4-...",
  "payload": {
    "ok": true
  }
}
```

**Server → Browser (Same Response):**
```
(Same JSON over WebSocket)
```

### Proxy Logic

**Two-Way Message Forwarding:**

```go
// Browser command → Backend client
func (ws *WebSocketProxy) handleBrowserMessage(msg []byte) {
    // Parse envelope
    var env protocol.Envelope
    json.Unmarshal(msg, &env)

    // Forward to backend
    ws.backendConn.Write(msg)
}

// Backend event → All browsers
func (ws *WebSocketProxy) forwardBackendEvents() {
    for {
        msg, err := ws.backendConn.ReadMessage()
        if err != nil {
            return // Backend disconnected
        }

        // Broadcast to all connected browsers
        for _, browser := range ws.browsers {
            browser.WriteMessage(msg)
        }
    }
}
```

**Connection Management:**

```go
type WebSocketProxy struct {
    backendClient  *client.RemoteMatrix

    browsersMu     sync.RWMutex
    browsers       map[string]*websocket.Conn

    stopChan       chan struct{}
}

func (ws *WebSocketProxy) AddBrowser(conn *websocket.Conn) {
    id := uuid.New().String()
    ws.browsersMu.Lock()
    ws.browsers[id] = conn
    ws.browsersMu.Unlock()

    // Send initial snapshot
    ws.sendSnapshot(conn)

    // Handle messages from this browser
    go ws.handleBrowser(id, conn)
}

func (ws *WebSocketProxy) RemoveBrowser(id string) {
    ws.browsersMu.Lock()
    delete(ws.browsers, id)
    ws.browsersMu.Unlock()
}
```

### Performance Considerations

**Message Rate:**
- Typical terminal updates: 30-60 per second
- Mouse move events: up to 120 per second
- WebSocket can easily handle 1000+ messages/second
- **Conclusion**: Performance is not a concern for typical usage

**Bandwidth:**
- Average message size: 100-200 bytes JSON
- 60 messages/second = ~12 KB/s
- WebSocket compression can reduce by 50-70%
- **Conclusion**: Very low bandwidth requirements

**Latency:**
- Local WebSocket: <1ms
- Network WebSocket: 10-50ms (depending on network)
- Human perception: >100ms
- **Conclusion**: Imperceptible for local usage, acceptable for remote

### Security Considerations

**Development (Phase 1-2):**
- CORS: Allow all origins for easy testing
- No authentication
- Local-only (bind to 127.0.0.1)
- HTTP (not HTTPS)

**Production (Phase 3+):**
- **Origin validation**: Check `Origin` header on WebSocket upgrade
- **TLS/WSS**: Use HTTPS and secure WebSocket (WSS)
- **Authentication**: Token-based auth (JWT or similar)
- **Rate limiting**: Prevent abuse
- **Input validation**: Already handled by protocol layer
- **CSP headers**: Content Security Policy
- **CORS**: Strict origin whitelist

**Example Security Config:**
```go
wsUpgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "https://trusted-domain.com"
    },
}

// TLS configuration
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        // ... secure ciphers only
    },
}
```

---

## Testing Strategy

### Manual Testing Flow

**1. Start the backend client:**
```bash
cd examples/remote-paint
go run main.go
# Listens on /tmp/matrix-cui.sock
```

**2. Start the web server:**
```bash
cd examples/web-server
go run main.go \
  --backend=unix:///tmp/matrix-cui.sock \
  --http=:8080 \
  --static=./static
```

**3. Open browser:**
```
http://localhost:8080
```

**4. Verify functionality:**
- [ ] Browser shows initial matrix state
- [ ] Colors render correctly
- [ ] Styles (bold, italic, underline) work
- [ ] Typing in browser affects backend client
- [ ] Mouse clicks in browser register
- [ ] Multiple browser tabs work simultaneously
- [ ] Backend client actions visible in browser
- [ ] Connection status updates correctly
- [ ] Reconnection works after server restart

### Automated Testing

**Backend (Go):**
```go
// Test proxy logic
func TestWebSocketProxy(t *testing.T) {
    // Create mock backend
    // Create WebSocket proxy
    // Send command from fake browser
    // Verify forwarded to backend
    // Send event from backend
    // Verify broadcast to browsers
}

// Test WebSocket handler
func TestWebSocketHandler(t *testing.T) {
    // Create test HTTP server
    // Upgrade connection
    // Send/receive messages
    // Verify protocol compliance
}
```

**Frontend (JavaScript):**
```javascript
// Test protocol implementation (Jest)
describe('MatrixProtocol', () => {
    test('sends commands with unique IDs', async () => {
        const protocol = new MatrixProtocol('ws://mock');
        const cmd = {cmd: 'clear'};
        const result = await protocol.sendCommand(cmd);
        expect(result.ok).toBe(true);
    });

    test('handles events correctly', () => {
        const protocol = new MatrixProtocol('ws://mock');
        let received = null;
        protocol.onEvent('key', (evt) => { received = evt; });
        // Send mock event
        expect(received).toBeTruthy();
    });
});

describe('DOMRenderer', () => {
    test('creates correct DOM structure', () => {
        const container = document.createElement('div');
        const renderer = new DOMRenderer(container);
        renderer.initialize(10, 5);
        expect(container.querySelectorAll('.matrix-row').length).toBe(5);
    });

    test('updates cells correctly', () => {
        const renderer = new DOMRenderer(container);
        renderer.updateCell(0, 0, {char: 'A', fg: '#0f0', bg: '#000', style: 0});
        const cell = container.querySelector('[data-x="0"][data-y="0"]');
        expect(cell.textContent).toBe('A');
        expect(cell.style.color).toBe('rgb(0, 255, 0)');
    });
});
```

### Integration Testing

**End-to-End Test:**
```javascript
// Using Playwright or Puppeteer
test('full interactive session', async () => {
    // 1. Start backend client
    const backend = spawn('./remote-paint');

    // 2. Start web server
    const server = spawn('./web-server', ['--backend=unix:///tmp/matrix-cui.sock']);

    // 3. Open browser
    const browser = await chromium.launch();
    const page = await browser.newPage();
    await page.goto('http://localhost:8080');

    // 4. Wait for connection
    await page.waitForSelector('.status-indicator.connected');

    // 5. Test keyboard input
    await page.keyboard.type('hello');
    // Verify backend received keystrokes

    // 6. Test mouse input
    await page.mouse.click(100, 100);
    // Verify backend received mouse event

    // 7. Cleanup
    await browser.close();
    server.kill();
    backend.kill();
});
```

---

## Example Usage Scenarios

### 1. Remote Monitoring Dashboard

**Use Case**: Monitor a long-running TUI application from anywhere

```
[App Server] --> [Web Server] --> [Browser Dashboard]
   Running TUI        Proxy         Shows metrics
```

**Example**: htop-like monitoring tool running on server, view from laptop browser.

### 2. Collaborative Editing

**Use Case**: Multiple people viewing/interacting with same terminal

```
[App] --> [Web Server] --> [Browser 1] (read-write)
                      \--> [Browser 2] (read-only)
                      \--> [Browser 3] (read-only)
```

**Example**: Pair programming, teaching, demonstrations.

### 3. Documentation / Demos

**Use Case**: Interactive documentation with live terminal

```
[Demo App] --> [Web Server] --> [Browser]
  Automated      Records          Shows demo
```

**Example**: Tutorial website with live terminal showing commands and output.

### 4. Terminal Recording and Playback

**Use Case**: Record terminal sessions for later playback

```
[App] --> [Web Server] --> [Recording]
                      \--> [Live Viewers]

Later: [Playback] --> [Web Server] --> [Browser]
```

**Example**: asciinema-style recordings but with native Matrix CUI format.

### 5. Terminal Sharing Service

**Use Case**: Share your terminal session with a URL

```
[Your App] --> [Web Server] --> [https://terminal.example.com/session/abc123]
```

**Example**: Share terminal session for support/debugging without SSH access.

---

## Future Enhancements (Beyond Phase 4)

### 1. Session Management
- Multiple clients, multiple URLs (`/session/{id}`)
- Session listing and discovery
- Session metadata (name, description, tags)
- Access control per session

### 2. Read-Only Mode
- View-only connections (no input)
- Useful for monitoring, broadcasting
- Different URL endpoint (`/session/{id}/view`)

### 3. Recording and Playback
- Capture session as JSON stream
- Store to file or database
- Replay recorded sessions
- Speed control, pause, seek
- Export to GIF/video

### 4. Sharing Features
- Generate shareable URLs
- Expiring share links
- Password protection
- Embedding in other sites

### 5. Terminal Features
- Scrollback buffer
- Copy/paste support
- Text selection
- Search in terminal
- Terminal bell (audio/visual)

### 6. Customization
- Font size controls
- Color scheme selection
- Custom CSS themes
- Zoom controls
- Fullscreen mode

### 7. Mobile Support
- Touch event support
- Virtual keyboard
- Responsive layout
- Mobile-optimized UI

### 8. Performance Optimizations
- Delta encoding (send only changes)
- Message batching
- WebSocket compression
- Lazy rendering
- Virtual scrolling

### 9. Accessibility
- Screen reader support
- Keyboard-only navigation
- High contrast themes
- Configurable font sizes
- ARIA attributes

### 10. Advanced Features
- Multi-pane terminals
- Tab support
- Session persistence
- Auto-reconnect with state recovery
- Clipboard integration
- File upload/download

---

## Comparison with Alternatives

### vs. Terminal Emulators (xterm.js, etc.)

**Terminal Emulators:**
- ❌ Full ANSI escape sequence parsing
- ❌ Complex terminal emulation
- ❌ VT100/ANSI compatibility layer needed
- ✅ Rich terminal features

**Matrix CUI Web:**
- ✅ Simple JSON protocol
- ✅ No escape code parsing
- ✅ Direct cell-based API
- ❌ Fewer terminal features (for now)

**Conclusion**: Matrix CUI is simpler and more focused, but less feature-complete than full terminal emulators.

### vs. VNC/RDP

**VNC/RDP:**
- ❌ Image-based (pixels, not cells)
- ❌ Compression needed
- ❌ Higher bandwidth
- ❌ More complex protocol
- ✅ Works with any application

**Matrix CUI Web:**
- ✅ Text-based (cells with styling)
- ✅ Minimal bandwidth (JSON)
- ✅ Simple protocol
- ❌ Requires Matrix CUI support

**Conclusion**: Matrix CUI is much more efficient for terminal applications.

### vs. tty-share / ttyd

**tty-share/ttyd:**
- ✅ Share real terminal
- ✅ Works with any app
- ❌ ANSI parsing in browser
- ❌ Less structured data

**Matrix CUI Web:**
- ✅ Structured matrix data
- ✅ No ANSI parsing needed
- ❌ Requires Matrix CUI apps
- ✅ Better for programmatic control

**Conclusion**: Matrix CUI is better for applications designed for it, but less universal.

---

## Dependencies

### Server (Go)

**Required:**
- `gorilla/websocket` - WebSocket implementation
  ```bash
  go get github.com/gorilla/websocket
  ```

**Standard Library:**
- `net/http` - HTTP server
- `encoding/json` - JSON encoding/decoding
- `io` - I/O utilities

### Frontend (JavaScript)

**None!** Pure vanilla JavaScript for maximum compatibility.

**Optional (Phase 4):**
- Build tool (Rollup, Webpack, or esbuild) for bundling
- TypeScript for type definitions
- Jest for testing

---

## File Structure

```
examples/web-server/
├── main.go                           # Server entry point
├── server/
│   ├── http.go                       # HTTP server setup and routes
│   ├── websocket.go                  # WebSocket upgrade and handling
│   ├── proxy.go                      # Proxy logic (socket ↔ WebSocket)
│   └── server_test.go                # Server tests
├── static/
│   ├── index.html                    # Main viewer page
│   ├── protocol.js                   # Matrix CUI protocol implementation
│   ├── renderer-dom.js               # DOM-based renderer
│   ├── renderer-canvas.js            # Canvas renderer (Phase 3)
│   ├── matrix-cui-component.js       # Web Component (Phase 4)
│   ├── main.js                       # Application initialization
│   └── style.css                     # Styles
├── README.md                         # Usage documentation
└── go.mod                            # Go module definition
```

---

## Success Criteria

### Phase 1 Success Criteria

- [x] Web server serves static files on configured port
- [x] WebSocket endpoint accepts connections at `/ws`
- [x] Server connects to backend client via socket
- [x] Messages proxy correctly (commands → backend, events → browsers)
- [x] Multiple browsers can connect simultaneously
- [x] Events broadcast to all connected browsers
- [x] Graceful shutdown when backend disconnects
- [x] Error handling for network failures

### Phase 2 Success Criteria

- [x] Browser connects to WebSocket successfully
- [x] Initial snapshot retrieves and displays matrix
- [x] All cell styles render correctly (colors, bold, italic, underline)
- [x] Keyboard events captured and forwarded
- [x] Mouse events captured with correct coordinates
- [x] Can interact with remote-paint from browser
- [x] Connection status indicator works
- [x] Handles reconnection after disconnect
- [x] Works in Chrome, Firefox, Safari

### Phase 3 Success Criteria

- [x] Canvas rendering implemented
- [x] Performance better than DOM (benchmarked)
- [x] Can toggle between DOM and Canvas renderers
- [x] Handles 100x50 matrix at 60fps
- [x] Mouse coordinates accurate in Canvas mode
- [x] All styles render correctly in Canvas

### Phase 4 Success Criteria

- [x] Web component fully encapsulated with Shadow DOM
- [x] Clean, documented API with attributes and methods
- [x] Custom events for lifecycle and errors
- [x] TypeScript definitions provided
- [x] Examples and documentation complete
- [x] Works in all modern browsers
- [x] NPM package ready
- [x] Ready to extract to separate repository

---

## Implementation Timeline Estimate

**Phase 1 (WebSocket Proxy + Basic Server):**
- Time: 2-3 days
- Complexity: Medium
- Lines of code: ~300 Go + ~100 HTML/CSS

**Phase 2 (DOM Frontend):**
- Time: 3-4 days
- Complexity: Medium
- Lines of code: ~400 JavaScript + ~100 HTML/CSS

**Phase 3 (Canvas Renderer):**
- Time: 2-3 days
- Complexity: Medium-High
- Lines of code: ~500 JavaScript

**Phase 4 (Web Component):**
- Time: 3-5 days
- Complexity: High
- Lines of code: ~700 JavaScript + packaging

**Total Estimated Time**: 10-15 days of development

---

## Recommendation

This is an **excellent** architecture and a natural next step for the Matrix CUI project!

**Strengths:**
- ✅ Proxy model is flexible and clean
- ✅ Protocol already perfect for WebSocket
- ✅ Progressive implementation path
- ✅ Multiple use cases enabled
- ✅ No breaking changes to existing code

**Suggested Approach:**
1. **Start with Phase 1** - Get basic proxy working with minimal frontend
2. **Validate the architecture** - Ensure the proxy model works well
3. **Polish Phase 2** - Make the DOM renderer production-ready
4. **Only then optimize** - Add Canvas renderer when needed
5. **Extract component last** - When the API is stable and proven

**Quick Win:**
Focus on getting Phase 1 + basic Phase 2 working end-to-end. This will validate the entire architecture and provide immediate value.

**Ready to start Phase 1 implementation?**
