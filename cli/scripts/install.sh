#!/usr/bin/env bash
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/guionardo/multi-git/main/cli/scripts/install.sh | bash
#   curl -fsSL .../install.sh | bash -s -- apply --dry-run
set -euo pipefail

REPO="${MULTI_GIT_REPO:-guionardo/multi-git}"
VERSION="${MULTI_GIT_VERSION:-latest}"
BASE="https://github.com/${REPO}/releases/${VERSION}/download"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac
case "$OS" in
  linux|darwin) ;;
  *) echo "unsupported os: $OS" >&2; exit 1 ;;
esac

BIN="multi-git_${OS}_${ARCH}"
if [ "$OS" = "windows" ]; then BIN="${BIN}.exe"; fi
URL="${BASE}/${BIN}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT
OUT="${TMPDIR}/multi-git"
echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "$OUT"
chmod +x "$OUT"

# Optional checksum verification
if curl -fsSL "${BASE}/SHA256SUMS" -o "${TMPDIR}/SHA256SUMS" 2>/dev/null; then
  (cd "$TMPDIR" && sha256sum -c SHA256SUMS 2>/dev/null | grep -q "$BIN: OK") || echo "checksum not verified (continuing)"
fi

exec "$OUT" "$@"