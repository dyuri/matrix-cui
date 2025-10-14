package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	matrixcui "github.com/dyuri/matrix-cui"
	"github.com/dyuri/matrix-cui/protocol"
)

// WebServer manages Matrix, socket server, HTTP server, and WebSocket connections.
type WebServer struct {
	matrix     *matrixcui.Matrix
	socketAddr string
	httpAddr   string
	staticDir  string

	httpServer *http.Server
	wsUpgrader websocket.Upgrader

	listenerMu sync.Mutex
	listener   net.Listener

	// Socket client sessions
	sessionsMu sync.RWMutex
	sessions   map[string]*SocketSession

	// Browser WebSocket connections
	browsersMu sync.RWMutex
	browsers   map[string]*websocket.Conn

	// Event channel for browser events (Phase 2)
	eventChan chan matrixcui.Event

	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewWebServer creates a new web server.
func NewWebServer(socketAddr, httpAddr, staticDir string, width, height int) (*WebServer, error) {
	ws := &WebServer{
		matrix:     matrixcui.NewMatrix(width, height),
		socketAddr: socketAddr,
		httpAddr:   httpAddr,
		staticDir:  staticDir,
		sessions:   make(map[string]*SocketSession),
		browsers:   make(map[string]*websocket.Conn),
		eventChan:  make(chan matrixcui.Event, 10),
		stopChan:   make(chan struct{}),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				// TODO: In production, validate origin
				return true
			},
		},
	}

	return ws, nil
}

// Start starts the web server, socket server, and HTTP server.
func (ws *WebServer) Start() error {
	// Start socket listener
	if err := ws.startSocketListener(); err != nil {
		return fmt.Errorf("failed to start socket listener: %w", err)
	}

	// Start event forwarding loop (browser events to socket clients)
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.forwardBrowserEvents()
	}()

	// Start periodic snapshot broadcaster (matrix updates to browsers)
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.broadcastSnapshotsPeriodically()
	}()

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.handleWebSocket)
	mux.Handle("/", http.FileServer(http.Dir(ws.staticDir)))

	// Create HTTP server
	ws.httpServer = &http.Server{
		Addr:    ws.httpAddr,
		Handler: mux,
	}

	// Start HTTP server in background
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		if err := ws.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Printf("Web server started: socket=%s http=%s", ws.socketAddr, ws.httpAddr)
	return nil
}

// startSocketListener starts listening for socket connections.
func (ws *WebServer) startSocketListener() error {
	var listener net.Listener
	var err error

	// Parse socket address (unix:// or tcp://)
	if strings.HasPrefix(ws.socketAddr, "unix://") {
		socketPath := strings.TrimPrefix(ws.socketAddr, "unix://")
		// Remove existing socket file
		if err := os.RemoveAll(socketPath); err != nil {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen on unix socket: %w", err)
		}
	} else if strings.HasPrefix(ws.socketAddr, "tcp://") {
		tcpAddr := strings.TrimPrefix(ws.socketAddr, "tcp://")
		listener, err = net.Listen("tcp", tcpAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on tcp: %w", err)
		}
	} else {
		return fmt.Errorf("invalid socket address format (use unix:// or tcp://): %s", ws.socketAddr)
	}

	ws.listenerMu.Lock()
	ws.listener = listener
	ws.listenerMu.Unlock()

	// Start accept loop
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.acceptLoop()
	}()

	return nil
}

// acceptLoop accepts incoming socket connections.
func (ws *WebServer) acceptLoop() {
	for {
		select {
		case <-ws.stopChan:
			return
		default:
		}

		conn, err := ws.listener.Accept()
		if err != nil {
			select {
			case <-ws.stopChan:
				return
			default:
				log.Printf("Socket accept error: %v", err)
				continue
			}
		}

		session := NewSocketSession(conn, ws)
		ws.registerSession(session)

		ws.wg.Add(1)
		go func() {
			defer ws.wg.Done()
			session.Handle()
			ws.unregisterSession(session)
		}()
	}
}

// Stop stops the web server gracefully.
func (ws *WebServer) Stop() {
	close(ws.stopChan)

	// Close socket listener
	ws.listenerMu.Lock()
	if ws.listener != nil {
		ws.listener.Close()
	}
	ws.listenerMu.Unlock()

	// Close all socket sessions
	ws.sessionsMu.Lock()
	for _, session := range ws.sessions {
		session.conn.Close()
	}
	ws.sessionsMu.Unlock()

	// Close all browser connections
	ws.browsersMu.Lock()
	for id, conn := range ws.browsers {
		conn.Close()
		delete(ws.browsers, id)
	}
	ws.browsersMu.Unlock()

	// Shutdown HTTP server
	if ws.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ws.httpServer.Shutdown(ctx)
	}

	ws.wg.Wait()
}

// registerSession adds a socket session.
func (ws *WebServer) registerSession(session *SocketSession) {
	ws.sessionsMu.Lock()
	defer ws.sessionsMu.Unlock()
	ws.sessions[session.id] = session
	log.Printf("Socket client connected: %s (total: %d)", session.id, len(ws.sessions))
}

// unregisterSession removes a socket session.
func (ws *WebServer) unregisterSession(session *SocketSession) {
	ws.sessionsMu.Lock()
	defer ws.sessionsMu.Unlock()
	delete(ws.sessions, session.id)
	log.Printf("Socket client disconnected: %s (remaining: %d)", session.id, len(ws.sessions))
}

// forwardBrowserEvents forwards events from browsers to socket clients.
func (ws *WebServer) forwardBrowserEvents() {
	for {
		select {
		case <-ws.stopChan:
			return
		case event := <-ws.eventChan:
			ws.broadcastEventToSockets(event)
		}
	}
}

// broadcastEventToSockets sends an event to all subscribed socket clients.
func (ws *WebServer) broadcastEventToSockets(event matrixcui.Event) {
	ws.sessionsMu.RLock()
	defer ws.sessionsMu.RUnlock()

	for _, session := range ws.sessions {
		session.sendEvent(event)
	}
}

// broadcastSnapshotsPeriodically sends matrix snapshots to all browsers periodically.
func (ws *WebServer) broadcastSnapshotsPeriodically() {
	ticker := time.NewTicker(50 * time.Millisecond) // ~20 FPS
	defer ticker.Stop()

	for {
		select {
		case <-ws.stopChan:
			return
		case <-ticker.C:
			// Only broadcast if there are browsers connected
			ws.browsersMu.RLock()
			browserCount := len(ws.browsers)
			ws.browsersMu.RUnlock()

			if browserCount > 0 {
				ws.broadcastSnapshotToAll()
			}
		}
	}
}

// broadcastSnapshotToAll sends the current matrix state to all connected browsers.
func (ws *WebServer) broadcastSnapshotToAll() {
	width := ws.matrix.Width()
	height := ws.matrix.Height()

	// Get full snapshot
	cells := protocol.MatrixToProto(ws.matrix)

	// Create snapshot response
	resp := protocol.NewResponseSnapshot(width, height, cells)
	env, err := protocol.WrapResponse("snapshot", resp)
	if err != nil {
		log.Printf("Failed to wrap snapshot: %v", err)
		return
	}

	// Broadcast to all browsers
	ws.browsersMu.RLock()
	defer ws.browsersMu.RUnlock()

	for id, conn := range ws.browsers {
		if err := conn.WriteJSON(env); err != nil {
			log.Printf("Failed to send snapshot to browser %s: %v", id, err)
		}
	}
}

// handleWebSocket handles WebSocket upgrade and connection.
func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := ws.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Generate browser ID
	browserID := fmt.Sprintf("browser-%d", time.Now().UnixNano())

	// Register browser
	ws.browsersMu.Lock()
	ws.browsers[browserID] = conn
	ws.browsersMu.Unlock()

	log.Printf("Browser connected: %s (total: %d)", browserID, len(ws.browsers))

	// Send initial snapshot
	ws.sendSnapshot(conn)

	// Handle messages from this browser
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		defer func() {
			ws.browsersMu.Lock()
			delete(ws.browsers, browserID)
			browserCount := len(ws.browsers)
			ws.browsersMu.Unlock()
			conn.Close()
			log.Printf("Browser disconnected: %s (remaining: %d)", browserID, browserCount)
		}()

		ws.handleBrowserMessages(conn)
	}()
}

// sendSnapshot sends the current matrix state to a browser.
func (ws *WebServer) sendSnapshot(conn *websocket.Conn) {
	width := ws.matrix.Width()
	height := ws.matrix.Height()

	// Get full snapshot
	cells := protocol.MatrixToProto(ws.matrix)

	// Create snapshot response
	resp := protocol.NewResponseSnapshot(width, height, cells)
	env, err := protocol.WrapResponse("snapshot", resp)
	if err != nil {
		log.Printf("Failed to wrap snapshot: %v", err)
		return
	}

	// Send to browser
	if err := conn.WriteJSON(env); err != nil {
		log.Printf("Failed to send snapshot: %v", err)
	}
}

// handleBrowserMessages handles incoming messages from a browser.
func (ws *WebServer) handleBrowserMessages(conn *websocket.Conn) {
	for {
		select {
		case <-ws.stopChan:
			return
		default:
		}

		// Read message from browser
		var env protocol.Envelope
		if err := conn.ReadJSON(&env); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		// Handle different message types
		switch env.Type {
		case protocol.MessageTypeCommand:
			// Parse command
			cmd, err := protocol.UnwrapCommand(env)
			if err != nil {
				log.Printf("Failed to unwrap command: %v", err)
				continue
			}

			// Execute command on matrix and get response
			resp := ws.executeCommand(cmd)

			// Send response back to browser
			respEnv, err := protocol.WrapResponse(env.ID, resp)
			if err != nil {
				log.Printf("Failed to wrap response: %v", err)
				continue
			}

			if err := conn.WriteJSON(respEnv); err != nil {
				log.Printf("Failed to send response: %v", err)
				return
			}

		case protocol.MessageTypeEvent:
			// Parse event from browser
			eventPayload, err := protocol.UnwrapEvent(env)
			if err != nil {
				log.Printf("Failed to unwrap event: %v", err)
				continue
			}

			// Convert protocol event to matrixcui event
			event := ws.convertProtocolEvent(eventPayload)
			if event != nil {
				// Send to event channel for forwarding to socket clients
				select {
				case ws.eventChan <- event:
				default:
					log.Printf("Event channel full, dropping event")
				}
			}

		default:
			log.Printf("Unknown message type from browser: %s", env.Type)
		}
	}
}

// convertProtocolEvent converts a protocol.EventPayload to a matrixcui.Event.
func (ws *WebServer) convertProtocolEvent(payload protocol.EventPayload) matrixcui.Event {
	switch payload.Event {
	case "key":
		return matrixcui.KeyEvent{
			Key:   ws.parseKey(payload.Key),
			Rune:  rune(payload.Rune),
			Alt:   payload.Alt,
			Ctrl:  payload.Ctrl,
			Shift: payload.Shift,
		}

	case "mouse":
		x := 0
		y := 0
		if payload.X != nil {
			x = *payload.X
		}
		if payload.Y != nil {
			y = *payload.Y
		}
		return matrixcui.MouseEvent{
			X:      x,
			Y:      y,
			Button: ws.parseMouseButton(payload.Button),
			Action: ws.parseMouseAction(payload.Action),
			Alt:    payload.Alt,
			Ctrl:   payload.Ctrl,
			Shift:  payload.Shift,
		}

	case "resize":
		width := 0
		height := 0
		if payload.Width != nil {
			width = *payload.Width
		}
		if payload.Height != nil {
			height = *payload.Height
		}
		return matrixcui.ResizeEvent{
			Width:  width,
			Height: height,
		}

	default:
		return nil
	}
}

// parseKey converts a string key name to matrixcui.Key.
func (ws *WebServer) parseKey(keyStr string) matrixcui.Key {
	switch keyStr {
	case "Enter":
		return matrixcui.KeyEnter
	case "Backspace":
		return matrixcui.KeyBackspace
	case "Tab":
		return matrixcui.KeyTab
	case "Escape":
		return matrixcui.KeyEscape
	case "Space":
		return matrixcui.KeySpace
	case "Up":
		return matrixcui.KeyUp
	case "Down":
		return matrixcui.KeyDown
	case "Left":
		return matrixcui.KeyLeft
	case "Right":
		return matrixcui.KeyRight
	case "Home":
		return matrixcui.KeyHome
	case "End":
		return matrixcui.KeyEnd
	case "PageUp":
		return matrixcui.KeyPageUp
	case "PageDown":
		return matrixcui.KeyPageDown
	case "Insert":
		return matrixcui.KeyInsert
	case "Delete":
		return matrixcui.KeyDelete
	case "F1":
		return matrixcui.KeyF1
	case "F2":
		return matrixcui.KeyF2
	case "F3":
		return matrixcui.KeyF3
	case "F4":
		return matrixcui.KeyF4
	case "F5":
		return matrixcui.KeyF5
	case "F6":
		return matrixcui.KeyF6
	case "F7":
		return matrixcui.KeyF7
	case "F8":
		return matrixcui.KeyF8
	case "F9":
		return matrixcui.KeyF9
	case "F10":
		return matrixcui.KeyF10
	case "F11":
		return matrixcui.KeyF11
	case "F12":
		return matrixcui.KeyF12
	default:
		return matrixcui.KeyNone
	}
}

// parseMouseButton converts a string button name to matrixcui.MouseButton.
func (ws *WebServer) parseMouseButton(buttonStr string) matrixcui.MouseButton {
	switch buttonStr {
	case "Left":
		return matrixcui.MouseButtonLeft
	case "Middle":
		return matrixcui.MouseButtonMiddle
	case "Right":
		return matrixcui.MouseButtonRight
	case "WheelUp":
		return matrixcui.MouseButtonWheelUp
	case "WheelDown":
		return matrixcui.MouseButtonWheelDown
	default:
		return matrixcui.MouseButtonNone
	}
}

// parseMouseAction converts a string action name to matrixcui.MouseAction.
func (ws *WebServer) parseMouseAction(actionStr string) matrixcui.MouseAction {
	switch actionStr {
	case "Press":
		return matrixcui.MouseActionPress
	case "Release":
		return matrixcui.MouseActionRelease
	case "Move":
		return matrixcui.MouseActionMove
	default:
		return matrixcui.MouseActionPress
	}
}

// executeCommand executes a command on the matrix.
func (ws *WebServer) executeCommand(cmd protocol.Command) protocol.Response {
	switch cmd.Cmd {
	case "put":
		if cmd.X == nil || cmd.Y == nil || cmd.Cell == nil {
			return protocol.NewResponseError(errors.New("put requires x, y, and cell"))
		}
		cell := protocol.CellFromProto(*cmd.Cell)
		if !ws.matrix.Put(*cmd.X, *cmd.Y, cell) {
			return protocol.NewResponseError(fmt.Errorf("out of bounds: (%d, %d)", *cmd.X, *cmd.Y))
		}
		return protocol.NewResponseOK()

	case "get":
		if cmd.X == nil || cmd.Y == nil {
			return protocol.NewResponseError(errors.New("get requires x and y"))
		}
		cell := ws.matrix.Get(*cmd.X, *cmd.Y)
		cellProto := protocol.CellToProto(cell)
		return protocol.NewResponseCell(cellProto)

	case "clear":
		ws.matrix.Clear()
		return protocol.NewResponseOK()

	case "fill":
		if cmd.Cell == nil {
			return protocol.NewResponseError(errors.New("fill requires cell"))
		}
		cell := protocol.CellFromProto(*cmd.Cell)
		ws.matrix.Fill(cell)
		return protocol.NewResponseOK()

	case "putString":
		if cmd.X == nil || cmd.Y == nil || cmd.Cell == nil {
			return protocol.NewResponseError(errors.New("putString requires x, y, text, and cell"))
		}
		cell := protocol.CellFromProto(*cmd.Cell)
		ws.matrix.PutString(*cmd.X, *cmd.Y, cmd.Text, cell)
		return protocol.NewResponseOK()

	case "resize":
		if cmd.Width == nil || cmd.Height == nil {
			return protocol.NewResponseError(errors.New("resize requires width and height"))
		}
		ws.matrix.Resize(*cmd.Width, *cmd.Height)
		return protocol.NewResponseOK()

	case "getSize":
		width, height := ws.matrix.Width(), ws.matrix.Height()
		return protocol.NewResponseSize(width, height)

	case "getSnapshot":
		width, height := ws.matrix.Width(), ws.matrix.Height()
		cells := protocol.MatrixToProto(ws.matrix)
		return protocol.NewResponseSnapshot(width, height, cells)

	default:
		return protocol.NewResponseError(fmt.Errorf("unknown command: %s", cmd.Cmd))
	}
}

// SocketSession represents a socket client connection.
type SocketSession struct {
	id     string
	conn   net.Conn
	server *WebServer

	subscribedMu sync.RWMutex
	subscribed   bool
	eventFilter  map[string]bool // Which event types to forward

	writerMu sync.Mutex // Protects writer from concurrent access
	writer   *bufio.Writer
	reader   *bufio.Reader
}

// NewSocketSession creates a new socket session.
func NewSocketSession(conn net.Conn, server *WebServer) *SocketSession {
	return &SocketSession{
		id:          uuid.New().String(),
		conn:        conn,
		server:      server,
		subscribed:  false,
		eventFilter: make(map[string]bool),
		writer:      bufio.NewWriter(conn),
		reader:      bufio.NewReader(conn),
	}
}

// Handle processes messages from the socket client.
func (sess *SocketSession) Handle() {
	defer sess.conn.Close()

	for {
		env, err := sess.receiveEnvelope()
		if err != nil {
			// Connection closed or error
			return
		}

		if env.Type != protocol.MessageTypeCommand {
			// Ignore non-command messages from client
			continue
		}

		cmd, err := protocol.UnwrapCommand(env)
		if err != nil {
			log.Printf("Failed to unwrap command: %v", err)
			continue
		}

		// Handle subscribe/unsubscribe specially
		if cmd.Cmd == "subscribe" {
			sess.handleSubscribe(env.ID, cmd)
			continue
		}
		if cmd.Cmd == "unsubscribe" {
			sess.handleUnsubscribe(env.ID, cmd)
			continue
		}

		// Execute command and send response
		resp := sess.server.executeCommand(cmd)
		respEnv, err := protocol.WrapResponse(env.ID, resp)
		if err != nil {
			log.Printf("Failed to wrap response: %v", err)
			continue
		}

		sess.writerMu.Lock()
		err = sess.sendEnvelope(respEnv)
		sess.writerMu.Unlock()

		if err != nil {
			log.Printf("Failed to send response: %v", err)
			return
		}
	}
}

// handleSubscribe processes a subscribe command.
func (sess *SocketSession) handleSubscribe(id string, cmd protocol.Command) {
	sess.subscribedMu.Lock()
	defer sess.subscribedMu.Unlock()

	sess.subscribed = true
	sess.eventFilter = make(map[string]bool)

	// If no events specified, subscribe to all
	if len(cmd.Events) == 0 {
		sess.eventFilter["key"] = true
		sess.eventFilter["mouse"] = true
		sess.eventFilter["resize"] = true
	} else {
		for _, eventType := range cmd.Events {
			sess.eventFilter[eventType] = true
		}
	}

	// Send OK response
	resp := protocol.NewResponseOK()
	respEnv, err := protocol.WrapResponse(id, resp)
	if err != nil {
		log.Printf("Failed to wrap subscribe response: %v", err)
		return
	}

	sess.writerMu.Lock()
	err = sess.sendEnvelope(respEnv)
	sess.writerMu.Unlock()

	if err != nil {
		log.Printf("Failed to send subscribe response: %v", err)
	}
}

// handleUnsubscribe processes an unsubscribe command.
func (sess *SocketSession) handleUnsubscribe(id string, cmd protocol.Command) {
	sess.subscribedMu.Lock()
	defer sess.subscribedMu.Unlock()

	sess.subscribed = false
	sess.eventFilter = make(map[string]bool)

	// Send OK response
	resp := protocol.NewResponseOK()
	respEnv, err := protocol.WrapResponse(id, resp)
	if err != nil {
		log.Printf("Failed to wrap unsubscribe response: %v", err)
		return
	}

	sess.writerMu.Lock()
	err = sess.sendEnvelope(respEnv)
	sess.writerMu.Unlock()

	if err != nil {
		log.Printf("Failed to send unsubscribe response: %v", err)
	}
}

// sendEvent sends an event to the socket client if subscribed.
func (sess *SocketSession) sendEvent(event matrixcui.Event) {
	sess.subscribedMu.RLock()
	defer sess.subscribedMu.RUnlock()

	if !sess.subscribed {
		return
	}

	// Convert event to protocol event payload
	var eventPayload protocol.EventPayload

	switch e := event.(type) {
	case matrixcui.KeyEvent:
		if !sess.eventFilter["key"] {
			return
		}
		eventPayload = protocol.NewKeyEventPayload(
			e.Key.String(),
			e.Rune,
			e.Alt,
			e.Ctrl,
			e.Shift,
		)

	case matrixcui.MouseEvent:
		if !sess.eventFilter["mouse"] {
			return
		}
		eventPayload = protocol.NewMouseEventPayload(
			e.X,
			e.Y,
			e.Button.String(),
			e.Action.String(),
			e.Alt,
			e.Ctrl,
			e.Shift,
		)

	case matrixcui.ResizeEvent:
		if !sess.eventFilter["resize"] {
			return
		}
		eventPayload = protocol.NewResizeEventPayload(e.Width, e.Height)

	default:
		return
	}

	// Wrap and send event
	env, err := protocol.WrapEvent(eventPayload)
	if err != nil {
		log.Printf("Failed to wrap event: %v", err)
		return
	}

	sess.writerMu.Lock()
	err = sess.sendEnvelope(env)
	sess.writerMu.Unlock()

	if err != nil {
		// If we can't send, the connection is probably dead
		// The session will be cleaned up when Handle() returns
		return
	}
}

// sendEnvelope sends an envelope to the socket client.
func (sess *SocketSession) sendEnvelope(env protocol.Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	// Write JSON data and check for short writes
	n, err := sess.writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("short write: wrote %d bytes, expected %d", n, len(data))
	}

	// Write newline
	if _, err := sess.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush to ensure data is sent
	if err := sess.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	return nil
}

// receiveEnvelope receives an envelope from the socket client.
func (sess *SocketSession) receiveEnvelope() (protocol.Envelope, error) {
	line, err := sess.reader.ReadString('\n')
	if err != nil {
		return protocol.Envelope{}, fmt.Errorf("failed to read line: %w", err)
	}

	var env protocol.Envelope
	if err := json.Unmarshal([]byte(line), &env); err != nil {
		return protocol.Envelope{}, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	return env, nil
}
