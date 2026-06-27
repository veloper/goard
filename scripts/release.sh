#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────────────────────────────────────
# Goard Release Script
# ──────────────────────────────────────────────────────────────────────────────
# Creates a semver release tag and pushes it to GitHub.
# The tag triggers .github/workflows/release.yml which:
#   1. Cross-compiles goard + goardctl for linux/darwin × amd64/arm64
#   2. Creates a GitHub Release with archives + checksums
#   3. Builds + pushes Docker image to veloper/goard:<tag> + latest
# ──────────────────────────────────────────────────────────────────────────────

if [ $# -ne 1 ]; then
  echo "Usage: $0 <version>"
  echo "  e.g. $0 v0.2.0"
  exit 1
fi

VERSION="$1"

# Validate semver format
if ! echo "$VERSION" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "Error: version must be in semver format (e.g. v0.2.0)"
  exit 1
fi

# Make sure we have the latest tags
git fetch --tags origin 2>/dev/null || true

# Find the highest existing semver tag (vX.Y.Z only)
LATEST=$(git tag -l 'v*' --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n1)

if [ -n "$LATEST" ]; then
  # Compare using sort -V (version sort)
  HIGHER=$(printf '%s\n%s\n' "$LATEST" "$VERSION" | sort -V | tail -n1)
  if [ "$HIGHER" != "$VERSION" ]; then
    echo "Error: $VERSION is not higher than latest tag $LATEST"
    echo "  Current latest: $LATEST"
    echo "  You tried:      $VERSION"
    exit 1
  fi
fi

echo "==> Tagging $VERSION..."
git tag "$VERSION"
git push origin "$VERSION"

echo ""
echo "✅ Release $VERSION pushed!"
echo "   Watch: https://github.com/veloper/goard/actions"
echo "   Docker: https://hub.docker.com/r/veloper/goard/tags"
