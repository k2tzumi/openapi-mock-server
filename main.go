package main

import (
	"flag"
	"fmt"
	"net/http"
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
		port     int
		baseURL  string
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

	// baseURL must start with "/"
	if baseURL[0] != '/' {
		baseURL = "/" + baseURL
	}

	// Create httpstub server
	ts := httpstub.NewServer(
		nt,
		httpstub.OpenApi3(specFile),
		httpstub.Addr(addr),
		httpstub.BaseURL(baseURL),
	)
	nt.Cleanup(func() {
		if ts != nil {
			ts.Close()
		}
	})

	// If ts implements http.Handler (it does export ServeHTTP in the stack trace),
	// wrap it with a panic recovery middleware so handler goroutine panics become a 500 response.
	// Note: if NewServer already started its own net/http server internally, this wrapper may not be used.
	var handler http.Handler

	// Capture the original handler to avoid recursive closure calls.
	var orig http.Handler
	if ts != nil {
		orig = ts
	}

	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Fprintf(os.Stderr, "recovered panic in handler: %v\n", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		if orig == nil {
			http.Error(w, "no OpenAPI handler available", http.StatusInternalServerError)
			return
		}

		orig.ServeHTTP(w, r)
	})

	// Start a server using our wrapped handler only if NewServer did not already start one.
	// If NewServer already started listening, this ListenAndServe will fail with "address already in use".
	go func() {
		if err := http.ListenAndServe(addr, handler); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	// keep ResponseExample behavior if needed
	if ts != nil {
		ts.ResponseExample()
	}

	fmt.Printf("Mock server started at http://%s%s\n", addr, baseURL)
	fmt.Printf("OpenAPI spec: %s\n", specFile)
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	fmt.Println("\nShutting down server...")
}
