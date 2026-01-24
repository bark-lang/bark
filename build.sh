#!/bin/bash
# Cross-platform build script for Bark
# Builds executables for macOS, Linux, and Windows

set -e

VERSION=${1:-"dev"}
BUILD_DIR="dist"
CMD_DIR="cmd/bark"

echo "Building Bark v${VERSION} for multiple platforms..."

# Create build directory
mkdir -p ${BUILD_DIR}

# macOS Apple Silicon (M1/M2/M3)
echo "Building for macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ${BUILD_DIR}/bark-darwin-arm64 ${CMD_DIR}/main.go

# macOS Intel
echo "Building for macOS AMD64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/bark-darwin-amd64 ${CMD_DIR}/main.go

# Linux x86-64
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/bark-linux-amd64 ${CMD_DIR}/main.go

# Linux ARM64 (Raspberry Pi, AWS Graviton, etc.)
echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ${BUILD_DIR}/bark-linux-arm64 ${CMD_DIR}/main.go

# Windows x86-64
echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/bark-windows-amd64.exe ${CMD_DIR}/main.go

echo ""
echo "Build complete! Binaries created in ${BUILD_DIR}/:"
ls -lh ${BUILD_DIR}/

echo ""
echo "To create a release archive:"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-darwin-arm64.tar.gz bark-darwin-arm64"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-darwin-amd64.tar.gz bark-darwin-amd64"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-linux-amd64.tar.gz bark-linux-amd64"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-linux-arm64.tar.gz bark-linux-arm64"
echo "  cd ${BUILD_DIR} && zip bark-${VERSION}-windows-amd64.zip bark-windows-amd64.exe"
