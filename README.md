# NekoPresenceKey — Windows

This branch contains the Windows desktop agent for NekoPresenceKey.

It receives cryptographically signed presence heartbeats from the paired Android app and locks Windows when that phone is no longer reachable for the configured timeout.

## Requirements

- Windows 10 or Windows 11 x64.
- Go 1.23 or newer to build from source.
- PowerShell 5.1+ for the installer.
- Android companion app from the repository's `android` branch.

## Build on Windows

Install Go from the official Go installer, then open PowerShell in this repository.

```powershell
New-Item -ItemType Directory -Force dist
cd desktop-agent-go
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath -ldflags="-s -w" -o ..\dist\nekopresence-windows-amd64.exe .
cd ..
```

## Test before installing

```powershell
.\dist\nekopresence-windows-amd64.exe --pair --dry-run --timeout 20
```

The console displays a 6-digit pairing code. In the Android app, enter this PC's private LAN IP, for example `192.168.1.50`, and the code.

## Run normally

```powershell
.\dist\nekopresence-windows-amd64.exe --pair --timeout 20
```

After pairing is saved:

```powershell
.\dist\nekopresence-windows-amd64.exe --listen 0.0.0.0:45873 --timeout 20
```

## Install at Windows logon

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\install\windows\install.ps1
```

This copies the agent to `%LOCALAPPDATA%\NekoPresenceKey` and creates a Scheduled Task for the current user.

## Firewall

Allow TCP port `45873` only on your **Private** network profile. Do not port-forward this port.

## Windows unlock limitation

The current agent can call `LockWorkStation`, but it intentionally does not store or type your password to bypass LogonUI. Full secure automatic unlock requires a proper Windows Credential Provider. Keep Windows Hello, PIN or password enabled as recovery authentication.
