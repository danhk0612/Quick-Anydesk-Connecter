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
- Checked Windows startup toggle
- Manual update checks through GitHub Releases with SHA-256 verification, self-replacement, and restart
- Clipboard image preview with explicit approval before OpenRouter Vision analysis
- OpenRouter API Key stored in Windows Credential Manager
- Settings reset, backup, and restore
- Context-specific OpenRouter, network, and analysis errors
- Korean and English UI
- Language switching from the tray menu without restarting
- Korean is the default language
- Embedded application and tray icon
- Single EXE deployment; user settings are stored under `%LOCALAPPDATA%`

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
- Image Auto Analysis
- Run at Startup
- Change Default Password
- OpenRouter Settings
- Check for Updates
- Back Up Settings
- Restore Settings
- Reset Settings
- Language
  - 한국어
  - English
- Exit

## Configuration

`config.ini` is stored under the current Windows user's Local AppData directory instead of next to the executable:

```text
%LOCALAPPDATA%\QuickAnydeskConnect\config.ini
```

If an older version left a valid `config.ini` beside the EXE and no user-profile configuration exists yet, it is migrated automatically once on the next launch.

```ini
[anydesk]
password=YOUR_PASSWORD

[general]
language=ko
image_analysis=false

[openrouter]
model=google/gemma-4-26b-a4b-it:free
```

Supported language values:

- `ko`
- `en`

If `[general]` or `language` is missing, Korean is used.

> [!WARNING]
> The unattended-access password is stored in plain text in `config.ini`. The file is kept in the current Windows user's Local AppData area.

### Settings backup, restore, and reset

**Back Up Settings** stores the current AnyDesk default password, language, image-analysis setting, OpenRouter model, and the OpenRouter API Key from Windows Credential Manager in a `.qacbackup` file. The **Run at Startup** registration is not included in backup, restore, or reset.

**Restore Settings** validates the selected backup before replacing the active settings and OpenRouter API Key. **Reset Settings** clears application settings and the OpenRouter API Key, then asks for a new default AnyDesk password.

> [!IMPORTANT]
> `.qacbackup` files contain the AnyDesk default password and OpenRouter API Key in readable form. Do not share them and keep them in a secure location.

## Image Auto Analysis

Image Auto Analysis is disabled by default and can be enabled from the tray menu.

On first activation, enter an OpenRouter API Key. The key is validated with the non-inference `GET /api/v1/key` endpoint and stored in Windows Credential Manager instead of `config.ini`.

Clipboard images are **not** sent automatically. The application first displays the copied image with **Analyze / Ignore** controls. The image is sent to OpenRouter only after you explicitly choose **Analyze**. The preview uses high-quality interpolation when scaled to fit the window. For analysis, only oversized images are downscaled while preserving aspect ratio, with the longest side capped at 1600 px; smaller images are never enlarged.

In **OpenRouter Settings**, you can enter both the API Key and the model ID directly. The dialog also includes a button that opens the OpenRouter API Keys page. The default model is `google/gemma-4-26b-a4b-it:free`, but the model field is not restricted to a predefined list. Any OpenRouter model that accepts image input can be entered.

### Suggested image-capable models

The following are practical examples for reading a short AnyDesk address from a screenshot. Model availability, pricing, rate limits, and IDs can change, so verify the current model page on OpenRouter before relying on one.

| Model ID | Cost | Notes |
| --- | --- | --- |
| `google/gemma-4-26b-a4b-it:free` | Free | Default. Multimodal and inexpensive to try, but free upstream providers may be temporarily rate-limited. |
| `google/gemma-4-31b-it:free` | Free | Larger free multimodal alternative; useful when the default free endpoint is unavailable. |
| `nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free` | Free | General multimodal model that accepts image input. |
| `google/gemini-2.5-flash-lite` | Paid, low cost | Stable fallback observed to work well for this simple number-extraction task. |

If a free model returns an upstream provider rate-limit error, wait and retry or change the model in **OpenRouter Settings**. The program distinguishes this from a general OpenRouter/account rate limit and shows the provider response when available.

> [!IMPORTANT]
> Approved clipboard images are sent to OpenRouter and the selected model provider. Free model providers may apply their own logging, retention, availability, and rate-limit policies. Avoid analyzing sensitive images unless you are comfortable sending them to the configured external service.

OpenRouter is a separate service and API usage may incur charges. Create or manage an API key at <https://openrouter.ai/settings/keys>.

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
- Tags matching `v*` build `QuickAnydeskConnect.exe` and `QuickAnydeskConnect.exe.sha256` and attach them to a GitHub Release.
- No general build artifact is uploaded, minimizing Actions artifact storage usage.

## License

MIT License. See [LICENSE](LICENSE).

## Disclaimer

Quick Anydesk Connect is an independent utility and is not affiliated with, endorsed by, or sponsored by AnyDesk Software GmbH. AnyDesk is a trademark of its respective owner.
