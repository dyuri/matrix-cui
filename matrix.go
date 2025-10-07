package matrixcui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"
)

// Matrix represents an NxM character matrix for rendering.
type Matrix struct {
	width  int
	height int
	cells  [][]Cell
}

// NewMatrix creates a new matrix with the specified dimensions.
// All cells are initialized to EmptyCell().
func NewMatrix(width, height int) *Matrix {
	if width <= 0 || height <= 0 {
		panic("matrix dimensions must be positive")
	}

	m := &Matrix{
		width:  width,
		height: height,
		cells:  make([][]Cell, height),
	}

	for y := 0; y < height; y++ {
		m.cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			m.cells[y][x] = EmptyCell()
		}
	}

	return m
}

// NewMatrixAuto creates a matrix that fills the current terminal size.
// Falls back to 80x24 if terminal size cannot be detected.
func NewMatrixAuto() *Matrix {
	width, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		// Fallback to common default
		width, height = 80, 24
	}
	return NewMatrix(width, height)
}

// Width returns the width of the matrix.
func (m *Matrix) Width() int {
	return m.width
}

// Height returns the height of the matrix.
func (m *Matrix) Height() int {
	return m.height
}

// Put sets the cell at the specified position.
// Returns false if coordinates are out of bounds.
func (m *Matrix) Put(x, y int, cell Cell) bool {
	if !m.InBounds(x, y) {
		return false
	}
	m.cells[y][x] = cell
	return true
}

// Get retrieves the cell at the specified position.
// Returns EmptyCell if coordinates are out of bounds.
func (m *Matrix) Get(x, y int) Cell {
	if !m.InBounds(x, y) {
		return EmptyCell()
	}
	return m.cells[y][x]
}

// InBounds checks if the coordinates are within the matrix bounds.
func (m *Matrix) InBounds(x, y int) bool {
	return x >= 0 && x < m.width && y >= 0 && y < m.height
}

// Clear resets all cells to EmptyCell().
func (m *Matrix) Clear() {
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			m.cells[y][x] = EmptyCell()
		}
	}
}

// Fill fills the entire matrix with the specified cell.
func (m *Matrix) Fill(cell Cell) {
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			m.cells[y][x] = cell
		}
	}
}

// Resize changes the matrix dimensions, preserving existing content where possible.
// New areas are filled with EmptyCell().
func (m *Matrix) Resize(width, height int) {
	if width <= 0 || height <= 0 {
		panic("matrix dimensions must be positive")
	}

	newCells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		newCells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			if y < m.height && x < m.width {
				newCells[y][x] = m.cells[y][x]
			} else {
				newCells[y][x] = EmptyCell()
			}
		}
	}

	m.cells = newCells
	m.width = width
	m.height = height
}

// PutString writes a string starting at the specified position with the given styling.
// Returns the number of characters successfully written.
func (m *Matrix) PutString(x, y int, s string, cell Cell) int {
	count := 0
	for i, ch := range s {
		if !m.Put(x+i, y, cell.WithChar(ch)) {
			break
		}
		count++
	}
	return count
}

// Display renders the matrix to stdout using ANSI escape sequences.
func (m *Matrix) Display() error {
	output := m.Render()
	_, err := fmt.Print(output)
	return err
}

// DisplayAt renders the matrix at the specified terminal position.
// Position (1,1) is the top-left corner.
func (m *Matrix) DisplayAt(x, y int) error {
	// Move cursor to position
	fmt.Printf("\033[%d;%dH", y, x)
	return m.Display()
}

// Clone creates a deep copy of the matrix.
func (m *Matrix) Clone() *Matrix {
	clone := NewMatrix(m.width, m.height)
	for y := 0; y < m.height; y++ {
		copy(clone.cells[y], m.cells[y])
	}
	return clone
}
