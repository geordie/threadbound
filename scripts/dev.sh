#!/bin/bash
set -e

# Development script for Threadbound
# Builds the Go backend and starts Tauri in dev mode

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DESKTOP_DIR="$PROJECT_ROOT/desktop"

echo "🔧 Starting Threadbound in development mode..."

# Build the Go backend first
echo ""
echo "Building Go backend..."
"$SCRIPT_DIR/build-backend.sh"

# Start Tauri in dev mode
echo ""
echo "Starting Tauri dev server..."
cd "$DESKTOP_DIR"
npm run tauri dev
