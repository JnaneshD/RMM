#!/usr/bin/env bash
set -e

OUTPUT_DIR="build"

echo "Compiling Linux binaries (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/linux/agent" ./cmd/agent
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/linux/server" ./cmd/server
echo "  > Linux build complete."

echo ""

echo "Compiling Windows binaries (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/windows/agent.exe" ./cmd/agent
GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/windows/server.exe" ./cmd/server
echo "  > Windows build complete."

echo ""
echo "All builds finished successfully! Binaries are located in '$OUTPUT_DIR'."
