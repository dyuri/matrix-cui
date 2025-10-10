package main

import (
	"fmt"
	"os"
	"time"

	matrix "github.com/dyuri/matrix-cui"
)

func main() {
	// Get size BEFORE terminal setup
	w1, h1, _ := matrix.GetTerminalSize()
	fmt.Printf("Size BEFORE terminal setup: %dx%d\n", w1, h1)
	time.Sleep(1 * time.Second)

	// Setup terminal
	term, err := matrix.NewTerminal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create terminal: %v\n", err)
		os.Exit(1)
	}
	defer term.Close()

	if err := term.SetupFullscreen(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup fullscreen: %v\n", err)
		os.Exit(1)
	}

	// Get size AFTER terminal setup
	w2, h2, _ := matrix.GetTerminalSize()

	// Print to alternate screen
	fmt.Printf("Size AFTER terminal setup: %dx%d\n", w2, h2)
	fmt.Printf("Press Enter to continue...")

	// Wait for enter
	var input string
	fmt.Scanln(&input)

	// Create matrix and test
	m := matrix.NewMatrixAuto()
	fmt.Printf("Matrix created with size: %dx%d\n", m.Width(), m.Height())

	// Wait for enter again
	fmt.Printf("Press Enter to exit...")
	fmt.Scanln(&input)
}
