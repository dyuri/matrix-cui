package protocol

import (
	"github.com/charmbracelet/lipgloss"
	matrixcui "github.com/dyuri/matrix-cui"
)

// CellToProto converts a library Cell to protocol CellProto.
func CellToProto(cell matrixcui.Cell) CellProto {
	return CellProto{
		Char:  string(cell.Char),
		FG:    string(cell.FG),
		BG:    string(cell.BG),
		Style: int(cell.Style),
	}
}

// CellFromProto converts a protocol CellProto to library Cell.
func CellFromProto(proto CellProto) matrixcui.Cell {
	var char rune = ' '
	if len(proto.Char) > 0 {
		char = []rune(proto.Char)[0]
	}

	return matrixcui.NewStyledCell(
		char,
		lipgloss.Color(proto.FG),
		lipgloss.Color(proto.BG),
		matrixcui.CellStyle(proto.Style),
	)
}

// MatrixToProto converts a full matrix to a 2D slice of CellProto.
func MatrixToProto(m *matrixcui.Matrix) [][]CellProto {
	width, height := m.Width(), m.Height()
	cells := make([][]CellProto, height)

	for y := 0; y < height; y++ {
		cells[y] = make([]CellProto, width)
		for x := 0; x < width; x++ {
			cells[y][x] = CellToProto(m.Get(x, y))
		}
	}

	return cells
}
