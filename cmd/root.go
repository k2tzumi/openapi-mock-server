package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/k1LoW/httpstub"
	"github.com/k1LoW/nontest"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	specFile string
	host     string
	port     int
	basePath string
}

func NewRootCmd() *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:   "openapi-mock-server",
		Short: "Start an OpenAPI-based mock server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Start the server
			server, err := RunServer(cmd.Context(), opts)
			if err != nil {
				return err
			}

			// 2. Wait for the signal to shut down
			return WaitSignalAndShutdown(server)
			// This return value will be nil or return a shutdown error
		},
	}

	cmd.Flags().StringVarP(&opts.specFile, "spec", "f", "", "Path to OpenAPI specification file (required)")
	cmd.Flags().StringVar(&opts.host, "host", "localhost", "Host to bind the server to")
	cmd.Flags().IntVar(&opts.port, "port", 8080, "Port to bind the server to")
	cmd.Flags().StringVar(&opts.basePath, "base-path", "/", "Base path for the mock server")
	_ = cmd.MarkFlagRequired("spec")

	return cmd
}

// Runs and monitors the server, returning the started server instance and an error channel that notifies when the server stops.
func RunServer(ctx context.Context, opts *rootOptions) (*http.Server, error) {
	// Validate spec file
	if opts.specFile == "" {
		return nil, fmt.Errorf("OpenAPI specification file is required")
	}
	if _, err := os.Stat(opts.specFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("OpenAPI specification file not found: %s", opts.specFile)
	}

	nt := nontest.New()

	addr := fmt.Sprintf("%s:%d", opts.host, opts.port)

	basePath := opts.basePath
	if basePath == "" {
		basePath = "/"
	}
	if basePath[0] != '/' {
		basePath = "/" + basePath
	}

	// Create httpstub server
	ts := httpstub.NewServer(
		nt,
		httpstub.OpenApi3(opts.specFile),
		httpstub.Addr(addr),
		httpstub.BasePath(basePath),
	)
	nt.Cleanup(func() {
		if ts != nil {
			ts.Close()
		}
	})

	// Wrap handler with panic recovery
	var orig http.Handler
	if ts != nil {
		orig = ts
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// HTTP server with timeouts
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	if ts != nil {
		ts.ResponseExample()
	}

	// Start server (run in the background)
	serverErrCh := make(chan error, 1) // Channel for notifying server startup errors
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- fmt.Errorf("server error: %w", err)
		}
		close(serverErrCh) // Close the channel when the server has completely shut down
	}()

	fmt.Printf("Mock server started at http://%s%s\n", addr, basePath)
	fmt.Printf("OpenAPI spec: %s\n", opts.specFile)
	fmt.Println("Press Ctrl+C to stop the server")

	return server, nil
}

// WaitSignalAndShutdown waits for a signal and performs a graceful shutdown.
func WaitSignalAndShutdown(server *http.Server) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh // Wait for a signal

	fmt.Println("\nShutting down server...")

	// Graceful shutdown
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		// If Shutdown fails, force Close
		fmt.Fprintf(os.Stderr, "graceful shutdown failed: %v\n", err)
		_ = server.Close()
		return err
	}

	return nil
}
