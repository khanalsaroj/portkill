# Port kill
A simple, cross-platform CLI tool to find and kill processes listening on specific ports.
Tired of killing ports manually using long commands? No more `netstat`, `lsof`, or hunting for PIDs.
Try `portkill`!

---

## The Problem

Killing a port usually looks like this:

### Windows

```bash
netstat -ano | findstr :<PORT>
taskkill /PID <PID> /F
```

### macOS /  Linux

```bash
lsof -i :<PORT>
kill -9 <PID>
```

Or:

```bash
ss -lptn 'sport = :8080'
```

- 👉 Too many steps
- 👉 Too much time wasted
- 👉 Easy to mess up

## Just run

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

---

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

---

##  Features

*  Kill any port instantly
*  Cross-platform (Windows, macOS, Linux)
*  Fast and lightweight
*  Automatically finds and kills the process
*  Supports multiple ports