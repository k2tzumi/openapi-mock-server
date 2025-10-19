.PHONY: build run clean test deps

# Binary name
BINARY_NAME=openapi-mock-server

# Build the binary
build:
	go build -o $(BINARY_NAME) main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run with example OpenAPI spec
run: build
	./$(BINARY_NAME) -spec example-openapi.yaml

# Run with custom spec
run-custom: build
	@if [ -z "$(SPEC)" ]; then \
		echo "Usage: make run-custom SPEC=your-openapi-spec.yaml"; \
		exit 1; \
	fi
	./$(BINARY_NAME) -spec $(SPEC)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

# Test the server with curl examples
test-server:
	@echo "Testing GET /users..."
	curl -X GET http://localhost:8080/users
	@echo "\n\nTesting GET /users/1..."
	curl -X GET http://localhost:8080/users/1
	@echo "\n\nTesting POST /users..."
	curl -X POST http://localhost:8080/users \
		-H "Content-Type: application/json" \
		-d '{"name": "Test User", "email": "test@example.com"}'
	@echo "\n"
