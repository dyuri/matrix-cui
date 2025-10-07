package main

import (
	"fmt"
	"math/rand"
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
	// Create matrix filling the terminal
	m := matrix.NewMatrixAuto()

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

	// Clear screen and hide cursor
	fmt.Print("\033[2J\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	// Animation loop
	frame := 0
	for {
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

		// Render and display
		fmt.Print("\033[H") // Move cursor to home position
		fmt.Print(m.Render())

		// Sleep for animation timing
		time.Sleep(50 * time.Millisecond)
		frame++
	}
}
