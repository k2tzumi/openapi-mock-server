# OpenAPI Mock Server CLI

A command-line tool that uses [k1LoW/nontest](https://github.com/k1LoW/nontest) and [k1LoW/httpstub](https://github.com/k1LoW/httpstub) to run an OpenAPI mock server.

## Installation

```bash
go get -u github.com/k1LoW/httpstub
go get -u github.com/k1LoW/nontest
go build -o openapi-mock-server main.go
```

## Usage

```bash
# Basic usage
./openapi-mock-server -spec openapi.yaml

# Specify custom port
./openapi-mock-server -spec openapi.yaml -port 3000

# Specify custom host
./openapi-mock-server -spec openapi.yaml -host 0.0.0.0

# Short options
./openapi-mock-server -f openapi.yaml -p 3000 -h localhost
```

## Command Line Options

- `-spec`, `-f`: Path to OpenAPI specification file (required)
- `-port`, `-p`: Port to run the mock server on (default: 8080)
- `-host`, `-h`: Host to bind the server to (default: localhost)

## Example

If you have an OpenAPI specification file `petstore.yaml`:

```bash
./openapi-mock-server -spec petstore.yaml -port 8080
```

This will start a mock server at `http://localhost:8080` that responds according to the OpenAPI specification.

## Features

- Automatically generates mock responses based on OpenAPI specification
- Supports OpenAPI 3.0 specifications
- Validates requests against the specification
- Returns example responses defined in the spec
- Supports path parameters, query parameters, and request bodies

## Stopping the Server

Press `Ctrl+C` to gracefully shut down the server.
