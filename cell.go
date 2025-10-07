package matrixcui

import "github.com/charmbracelet/lipgloss"

// Cell represents a single cell in the matrix with character and styling.
type Cell struct {
	Char  rune
	FG    lipgloss.Color // Foreground color
	BG    lipgloss.Color // Background color
	Style CellStyle      // Additional style flags
}

// CellStyle represents text styling options (bold, italic, etc.)
type CellStyle uint8

const (
	StyleNormal CellStyle = 0
	StyleBold   CellStyle = 1 << iota
	StyleItalic
	StyleUnderline
	StyleBlink
	StyleReverse
)

// NewCell creates a new cell with the specified character and colors.
func NewCell(char rune, fg, bg lipgloss.Color) Cell {
	return Cell{
		Char:  char,
		FG:    fg,
		BG:    bg,
		Style: StyleNormal,
	}
}

// NewStyledCell creates a new cell with character, colors, and style.
func NewStyledCell(char rune, fg, bg lipgloss.Color, style CellStyle) Cell {
	return Cell{
		Char:  char,
		FG:    fg,
		BG:    bg,
		Style: style,
	}
}

// EmptyCell returns a default empty cell (space with no styling).
func EmptyCell() Cell {
	return Cell{
		Char:  ' ',
		FG:    lipgloss.Color(""),
		BG:    lipgloss.Color(""),
		Style: StyleNormal,
	}
}

// WithChar returns a copy of the cell with a different character.
func (c Cell) WithChar(char rune) Cell {
	c.Char = char
	return c
}

// WithFG returns a copy of the cell with a different foreground color.
func (c Cell) WithFG(fg lipgloss.Color) Cell {
	c.FG = fg
	return c
}

// WithBG returns a copy of the cell with a different background color.
func (c Cell) WithBG(bg lipgloss.Color) Cell {
	c.BG = bg
	return c
}

// WithStyle returns a copy of the cell with a different style.
func (c Cell) WithStyle(style CellStyle) Cell {
	c.Style = style
	return c
}

// AddStyle returns a copy of the cell with additional style flags.
func (c Cell) AddStyle(style CellStyle) Cell {
	c.Style |= style
	return c
}

// HasStyle checks if the cell has the specified style flag.
func (c Cell) HasStyle(style CellStyle) bool {
	return c.Style&style != 0
}
