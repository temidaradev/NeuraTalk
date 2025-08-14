#!/bin/bash

# Build script for NeuraTalk Web

set -e

echo "Building NeuraTalk Web..."

# Create web directory if it doesn't exist
mkdir -p web

# Build the WASM binary
echo "Building WASM binary..."
GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/web

# Build the server
echo "Building server..."
go build -o bin/neuratalk-server ./cmd/server

echo "Build complete!"
echo ""
echo "To run:"
echo "1. Start the server: ./bin/neuratalk-server"
echo "2. Open your browser to: http://localhost:8080"
echo ""
echo "Make sure Ollama is running locally before starting the server."
