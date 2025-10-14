package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyuri/matrix-cui/examples/web-server/server"
)

func main() {
	// Parse command-line flags
	socketAddr := flag.String("socket", "unix:///tmp/matrix-cui-web.sock", "Socket address for clients (unix:///path or tcp://host:port)")
	httpAddr := flag.String("http", ":8080", "HTTP server address")
	staticDir := flag.String("static", "./static", "Static files directory")
	width := flag.Int("width", 80, "Matrix width")
	height := flag.Int("height", 24, "Matrix height")
	flag.Parse()

	// Create web server
	srv, err := server.NewWebServer(*socketAddr, *httpAddr, *staticDir, *width, *height)
	if err != nil {
		log.Fatalf("Failed to create web server: %v", err)
	}

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("Matrix CUI Web Server started")
	log.Printf("  Socket:   %s (for clients like remote-paint)", *socketAddr)
	log.Printf("  HTTP:     http://localhost%s (for browsers)", *httpAddr)
	log.Printf("  Matrix:   %dx%d", *width, *height)
	log.Printf("  Static:   %s", *staticDir)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down...")
	srv.Stop()
}
