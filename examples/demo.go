package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	matrixcui "github.com/dyuri/matrix-cui"
)

func main() {
	// Create a 40x10 matrix
	m := matrixcui.NewMatrix(40, 10)

	// Example 1: Basic text with colors
	titleCell := matrixcui.NewCell('H', lipgloss.Color("15"), lipgloss.Color("99"))
	titleCell = titleCell.AddStyle(matrixcui.StyleBold)

	m.PutString(2, 1, "Hello, Matrix CUI!", titleCell)

	// Example 2: Draw a colored box
	boxCell := matrixcui.NewCell('█', lipgloss.Color("10"), lipgloss.Color(""))
	for x := 2; x < 20; x++ {
		m.Put(x, 3, boxCell)
		m.Put(x, 7, boxCell)
	}
	for y := 3; y < 8; y++ {
		m.Put(2, y, boxCell)
		m.Put(19, y, boxCell)
	}

	// Example 3: Some styled text inside the box
	infoCell := matrixcui.NewCell(' ', lipgloss.Color("0"), lipgloss.Color("15"))
	m.PutString(4, 5, "Styled Text!", infoCell)

	// Example 4: Different colors
	colors := []string{"196", "208", "226", "46", "51", "21", "201"}
	y := 9
	for i, color := range colors {
		cell := matrixcui.NewCell('●', lipgloss.Color(color), lipgloss.Color(""))
		m.Put(2+i*2, y, cell)
	}

	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	// Display the matrix
	if err := m.Display(); err != nil {
		fmt.Printf("\nError displaying matrix: %v\n", err)
		return
	}

	// Example 5: Animation - update a cell over time
	fmt.Println("\n\n--- Animation Example ---")
	animMatrix := matrixcui.NewMatrix(20, 3)
	animMatrix.Fill(matrixcui.NewCell('.', lipgloss.Color("240"), lipgloss.Color("")))

	ballCell := matrixcui.NewCell('O', lipgloss.Color("226"), lipgloss.Color(""))
	ballCell = ballCell.AddStyle(matrixcui.StyleBold)

	for x := 0; x < 20; x++ {
		// Clear previous position
		if x > 0 {
			animMatrix.Put(x-1, 1, matrixcui.NewCell('.', lipgloss.Color("240"), lipgloss.Color("")))
		}
		// Draw new position
		animMatrix.Put(x, 1, ballCell)

		// Render and display
		fmt.Print("\033[2J\033[H")
		animMatrix.Display()
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n\nDemo complete!")
}
