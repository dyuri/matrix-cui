# Matrix CUI

A Go library for creating character-based user interfaces using an NxM matrix API, with ANSI terminal emulation.

## Overview

Matrix CUI provides a simple, modern API for building terminal UIs without dealing directly with ANSI escape sequences and terminal protocols. Instead of managing cursor positions and escape codes, you work with a 2D matrix of cells, each with its own character, colors, and styling.

This initial version serves as an emulation layer, converting the matrix API to classic ANSI terminal output using [Charm's Lipgloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Simple API**: Set cells by position with `Put(x, y, cell)`
- **Rich Styling**: Foreground/background colors, bold, italic, underline, and more
- **Flexible**: Create matrices of any size or auto-detect terminal dimensions
- **Efficient**: ANSI rendering optimized with Lipgloss
- **Lightweight**: Minimal dependencies

## Installation

```bash
go get github.com/dyuri/matrix-cui
```

## Quick Start

```go
package main

import (
    "github.com/charmbracelet/lipgloss"
    matrixcui "github.com/dyuri/matrix-cui"
)

func main() {
    // Create a 40x10 matrix
    m := matrixcui.NewMatrix(40, 10)

    // Create a styled cell
    cell := matrixcui.NewCell('X', lipgloss.Color("10"), lipgloss.Color("0"))
    cell = cell.AddStyle(matrixcui.StyleBold)

    // Put cell at position
    m.Put(5, 5, cell)

    // Write a string
    textCell := matrixcui.NewCell(' ', lipgloss.Color("15"), lipgloss.Color("99"))
    m.PutString(2, 2, "Hello, World!", textCell)

    // Display the matrix
    m.Display()
}
```

## API Reference

### Core Types

#### `Cell`
Represents a single character cell with styling.

```go
type Cell struct {
    Char  rune
    FG    lipgloss.Color // Foreground color
    BG    lipgloss.Color // Background color
    Style CellStyle      // Style flags (bold, italic, etc.)
}
```

#### `Matrix`
The main matrix type for managing cells.

```go
type Matrix struct {
    // private fields
}
```

### Creating Matrices

```go
// Fixed size matrix
m := matrixcui.NewMatrix(80, 24)

// Auto-detect terminal size
m := matrixcui.NewMatrixAuto()
```

### Working with Cells

```go
// Create cells
cell := matrixcui.NewCell('A', lipgloss.Color("10"), lipgloss.Color("0"))
styledCell := matrixcui.NewStyledCell('B', fg, bg, matrixcui.StyleBold)
empty := matrixcui.EmptyCell()

// Modify cells (returns new cell)
cell2 := cell.WithChar('X')
cell3 := cell.WithFG(lipgloss.Color("196"))
cell4 := cell.AddStyle(matrixcui.StyleItalic)

// Check styles
if cell.HasStyle(matrixcui.StyleBold) {
    // ...
}
```

### Available Styles

```go
matrixcui.StyleNormal
matrixcui.StyleBold
matrixcui.StyleItalic
matrixcui.StyleUnderline
matrixcui.StyleBlink
matrixcui.StyleReverse
```

### Matrix Operations

```go
// Set/Get cells
m.Put(x, y, cell)           // Returns bool (success)
cell := m.Get(x, y)         // Returns EmptyCell if out of bounds

// Write strings
m.PutString(x, y, "text", cell)  // Returns chars written

// Query
width := m.Width()
height := m.Height()
inBounds := m.InBounds(x, y)

// Modify matrix
m.Clear()                   // Reset to empty cells
m.Fill(cell)                // Fill with cell
m.Resize(width, height)     // Resize, preserving content

// Clone
m2 := m.Clone()
```

### Rendering

```go
// Render to string
output := m.Render()

// Display to terminal
m.Display()                 // Print at current position
m.DisplayAt(x, y)          // Print at specific position

// Partial rendering
m.RenderRegion(x, y, width, height)
m.RenderRow(y)
m.RenderCol(x)
```

## Examples

See the [examples/demo.go](examples/demo.go) file for a complete demonstration including:
- Basic text with colors
- Drawing boxes
- Multiple styles
- Simple animation

Run the demo:
```bash
cd examples
go run demo.go
```

## Colors

Matrix CUI uses Lipgloss colors, which supports:
- ANSI 256 colors: `lipgloss.Color("10")` (0-255)
- Hex colors: `lipgloss.Color("#FF5F87")`
- RGB: Via Lipgloss utilities

See [Lipgloss documentation](https://github.com/charmbracelet/lipgloss) for more color options.

## Roadmap

This is the first step in experimenting with the Matrix CUI API. Future plans include:

- [ ] Native rendering backends (bypass ANSI emulation)
- [ ] Event handling system
- [ ] Layout helpers and containers
- [ ] Diff-based rendering for efficiency
- [ ] Unicode and wide character support improvements
- [ ] Integration with other TUI frameworks

## Contributing

This is an experimental project. Feedback and contributions are welcome!

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Built with [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charm
- Inspired by modern terminal UI libraries and the need for simpler abstractions
