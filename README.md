# Quick Anydesk Connect

[한국어](README.ko.md)

Quick Anydesk Connect is a lightweight Windows tray utility that streamlines repeated AnyDesk remote-support connections.

It watches the clipboard for a 9- or 10-digit AnyDesk address, asks for confirmation, and starts a connection using a configured unattended-access password. If the configured password does not match the remote PC, AnyDesk can prompt for the actual password as usual.

## Features

- Runs in the Windows notification area (system tray)
- Detects newly copied 9- or 10-digit AnyDesk addresses
- Always-on-top confirmation before connecting
- Manual connection from the tray menu or by double-clicking the tray icon
- Automatically submits a configured unattended-access password
- Starts AnyDesk first when AnyDesk has been fully exited, then connects after initialization
- Add/remove the launcher from Windows startup
- Korean and English UI
- Language switching from the tray menu without restarting
- Korean is the default language
- Embedded application and tray icon
- Single EXE deployment; `config.ini` is created on first run if missing

## Requirements

- Windows
- AnyDesk installed in one of the supported locations
- For building from source: Go 1.22 or newer

The launcher currently searches these locations:

```text
C:\Program Files (x86)\AnyDesk-*\AnyDesk-*.exe
C:\Program Files\AnyDesk-*\AnyDesk-*.exe
C:\Program Files (x86)\AnyDesk\AnyDesk.exe
C:\Program Files\AnyDesk\AnyDesk.exe
```

## Usage

1. Run `QuickAnydeskConnect.exe`.
2. On first launch, enter the shared unattended-access password.
3. Copy a customer's AnyDesk address from a messenger or other application.
4. When the confirmation dialog appears, choose **Yes** to connect.

You can also open a manual connection dialog by:

- Double-clicking the tray icon, or
- Right-clicking the tray icon and selecting **Remote Connection**.

### Tray menu

- Remote Connection
- Add to Startup
- Remove from Startup
- Language
  - 한국어
  - English
- Exit

## Configuration

`config.ini` is stored next to the executable.

```ini
[anydesk]
password=YOUR_PASSWORD

[general]
language=ko
```

Supported language values:

- `ko`
- `en`

If `[general]` or `language` is missing, Korean is used.

> [!WARNING]
> The unattended-access password is stored in plain text in `config.ini`. Keep the file in a location accessible only to appropriate users.

## AnyDesk address detection

Clipboard text is normalized by removing spaces, tabs, line breaks, and hyphens. The resulting value must be exactly 9 or 10 digits.

Examples that are accepted:

```text
123456789
123 456 789
123-456-789
```

## Building

Run `build.bat` on Windows.

The batch file:

1. Removes old generated Windows resource files.
2. Uses [`go-winres`](https://github.com/tc-hib/go-winres) to embed the Windows executable icon.
3. Builds `QuickAnydeskConnect.exe` as a GUI executable.

`build.bat` is intentionally stored as **EUC-KR with CRLF line endings**.

The tray icon is also embedded through Go's `embed` package, so `app.ico` is not required after the executable has been built.

## GitHub Actions

- Pushes and pull requests run a Windows build check.
- Tags matching `v*` build `QuickAnydeskConnect.exe` and attach it to a GitHub Release.
- No general build artifact is uploaded, minimizing Actions artifact storage usage.

## License

MIT License. See [LICENSE](LICENSE).

## Disclaimer

Quick Anydesk Connect is an independent utility and is not affiliated with, endorsed by, or sponsored by AnyDesk Software GmbH. AnyDesk is a trademark of its respective owner.
