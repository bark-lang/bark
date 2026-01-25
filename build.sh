#!/bin/bash
# Cross-platform build script for bark
# Builds executables for macOS, Linux, and Windows
# Creates both full (with SQL) and lite (without SQL) versions

set -e

VERSION=${1:-"dev"}
BUILD_DIR="dist"

echo "Building bark ${VERSION} for multiple platforms..."

# Create build directory
mkdir -p ${BUILD_DIR}

# Platforms to build
PLATFORMS=(
    "darwin arm64"
    "darwin amd64"
    "linux amd64"
    "linux arm64"
    "windows amd64"
)

# Build lite versions (without SQL - smaller binary)
echo ""
echo "=== Building LITE versions (without SQL) ==="
for platform in "${PLATFORMS[@]}"; do
    read -r os arch <<< "$platform"
    output="bark-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi
    echo "Building lite for ${os}/${arch}..."
    GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o ${BUILD_DIR}/${output} ./cmd/bark
done

# Build full versions (with SQL)
echo ""
echo "=== Building FULL versions (with SQL) ==="
for platform in "${PLATFORMS[@]}"; do
    read -r os arch <<< "$platform"
    output="bark-full-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi
    echo "Building full for ${os}/${arch}..."
    GOOS=$os GOARCH=$arch go build -tags=sql -ldflags="-s -w" -o ${BUILD_DIR}/${output} ./cmd/bark
done

echo ""
echo "Build complete! Binaries created in ${BUILD_DIR}/:"
ls -lh ${BUILD_DIR}/

echo ""
echo "To create release archives:"
echo "  # Lite versions"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-darwin-arm64.tar.gz bark-darwin-arm64"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-darwin-amd64.tar.gz bark-darwin-amd64"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-linux-amd64.tar.gz bark-linux-amd64"
echo "  cd ${BUILD_DIR} && tar -czf bark-${VERSION}-linux-arm64.tar.gz bark-linux-arm64"
echo "  cd ${BUILD_DIR} && zip bark-${VERSION}-windows-amd64.zip bark-windows-amd64.exe"
echo ""
echo "  # Full versions (with SQL)"
echo "  cd ${BUILD_DIR} && tar -czf bark-full-${VERSION}-darwin-arm64.tar.gz bark-full-darwin-arm64"
echo "  cd ${BUILD_DIR} && tar -czf bark-full-${VERSION}-darwin-amd64.tar.gz bark-full-darwin-amd64"
echo "  cd ${BUILD_DIR} && tar -czf bark-full-${VERSION}-linux-amd64.tar.gz bark-full-linux-amd64"
echo "  cd ${BUILD_DIR} && tar -czf bark-full-${VERSION}-linux-arm64.tar.gz bark-full-linux-arm64"
echo "  cd ${BUILD_DIR} && zip bark-full-${VERSION}-windows-amd64.zip bark-full-windows-amd64.exe"
