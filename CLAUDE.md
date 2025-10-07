# Matrix CUI - Project Context for Claude

## Project Overview

Matrix CUI is a Go library that provides a simplified API for creating character-based terminal user interfaces. Instead of dealing with ANSI escape sequences directly, developers work with an NxM matrix of cells.

### Goals

1. **Experiment with API design**: This is the first iteration to test the matrix-based API concept
2. **Emulation layer**: Currently emulates the matrix API using ANSI terminal output via Lipgloss
3. **Replace classic TUIs**: Eventually replace traditional TUI implementations that require manual ANSI sequence management
4. **Future-proof**: Design an API that can later support different rendering backends

### Design Philosophy

- **Simplicity first**: API should be intuitive - `Put(x, y, content)` instead of cursor movement
- **Cell-centric**: Each cell is independent with its own character, colors, and styles
- **Immutable patterns**: Cell methods return new cells rather than modifying in place
- **Bounds-safe**: Operations on out-of-bounds coordinates fail gracefully

## Architecture

### Core Components

1. **cell.go**: Defines `Cell` type and styling
   - `Cell` struct: character + foreground + background + style flags
   - Style constants: Bold, Italic, Underline, Blink, Reverse
   - Helper constructors and immutable modifier methods

2. **matrix.go**: The main `Matrix` type
   - 2D array of cells with width/height
   - CRUD operations: Put, Get, Clear, Fill, Resize
   - Utility: InBounds checking, PutString for text
   - Auto-sizing from terminal dimensions

3. **render.go**: ANSI rendering/emulation layer
   - Converts matrix to ANSI string using Lipgloss
   - Supports partial rendering (regions, rows, columns)
   - Applies all cell styling through Lipgloss style builder

4. **examples/demo.go**: Demonstration program
   - Shows basic text, boxes, colors, styling
   - Includes simple animation example

### Dependencies

- **Lipgloss** (`github.com/charmbracelet/lipgloss`): Color and ANSI styling
  - Chosen for: Lightweight, well-maintained, great API, Charm ecosystem
  - Used in: Cell rendering, color definitions

- **x/term** (`github.com/charmbracelet/x/term`): Terminal size detection
  - Used in: `NewMatrixAuto()` for auto-sizing

### Key Design Decisions

1. **Lipgloss over Bubbletea**: Bubbletea is a full framework with event loops - too heavyweight for just rendering
2. **Color type**: Uses `lipgloss.Color` directly for maximum flexibility (256-color, hex, etc.)
3. **Coordinate system**: (0,0) is top-left, x is horizontal, y is vertical
4. **Error handling**: Out-of-bounds operations return false/empty rather than panicking (except constructors)
5. **Immutable cells**: Cell modification methods return new cells to avoid shared state bugs

## Common Tasks

### Adding New Cell Styles

1. Add constant to `CellStyle` type in cell.go
2. Update `HasStyle()` and `AddStyle()` methods if needed
3. Add rendering logic in `renderCell()` in render.go
4. Update documentation in README.md

### Adding New Matrix Operations

1. Add method to Matrix type in matrix.go
2. Ensure bounds checking for safety
3. Add corresponding test in matrix_test.go
4. Document in README.md

### Adding New Rendering Modes

1. Add method to render.go
2. Use `renderCell()` for consistent styling
3. Consider performance for large matrices

## Testing

- Unit tests in matrix_test.go
- Test coverage priorities:
  1. Bounds checking (most critical for safety)
  2. Cell operations (immutability, style flags)
  3. Matrix operations (Put, Get, Resize)
  4. Rendering (correctness of ANSI output)

## Future Considerations

### Planned Features

1. **Native backends**: Skip ANSI emulation entirely
2. **Event system**: Mouse, keyboard input handling
3. **Layout engine**: Containers, flexbox-like layouts
4. **Diff rendering**: Only update changed cells
5. **Double buffering**: Smooth animations
6. **Wide char support**: Better Unicode/emoji handling

### API Stability

This is v0 - API may change based on usage feedback. Once we have real-world usage in the dependent projects, we'll stabilize to v1.

### Performance Notes

- Current implementation is not optimized - acceptable for initial experimentation
- For large matrices, consider lazy rendering or dirty tracking
- String builder pre-allocation helps but could be tuned

## Code Style

- Standard Go formatting (gofmt)
- Exported types/functions have doc comments
- Error returns for fallible operations (except constructors which panic)
- Prefer composition over inheritance
- Keep dependencies minimal

## Related Projects

Projects that will eventually use Matrix CUI:
- (List user's projects here as they adopt the library)

## Development Workflow

1. Make changes to library code
2. Run `go test ./...`
3. Test with `examples/demo.go`: `cd examples && go run demo.go`
4. Update README.md if API changed
5. Update this file if architecture changed

## Questions for Future

- Should we support z-index/layers?
- Do we need a cursor abstraction?
- Should cells support multiple characters (grapheme clusters)?
- Animation API - should it be built-in or separate?
- Terminal capability detection - how much to support?
