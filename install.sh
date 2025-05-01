#!/bin/bash

set -e

APP_NAME="neuratalk"
OUTPUT_DIR="bin"

echo "🔍 Detecting OS and architecture..."
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Linux)
        GOOS=linux
        ;;
    Darwin)
        GOOS=darwin
        ;;
    *)
        echo "❌ Unsupported operating system: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64)
        GOARCH=amd64
        ;;
    arm64|aarch64)
        GOARCH=arm64
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "📦 Ensuring Go modules are ready..."
go mod tidy

echo "📁 Creating output directory: $OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

echo "🛠️ Building for $GOOS/$GOARCH..."
GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT_DIR/$APP_NAME" .

echo "✅ Build complete: $OUTPUT_DIR/$APP_NAME"
echo "🚀 To run: ./$OUTPUT_DIR/$APP_NAME"
