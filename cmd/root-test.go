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
		baseURL:  "/",
	}
	t.Context()
	err := Run(context.Background(), opts)
	assert.Error(t, err)
	assert.Equal(t, "OpenAPI specification file is required", err.Error())
}

