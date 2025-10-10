package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	matrix "github.com/dyuri/matrix-cui"
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
	// Setup terminal for fullscreen TUI
	term, err := matrix.NewTerminal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create terminal: %v\n", err)
		os.Exit(1)
	}
	defer term.Close()

	if err := term.SetupFullscreen(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup fullscreen: %v\n", err)
		os.Exit(1)
	}

	// Setup cleanup on interrupt signals (Ctrl+C)
	term.SetupCleanupOnSignal()

	// Create matrix filling the terminal
	m := matrix.NewMatrixAuto()

	width, height := m.Width(), m.Height()

	// Start event reader
	reader := matrix.NewEventReader()
	eventChan, cleanup := matrix.StartEventChannel(reader)
	defer cleanup()

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
	animationSpeed := 50 * time.Millisecond // Base animation speed
	paused := false

	// Animation ticker
	ticker := time.NewTicker(animationSpeed)
	defer ticker.Stop()

	for {
		select {
		case event := <-eventChan:
			// Handle keyboard events
			if keyEvent, ok := event.(matrix.KeyEvent); ok {
				switch keyEvent.Key {
				case matrix.KeyEscape:
					// ESC to quit
					return
				case matrix.KeyUp, matrix.KeyRight:
					// Speed up (decrease delay)
					animationSpeed -= 10 * time.Millisecond
					if animationSpeed < 10*time.Millisecond {
						animationSpeed = 10 * time.Millisecond
					}
					ticker.Reset(animationSpeed)
				case matrix.KeyDown, matrix.KeyLeft:
					// Slow down (increase delay)
					animationSpeed += 10 * time.Millisecond
					if animationSpeed > 200*time.Millisecond {
						animationSpeed = 200 * time.Millisecond
					}
					ticker.Reset(animationSpeed)
				case matrix.KeySpace:
					// Toggle pause
					paused = !paused
				}

				// Handle character keys
				if keyEvent.Key == matrix.KeyNone {
					switch keyEvent.Rune {
					case 'p', 'P', ' ':
						// Toggle pause
						paused = !paused
					case 'q', 'Q':
						return
					case '+', '=':
						// Speed up
						animationSpeed -= 10 * time.Millisecond
						if animationSpeed < 10*time.Millisecond {
							animationSpeed = 10 * time.Millisecond
						}
						ticker.Reset(animationSpeed)
					case '-', '_':
						// Slow down
						animationSpeed += 10 * time.Millisecond
						if animationSpeed > 200*time.Millisecond {
							animationSpeed = 200 * time.Millisecond
						}
						ticker.Reset(animationSpeed)
					}
				}
			}

			// Handle resize events
			if resizeEvent, ok := event.(matrix.ResizeEvent); ok {
				// Only resize if dimensions actually changed
				if resizeEvent.Width != width || resizeEvent.Height != height {
					m.Resize(resizeEvent.Width, resizeEvent.Height)
					width, height = resizeEvent.Width, resizeEvent.Height
					// Reinitialize columns for new width
					newColumns := make([]*Column, width)
					for i := 0; i < width; i++ {
						if i < len(columns) {
							newColumns[i] = columns[i]
							newColumns[i].x = i // Update x position
						} else {
							newColumns[i] = &Column{
								x:      i,
								y:      -rand.Intn(height),
								speed:  1 + rand.Intn(3),
								length: 10 + rand.Intn(20),
								chars:  make([]rune, 30),
							}
							for j := range newColumns[i].chars {
								newColumns[i].chars[j] = chars[rand.Intn(len(chars))]
							}
						}
					}
					columns = newColumns
				}
			}

		case <-ticker.C:
			// Animation frame
			if paused {
				continue
			}
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

			// Draw FPS counter and controls in top-right corner
			statusText := fmt.Sprintf("FPS: %.0f | Speed: %dms | Size: %dx%d | ESC/Q:Quit | +/-:Speed | Space:Pause",
				fps, animationSpeed.Milliseconds(), width, height)
			statusCell := matrix.NewCell(' ', lipgloss.Color("#FFFFFF"), lipgloss.Color("#000000"))
			statusX := width - len(statusText)
			if statusX < 0 {
				statusX = 0
			}
			if statusX < width {
				m.PutString(statusX, 0, statusText, statusCell)
			}

			// Show pause indicator
			if paused {
				pauseText := "PAUSED"
				pauseCell := matrix.NewCell(' ', lipgloss.Color("#FF0000"), lipgloss.Color("#000000")).
					AddStyle(matrix.StyleBold)
				pauseX := (width - len(pauseText)) / 2
				if pauseX >= 0 && height > 0 {
					m.PutString(pauseX, height/2, pauseText, pauseCell)
				}
			}

			// Render and display
			fmt.Print("\033[H") // Move cursor to home position
			fmt.Print(m.Render())

			frame++
		}
	}
}
