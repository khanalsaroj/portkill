# portkill
A simple, cross-platform CLI tool to find and kill processes listening on specific ports.
## Usage

```bash
# Kill a single port
portkill kill 8080

# Kill multiple ports at once
portkill kill 3000 8080 9090
```

### Example output

```
  →  scanning port 8080 ...
  ⚡  found PID 54321 on port 8080 — terminating ...
  ✓  PID 54321 on port 8080 terminated successfully

  →  scanning port 3000 ...
  ℹ  port 3000 is not in use
```

## Installation

### 🐧 Linux  (requires `curl`)

```bash
curl -fsSL https://raw.githubusercontent.com/khanalsaroj/portkill/refs/heads/main/main/install.sh | bash
```

### 🪟 Windows (PowerShell installer)

Open **PowerShell as Administrator**:

```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

```powershell
iwr -useb https://raw.githubusercontent.com/khanalsaroj/portkill/refs/heads/main/main/install.ps1 | iex
```

> ***Restart your terminal after installation.***

### Verify Installation

```bash
portkill -h
```

Or download a prebuilt binary for your platform from the [Releases](https://github.com/khanalsaroj/portkill/releases)
page.


> **Elevated permissions** — killing processes owned by other users or system
> processes requires `sudo` on Unix or an Administrator prompt on Windows.
