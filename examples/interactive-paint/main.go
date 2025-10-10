package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	matrix "github.com/dyuri/matrix-cui"
)

// Color palette for painting
var colors = []lipgloss.Color{
	lipgloss.Color("#FF0000"), // 1: Red
	lipgloss.Color("#00FF00"), // 2: Green
	lipgloss.Color("#0000FF"), // 3: Blue
	lipgloss.Color("#FFFF00"), // 4: Yellow
	lipgloss.Color("#FF00FF"), // 5: Magenta
	lipgloss.Color("#00FFFF"), // 6: Cyan
	lipgloss.Color("#FFFFFF"), // 7: White
	lipgloss.Color("#FFA500"), // 8: Orange
	lipgloss.Color("#800080"), // 9: Purple
}

func main() {
	// Setup terminal for fullscreen TUI with mouse tracking
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

	// Enable mouse tracking
	if err := term.EnableMouseTracking(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enable mouse tracking: %v\n", err)
		os.Exit(1)
	}

	// Setup cleanup on interrupt signals
	term.SetupCleanupOnSignal()

	// Create matrix filling the terminal
	m := matrix.NewMatrixAuto()
	width, height := m.Width(), m.Height()

	// Initialize with black background
	clearCanvas(m, width, height)

	// Start event reader
	reader := matrix.NewEventReader()
	eventChan, cleanup := matrix.StartEventChannel(reader)
	defer cleanup()

	// Current drawing state
	currentColorIdx := 0
	isPainting := false

	// Main event loop
	for {
		select {
		case event := <-eventChan:
			// Handle mouse events
			if mouseEvent, ok := event.(matrix.MouseEvent); ok {
				switch mouseEvent.Action {
				case matrix.MouseActionPress:
					if mouseEvent.Button == matrix.MouseButtonLeft {
						isPainting = true
						paintCell(m, mouseEvent.X, mouseEvent.Y, colors[currentColorIdx])
					}
				case matrix.MouseActionRelease:
					if mouseEvent.Button == matrix.MouseButtonLeft {
						isPainting = false
					}
				case matrix.MouseActionMove:
					if isPainting {
						paintCell(m, mouseEvent.X, mouseEvent.Y, colors[currentColorIdx])
					}
				}
			}

			// Handle keyboard events
			if keyEvent, ok := event.(matrix.KeyEvent); ok {
				switch keyEvent.Key {
				case matrix.KeyEscape:
					return
				}

				// Handle character keys
				if keyEvent.Key == matrix.KeyNone {
					switch keyEvent.Rune {
					case 'q', 'Q':
						return
					case 'c', 'C':
						// Clear canvas
						clearCanvas(m, width, height)
					case '1', '2', '3', '4', '5', '6', '7', '8', '9':
						// Change color
						colorNum := int(keyEvent.Rune - '1')
						if colorNum >= 0 && colorNum < len(colors) {
							currentColorIdx = colorNum
						}
					}
				}
			}

			// Handle resize events
			if resizeEvent, ok := event.(matrix.ResizeEvent); ok {
				m.Resize(resizeEvent.Width, resizeEvent.Height)
				width, height = resizeEvent.Width, resizeEvent.Height
				clearCanvas(m, width, height)
			}

			// Draw status bar
			drawStatusBar(m, width, height, currentColorIdx, colors)

			// Render
			fmt.Print("\033[H")
			fmt.Print(m.Render())
		}
	}
}

func clearCanvas(m *matrix.Matrix, width, height int) {
	// Fill with black background
	emptyCell := matrix.NewCell(' ', lipgloss.Color(""), lipgloss.Color("#000000"))
	m.Fill(emptyCell)
}

func paintCell(m *matrix.Matrix, x, y int, color lipgloss.Color) {
	// Paint with a colored block character
	cell := matrix.NewCell('█', color, lipgloss.Color("#000000"))
	m.Put(x, y, cell)
}

func drawStatusBar(m *matrix.Matrix, width, height, colorIdx int, colors []lipgloss.Color) {
	// Don't draw if there's no room
	if height < 1 {
		return
	}

	// Clear status bar area (top line)
	statusCell := matrix.NewCell(' ', lipgloss.Color("#FFFFFF"), lipgloss.Color("#222222"))
	for x := 0; x < width; x++ {
		m.Put(x, 0, statusCell)
	}

	// Draw title
	title := "Interactive Paint"
	titleCell := matrix.NewCell(' ', lipgloss.Color("#FFFFFF"), lipgloss.Color("#222222")).AddStyle(matrix.StyleBold)
	m.PutString(2, 0, title, titleCell)

	// Draw instructions
	instructions := "1-9:Color | C:Clear | Q/ESC:Quit | Click+Drag:Draw"
	instrCell := matrix.NewCell(' ', lipgloss.Color("#AAAAAA"), lipgloss.Color("#222222"))
	instrX := width - len(instructions) - 2
	if instrX > len(title)+4 {
		m.PutString(instrX, 0, instructions, instrCell)
	}

	// Draw color palette preview
	if height >= 3 {
		paletteY := 1
		paletteX := 2

		// Label
		labelCell := matrix.NewCell(' ', lipgloss.Color("#FFFFFF"), lipgloss.Color("#000000"))
		m.PutString(paletteX, paletteY, "Colors: ", labelCell)
		paletteX += 8

		// Draw each color swatch
		for i, color := range colors {
			// Current color indicator
			char := '█'
			if i == colorIdx {
				// Highlight current color with brackets
				bracketCell := matrix.NewCell('[', lipgloss.Color("#FFFF00"), lipgloss.Color("#000000"))
				m.Put(paletteX, paletteY, bracketCell)
				paletteX++
			}

			// Color swatch
			colorCell := matrix.NewCell(char, color, lipgloss.Color("#000000"))
			m.Put(paletteX, paletteY, colorCell)
			paletteX++

			if i == colorIdx {
				bracketCell := matrix.NewCell(']', lipgloss.Color("#FFFF00"), lipgloss.Color("#000000"))
				m.Put(paletteX, paletteY, bracketCell)
				paletteX++
			}

			// Space between colors
			m.Put(paletteX, paletteY, labelCell)
			paletteX++
		}
	}
}
