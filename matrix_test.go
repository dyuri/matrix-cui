package matrixcui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewMatrix(t *testing.T) {
	m := NewMatrix(10, 5)

	if m.Width() != 10 {
		t.Errorf("Expected width 10, got %d", m.Width())
	}
	if m.Height() != 5 {
		t.Errorf("Expected height 5, got %d", m.Height())
	}

	// All cells should be empty
	for y := 0; y < m.Height(); y++ {
		for x := 0; x < m.Width(); x++ {
			cell := m.Get(x, y)
			if cell.Char != ' ' {
				t.Errorf("Expected empty cell at (%d,%d), got char '%c'", x, y, cell.Char)
			}
		}
	}
}

func TestNewMatrixPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for zero width")
		}
	}()
	NewMatrix(0, 5)
}

func TestPutGet(t *testing.T) {
	m := NewMatrix(10, 10)
	cell := NewCell('X', lipgloss.Color("10"), lipgloss.Color("0"))

	if !m.Put(5, 5, cell) {
		t.Error("Put failed for valid coordinates")
	}

	retrieved := m.Get(5, 5)
	if retrieved.Char != 'X' {
		t.Errorf("Expected 'X', got '%c'", retrieved.Char)
	}
}

func TestPutOutOfBounds(t *testing.T) {
	m := NewMatrix(10, 10)
	cell := NewCell('X', lipgloss.Color("10"), lipgloss.Color("0"))

	if m.Put(-1, 5, cell) {
		t.Error("Put should fail for negative x")
	}
	if m.Put(5, -1, cell) {
		t.Error("Put should fail for negative y")
	}
	if m.Put(10, 5, cell) {
		t.Error("Put should fail for x >= width")
	}
	if m.Put(5, 10, cell) {
		t.Error("Put should fail for y >= height")
	}
}

func TestGetOutOfBounds(t *testing.T) {
	m := NewMatrix(10, 10)

	cell := m.Get(-1, 5)
	if cell.Char != ' ' {
		t.Error("Get should return EmptyCell for out of bounds")
	}

	cell = m.Get(10, 5)
	if cell.Char != ' ' {
		t.Error("Get should return EmptyCell for out of bounds")
	}
}

func TestInBounds(t *testing.T) {
	m := NewMatrix(10, 10)

	tests := []struct {
		x, y     int
		expected bool
	}{
		{0, 0, true},
		{9, 9, true},
		{5, 5, true},
		{-1, 5, false},
		{5, -1, false},
		{10, 5, false},
		{5, 10, false},
	}

	for _, tt := range tests {
		result := m.InBounds(tt.x, tt.y)
		if result != tt.expected {
			t.Errorf("InBounds(%d, %d) = %v, expected %v", tt.x, tt.y, result, tt.expected)
		}
	}
}

func TestClear(t *testing.T) {
	m := NewMatrix(10, 10)
	cell := NewCell('X', lipgloss.Color("10"), lipgloss.Color("0"))

	// Fill with non-empty cells
	for y := 0; y < m.Height(); y++ {
		for x := 0; x < m.Width(); x++ {
			m.Put(x, y, cell)
		}
	}

	m.Clear()

	// All cells should be empty
	for y := 0; y < m.Height(); y++ {
		for x := 0; x < m.Width(); x++ {
			c := m.Get(x, y)
			if c.Char != ' ' {
				t.Errorf("Expected empty cell at (%d,%d) after Clear", x, y)
			}
		}
	}
}

func TestFill(t *testing.T) {
	m := NewMatrix(10, 10)
	fillCell := NewCell('*', lipgloss.Color("10"), lipgloss.Color("0"))

	m.Fill(fillCell)

	for y := 0; y < m.Height(); y++ {
		for x := 0; x < m.Width(); x++ {
			c := m.Get(x, y)
			if c.Char != '*' {
				t.Errorf("Expected '*' at (%d,%d) after Fill, got '%c'", x, y, c.Char)
			}
		}
	}
}

func TestResize(t *testing.T) {
	m := NewMatrix(5, 5)
	cell := NewCell('X', lipgloss.Color("10"), lipgloss.Color("0"))
	m.Put(2, 2, cell)

	// Resize larger
	m.Resize(10, 10)
	if m.Width() != 10 || m.Height() != 10 {
		t.Error("Resize failed to update dimensions")
	}

	// Original content should be preserved
	c := m.Get(2, 2)
	if c.Char != 'X' {
		t.Error("Resize lost original content")
	}

	// New area should be empty
	c = m.Get(7, 7)
	if c.Char != ' ' {
		t.Error("New area after resize should be empty")
	}

	// Resize smaller
	m.Resize(3, 3)
	if m.Width() != 3 || m.Height() != 3 {
		t.Error("Resize smaller failed")
	}

	// Content within new bounds should be preserved
	c = m.Get(2, 2)
	if c.Char != 'X' {
		t.Error("Resize smaller lost content within new bounds")
	}
}

func TestPutString(t *testing.T) {
	m := NewMatrix(20, 5)
	cell := NewCell(' ', lipgloss.Color("10"), lipgloss.Color("0"))

	count := m.PutString(2, 2, "Hello", cell)
	if count != 5 {
		t.Errorf("Expected PutString to write 5 chars, got %d", count)
	}

	// Check each character
	expected := "Hello"
	for i, ch := range expected {
		c := m.Get(2+i, 2)
		if c.Char != ch {
			t.Errorf("Expected '%c' at position %d, got '%c'", ch, i, c.Char)
		}
	}
}

func TestPutStringOutOfBounds(t *testing.T) {
	m := NewMatrix(10, 5)
	cell := NewCell(' ', lipgloss.Color("10"), lipgloss.Color("0"))

	// String that extends beyond width
	count := m.PutString(8, 2, "HelloWorld", cell)
	if count != 2 {
		t.Errorf("Expected PutString to write 2 chars before boundary, got %d", count)
	}
}

func TestClone(t *testing.T) {
	m := NewMatrix(5, 5)
	cell := NewCell('X', lipgloss.Color("10"), lipgloss.Color("0"))
	m.Put(2, 2, cell)

	clone := m.Clone()

	// Clone should have same dimensions
	if clone.Width() != m.Width() || clone.Height() != m.Height() {
		t.Error("Clone has different dimensions")
	}

	// Clone should have same content
	if clone.Get(2, 2).Char != 'X' {
		t.Error("Clone missing content from original")
	}

	// Modifying clone should not affect original
	clone.Put(2, 2, NewCell('Y', lipgloss.Color("11"), lipgloss.Color("0")))
	if m.Get(2, 2).Char != 'X' {
		t.Error("Modifying clone affected original")
	}
}

func TestCellWithMethods(t *testing.T) {
	cell := NewCell('A', lipgloss.Color("10"), lipgloss.Color("0"))

	// Test WithChar
	cell2 := cell.WithChar('B')
	if cell2.Char != 'B' || cell.Char != 'A' {
		t.Error("WithChar not working correctly")
	}

	// Test WithFG
	cell3 := cell.WithFG(lipgloss.Color("20"))
	if cell3.FG != lipgloss.Color("20") || cell.FG != lipgloss.Color("10") {
		t.Error("WithFG not working correctly")
	}

	// Test WithBG
	cell4 := cell.WithBG(lipgloss.Color("30"))
	if cell4.BG != lipgloss.Color("30") {
		t.Error("WithBG not working correctly")
	}

	// Test WithStyle
	cell5 := cell.WithStyle(StyleBold)
	if !cell5.HasStyle(StyleBold) {
		t.Error("WithStyle not working correctly")
	}
}

func TestCellStyles(t *testing.T) {
	cell := NewCell('A', lipgloss.Color("10"), lipgloss.Color("0"))

	// Test AddStyle
	cell = cell.AddStyle(StyleBold)
	if !cell.HasStyle(StyleBold) {
		t.Error("AddStyle(StyleBold) failed")
	}

	cell = cell.AddStyle(StyleItalic)
	if !cell.HasStyle(StyleItalic) {
		t.Error("AddStyle(StyleItalic) failed")
	}

	// Should still have Bold
	if !cell.HasStyle(StyleBold) {
		t.Error("Lost StyleBold after adding StyleItalic")
	}

	// Test multiple styles
	if !cell.HasStyle(StyleBold | StyleItalic) {
		t.Error("HasStyle with multiple flags failed")
	}

	// Should not have other styles
	if cell.HasStyle(StyleUnderline) {
		t.Error("HasStyle returned true for style not added")
	}
}

func TestRender(t *testing.T) {
	m := NewMatrix(5, 3)
	cell := NewCell('X', lipgloss.Color(""), lipgloss.Color(""))
	m.Fill(cell)

	output := m.Render()

	// Should have 3 lines (height=3)
	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines in output, got %d", len(lines))
	}
}

func TestEmptyCell(t *testing.T) {
	empty := EmptyCell()
	if empty.Char != ' ' {
		t.Error("EmptyCell should have space character")
	}
	if empty.Style != StyleNormal {
		t.Error("EmptyCell should have StyleNormal")
	}
}

func TestNewStyledCell(t *testing.T) {
	cell := NewStyledCell('X', lipgloss.Color("10"), lipgloss.Color("0"), StyleBold|StyleItalic)

	if cell.Char != 'X' {
		t.Error("NewStyledCell char not set correctly")
	}
	if !cell.HasStyle(StyleBold) {
		t.Error("NewStyledCell missing StyleBold")
	}
	if !cell.HasStyle(StyleItalic) {
		t.Error("NewStyledCell missing StyleItalic")
	}
}
