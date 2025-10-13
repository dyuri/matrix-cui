package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	matrixcui "github.com/dyuri/matrix-cui"
	"github.com/dyuri/matrix-cui/protocol"
)

// Server manages a Matrix and handles client connections.
type Server struct {
	matrix    *matrixcui.Matrix
	terminal  *matrixcui.Terminal
	eventChan <-chan matrixcui.Event

	listenerMu sync.Mutex
	listener   net.Listener

	sessionsMu    sync.RWMutex
	sessions      map[string]*Session
	hadClient     bool
	noClientsChan chan struct{} // Closed when all clients disconnect after having at least one

	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewServer creates a new server with the given matrix and terminal.
func NewServer(matrix *matrixcui.Matrix, terminal *matrixcui.Terminal) *Server {
	return &Server{
		matrix:        matrix,
		terminal:      terminal,
		sessions:      make(map[string]*Session),
		noClientsChan: make(chan struct{}),
		stopChan:      make(chan struct{}),
	}
}

// ListenUnix starts listening on a Unix socket.
func (s *Server) ListenUnix(socketPath string) error {
	// Remove existing socket file if it exists
	if err := os.RemoveAll(socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	s.listenerMu.Lock()
	s.listener = listener
	s.listenerMu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptLoop(listener)
	}()

	return nil
}

// StartEventForwarding starts forwarding terminal events to subscribed clients.
func (s *Server) StartEventForwarding(eventChan <-chan matrixcui.Event) {
	s.eventChan = eventChan
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.eventForwardLoop()
	}()
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	close(s.stopChan)

	// Close the listener to unblock Accept()
	s.listenerMu.Lock()
	if s.listener != nil {
		s.listener.Close()
	}
	s.listenerMu.Unlock()

	s.wg.Wait()
}

// NoClientsChannel returns a channel that gets closed when all clients disconnect
// after at least one client has connected.
func (s *Server) NoClientsChannel() <-chan struct{} {
	return s.noClientsChan
}

// acceptLoop accepts incoming connections.
func (s *Server) acceptLoop(listener net.Listener) {
	defer listener.Close()

	for {
		select {
		case <-s.stopChan:
			return
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return
			default:
				fmt.Fprintf(os.Stderr, "Accept error: %v\n", err)
				continue
			}
		}

		session := NewSession(conn, s)
		s.registerSession(session)

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			session.Handle()
			s.unregisterSession(session)
		}()
	}
}

// eventForwardLoop forwards events from the terminal to subscribed clients.
func (s *Server) eventForwardLoop() {
	if s.eventChan == nil {
		return
	}

	for {
		select {
		case <-s.stopChan:
			return
		case event, ok := <-s.eventChan:
			if !ok {
				return
			}
			s.broadcastEvent(event)
		}
	}
}

// registerSession adds a session to the server.
func (s *Server) registerSession(session *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[session.id] = session
	s.hadClient = true
}

// unregisterSession removes a session from the server.
func (s *Server) unregisterSession(session *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, session.id)

	// If we had a client and now have none, signal that all clients disconnected
	if s.hadClient && len(s.sessions) == 0 {
		select {
		case <-s.noClientsChan:
			// Already closed
		default:
			close(s.noClientsChan)
		}
	}
}

// broadcastEvent sends an event to all subscribed clients.
func (s *Server) broadcastEvent(event matrixcui.Event) {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	for _, session := range s.sessions {
		session.sendEvent(event)
	}
}

// executeCommand executes a command and returns a response.
func (s *Server) executeCommand(cmd protocol.Command) protocol.Response {
	switch cmd.Cmd {
	case "put":
		return s.cmdPut(cmd)
	case "get":
		return s.cmdGet(cmd)
	case "clear":
		return s.cmdClear(cmd)
	case "fill":
		return s.cmdFill(cmd)
	case "putString":
		return s.cmdPutString(cmd)
	case "resize":
		return s.cmdResize(cmd)
	case "getSize":
		return s.cmdGetSize(cmd)
	case "getSnapshot":
		return s.cmdGetSnapshot(cmd)
	default:
		return protocol.NewResponseError(fmt.Errorf("unknown command: %s", cmd.Cmd))
	}
}

func (s *Server) cmdPut(cmd protocol.Command) protocol.Response {
	if cmd.X == nil || cmd.Y == nil || cmd.Cell == nil {
		return protocol.NewResponseError(errors.New("put requires x, y, and cell"))
	}

	cell := protocol.CellFromProto(*cmd.Cell)
	if !s.matrix.Put(*cmd.X, *cmd.Y, cell) {
		return protocol.NewResponseError(fmt.Errorf("out of bounds: (%d, %d)", *cmd.X, *cmd.Y))
	}

	return protocol.NewResponseOK()
}

func (s *Server) cmdGet(cmd protocol.Command) protocol.Response {
	if cmd.X == nil || cmd.Y == nil {
		return protocol.NewResponseError(errors.New("get requires x and y"))
	}

	cell := s.matrix.Get(*cmd.X, *cmd.Y)
	cellProto := protocol.CellToProto(cell)
	return protocol.NewResponseCell(cellProto)
}

func (s *Server) cmdClear(cmd protocol.Command) protocol.Response {
	s.matrix.Clear()
	return protocol.NewResponseOK()
}

func (s *Server) cmdFill(cmd protocol.Command) protocol.Response {
	if cmd.Cell == nil {
		return protocol.NewResponseError(errors.New("fill requires cell"))
	}

	cell := protocol.CellFromProto(*cmd.Cell)
	s.matrix.Fill(cell)
	return protocol.NewResponseOK()
}

func (s *Server) cmdPutString(cmd protocol.Command) protocol.Response {
	if cmd.X == nil || cmd.Y == nil || cmd.Cell == nil {
		return protocol.NewResponseError(errors.New("putString requires x, y, text, and cell"))
	}

	cell := protocol.CellFromProto(*cmd.Cell)
	count := s.matrix.PutString(*cmd.X, *cmd.Y, cmd.Text, cell)
	_ = count // We don't return count in response yet, but could add it

	return protocol.NewResponseOK()
}

func (s *Server) cmdResize(cmd protocol.Command) protocol.Response {
	if cmd.Width == nil || cmd.Height == nil {
		return protocol.NewResponseError(errors.New("resize requires width and height"))
	}

	s.matrix.Resize(*cmd.Width, *cmd.Height)
	return protocol.NewResponseOK()
}

func (s *Server) cmdGetSize(cmd protocol.Command) protocol.Response {
	width, height := s.matrix.Width(), s.matrix.Height()
	return protocol.NewResponseSize(width, height)
}

func (s *Server) cmdGetSnapshot(cmd protocol.Command) protocol.Response {
	width, height := s.matrix.Width(), s.matrix.Height()
	cells := protocol.MatrixToProto(s.matrix)
	return protocol.NewResponseSnapshot(width, height, cells)
}

// sendEnvelope is a helper to send an envelope over a connection.
func sendEnvelope(writer *bufio.Writer, env protocol.Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	// Write JSON data and check for short writes
	n, err := writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("short write: wrote %d bytes, expected %d", n, len(data))
	}

	// Write newline
	if _, err := writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush to ensure data is sent
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	return nil
}

// receiveEnvelope is a helper to receive an envelope from a connection.
func receiveEnvelope(reader *bufio.Reader) (protocol.Envelope, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return protocol.Envelope{}, fmt.Errorf("failed to read line: %w", err)
	}

	var env protocol.Envelope
	if err := json.Unmarshal([]byte(line), &env); err != nil {
		return protocol.Envelope{}, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	return env, nil
}
