#!/usr/bin/env bash
set -euo pipefail
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) BIN=nekopresence-linux-amd64 ;;
  aarch64|arm64) BIN=nekopresence-linux-arm64 ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
mkdir -p "$HOME/.local/bin" "$HOME/.config/systemd/user"
install -m755 "$ROOT/dist/$BIN" "$HOME/.local/bin/nekopresence"
cp "$ROOT/install/linux/nekopresence.service" "$HOME/.config/systemd/user/nekopresence.service"
systemctl --user daemon-reload
systemctl --user enable --now nekopresence.service
echo "Installed. Run: systemctl --user status nekopresence"
echo "To re-open pairing: systemctl --user stop nekopresence && ~/.local/bin/nekopresence --pair"
