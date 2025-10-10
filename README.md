# Matrix CUI

A Go library for creating character-based user interfaces using an NxM matrix API, with ANSI terminal emulation.

## Overview

Matrix CUI provides a simple, modern API for building terminal UIs without dealing directly with ANSI escape sequences and terminal protocols. Instead of managing cursor positions and escape codes, you work with a 2D matrix of cells, each with its own character, colors, and styling.

This initial version serves as an emulation layer, converting the matrix API to classic ANSI terminal output using [Charm's Lipgloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Simple API**: Set cells by position with `Put(x, y, cell)`
- **Rich Styling**: Foreground/background colors, bold, italic, underline, and more
- **Flexible**: Create matrices of any size or auto-detect terminal dimensions
- **Event Handling**: Channel-based keyboard and mouse input
- **Fullscreen TUI**: Raw mode, alternate screen buffer, mouse tracking
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

## Event Handling

Matrix CUI provides a complete event handling system for interactive TUI applications.

### Terminal Setup

```go
// Create and setup terminal for fullscreen TUI
term, err := matrixcui.NewTerminal()
if err != nil {
    log.Fatal(err)
}
defer term.Close()

// Enter fullscreen mode (raw mode + alternate screen + hide cursor)
if err := term.SetupFullscreen(); err != nil {
    log.Fatal(err)
}

// Optional: Enable mouse tracking
if err := term.EnableMouseTracking(); err != nil {
    log.Fatal(err)
}

// Optional: Setup cleanup on Ctrl+C
term.SetupCleanupOnSignal()
```

### Event Loop

```go
// Create event reader
reader := matrixcui.NewEventReader()
eventChan, cleanup := matrixcui.StartEventChannel(reader)
defer cleanup()

// Event loop
for event := range eventChan {
    switch e := event.(type) {
    case matrixcui.KeyEvent:
        // Handle keyboard input
        if e.Key == matrixcui.KeyEscape {
            return // Exit
        }
        if e.Key == matrixcui.KeyNone {
            // Regular character
            fmt.Printf("Typed: %c\n", e.Rune)
        }

    case matrixcui.MouseEvent:
        // Handle mouse input
        if e.Action == matrixcui.MouseActionPress {
            fmt.Printf("Clicked at (%d, %d)\n", e.X, e.Y)
        }

    case matrixcui.ResizeEvent:
        // Handle terminal resize
        m.Resize(e.Width, e.Height)
    }
}
```

### Keyboard Events

```go
type KeyEvent struct {
    Key   Key    // Special key (KeyEnter, KeyEscape, KeyUp, etc.)
    Rune  rune   // Character typed (if printable)
    Alt   bool   // Alt modifier
    Ctrl  bool   // Ctrl modifier
    Shift bool   // Shift modifier
}
```

**Supported Keys:**
- Navigation: `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`, `KeyHome`, `KeyEnd`, `KeyPageUp`, `KeyPageDown`
- Control: `KeyEnter`, `KeyEscape`, `KeyBackspace`, `KeyTab`, `KeyDelete`, `KeyInsert`
- Function: `KeyF1` through `KeyF12`
- Shortcuts: `KeyCtrlC`, `KeyCtrlD`, `KeyCtrlZ`

### Mouse Events

```go
type MouseEvent struct {
    X      int          // Column position (0-based)
    Y      int          // Row position (0-based)
    Button MouseButton  // Which button
    Action MouseAction  // Press, Release, or Move
    Alt    bool         // Alt modifier
    Ctrl   bool         // Ctrl modifier
    Shift  bool         // Shift modifier
}
```

**Mouse Buttons:**
- `MouseButtonLeft`, `MouseButtonMiddle`, `MouseButtonRight`
- `MouseButtonWheelUp`, `MouseButtonWheelDown`

**Mouse Actions:**
- `MouseActionPress` - Button pressed
- `MouseActionRelease` - Button released
- `MouseActionMove` - Mouse moved while button held (drag)

### Complete Example

```go
package main

import (
    "fmt"
    "os"

    "github.com/charmbracelet/lipgloss"
    matrixcui "github.com/dyuri/matrix-cui"
)

func main() {
    // Setup terminal
    term, _ := matrixcui.NewTerminal()
    defer term.Close()
    term.SetupFullscreen()
    term.EnableMouseTracking()
    term.SetupCleanupOnSignal()

    // Create matrix
    m := matrixcui.NewMatrixAuto()

    // Event loop
    reader := matrixcui.NewEventReader()
    eventChan, cleanup := matrixcui.StartEventChannel(reader)
    defer cleanup()

    for event := range eventChan {
        if keyEvent, ok := event.(matrixcui.KeyEvent); ok {
            if keyEvent.Key == matrixcui.KeyEscape {
                return
            }
        }

        if mouseEvent, ok := event.(matrixcui.MouseEvent); ok {
            if mouseEvent.Action == matrixcui.MouseActionPress {
                // Paint on click
                cell := matrixcui.NewCell('█', lipgloss.Color("#00FF00"), lipgloss.Color("#000000"))
                m.Put(mouseEvent.X, mouseEvent.Y, cell)
            }
        }

        // Render
        fmt.Print("\033[H")
        fmt.Print(m.Render())
    }
}
```

## Examples

The repository includes several example programs demonstrating different features:

### Simple Demo (`examples/simple-demo/`)
Basic demonstration of the matrix API:
- Text rendering with colors
- Drawing boxes
- Multiple text styles
- Simple animation

```bash
cd examples/simple-demo
go run demo.go
```

### Matrix Rain (`examples/matrix-rain/`)
Interactive Matrix-style falling text animation:
- Full terminal size usage
- Keyboard event handling (ESC/Q to quit, +/- for speed, Space to pause)
- Terminal resize handling
- Real-time FPS display

```bash
cd examples/matrix-rain
go run main.go
```

### Interactive Paint (`examples/interactive-paint/`)
Mouse-driven drawing application:
- Click and drag to paint
- Number keys (1-9) to change colors
- 'C' to clear canvas
- Full mouse tracking with drag support

```bash
cd examples/interactive-paint
go run main.go
```

### Key Test (`examples/key-test/`)
Event debugging tool:
- Displays all keyboard and mouse events
- Useful for testing input handling
- Shows event details (modifiers, coordinates, etc.)

```bash
cd examples/key-test
go run main.go
```

## Colors

Matrix CUI uses Lipgloss colors, which supports:
- ANSI 256 colors: `lipgloss.Color("10")` (0-255)
- Hex colors: `lipgloss.Color("#FF5F87")`
- RGB: Via Lipgloss utilities

See [Lipgloss documentation](https://github.com/charmbracelet/lipgloss) for more color options.

## Roadmap

This is an experimental project under active development. Current status:

- [x] **Matrix API** - Simple 2D cell-based interface
- [x] **ANSI Rendering** - Lipgloss-based terminal output with raw mode support
- [x] **Event Handling** - Keyboard and mouse input via channels
- [x] **Terminal Management** - Raw mode, alternate screen, mouse tracking
- [ ] Native rendering backends (bypass ANSI emulation)
- [ ] Layout helpers and containers
- [ ] Diff-based rendering for efficiency
- [ ] Unicode and wide character support improvements
- [ ] Integration with other TUI frameworks

## Design Philosophy

Matrix CUI is designed to be:
- **Simple**: Intuitive API without forcing architectural patterns
- **Unopinionated**: Provides primitives, you build the patterns
- **Go-idiomatic**: Channel-based events, value semantics for cells
- **Minimal**: Small dependency footprint, focused scope

## Contributing

This is an experimental project. Feedback and contributions are welcome!

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Built with [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charm
- Inspired by modern terminal UI libraries and the need for simpler abstractions
