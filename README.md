# NekoPresenceKey — Linux & Raspberry Pi

This branch contains the Linux desktop agent for x86-64 PCs and 64-bit ARM systems such as Raspberry Pi OS.

The agent verifies signed Android presence heartbeats and locks the current graphical session when the paired phone disappears for the configured timeout.

## Supported targets

- Linux x86-64 (`amd64`).
- Linux ARM64 (`arm64` / `aarch64`), including 64-bit Raspberry Pi OS.
- systemd/logind graphical sessions are recommended.

## Requirements

- Go 1.23 or newer.
- systemd/logind for the included user service and `loginctl` lock support.
- Android companion app from the repository's `android` branch.

## Build on Linux x64

```bash
mkdir -p dist
cd desktop-agent-go
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../dist/nekopresence-linux-amd64 .
cd ..
chmod +x dist/nekopresence-linux-amd64
```

## Build on Raspberry Pi 64-bit

Check the architecture:

```bash
uname -m
```

`aarch64` means ARM64. Then:

```bash
mkdir -p dist
cd desktop-agent-go
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ../dist/nekopresence-linux-arm64 .
cd ..
chmod +x dist/nekopresence-linux-arm64
```

## Build both Linux targets

```bash
./scripts/build-desktop.sh
```

This produces Linux x64 and Linux ARM64/Raspberry Pi binaries in `dist/`.

## Test safely

Linux x64:

```bash
./dist/nekopresence-linux-amd64 --pair --dry-run --timeout 20
```

Raspberry Pi ARM64:

```bash
./dist/nekopresence-linux-arm64 --pair --dry-run --timeout 20
```

Pair the Android app using this machine's private LAN address and the displayed 6-digit pairing code.

## Run normally

```bash
./dist/nekopresence-linux-amd64 --pair --timeout 20
```

or on Pi/ARM64:

```bash
./dist/nekopresence-linux-arm64 --pair --timeout 20
```

## Install as a systemd user service

```bash
chmod +x install/linux/install.sh
./install/linux/install.sh
```

The installer selects `amd64` or `arm64` automatically.

Check status:

```bash
systemctl --user status nekopresence
```

View logs:

```bash
journalctl --user -u nekopresence -f
```

Re-pair:

```bash
systemctl --user stop nekopresence
~/.local/bin/nekopresence --pair
systemctl --user start nekopresence
```

## Raspberry Pi notes

Use 64-bit Raspberry Pi OS for the ARM64 target. Guest Wi-Fi/client isolation can prevent the phone and Pi from reaching one another. A headless Pi without an active graphical session has no desktop session to lock.

## Firewall

Allow TCP `45873` from your trusted local subnet only. Do not port-forward it from your router.

## GitHub Actions

The workflow on this branch builds both Linux x64 and Raspberry Pi/Linux ARM64 and uploads them as an Actions artifact.
