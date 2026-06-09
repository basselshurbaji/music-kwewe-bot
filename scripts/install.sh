#!/usr/bin/env bash
# Install the Homebrew dependencies needed to run music-queue.
set -euo pipefail

if ! command -v brew >/dev/null 2>&1; then
  echo "Error: Homebrew is not installed." >&2
  echo "Install it first: https://brew.sh" >&2
  exit 1
fi

echo "==> Installing Homebrew dependencies: mpv yt-dlp"
brew install mpv yt-dlp

if ! command -v go >/dev/null 2>&1; then
  echo "Warning: Go is not installed. Install Go 1.26+ ('brew install go' or https://go.dev/dl/)." >&2
fi

echo "==> Dependencies installed."
