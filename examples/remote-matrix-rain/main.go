package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	matrix "github.com/dyuri/matrix-cui"
	"github.com/dyuri/matrix-cui/client"
)

// Column represents a falling stream of characters
type Column struct {
	x      int
	y      int
	speed  int
	length int
	chars  []rune
}

func main() {
	// Parse command-line flags
	socketAddr := flag.String("socket", "unix:///tmp/matrix-cui-web.sock", "Socket address to connect to")
	animationSpeed := flag.Duration("speed", 50*time.Millisecond, "Animation speed (lower is faster)")
	flag.Parse()

	// Connect to remote matrix server
	m, err := client.NewMatrixRemote(*socketAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure the web-server is running: cd examples/web-server && go run main.go\n")
		os.Exit(1)
	}
	defer m.Close()

	fmt.Fprintf(os.Stderr, "Connected to %s\n", *socketAddr)
	fmt.Fprintf(os.Stderr, "Matrix Rain animation running. Press Ctrl+C to quit.\n")
	fmt.Fprintf(os.Stderr, "Open http://localhost:8080 in your browser to view!\n")

	width, height := m.Width(), m.Height()

	// Character set for the rain
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789@#$%^&*()_+-=[]{}|;:,.<>?/~`")

	// Initialize columns
	columns := make([]*Column, width)
	for i := 0; i < width; i++ {
		columns[i] = &Column{
			x:      i,
			y:      -rand.Intn(height), // Start at random heights (some off-screen)
			speed:  1 + rand.Intn(3),   // Speed varies from 1-3
			length: 10 + rand.Intn(20), // Length varies from 10-30
			chars:  make([]rune, 30),
		}
		// Fill with random characters
		for j := range columns[i].chars {
			columns[i].chars[j] = chars[rand.Intn(len(chars))]
		}
	}

	// Color gradients for the rain (bright green at head, fading to dark)
	colors := []lipgloss.Color{
		lipgloss.Color("#FFFFFF"), // Bright white at very front
		lipgloss.Color("#CCFFCC"), // Very bright green
		lipgloss.Color("#66FF66"), // Bright green
		lipgloss.Color("#33FF33"), // Green
		lipgloss.Color("#00DD00"), // Medium green
		lipgloss.Color("#00AA00"), // Darker green
		lipgloss.Color("#008800"), // Even darker
		lipgloss.Color("#006600"), // Dark green
		lipgloss.Color("#004400"), // Very dark
		lipgloss.Color("#002200"), // Almost black
	}

	// Animation loop state
	frame := 0
	lastFrameTime := time.Now()
	fps := 0.0

	// Animation ticker
	ticker := time.NewTicker(*animationSpeed)
	defer ticker.Stop()

	// Setup signal handling for clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			// Ctrl+C or SIGTERM
			fmt.Fprintf(os.Stderr, "\nShutting down...\n")
			return

		case <-ticker.C:
			// Animation frame
			// Calculate FPS
			now := time.Now()
			frameDuration := now.Sub(lastFrameTime).Seconds()
			if frameDuration > 0 {
				fps = 1.0 / frameDuration
			}
			lastFrameTime = now

			// Clear matrix with black background
			m.Fill(matrix.NewCell(' ', lipgloss.Color(""), lipgloss.Color("#000000")))

			// Update and draw each column
			for _, col := range columns {
				// Occasionally change characters
				if rand.Float32() < 0.1 {
					idx := rand.Intn(len(col.chars))
					col.chars[idx] = chars[rand.Intn(len(chars))]
				}

				// Draw the column
				for i := 0; i < col.length; i++ {
					y := col.y - i
					if y >= 0 && y < height {
						// Determine color based on position in the stream
						colorIdx := i
						if colorIdx >= len(colors) {
							colorIdx = len(colors) - 1
						}

						char := col.chars[i%len(col.chars)]
						cell := matrix.NewCell(char, colors[colorIdx], lipgloss.Color("#000000"))

						// Make the head brighter/bold
						if i == 0 {
							cell = cell.AddStyle(matrix.StyleBold)
						}

						m.Put(col.x, y, cell)
					}
				}

				// Move column down
				if frame%col.speed == 0 {
					col.y++
				}

				// Reset column when it goes off screen
				if col.y-col.length > height {
					col.y = -rand.Intn(height / 2)
					col.speed = 1 + rand.Intn(3)
					col.length = 10 + rand.Intn(20)
				}
			}

			// Draw status bar in top-left corner
			statusText := fmt.Sprintf("Remote Matrix Rain | FPS: %.0f | %dx%d", fps, width, height)
			statusCell := matrix.NewCell(' ', lipgloss.Color("#00FF00"), lipgloss.Color("#000000")).
				AddStyle(matrix.StyleBold)
			m.PutString(1, 0, statusText, statusCell)

			frame++
		}
	}
}
