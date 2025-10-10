package server

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	matrixcui "github.com/dyuri/matrix-cui"
	"github.com/dyuri/matrix-cui/client"
)

func TestServerClientIntegration(t *testing.T) {
	t.Skip("Integration test - run manually")

	// Create a test matrix
	m := matrixcui.NewMatrix(20, 10)

	// Create server without terminal (for testing)
	srv := NewServer(m, nil)

	// Start server on test socket
	socketPath := "/tmp/matrix-cui-test.sock"
	if err := srv.ListenUnix(socketPath); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Connect client
	remote, err := client.NewMatrixRemote("unix://" + socketPath)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer remote.Close()

	// Test: Get size
	width := remote.Width()
	height := remote.Height()
	if width != 20 || height != 10 {
		t.Errorf("Expected size 20x10, got %dx%d", width, height)
	}

	// Test: Put and Get
	testCell := matrixcui.NewCell('A', lipgloss.Color("#FF0000"), lipgloss.Color("#000000"))
	if !remote.Put(5, 5, testCell) {
		t.Errorf("Put failed")
	}

	// Give time for command to process
	time.Sleep(10 * time.Millisecond)

	// Verify on server side
	serverCell := m.Get(5, 5)
	if serverCell.Char != 'A' {
		t.Errorf("Expected 'A', got '%c'", serverCell.Char)
	}

	// Test: Get from remote
	remoteCell := remote.Get(5, 5)
	if remoteCell.Char != 'A' {
		t.Errorf("Expected 'A' from remote, got '%c'", remoteCell.Char)
	}

	// Test: PutString
	styleCell := matrixcui.NewCell(' ', lipgloss.Color("#00FF00"), lipgloss.Color(""))
	remote.PutString(0, 0, "Hello", styleCell)
	time.Sleep(10 * time.Millisecond)

	// Verify on server
	if m.Get(0, 0).Char != 'H' {
		t.Errorf("Expected 'H' at (0,0)")
	}
	if m.Get(4, 0).Char != 'o' {
		t.Errorf("Expected 'o' at (4,0)")
	}

	// Test: Clear
	remote.Clear()
	time.Sleep(10 * time.Millisecond)

	// Verify all cells are empty
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			cell := m.Get(x, y)
			if cell.Char != ' ' {
				t.Errorf("Expected empty cell at (%d,%d), got '%c'", x, y, cell.Char)
			}
		}
	}

	// Test: Fill
	fillCell := matrixcui.NewCell('X', lipgloss.Color("#0000FF"), lipgloss.Color(""))
	remote.Fill(fillCell)
	time.Sleep(10 * time.Millisecond)

	// Verify all cells are 'X'
	if m.Get(10, 5).Char != 'X' {
		t.Errorf("Expected 'X' after fill, got '%c'", m.Get(10, 5).Char)
	}

	// Test: InBounds
	if !remote.InBounds(19, 9) {
		t.Errorf("InBounds(19, 9) should be true")
	}
	if remote.InBounds(20, 10) {
		t.Errorf("InBounds(20, 10) should be false")
	}

	// Test: Resize
	remote.Resize(30, 15)
	time.Sleep(10 * time.Millisecond)

	if m.Width() != 30 || m.Height() != 15 {
		t.Errorf("Expected size 30x15 after resize, got %dx%d", m.Width(), m.Height())
	}
}

func TestServerMultipleClients(t *testing.T) {
	t.Skip("Integration test - run manually")
	// Create a test matrix
	m := matrixcui.NewMatrix(10, 10)

	// Create server
	srv := NewServer(m, nil)

	// Start server
	socketPath := "/tmp/matrix-cui-test-multi.sock"
	if err := srv.ListenUnix(socketPath); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	time.Sleep(50 * time.Millisecond)

	// Connect two clients
	client1, err := client.NewMatrixRemote("unix://" + socketPath)
	if err != nil {
		t.Fatalf("Failed to connect client1: %v", err)
	}
	defer client1.Close()

	client2, err := client.NewMatrixRemote("unix://" + socketPath)
	if err != nil {
		t.Fatalf("Failed to connect client2: %v", err)
	}
	defer client2.Close()

	// Client 1 writes
	cell1 := matrixcui.NewCell('1', lipgloss.Color("#FF0000"), lipgloss.Color(""))
	client1.Put(0, 0, cell1)
	time.Sleep(10 * time.Millisecond)

	// Client 2 reads
	retrieved := client2.Get(0, 0)
	if retrieved.Char != '1' {
		t.Errorf("Client 2 should see '1' written by client 1, got '%c'", retrieved.Char)
	}

	// Client 2 writes
	cell2 := matrixcui.NewCell('2', lipgloss.Color("#00FF00"), lipgloss.Color(""))
	client2.Put(5, 5, cell2)
	time.Sleep(10 * time.Millisecond)

	// Client 1 reads
	retrieved = client1.Get(5, 5)
	if retrieved.Char != '2' {
		t.Errorf("Client 1 should see '2' written by client 2, got '%c'", retrieved.Char)
	}
}

func TestServerEventSubscription(t *testing.T) {
	t.Skip("Integration test - run manually")
	// Create a test matrix
	m := matrixcui.NewMatrix(10, 10)

	// Create server
	srv := NewServer(m, nil)

	// Create event channel for testing
	testEventChan := make(chan matrixcui.Event, 10)

	// Start event forwarding
	srv.StartEventForwarding(testEventChan)

	// Start server
	socketPath := "/tmp/matrix-cui-test-events.sock"
	if err := srv.ListenUnix(socketPath); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	time.Sleep(50 * time.Millisecond)

	// Connect client
	remote, err := client.NewMatrixRemote("unix://" + socketPath)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer remote.Close()

	// Subscribe to events
	if err := remote.Subscribe([]string{"key", "resize"}); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// Send a test event
	testEvent := matrixcui.KeyEvent{
		Key:  matrixcui.KeyEnter,
		Rune: 13,
	}
	testEventChan <- testEvent

	// Wait for event to be received by client
	select {
	case event := <-remote.EventChannel():
		keyEvent, ok := event.(matrixcui.KeyEvent)
		if !ok {
			t.Fatalf("Expected KeyEvent, got %T", event)
		}
		if keyEvent.Key != matrixcui.KeyEnter {
			t.Errorf("Expected KeyEnter, got %v", keyEvent.Key)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timeout waiting for event")
	}

	// Send another event
	resizeEvent := matrixcui.ResizeEvent{Width: 50, Height: 25}
	testEventChan <- resizeEvent

	// Wait for resize event
	select {
	case event := <-remote.EventChannel():
		resizeEv, ok := event.(matrixcui.ResizeEvent)
		if !ok {
			t.Fatalf("Expected ResizeEvent, got %T", event)
		}
		if resizeEv.Width != 50 || resizeEv.Height != 25 {
			t.Errorf("Expected resize to 50x25, got %dx%d", resizeEv.Width, resizeEv.Height)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timeout waiting for resize event")
	}

	// Unsubscribe
	if err := remote.Unsubscribe(); err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}
}
