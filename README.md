# NekoPresenceKey — Android

This branch contains the Android companion app used as the cryptographic presence key for a paired NekoPresenceKey computer.

The app creates an Ed25519 identity in Android Keystore, pins the paired computer key and signs fresh heartbeat challenges. Simply joining a Wi-Fi network does not authenticate a computer.

## Supported Android versions

- Minimum SDK: Android 9 / API 28.
- Target/compile SDK: Android 16 / API 36.
- Android 13 and Android 14 are supported, including Samsung Galaxy A52 5G on Android 14.

## Requirements

- Android Studio with Android SDK 36 installed, or command-line Gradle/JDK setup.
- JDK 17+; JDK 21 is recommended.
- A Windows/Linux/Raspberry Pi agent from the corresponding repository branch.

## Build with Android Studio

1. Clone/switch to the `android` branch.
2. Open the repository folder in Android Studio.
3. Wait for Gradle sync.
4. Select **Build → Build Bundle(s) / APK(s) → Build APK(s)**.
5. The debug APK will be under `app/build/outputs/apk/debug/app-debug.apk`.

## Build from command line

With Gradle 8.11.1 available:

```bash
gradle assembleDebug
```

The APK is written to:

```text
app/build/outputs/apk/debug/app-debug.apk
```

## Install using ADB

```bash
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

## Pair with a computer

1. Put the Android device and computer on the same trusted home/local network.
2. Start the desktop agent with `--pair`.
3. Enter the computer's private LAN IP, for example `192.168.1.50`.
4. Enter the displayed 6-digit pairing code.
5. Complete pairing and press **Start presence key**.

The app requires Wi-Fi/private LAN transport and rejects VPN transport for presence checks.

## Samsung background operation

If Samsung stops the service with the screen off, open **Settings → Apps → NekoPresenceKey → Battery → Unrestricted** and make sure the app is not in Sleeping apps or Deep sleeping apps.

## GitHub Actions

The workflow on this branch builds the debug APK on pushes and manual workflow runs and uploads an `android-apk` artifact.

## Security

Only pair on your trusted local network. Never expose the desktop agent port to the Internet. Re-pair the computer if the phone is lost or replaced.
