#!/bin/bash
set -e

# Build script for the Go backend
# This builds the backend binary for macOS (both Intel and Apple Silicon)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SRC_DIR="$PROJECT_ROOT/src"
OUTPUT_DIR="$PROJECT_ROOT/desktop/src-tauri/binaries"

echo "🔨 Building Go backend..."
echo "Project root: $PROJECT_ROOT"
echo "Output directory: $OUTPUT_DIR"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build for macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64 (Apple Silicon)..."
cd "$SRC_DIR"
GOOS=darwin GOARCH=arm64 go build -o "$OUTPUT_DIR/threadbound-aarch64-apple-darwin" ./cmd/threadbound
echo "✅ Built: threadbound-aarch64-apple-darwin"

# Build for macOS AMD64 (Intel)
echo "Building for macOS AMD64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o "$OUTPUT_DIR/threadbound-x86_64-apple-darwin" ./cmd/threadbound
echo "✅ Built: threadbound-x86_64-apple-darwin"

echo "✅ Backend build complete!"
