package protocol

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/charmbracelet/lipgloss"
	matrixcui "github.com/dyuri/matrix-cui"
)

func TestEnvelopeWrappingCommand(t *testing.T) {
	cmd := NewCommandPut(10, 5, CellProto{Char: "A", FG: "#00FF00", BG: "#000000", Style: 1})
	env, err := WrapCommand("test-id-123", cmd)
	if err != nil {
		t.Fatalf("WrapCommand failed: %v", err)
	}

	if env.Type != MessageTypeCommand {
		t.Errorf("expected type %s, got %s", MessageTypeCommand, env.Type)
	}
	if env.ID != "test-id-123" {
		t.Errorf("expected id 'test-id-123', got '%s'", env.ID)
	}

	// Unwrap and verify
	unwrapped, err := UnwrapCommand(env)
	if err != nil {
		t.Fatalf("UnwrapCommand failed: %v", err)
	}
	if unwrapped.Cmd != "put" {
		t.Errorf("expected cmd 'put', got '%s'", unwrapped.Cmd)
	}
	if *unwrapped.X != 10 {
		t.Errorf("expected x=10, got %d", *unwrapped.X)
	}
	if *unwrapped.Y != 5 {
		t.Errorf("expected y=5, got %d", *unwrapped.Y)
	}
	if unwrapped.Cell.Char != "A" {
		t.Errorf("expected char 'A', got '%s'", unwrapped.Cell.Char)
	}
}

func TestEnvelopeWrappingResponse(t *testing.T) {
	resp := NewResponseSize(80, 24)
	env, err := WrapResponse("test-id-456", resp)
	if err != nil {
		t.Fatalf("WrapResponse failed: %v", err)
	}

	if env.Type != MessageTypeResponse {
		t.Errorf("expected type %s, got %s", MessageTypeResponse, env.Type)
	}
	if env.ID != "test-id-456" {
		t.Errorf("expected id 'test-id-456', got '%s'", env.ID)
	}

	// Unwrap and verify
	unwrapped, err := UnwrapResponse(env)
	if err != nil {
		t.Fatalf("UnwrapResponse failed: %v", err)
	}
	if !unwrapped.OK {
		t.Errorf("expected OK=true, got false")
	}
	if *unwrapped.Width != 80 {
		t.Errorf("expected width=80, got %d", *unwrapped.Width)
	}
	if *unwrapped.Height != 24 {
		t.Errorf("expected height=24, got %d", *unwrapped.Height)
	}
}

func TestEnvelopeWrappingEvent(t *testing.T) {
	event := NewKeyEventPayload("Escape", 0, false, false, false)
	env, err := WrapEvent(event)
	if err != nil {
		t.Fatalf("WrapEvent failed: %v", err)
	}

	if env.Type != MessageTypeEvent {
		t.Errorf("expected type %s, got %s", MessageTypeEvent, env.Type)
	}
	if env.ID != "" {
		t.Errorf("expected empty id for event, got '%s'", env.ID)
	}

	// Unwrap and verify
	unwrapped, err := UnwrapEvent(env)
	if err != nil {
		t.Fatalf("UnwrapEvent failed: %v", err)
	}
	if unwrapped.Event != "key" {
		t.Errorf("expected event 'key', got '%s'", unwrapped.Event)
	}
	if unwrapped.Key != "Escape" {
		t.Errorf("expected key 'Escape', got '%s'", unwrapped.Key)
	}
}

func TestEnvelopeJSONRoundTrip(t *testing.T) {
	// Create a command envelope
	cmd := NewCommandClear()
	env, err := WrapCommand("json-test", cmd)
	if err != nil {
		t.Fatalf("WrapCommand failed: %v", err)
	}

	// Serialize to JSON
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Deserialize from JSON
	var decoded Envelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify
	if decoded.Type != MessageTypeCommand {
		t.Errorf("expected type %s, got %s", MessageTypeCommand, decoded.Type)
	}
	if decoded.ID != "json-test" {
		t.Errorf("expected id 'json-test', got '%s'", decoded.ID)
	}

	unwrapped, err := UnwrapCommand(decoded)
	if err != nil {
		t.Fatalf("UnwrapCommand failed: %v", err)
	}
	if unwrapped.Cmd != "clear" {
		t.Errorf("expected cmd 'clear', got '%s'", unwrapped.Cmd)
	}
}

func TestCellConversion(t *testing.T) {
	// Create a library cell
	cell := matrixcui.NewStyledCell(
		'A',
		lipgloss.Color("#00FF00"),
		lipgloss.Color("#000000"),
		matrixcui.StyleBold|matrixcui.StyleUnderline,
	)

	// Convert to proto
	proto := CellToProto(cell)
	if proto.Char != "A" {
		t.Errorf("expected char 'A', got '%s'", proto.Char)
	}
	if proto.FG != "#00FF00" {
		t.Errorf("expected fg '#00FF00', got '%s'", proto.FG)
	}
	if proto.BG != "#000000" {
		t.Errorf("expected bg '#000000', got '%s'", proto.BG)
	}
	expectedStyle := int(matrixcui.StyleBold | matrixcui.StyleUnderline)
	if proto.Style != expectedStyle {
		t.Errorf("expected style %d, got %d", expectedStyle, proto.Style)
	}

	// Convert back to library cell
	converted := CellFromProto(proto)
	if converted.Char != 'A' {
		t.Errorf("expected char 'A', got '%c'", converted.Char)
	}
	if converted.FG != lipgloss.Color("#00FF00") {
		t.Errorf("expected fg '#00FF00', got '%s'", converted.FG)
	}
	if converted.BG != lipgloss.Color("#000000") {
		t.Errorf("expected bg '#000000', got '%s'", converted.BG)
	}
	if converted.Style != (matrixcui.StyleBold | matrixcui.StyleUnderline) {
		t.Errorf("expected style %d, got %d", matrixcui.StyleBold|matrixcui.StyleUnderline, converted.Style)
	}
}

func TestCellProtoEmptyChar(t *testing.T) {
	// Test handling of empty char in proto
	proto := CellProto{Char: "", FG: "#FFFFFF", BG: "", Style: 0}
	cell := CellFromProto(proto)
	if cell.Char != ' ' {
		t.Errorf("expected space for empty char, got '%c'", cell.Char)
	}
}

func TestMatrixToProto(t *testing.T) {
	// Create a small test matrix
	m := matrixcui.NewMatrix(3, 2)
	m.Put(0, 0, matrixcui.NewCell('A', lipgloss.Color("#FF0000"), lipgloss.Color("")))
	m.Put(1, 0, matrixcui.NewCell('B', lipgloss.Color("#00FF00"), lipgloss.Color("")))
	m.Put(2, 1, matrixcui.NewCell('C', lipgloss.Color("#0000FF"), lipgloss.Color("")))

	// Convert to proto
	proto := MatrixToProto(m)

	// Verify dimensions
	if len(proto) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(proto))
	}
	if len(proto[0]) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(proto[0]))
	}

	// Verify cells
	if proto[0][0].Char != "A" {
		t.Errorf("expected cell[0][0] = 'A', got '%s'", proto[0][0].Char)
	}
	if proto[0][1].Char != "B" {
		t.Errorf("expected cell[0][1] = 'B', got '%s'", proto[0][1].Char)
	}
	if proto[1][2].Char != "C" {
		t.Errorf("expected cell[1][2] = 'C', got '%s'", proto[1][2].Char)
	}
}

func TestCommandHelpers(t *testing.T) {
	tests := []struct {
		name     string
		cmd      Command
		expected string
	}{
		{"Get", NewCommandGet(10, 5), "get"},
		{"Clear", NewCommandClear(), "clear"},
		{"Subscribe", NewCommandSubscribe([]string{"key", "mouse"}), "subscribe"},
		{"Unsubscribe", NewCommandUnsubscribe(), "unsubscribe"},
		{"GetSize", NewCommandGetSize(), "getSize"},
		{"GetSnapshot", NewCommandGetSnapshot(), "getSnapshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd.Cmd != tt.expected {
				t.Errorf("expected cmd '%s', got '%s'", tt.expected, tt.cmd.Cmd)
			}
		})
	}
}

func TestResponseHelpers(t *testing.T) {
	// OK response
	respOK := NewResponseOK()
	if !respOK.OK {
		t.Errorf("expected OK=true")
	}

	// Error response
	respErr := NewResponseError(errors.New("test error"))
	if respErr.OK {
		t.Errorf("expected OK=false")
	}
	if respErr.Error != "test error" {
		t.Errorf("expected error 'test error', got '%s'", respErr.Error)
	}

	// Cell response
	cell := CellProto{Char: "A", FG: "#00FF00", BG: "", Style: 0}
	respCell := NewResponseCell(cell)
	if !respCell.OK {
		t.Errorf("expected OK=true")
	}
	if respCell.Cell.Char != "A" {
		t.Errorf("expected cell char 'A', got '%s'", respCell.Cell.Char)
	}
}

func TestEventPayloadHelpers(t *testing.T) {
	// Key event
	keyEvent := NewKeyEventPayload("Enter", 13, false, true, false)
	if keyEvent.Event != "key" {
		t.Errorf("expected event 'key', got '%s'", keyEvent.Event)
	}
	if keyEvent.Key != "Enter" {
		t.Errorf("expected key 'Enter', got '%s'", keyEvent.Key)
	}
	if !keyEvent.Ctrl {
		t.Errorf("expected Ctrl=true")
	}

	// Mouse event
	mouseEvent := NewMouseEventPayload(10, 5, "Left", "Press", false, false, true)
	if mouseEvent.Event != "mouse" {
		t.Errorf("expected event 'mouse', got '%s'", mouseEvent.Event)
	}
	if *mouseEvent.X != 10 {
		t.Errorf("expected x=10, got %d", *mouseEvent.X)
	}
	if mouseEvent.Button != "Left" {
		t.Errorf("expected button 'Left', got '%s'", mouseEvent.Button)
	}
	if !mouseEvent.Shift {
		t.Errorf("expected Shift=true")
	}

	// Resize event
	resizeEvent := NewResizeEventPayload(100, 30)
	if resizeEvent.Event != "resize" {
		t.Errorf("expected event 'resize', got '%s'", resizeEvent.Event)
	}
	if *resizeEvent.Width != 100 {
		t.Errorf("expected width=100, got %d", *resizeEvent.Width)
	}
	if *resizeEvent.Height != 30 {
		t.Errorf("expected height=30, got %d", *resizeEvent.Height)
	}
}
