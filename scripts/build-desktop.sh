#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$ROOT/dist"
cd "$ROOT/desktop-agent-go"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o "$ROOT/dist/nekopresence-linux-amd64" .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o "$ROOT/dist/nekopresence-linux-arm64" .
sha256sum "$ROOT"/dist/nekopresence-linux-* > "$ROOT/dist/SHA256SUMS.txt"
