#!/bin/bash

# rq cross-platform build script
# Usage: ./build.sh

set -e

# Build configuration
APP_NAME="rq"
BUILD_DIR="dist"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Platform targets
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

echo "Building $APP_NAME v$VERSION for all platforms..."

# Clean and create build directory
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    
    output_name="$APP_NAME-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        output_name="$output_name.exe"
    fi
    
    echo "Building $output_name..."
    
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w -X main.version=$VERSION" \
        -o "$BUILD_DIR/$output_name" \
        .
done

echo ""
echo "Build complete! Binaries in $BUILD_DIR/:"
ls -la $BUILD_DIR/

echo ""
echo "File sizes:"
du -h $BUILD_DIR/*
