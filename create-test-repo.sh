#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CACHE_DIR="$SCRIPT_DIR/testdata/test-repo"
dest="${1:-./test-repo}"

if [ -d "$dest" ]; then
  echo "Removing existing $dest"
  rm -rf "$dest"
fi

if [ ! -d "$CACHE_DIR" ]; then
  echo "Cache not found. Generating test repo..."
  bash "$SCRIPT_DIR/generate-test-repo.sh"
fi

cp -r "$CACHE_DIR" "$dest"

cd "$dest"
echo ""
echo "Created test repo at $dest"
echo "Branch: $(git branch --show-current)"
echo ""
echo "Commits to review (main..HEAD):"
git log --oneline main..HEAD
