package protocol

import (
	"encoding/json"
	"fmt"
)

// MessageType identifies the category of message.
type MessageType string

const (
	MessageTypeCommand  MessageType = "command"
	MessageTypeResponse MessageType = "response"
	MessageTypeEvent    MessageType = "event"
)

// Envelope wraps all protocol messages.
type Envelope struct {
	Type    MessageType     `json:"type"`
	ID      string          `json:"id,omitempty"` // Present for commands/responses, omitted for events
	Payload json.RawMessage `json:"payload"`
}

// Command represents a request from client to server.
type Command struct {
	Cmd string `json:"cmd"`

	// Matrix operations
	X      *int       `json:"x,omitempty"`
	Y      *int       `json:"y,omitempty"`
	Cell   *CellProto `json:"cell,omitempty"`
	Text   string     `json:"text,omitempty"`
	Width  *int       `json:"width,omitempty"`
	Height *int       `json:"height,omitempty"`

	// Event subscription
	Events []string `json:"events,omitempty"`
}

// Response represents a reply from server to client.
type Response struct {
	OK     bool          `json:"ok"`
	Error  string        `json:"error,omitempty"`
	Cell   *CellProto    `json:"cell,omitempty"`
	Width  *int          `json:"width,omitempty"`
	Height *int          `json:"height,omitempty"`
	Cells  [][]CellProto `json:"cells,omitempty"`  // For full snapshot
	Delta  []CellUpdate  `json:"delta,omitempty"`  // For delta updates
}

// EventPayload represents an event broadcast from server to clients.
type EventPayload struct {
	Event string `json:"event"` // "key", "mouse", "resize"

	// KeyEvent fields
	Key   string `json:"key,omitempty"`
	Rune  int32  `json:"rune,omitempty"`
	Alt   bool   `json:"alt,omitempty"`
	Ctrl  bool   `json:"ctrl,omitempty"`
	Shift bool   `json:"shift,omitempty"`

	// MouseEvent fields
	X      *int   `json:"x,omitempty"`
	Y      *int   `json:"y,omitempty"`
	Button string `json:"button,omitempty"`
	Action string `json:"action,omitempty"`

	// ResizeEvent fields
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
}

// CellProto represents a cell in the protocol (JSON-serializable).
type CellProto struct {
	Char  string `json:"char"`
	FG    string `json:"fg"`
	BG    string `json:"bg"`
	Style int    `json:"style"`
}

// CellUpdate represents a single cell update for delta encoding.
type CellUpdate struct {
	X    int       `json:"x"`
	Y    int       `json:"y"`
	Cell CellProto `json:"cell"`
}

// Helper functions for creating commands

// NewCommandPut creates a Put command.
func NewCommandPut(x, y int, cell CellProto) Command {
	return Command{
		Cmd:  "put",
		X:    &x,
		Y:    &y,
		Cell: &cell,
	}
}

// NewCommandGet creates a Get command.
func NewCommandGet(x, y int) Command {
	return Command{
		Cmd: "get",
		X:   &x,
		Y:   &y,
	}
}

// NewCommandClear creates a Clear command.
func NewCommandClear() Command {
	return Command{Cmd: "clear"}
}

// NewCommandFill creates a Fill command.
func NewCommandFill(cell CellProto) Command {
	return Command{
		Cmd:  "fill",
		Cell: &cell,
	}
}

// NewCommandPutString creates a PutString command.
func NewCommandPutString(x, y int, text string, cell CellProto) Command {
	return Command{
		Cmd:  "putString",
		X:    &x,
		Y:    &y,
		Text: text,
		Cell: &cell,
	}
}

// NewCommandResize creates a Resize command.
func NewCommandResize(width, height int) Command {
	return Command{
		Cmd:    "resize",
		Width:  &width,
		Height: &height,
	}
}

// NewCommandSubscribe creates a Subscribe command.
func NewCommandSubscribe(events []string) Command {
	return Command{
		Cmd:    "subscribe",
		Events: events,
	}
}

// NewCommandUnsubscribe creates an Unsubscribe command.
func NewCommandUnsubscribe() Command {
	return Command{Cmd: "unsubscribe"}
}

// NewCommandGetSize creates a GetSize command.
func NewCommandGetSize() Command {
	return Command{Cmd: "getSize"}
}

// NewCommandGetSnapshot creates a GetSnapshot command.
func NewCommandGetSnapshot() Command {
	return Command{Cmd: "getSnapshot"}
}

// Helper functions for creating responses

// NewResponseOK creates a successful response.
func NewResponseOK() Response {
	return Response{OK: true}
}

// NewResponseError creates an error response.
func NewResponseError(err error) Response {
	return Response{OK: false, Error: err.Error()}
}

// NewResponseCell creates a response with cell data.
func NewResponseCell(cell CellProto) Response {
	return Response{OK: true, Cell: &cell}
}

// NewResponseSize creates a response with size data.
func NewResponseSize(width, height int) Response {
	return Response{OK: true, Width: &width, Height: &height}
}

// NewResponseSnapshot creates a response with full matrix snapshot.
func NewResponseSnapshot(width, height int, cells [][]CellProto) Response {
	return Response{OK: true, Width: &width, Height: &height, Cells: cells}
}

// NewResponseDelta creates a response with delta updates.
func NewResponseDelta(delta []CellUpdate) Response {
	return Response{OK: true, Delta: delta}
}

// Helper functions for creating event payloads

// NewKeyEventPayload creates a key event payload.
func NewKeyEventPayload(key string, r rune, alt, ctrl, shift bool) EventPayload {
	return EventPayload{
		Event: "key",
		Key:   key,
		Rune:  int32(r),
		Alt:   alt,
		Ctrl:  ctrl,
		Shift: shift,
	}
}

// NewMouseEventPayload creates a mouse event payload.
func NewMouseEventPayload(x, y int, button, action string, alt, ctrl, shift bool) EventPayload {
	return EventPayload{
		Event:  "mouse",
		X:      &x,
		Y:      &y,
		Button: button,
		Action: action,
		Alt:    alt,
		Ctrl:   ctrl,
		Shift:  shift,
	}
}

// NewResizeEventPayload creates a resize event payload.
func NewResizeEventPayload(width, height int) EventPayload {
	return EventPayload{
		Event:  "resize",
		Width:  &width,
		Height: &height,
	}
}

// Envelope construction helpers

// WrapCommand wraps a command in an envelope.
func WrapCommand(id string, cmd Command) (Envelope, error) {
	payload, err := json.Marshal(cmd)
	if err != nil {
		return Envelope{}, fmt.Errorf("failed to marshal command: %w", err)
	}
	return Envelope{
		Type:    MessageTypeCommand,
		ID:      id,
		Payload: payload,
	}, nil
}

// WrapResponse wraps a response in an envelope.
func WrapResponse(id string, resp Response) (Envelope, error) {
	payload, err := json.Marshal(resp)
	if err != nil {
		return Envelope{}, fmt.Errorf("failed to marshal response: %w", err)
	}
	return Envelope{
		Type:    MessageTypeResponse,
		ID:      id,
		Payload: payload,
	}, nil
}

// WrapEvent wraps an event in an envelope.
func WrapEvent(event EventPayload) (Envelope, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return Envelope{}, fmt.Errorf("failed to marshal event: %w", err)
	}
	return Envelope{
		Type:    MessageTypeEvent,
		Payload: payload,
	}, nil
}

// UnwrapCommand extracts a command from an envelope.
func UnwrapCommand(env Envelope) (Command, error) {
	if env.Type != MessageTypeCommand {
		return Command{}, fmt.Errorf("expected command, got %s", env.Type)
	}
	var cmd Command
	if err := json.Unmarshal(env.Payload, &cmd); err != nil {
		return Command{}, fmt.Errorf("failed to unmarshal command: %w", err)
	}
	return cmd, nil
}

// UnwrapResponse extracts a response from an envelope.
func UnwrapResponse(env Envelope) (Response, error) {
	if env.Type != MessageTypeResponse {
		return Response{}, fmt.Errorf("expected response, got %s", env.Type)
	}
	var resp Response
	if err := json.Unmarshal(env.Payload, &resp); err != nil {
		return Response{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return resp, nil
}

// UnwrapEvent extracts an event from an envelope.
func UnwrapEvent(env Envelope) (EventPayload, error) {
	if env.Type != MessageTypeEvent {
		return EventPayload{}, fmt.Errorf("expected event, got %s", env.Type)
	}
	var event EventPayload
	if err := json.Unmarshal(env.Payload, &event); err != nil {
		return EventPayload{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return event, nil
}
