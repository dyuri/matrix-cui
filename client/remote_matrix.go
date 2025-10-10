package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"

	matrixcui "github.com/dyuri/matrix-cui"
	"github.com/dyuri/matrix-cui/protocol"
)

// RemoteMatrix is a client that connects to a Matrix server.
// It implements the same interface as Matrix for transparent remote usage.
type RemoteMatrix struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer

	pendingMu sync.Mutex
	pending   map[string]chan protocol.Response

	eventChan chan matrixcui.Event
	eventMu   sync.RWMutex

	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewMatrixRemote connects to a Matrix server at the given address.
// Address format: "unix:///path/to/socket" or "tcp://host:port"
func NewMatrixRemote(address string) (*RemoteMatrix, error) {
	// Parse address
	network, addr, err := parseAddress(address)
	if err != nil {
		return nil, err
	}

	// Connect
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	rm := &RemoteMatrix{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		writer:    bufio.NewWriter(conn),
		pending:   make(map[string]chan protocol.Response),
		eventChan: make(chan matrixcui.Event, 10),
		stopChan:  make(chan struct{}),
	}

	// Start receive loop
	rm.wg.Add(1)
	go func() {
		defer rm.wg.Done()
		rm.receiveLoop()
	}()

	return rm, nil
}

// Close closes the connection to the server.
func (rm *RemoteMatrix) Close() error {
	close(rm.stopChan)
	rm.wg.Wait()
	return rm.conn.Close()
}

// Width returns the width of the matrix.
func (rm *RemoteMatrix) Width() int {
	resp, err := rm.sendCommand(protocol.NewCommandGetSize())
	if err != nil || !resp.OK || resp.Width == nil {
		return 80 // Fallback
	}
	return *resp.Width
}

// Height returns the height of the matrix.
func (rm *RemoteMatrix) Height() int {
	resp, err := rm.sendCommand(protocol.NewCommandGetSize())
	if err != nil || !resp.OK || resp.Height == nil {
		return 24 // Fallback
	}
	return *resp.Height
}

// Put sets the cell at the specified position.
func (rm *RemoteMatrix) Put(x, y int, cell matrixcui.Cell) bool {
	cellProto := protocol.CellToProto(cell)
	resp, err := rm.sendCommand(protocol.NewCommandPut(x, y, cellProto))
	return err == nil && resp.OK
}

// Get retrieves the cell at the specified position.
func (rm *RemoteMatrix) Get(x, y int) matrixcui.Cell {
	resp, err := rm.sendCommand(protocol.NewCommandGet(x, y))
	if err != nil || !resp.OK || resp.Cell == nil {
		return matrixcui.EmptyCell()
	}
	return protocol.CellFromProto(*resp.Cell)
}

// InBounds checks if the coordinates are within the matrix bounds.
func (rm *RemoteMatrix) InBounds(x, y int) bool {
	width, height := rm.Width(), rm.Height()
	return x >= 0 && x < width && y >= 0 && y < height
}

// Clear resets all cells to EmptyCell().
func (rm *RemoteMatrix) Clear() {
	rm.sendCommand(protocol.NewCommandClear())
}

// Fill fills the entire matrix with the specified cell.
func (rm *RemoteMatrix) Fill(cell matrixcui.Cell) {
	cellProto := protocol.CellToProto(cell)
	rm.sendCommand(protocol.NewCommandFill(cellProto))
}

// Resize changes the matrix dimensions.
func (rm *RemoteMatrix) Resize(width, height int) {
	rm.sendCommand(protocol.NewCommandResize(width, height))
}

// PutString writes a string starting at the specified position.
func (rm *RemoteMatrix) PutString(x, y int, s string, cell matrixcui.Cell) int {
	cellProto := protocol.CellToProto(cell)
	resp, err := rm.sendCommand(protocol.NewCommandPutString(x, y, s, cellProto))
	if err != nil || !resp.OK {
		return 0
	}
	// For now, we assume all characters were written
	// Could extend protocol to return count
	return len(s)
}

// Subscribe starts receiving events from the server.
func (rm *RemoteMatrix) Subscribe(events []string) error {
	resp, err := rm.sendCommand(protocol.NewCommandSubscribe(events))
	if err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	return nil
}

// Unsubscribe stops receiving events from the server.
func (rm *RemoteMatrix) Unsubscribe() error {
	resp, err := rm.sendCommand(protocol.NewCommandUnsubscribe())
	if err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	return nil
}

// EventChannel returns a channel that receives events from the server.
// You must call Subscribe() before events will be received.
func (rm *RemoteMatrix) EventChannel() <-chan matrixcui.Event {
	return rm.eventChan
}

// sendCommand sends a command and waits for the response.
func (rm *RemoteMatrix) sendCommand(cmd protocol.Command) (protocol.Response, error) {
	id := uuid.New().String()

	// Register for response
	respChan := make(chan protocol.Response, 1)
	rm.pendingMu.Lock()
	rm.pending[id] = respChan
	rm.pendingMu.Unlock()

	// Cleanup
	defer func() {
		rm.pendingMu.Lock()
		delete(rm.pending, id)
		rm.pendingMu.Unlock()
		close(respChan)
	}()

	// Send command
	env, err := protocol.WrapCommand(id, cmd)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to wrap command: %w", err)
	}

	if err := rm.sendEnvelope(env); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respChan:
		return resp, nil
	case <-rm.stopChan:
		return protocol.Response{}, errors.New("connection closed")
	}
}

// sendEnvelope sends an envelope to the server.
func (rm *RemoteMatrix) sendEnvelope(env protocol.Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	if _, err := rm.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	if _, err := rm.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	if err := rm.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	return nil
}

// receiveLoop processes incoming messages from the server.
func (rm *RemoteMatrix) receiveLoop() {
	for {
		select {
		case <-rm.stopChan:
			return
		default:
		}

		line, err := rm.reader.ReadString('\n')
		if err != nil {
			// Connection closed or error
			return
		}

		var env protocol.Envelope
		if err := json.Unmarshal([]byte(line), &env); err != nil {
			continue
		}

		switch env.Type {
		case protocol.MessageTypeResponse:
			rm.handleResponse(env)
		case protocol.MessageTypeEvent:
			rm.handleEvent(env)
		}
	}
}

// handleResponse delivers a response to the waiting command.
func (rm *RemoteMatrix) handleResponse(env protocol.Envelope) {
	resp, err := protocol.UnwrapResponse(env)
	if err != nil {
		return
	}

	rm.pendingMu.Lock()
	respChan, ok := rm.pending[env.ID]
	rm.pendingMu.Unlock()

	if ok {
		respChan <- resp
	}
}

// handleEvent converts a protocol event to a library event and sends it to the channel.
func (rm *RemoteMatrix) handleEvent(env protocol.Envelope) {
	eventPayload, err := protocol.UnwrapEvent(env)
	if err != nil {
		return
	}

	var event matrixcui.Event

	switch eventPayload.Event {
	case "key":
		event = matrixcui.KeyEvent{
			Key:   parseKey(eventPayload.Key),
			Rune:  rune(eventPayload.Rune),
			Alt:   eventPayload.Alt,
			Ctrl:  eventPayload.Ctrl,
			Shift: eventPayload.Shift,
		}

	case "mouse":
		if eventPayload.X == nil || eventPayload.Y == nil {
			return
		}
		event = matrixcui.MouseEvent{
			X:      *eventPayload.X,
			Y:      *eventPayload.Y,
			Button: parseMouseButton(eventPayload.Button),
			Action: parseMouseAction(eventPayload.Action),
			Alt:    eventPayload.Alt,
			Ctrl:   eventPayload.Ctrl,
			Shift:  eventPayload.Shift,
		}

	case "resize":
		if eventPayload.Width == nil || eventPayload.Height == nil {
			return
		}
		event = matrixcui.ResizeEvent{
			Width:  *eventPayload.Width,
			Height: *eventPayload.Height,
		}

	default:
		return
	}

	select {
	case rm.eventChan <- event:
	case <-rm.stopChan:
	default:
		// Channel full, drop event
	}
}

// parseAddress parses an address string into network and address.
func parseAddress(address string) (network, addr string, err error) {
	// Simple parsing for "unix:///path" or "tcp://host:port"
	if len(address) > 7 && address[:7] == "unix://" {
		return "unix", address[7:], nil
	}
	if len(address) > 6 && address[:6] == "tcp://" {
		return "tcp", address[6:], nil
	}
	return "", "", fmt.Errorf("invalid address format: %s (expected unix:// or tcp://)", address)
}

// parseKey converts a string key name to a Key constant.
func parseKey(keyStr string) matrixcui.Key {
	// Map string names back to Key constants
	switch keyStr {
	case "None":
		return matrixcui.KeyNone
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
	case "Ctrl+C":
		return matrixcui.KeyCtrlC
	case "Ctrl+D":
		return matrixcui.KeyCtrlD
	case "Ctrl+Z":
		return matrixcui.KeyCtrlZ
	default:
		return matrixcui.KeyNone
	}
}

// parseMouseButton converts a string button name to a MouseButton constant.
func parseMouseButton(buttonStr string) matrixcui.MouseButton {
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

// parseMouseAction converts a string action name to a MouseAction constant.
func parseMouseAction(actionStr string) matrixcui.MouseAction {
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

// Stub methods to satisfy potential interfaces

// Display renders the matrix (not supported for remote, returns error).
func (rm *RemoteMatrix) Display() error {
	return errors.New("Display() not supported on RemoteMatrix - rendering happens server-side")
}

// DisplayAt renders the matrix at position (not supported for remote, returns error).
func (rm *RemoteMatrix) DisplayAt(x, y int) error {
	return errors.New("DisplayAt() not supported on RemoteMatrix - rendering happens server-side")
}

// Clone creates a copy (not supported for remote, returns nil).
func (rm *RemoteMatrix) Clone() *matrixcui.Matrix {
	return nil
}

// Render returns the ANSI string (not supported for remote, returns empty).
func (rm *RemoteMatrix) Render() string {
	return ""
}

// RenderRegion returns a region ANSI string (not supported for remote, returns empty).
func (rm *RemoteMatrix) RenderRegion(x, y, width, height int) string {
	return ""
}

// RenderRow returns a row ANSI string (not supported for remote, returns empty).
func (rm *RemoteMatrix) RenderRow(y int) string {
	return ""
}

// RenderCol returns a column ANSI string (not supported for remote, returns empty).
func (rm *RemoteMatrix) RenderCol(x int) string {
	return ""
}

// Ensure RemoteMatrix has a compatible interface subset
var _ interface {
	Width() int
	Height() int
	Put(x, y int, cell matrixcui.Cell) bool
	Get(x, y int) matrixcui.Cell
	InBounds(x, y int) bool
	Clear()
	Fill(cell matrixcui.Cell)
	Resize(width, height int)
	PutString(x, y int, s string, cell matrixcui.Cell) int
} = (*RemoteMatrix)(nil)
