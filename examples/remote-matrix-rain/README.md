# Remote Matrix Rain

Non-interactive Matrix digital rain animation that connects to a Matrix CUI web-server. This is a demonstration client that automatically runs the animation without requiring any user input - perfect for testing the web-based interface.

## What it does

- Connects to a Matrix CUI server via socket (Unix or TCP)
- Runs a continuous Matrix-style digital rain animation
- Displays FPS counter
- No user interaction required - just runs automatically
- View the animation in your browser by connecting to the web-server

## Difference from matrix-rain

- **matrix-rain**: Uses Terminal for display, supports keyboard controls (pause, speed up/down, quit)
- **remote-matrix-rain**: Uses RemoteMatrix client, no keyboard controls, designed for web viewing

## Usage

### Quick Start

**Terminal 1** - Start web-server:
```bash
cd examples/web-server
go run main.go
# Web server started: socket=unix:///tmp/matrix-cui-web.sock http=:8080
```

**Terminal 2** - Start remote-matrix-rain:
```bash
cd examples/remote-matrix-rain
go run main.go
# Connected to unix:///tmp/matrix-cui-web.sock
# Matrix Rain animation running. Press Ctrl+C to quit.
# Open http://localhost:8080 in your browser to view!
```

**Browser** - Open http://localhost:8080

You should see the Matrix rain animation running in your browser! The animation will continue running until you press Ctrl+C in Terminal 2.

## Command-Line Options

```bash
go run main.go [options]

Options:
  -socket string
        Socket address to connect to (default "unix:///tmp/matrix-cui-web.sock")
        Examples:
          unix:///tmp/matrix-cui-web.sock
          tcp://localhost:9000

  -speed duration
        Animation speed (lower is faster) (default 50ms)
        Examples:
          30ms   (faster)
          100ms  (slower)
```

### Examples

**Connect to TCP server:**
```bash
go run main.go -socket=tcp://localhost:9000
```

**Faster animation:**
```bash
go run main.go -speed=30ms
```

**Slower animation:**
```bash
go run main.go -speed=100ms
```

## Building

```bash
go build
./remote-matrix-rain
```

## How It Works

1. Connects to a Matrix CUI server via socket (like web-server)
2. Gets the matrix dimensions from the server
3. Initializes falling "columns" of characters
4. Each frame:
   - Clears the matrix to black
   - Updates column positions
   - Draws each column with color gradient (bright at head, fading to dark)
   - Occasionally changes characters for variation
   - Sends all updates to the server
5. Server broadcasts updates to all connected browsers
6. Browser displays the animated matrix in real-time

## Technical Details

- **Animation Speed**: Default 50ms between frames (~20 FPS)
- **Column Count**: One column per matrix width
- **Column Length**: Random 10-30 characters
- **Column Speed**: Random 1-3 pixels per frame
- **Color Gradient**: 10 shades from white → bright green → dark green
- **Character Set**: A-Z, a-z, 0-9, and symbols

## Use Cases

- **Testing web-server**: Non-interactive client for testing without keyboard input
- **Demo**: Show off the web-based Matrix CUI viewer
- **Screensaver**: Run as a decorative animation
- **Performance testing**: Test how well the web-server handles continuous updates

## Troubleshooting

**Can't connect to server:**
- Make sure web-server is running first
- Check that socket address matches (default: unix:///tmp/matrix-cui-web.sock)
- For TCP, verify the port is correct

**Animation not visible:**
- Open http://localhost:8080 in your browser
- Check browser console for errors
- Verify web-server is running and shows "Socket client connected"

**Animation too fast/slow:**
- Use `-speed` flag to adjust (e.g., `-speed=100ms` for slower)

**Can't stop the animation:**
- Press Ctrl+C in the terminal running remote-matrix-rain

## Comparison with Other Examples

| Example | Display | Interaction | Use Case |
|---------|---------|-------------|----------|
| **matrix-rain** | Terminal | Keyboard controls | Standalone TUI demo |
| **remote-matrix-rain** | Browser (via web-server) | None | Web interface testing |
| **remote-paint** | Browser (via web-server) | Mouse/keyboard (Phase 2) | Interactive web client |
| **matrix-server** | Terminal | None (just forwards) | Socket server for clients |

## License

Same as Matrix CUI main project.
