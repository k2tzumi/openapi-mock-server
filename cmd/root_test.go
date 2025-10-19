package cmd

import (
	"context"
	"testing"

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
