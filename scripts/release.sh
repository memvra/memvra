#!/usr/bin/env bash
# Automates a GitHub release using goreleaser.
# Usage: ./scripts/release.sh v1.2.3
set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>" >&2
    echo "  Example: $0 v1.0.0" >&2
    exit 1
fi

VERSION="$1"

BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "${BRANCH}" != "main" ]; then
    echo "Must be on main branch (current: ${BRANCH})" >&2
    exit 1
fi

if ! git diff --quiet; then
    echo "Working tree is dirty. Commit or stash changes first." >&2
    exit 1
fi

echo "Tagging ${VERSION}..."
git tag -a "${VERSION}" -m "Release ${VERSION}"
git push origin "${VERSION}"

echo "Running goreleaser..."
goreleaser release --clean
