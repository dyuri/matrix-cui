package matrixcui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Render converts the matrix to a string with ANSI escape sequences.
// This is the emulation layer that converts our matrix API to classic terminal output.
func (m *Matrix) Render() string {
	var sb strings.Builder

	// Pre-allocate approximate size: each cell can be up to ~20 bytes with ANSI codes
	sb.Grow(m.width * m.height * 20)

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			cell := m.cells[y][x]
			sb.WriteString(renderCell(cell))
		}
		// Add newline except for the last row
		if y < m.height-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderCell converts a single cell to an ANSI-styled string using lipgloss.
func renderCell(cell Cell) string {
	// Start with the base style
	style := lipgloss.NewStyle()

	// Apply colors
	if cell.FG != "" {
		style = style.Foreground(cell.FG)
	}
	if cell.BG != "" {
		style = style.Background(cell.BG)
	}

	// Apply text styles
	if cell.HasStyle(StyleBold) {
		style = style.Bold(true)
	}
	if cell.HasStyle(StyleItalic) {
		style = style.Italic(true)
	}
	if cell.HasStyle(StyleUnderline) {
		style = style.Underline(true)
	}
	if cell.HasStyle(StyleBlink) {
		style = style.Blink(true)
	}
	if cell.HasStyle(StyleReverse) {
		style = style.Reverse(true)
	}

	return style.Render(string(cell.Char))
}

// RenderRegion renders a rectangular region of the matrix.
// Returns empty string if coordinates are invalid.
func (m *Matrix) RenderRegion(x, y, width, height int) string {
	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		return ""
	}
	if x >= m.width || y >= m.height {
		return ""
	}

	// Adjust dimensions if they exceed matrix bounds
	if x+width > m.width {
		width = m.width - x
	}
	if y+height > m.height {
		height = m.height - y
	}

	var sb strings.Builder
	sb.Grow(width * height * 20)

	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			cell := m.cells[y+row][x+col]
			sb.WriteString(renderCell(cell))
		}
		if row < height-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// RenderRow renders a single row of the matrix.
// Returns empty string if y is out of bounds.
func (m *Matrix) RenderRow(y int) string {
	if y < 0 || y >= m.height {
		return ""
	}
	return m.RenderRegion(0, y, m.width, 1)
}

// RenderCol renders a single column of the matrix.
// Returns empty string if x is out of bounds.
func (m *Matrix) RenderCol(x int) string {
	if x < 0 || x >= m.width {
		return ""
	}

	var sb strings.Builder
	sb.Grow(m.height * 20)

	for y := 0; y < m.height; y++ {
		cell := m.cells[y][x]
		sb.WriteString(renderCell(cell))
		if y < m.height-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
