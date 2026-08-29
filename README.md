# NekoPresenceKey

NekoPresenceKey is a local-network phone presence key for Windows, Linux and Raspberry Pi.

Your Android phone is paired cryptographically to a specific computer. Being connected to a Wi-Fi network by itself is **not** enough to authenticate: the phone and computer use separate Ed25519 identities, fresh signed challenges and a pinned pairing relationship.

When the paired phone stops responding for the configured timeout, the desktop agent locks the computer. The normal Windows/Linux password, PIN or other OS authentication should always remain available as a fallback.

## Platform branches

This repository keeps each platform in its own branch so users only need the source and build instructions for the platform they want:

| Branch | Platform | Source |
| --- | --- | --- |
| [`windows`](../../tree/windows) | Windows 10/11 x64 | Go desktop agent + PowerShell installer |
| [`android`](../../tree/android) | Android companion app | Kotlin / Android Studio |
| [`linux`](../../tree/linux) | Linux x64 + Raspberry Pi ARM64 | Go desktop agent + systemd user service |

## Current features

- One-time 6-digit pairing.
- Android Keystore-backed Ed25519 identity.
- Persistent computer Ed25519 identity.
- Signed challenge/response presence checks.
- Private-LAN-only desktop API.
- Android rejects cellular/VPN transport for presence.
- Configurable disconnect timeout.
- Automatic locking when the paired phone disappears.
- Windows x64 support.
- Linux x64 support.
- Raspberry Pi 64-bit / Linux ARM64 support.
- GitHub Actions builds for each branch.

## Important unlock note

The current Windows agent can lock the workstation but does **not** bypass the Windows lock screen. Secure automatic Windows unlocking requires a Windows Credential Provider. The project intentionally does not store or type your Windows password.

Linux desktop unlock support varies by desktop/session implementation; keep your normal password/PIN available.

## Quick start

1. Pick the branch for your computer and follow its README to build/install it.
2. Build/install the Android app from the `android` branch.
3. Start the computer agent with `--pair`.
4. Enter the PC's private LAN IP and displayed 6-digit code in the Android app.
5. Test with `--dry-run` before enabling real automatic locking.

## Security

Do **not** expose TCP port `45873` to the Internet or port-forward it from your router. Restrict it to your private/home LAN. A copied SSID or another network using the same private IP range is not sufficient to authenticate without the paired private key.

See each platform branch README for complete build and installation instructions.
