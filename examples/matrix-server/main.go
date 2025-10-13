package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	matrixcui "github.com/dyuri/matrix-cui"
	"github.com/dyuri/matrix-cui/server"
)

const socketPath = "/tmp/matrix-cui.sock"

func main() {
	// Create terminal
	term, err := matrixcui.NewTerminal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create terminal: %v\n", err)
		os.Exit(1)
	}

	// Setup fullscreen mode
	if err := term.SetupFullscreen(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup fullscreen: %v\n", err)
		os.Exit(1)
	}

	// Enable mouse tracking
	if err := term.EnableMouseTracking(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enable mouse tracking: %v\n", err)
		os.Exit(1)
	}

	// Create auto-sized matrix
	m := matrixcui.NewMatrixAuto()

	// Create server
	srv := server.NewServer(m, term)

	// Start event forwarding
	reader := matrixcui.NewEventReader()
	eventChan, cleanup := matrixcui.StartEventChannel(reader)
	defer cleanup()
	srv.StartEventForwarding(eventChan)

	// Listen on Unix socket
	if err := srv.ListenUnix(socketPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on socket: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Server listening on %s\n", socketPath)

	// Render loop
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Get channel that signals when all clients disconnect
	noClientsChan := srv.NoClientsChannel()

	for {
		select {
		case <-ticker.C:
			// Render matrix to terminal
			output := m.Render()
			fmt.Print("\033[H") // Move cursor to home
			fmt.Print(output)

		case event := <-eventChan:
			// Check for quit (Ctrl+C or 'q')
			if keyEvent, ok := event.(matrixcui.KeyEvent); ok {
				if keyEvent.Key == matrixcui.KeyCtrlC || keyEvent.Rune == 'q' {
					goto cleanup
				}
			}

			// Handle resize events
			if resizeEvent, ok := event.(matrixcui.ResizeEvent); ok {
				m.Resize(resizeEvent.Width, resizeEvent.Height)
			}

		case <-noClientsChan:
			// All clients disconnected
			fmt.Fprintf(os.Stderr, "\nAll clients disconnected\n")
			goto cleanup

		case <-sigChan:
			goto cleanup
		}
	}

cleanup:
	// Restore terminal first (exit alt screen)
	term.Close()

	fmt.Fprintf(os.Stderr, "\nShutting down server...\n")
	srv.Stop()
	os.Remove(socketPath)
}
