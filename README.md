# NetCheck: quick internet diagnostics for helpdesk

**English** | [Русский](README.ru.md)

A small Windows utility. A double click opens several `cmd` windows, each running its own diagnostic command (`ping`, `tracert`, `curl`, etc.). The windows are automatically arranged in a grid and never overlap.

Use case: an employee loses internet (or it is slow) while connected **via Ethernet cable**. They launch `NetCheck.exe`, and the helpdesk sees live results right away: is the problem at the gateway, beyond it, in DNS, and so on.

By default 4 windows open:

| Window | Command | What it shows |
|--------|---------|---------------|
| Ping Google DNS | `ping 8.8.8.8 -t` | whether there is internet connectivity |
| Ping Gateway | `ping {GATEWAY} -t` | whether the router/switch is reachable |
| Tracert Google | `tracert 8.8.8.8` | where the route breaks |
| External IP | `curl ipinfo.io` | public IP and ISP |

The `{GATEWAY}` placeholder is replaced automatically with the default gateway IP. If no gateway is found (cable unplugged, no DHCP lease), the window shows `GATEWAY_NOT_FOUND`.

## Requirements

- Windows 10 / 11
- [Go](https://go.dev/dl/) (only on the helpdesk machine, for building)
- VS Code or any text editor
- Nothing to install on the user's PC: just one `.exe`

## Quick start

1. Download the repository (`git clone` or **Code → Download ZIP**, then unzip).
2. Open the folder in VS Code.
3. Edit the command list in `main.go` (see below).
4. Run `build.bat` (double click). `NetCheck.exe` appears in the folder.
5. Copy `NetCheck.exe` to the user's PC and create a desktop shortcut.
6. When there is a problem, the user simply double-clicks the shortcut.

Manual build:

```bash
go build -ldflags="-H=windowsgui -s -w" -o NetCheck.exe
```

The `-H=windowsgui` flag hides the launcher's own console window, so only the command windows are shown.

## How to add or change commands

Open `main.go` and find the `terminals` list. Each line is one window:

```go
{"Window title", "command"},
```

Below the list there are ready-made examples in comments. Remove the leading `//` and put in your command or IP:

```go
// {"Ping server", "ping 10.0.0.1 -t"},       // put your IP here
// {"Tracert server", "tracert 10.0.0.1"},    // put your IP here
// {"DNS check", "nslookup google.com"},
// {"Network config", "ipconfig /all"},
```

Rules:

- **Every line must end with `},`** (closing curly bracket and a comma). A missing comma is the most common build error (`unexpected newline in composite literal`).
- Use straight quotes `"`, not curly quotes `“ ”` or `« »` (they appear when text is copied from Word or messengers).
- The window title must not contain `&`, `|` or `"`.
- If the command itself needs `&` or `|`, escape them: `^&`, `^|`.
- The number of windows is unlimited. The layout adapts automatically: 4 windows give a 2×2 grid, 6 windows 3×2, 9 windows 3×3, and so on.
- Windows stay open after the command finishes, so the output remains on screen.

## How it works

1. Detects the default gateway via `route print` (independent of the Windows language).
2. Starts each command in a separate window via `conhost.exe cmd.exe /k`. These are classic console windows, so they do not merge into Windows Terminal tabs on Windows 11.
3. Finds each window by a unique tag in its title and moves it into its own grid cell inside the screen work area (excluding the taskbar).

## Known quirks

- Test the built `.exe`, not `go run` from the VS Code terminal: window behavior may differ.
- Windows 10/11 windows have an invisible border, so a gap or overlap of a few pixels between windows is possible.
- Antivirus software may be suspicious of unsigned `.exe` files that launch `cmd`. If so, add the file to exclusions or sign it with your organization's certificate.
- `curl` is built into Windows 10 (1803) and newer. On older systems the command will not work.

## Repository structure

```
main.go        source code and command list
go.mod         Go module
build.bat      one-click build
README.md      this file (English)
README.ru.md   Russian version
.gitignore     keeps built .exe files out of the repository
```
