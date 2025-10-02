#!/bin/bash
set -e

# Build script for the complete Threadbound desktop app
# This builds both the Go backend and the Tauri frontend

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DESKTOP_DIR="$PROJECT_ROOT/desktop"

echo "🚀 Building Threadbound Desktop App..."

# Step 1: Build the Go backend
echo ""
echo "Step 1/2: Building Go backend..."
"$SCRIPT_DIR/build-backend.sh"

# Step 2: Build the Tauri app
echo ""
echo "Step 2/2: Building Tauri desktop app..."
cd "$DESKTOP_DIR"
npm run tauri build

echo ""
echo "✅ Build complete!"
echo ""
echo "📦 Your app bundle is ready in: $DESKTOP_DIR/src-tauri/target/release/bundle/"
