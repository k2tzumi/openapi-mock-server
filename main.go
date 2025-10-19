package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/k1LoW/httpstub"
	"github.com/k1LoW/nontest"
)

func main() {
	var (
		specFile string
		host     string
		port	 int
		baseURL     string
	)

	flag.StringVar(&specFile, "spec", "", "Path to OpenAPI specification file (required)")
	flag.StringVar(&specFile, "f", "", "Path to OpenAPI specification file (shorthand)")
	flag.StringVar(&host, "host", "localhost", "Host to bind the server to")
	flag.IntVar(&port, "port", 8000, "Port to bind the server to")
	flag.StringVar(&baseURL, "base-url", "/", "Base URL for the mock server")
	flag.Parse()

	// Check if spec file is provided
	if specFile == "" {
		fmt.Fprintf(os.Stderr, "Error: OpenAPI specification file is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if spec file exists
	if _, err := os.Stat(specFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: OpenAPI specification file not found: %s\n", specFile)
		os.Exit(1)
	}
	// Create nontest
	nt := nontest.New()

	addr := fmt.Sprintf("%s:%d", host, port)

	// Create httpstub router
	r := httpstub.NewRouter(
		nt,
		httpstub.OpenApi3(specFile),
		// httpstub.Addr(addr),
		httpstub.BaseURL(baseURL),
	)

	s := r.Server()

	// Start the server
	s.Start()

	fmt.Printf("Mock server started at http://%s\n", addr)
	fmt.Printf("OpenAPI spec: %s\n", specFile)
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	fmt.Println("\nShutting down server...")
}