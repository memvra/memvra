#!/usr/bin/env bash
# Build memvra for the current OS/arch.
set -euo pipefail

BINARY="memvra"
BUILD_DIR="dist"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

mkdir -p "${BUILD_DIR}"

echo "Building ${BINARY} ${VERSION}..."
CGO_ENABLED=1 go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/${BINARY}" ./cmd/memvra
echo "Done: ${BUILD_DIR}/${BINARY}"
