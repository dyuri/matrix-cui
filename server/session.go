package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/google/uuid"

	matrixcui "github.com/dyuri/matrix-cui"
	"github.com/dyuri/matrix-cui/protocol"
)

// Session represents a client connection session.
type Session struct {
	id     string
	conn   net.Conn
	server *Server

	subscribedMu sync.RWMutex
	subscribed   bool
	eventFilter  map[string]bool // Which event types to forward

	writerMu sync.Mutex // Protects writer from concurrent access
	writer   *bufio.Writer
	reader   *bufio.Reader
}

// NewSession creates a new session for a connection.
func NewSession(conn net.Conn, server *Server) *Session {
	return &Session{
		id:          uuid.New().String(),
		conn:        conn,
		server:      server,
		subscribed:  false,
		eventFilter: make(map[string]bool),
		writer:      bufio.NewWriter(conn),
		reader:      bufio.NewReader(conn),
	}
}

// Handle processes messages from the client.
func (sess *Session) Handle() {
	defer sess.conn.Close()

	for {
		env, err := receiveEnvelope(sess.reader)
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
			fmt.Fprintf(os.Stderr, "Failed to unwrap command: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "Failed to wrap response: %v\n", err)
			continue
		}

		sess.writerMu.Lock()
		err = sendEnvelope(sess.writer, respEnv)
		sess.writerMu.Unlock()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
			return
		}
	}
}

// handleSubscribe processes a subscribe command.
func (sess *Session) handleSubscribe(id string, cmd protocol.Command) {
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
		fmt.Fprintf(os.Stderr, "Failed to wrap subscribe response: %v\n", err)
		return
	}

	sess.writerMu.Lock()
	err = sendEnvelope(sess.writer, respEnv)
	sess.writerMu.Unlock()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send subscribe response: %v\n", err)
	}
}

// handleUnsubscribe processes an unsubscribe command.
func (sess *Session) handleUnsubscribe(id string, cmd protocol.Command) {
	sess.subscribedMu.Lock()
	defer sess.subscribedMu.Unlock()

	sess.subscribed = false
	sess.eventFilter = make(map[string]bool)

	// Send OK response
	resp := protocol.NewResponseOK()
	respEnv, err := protocol.WrapResponse(id, resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to wrap unsubscribe response: %v\n", err)
		return
	}

	sess.writerMu.Lock()
	err = sendEnvelope(sess.writer, respEnv)
	sess.writerMu.Unlock()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send unsubscribe response: %v\n", err)
	}
}

// sendEvent sends an event to the client if subscribed.
func (sess *Session) sendEvent(event matrixcui.Event) {
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
		fmt.Fprintf(os.Stderr, "Failed to wrap event: %v\n", err)
		return
	}

	sess.writerMu.Lock()
	err = sendEnvelope(sess.writer, env)
	sess.writerMu.Unlock()

	if err != nil {
		// If we can't send, the connection is probably dead
		// The session will be cleaned up when Handle() returns
		return
	}
}
