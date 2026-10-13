#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Make sure we have an embedded frontend built
if [ ! -f "$PROJECT_ROOT/internal/web/dist/index.html" ]; then
	echo "Building frontend..."
	mise run build:embed
fi

# Pre-build both binaries so tests don't pay compilation cost
echo "Building test binaries..."
export LIF_BINARY="$PROJECT_ROOT/lif"
export LIF_SERVE_BINARY="$PROJECT_ROOT/lif-serve"
mise exec -- go build -o "$LIF_BINARY" "$PROJECT_ROOT/cmd/lif"
mise exec -- go build -o "$LIF_SERVE_BINARY" "$PROJECT_ROOT/cmd/lif-serve"

cd "$SCRIPT_DIR/.."

# Run Playwright tests
npx playwright test "$@"
