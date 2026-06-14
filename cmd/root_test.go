package cmd

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun_MissingSpecFile(t *testing.T) {
	opts := &rootOptions{
		specFile: "",
		host:     "localhost",
		port:     8080,
		basePath: "/",
	}
	t.Context()
	_, err := RunServer(context.Background(), opts)
	assert.Error(t, err)
	assert.Equal(t, "OpenAPI specification file is required", err.Error())
}

func TestRun_EmptyBasePath(t *testing.T) {
	opts := &rootOptions{
		specFile: "../testdata/petstore.yaml",
		host:     "localhost",
		port:     8080,
		basePath: "",
	}
	_, err := RunServer(context.Background(), opts)
	assert.NoError(t, err)
}

func TestRunServer_StartsOnlyOnce(t *testing.T) {
	var mu sync.Mutex
	callCount := 0
	started := make(chan struct{}, 1)
	startFunc := func(server *http.Server) error {
		mu.Lock()
		callCount++
		mu.Unlock()

		select {
		case started <- struct{}{}:
		default:
		}

		return nil
	}

	opts := &rootOptions{
		specFile: "../testdata/petstore.yaml",
		host:     "localhost",
		port:     18080,
		basePath: "",
	}

	_, err := runServer(context.Background(), opts, startFunc)
	assert.NoError(t, err)

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not start")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, callCount)
}
