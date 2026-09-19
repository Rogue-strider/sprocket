#!/usr/bin/env bash
# sprocket installer — macOS, Linux, and Windows (Git Bash / WSL).
#
# Stages the prebuilt binary matching this machine's OS/arch as
# bin/sprocket-{learn,validate,hook}.exe (the .exe suffix is kept on every
# platform on purpose — Linux/macOS execute it fine since they don't care
# about file extensions, and it lets hooks.json reference one fixed path
# regardless of OS). No Go toolchain is required to install or run this.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Linux)              goos="linux" ;;
  Darwin)             goos="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) goos="windows" ;;  # Git Bash / MSYS2 on Windows
  *)                  echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64)  goarch="amd64" ;;
  arm64|aarch64) goarch="arm64" ;;
  *)             echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

platform="${goos}-${goarch}"
echo "==> Detected platform: $platform"

# Prebuilt Windows binaries are named with a .exe suffix already
# (sprocket-learn-windows-amd64.exe); Linux/macOS ones have no suffix.
src_ext=""
[ "$goos" = "windows" ] && src_ext=".exe"

for cmd in sprocket-learn sprocket-validate sprocket-hook; do
  src="$BIN_DIR/${cmd}-${platform}${src_ext}"
  dest="$BIN_DIR/${cmd}.exe"
  if [ ! -f "$src" ]; then
    echo "No prebuilt binary for $platform ($src missing)." >&2
    echo "If you have Go installed, build it yourself:" >&2
    echo "  go build -o $dest ./cmd/$cmd" >&2
    exit 1
  fi
  cp "$src" "$dest"
  chmod +x "$dest"
  echo "==> Staged $dest"
done

if ! command -v claude >/dev/null 2>&1; then
  echo "Claude Code CLI not found on PATH." >&2
  echo "Install it first: https://docs.claude.com/en/docs/claude-code" >&2
  exit 1
fi

echo "==> Registering local marketplace at $SCRIPT_DIR"
claude plugin marketplace add "$SCRIPT_DIR"

echo "==> Installing sprocket plugin"
claude plugin install sprocket@sprocket-marketplace

echo ""
echo "Installed. Restart Claude Code (or start a new session) in any project to pick it up."
echo "Try: /plan, /review, /fix-build, /learn"
