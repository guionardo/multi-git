#!/usr/bin/env bash
# Short URL (GitHub Pages): https://guionardo.github.io/multi-git/install.sh
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
URL="${BASE}/${BIN}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT
OUT="${TMPDIR}/multi-git"
echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "$OUT"
chmod +x "$OUT"
exec "$OUT" "$@"