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
    baseURL  string
}

func NewRootCmd() *cobra.Command {
    opts := &rootOptions{}
    cmd := &cobra.Command{
        Use:   "openapi-mock-server",
        Short: "Start an OpenAPI-based mock server",
        RunE: func(cmd *cobra.Command, args []string) error {
            return Run(cmd.Context(), opts)
        },
    }

    cmd.Flags().StringVarP(&opts.specFile, "spec", "f", "", "Path to OpenAPI specification file (required)")
    cmd.Flags().StringVar(&opts.host, "host", "localhost", "Host to bind the server to")
    cmd.Flags().IntVar(&opts.port, "port", 8080, "Port to bind the server to")
    cmd.Flags().StringVar(&opts.baseURL, "base-url", "/", "Base URL for the mock server")
    _ = cmd.MarkFlagRequired("spec")

    return cmd
}

func Run(ctx context.Context, opts *rootOptions) error {
    // Validate spec file
    if opts.specFile == "" {
        return fmt.Errorf("OpenAPI specification file is required")
    }
    if _, err := os.Stat(opts.specFile); os.IsNotExist(err) {
        return fmt.Errorf("OpenAPI specification file not found: %s", opts.specFile)
    }

    nt := nontest.New()

    addr := fmt.Sprintf("%s:%d", opts.host, opts.port)

    baseURL := opts.baseURL
    if baseURL == "" {
        baseURL = "/"
    }
    if baseURL[0] != '/' {
        baseURL = "/" + baseURL
    }

    // Create httpstub server
    ts := httpstub.NewServer(
        nt,
        httpstub.OpenApi3(opts.specFile),
        httpstub.Addr(addr),
        httpstub.BaseURL(baseURL),
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

    fmt.Printf("Mock server started at http://%s%s\n", addr, baseURL)
    fmt.Printf("OpenAPI spec: %s\n", opts.specFile)
    fmt.Println("Press Ctrl+C to stop the server")

    // Wait for interrupt
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    <-sigCh
    fmt.Println("\nShutting down server...")

    // Graceful shutdown
    ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Shutdown(ctxShutdown); err != nil {
        fmt.Fprintf(os.Stderr, "graceful shutdown failed: %v\n", err)
        _ = server.Close()
    }

    return nil
}