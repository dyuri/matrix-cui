package main

import (
	"fmt"
	"os"

	matrix "github.com/dyuri/matrix-cui"
)

func main() {
	// Setup terminal
	term, err := matrix.NewTerminal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create terminal: %v\n", err)
		os.Exit(1)
	}
	defer term.Close()

	if err := term.EnterRawMode(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enter raw mode: %v\n", err)
		os.Exit(1)
	}

	// Enable mouse tracking to test mouse events too
	if err := term.EnableMouseTracking(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to enable mouse tracking: %v\n", err)
	}

	fmt.Println("Key Test - Press keys or click mouse to see what is detected. Press Ctrl+C to exit.")
	fmt.Println("---")

	// Start event reader
	reader := matrix.NewEventReader()
	defer reader.Close()

	// Read events
	for {
		event, err := reader.ReadEvent()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\r\nError reading event: %v\n", err)
			break
		}

		switch e := event.(type) {
		case matrix.KeyEvent:
			if e.Key != matrix.KeyNone {
				fmt.Printf("\r\nKey: %s (Alt=%v, Ctrl=%v, Shift=%v)", e.Key.String(), e.Alt, e.Ctrl, e.Shift)
			} else {
				fmt.Printf("\r\nRune: %c (Alt=%v, Ctrl=%v)", e.Rune, e.Alt, e.Ctrl)
			}

			// Exit on Ctrl+C
			if e.Key == matrix.KeyCtrlC {
				fmt.Println("\r\nExiting...")
				return
			}
		case matrix.ResizeEvent:
			fmt.Printf("\r\nResize: %dx%d", e.Width, e.Height)
		case matrix.MouseEvent:
			fmt.Printf("\r\nMouse: %s %s at (%d,%d) [Alt=%v Ctrl=%v Shift=%v]",
				e.Button.String(), e.Action.String(), e.X, e.Y, e.Alt, e.Ctrl, e.Shift)
		}
	}
}
