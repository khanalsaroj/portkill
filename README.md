<div align="center">

# portkill

**Find and free the process holding a port — instantly.**

No more `netstat`, `lsof`, or hunting for PIDs. Just `portkill kill 8080`.

[![CI](https://github.com/khanalsaroj/portkill/actions/workflows/ci.yml/badge.svg)](https://github.com/khanalsaroj/portkill/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/khanalsaroj/portkill?sort=semver)](https://github.com/khanalsaroj/portkill/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/khanalsaroj/portkill)](https://goreportcard.com/report/github.com/khanalsaroj/portkill)
[![License: MIT](https://img.shields.io/github/license/khanalsaroj/portkill)](LICENSE)

</div>

---

## The problem

"Port already in use." Again. Freeing it usually means a two-step ritual:

<table>
<tr><th>Windows</th><th>macOS / Linux</th></tr>
<tr><td>

```powershell
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

</td><td>

```bash
lsof -i :8080
kill -9 <PID>
```

</td></tr>
</table>

Too many steps, too easy to kill the wrong PID, and different on every OS.

## The fix

```bash
# Free a single port
portkill kill 8080

# Free several at once
portkill kill 3000 8080 9090
```

```text
  →  scanning port 8080 …
  ✓  port 8080 freed — killed node (PID 54321)

  →  scanning port 3000 …
  ℹ  port 3000 is not in use
```

## Features

- ⚡ **Free any port instantly** — one command, no PID hunting.
- 🖥️ **Truly cross-platform** — Windows, macOS, Linux, and FreeBSD.
- 🔎 **`list` before you leap** — see exactly what's on a port (name, PID, protocol, address).
- 🎯 **Kills everything on the port** — IPv4 + IPv6, TCP + UDP, and ports shared by multiple processes.
- 🛡️ **Graceful by default** — asks nicely, then force-kills if needed, and verifies the port is actually freed.
- 🤖 **Scriptable** — `--json` output and meaningful exit codes.
- 🪶 **Zero dependencies** — a single small static binary; no runtime, no external tools required on Linux.

## Installation

### 🐧 Linux / 🍎 macOS / 😈 FreeBSD

```bash
curl -fsSL https://raw.githubusercontent.com/khanalsaroj/portkill/main/main/install.sh | bash
```

Installs to `/usr/local/bin` (using `sudo` if needed) or falls back to
`~/.local/bin`. Override with `PORTKILL_INSTALL_DIR` or pin a version with
`PORTKILL_VERSION`.

### 🪟 Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/khanalsaroj/portkill/main/main/install.ps1 | iex
```

> Restart your terminal afterwards so the updated `PATH` takes effect.

### 📦 Prebuilt binaries

Grab an archive for your platform from the
[Releases](https://github.com/khanalsaroj/portkill/releases) page (each release
ships a `checksums.txt`).

### 🛠️ From source

```bash
go install github.com/khanalsaroj/portkill/cmd/portkill@latest
```

### Verify

```bash
portkill version
```

## Usage

```text
portkill <command> [flags] <port> [port...]

COMMANDS
  kill        Terminate the process(es) bound to one or more ports
  list, ls    Show what is listening on one or more ports
  version     Print version information
  help        Show this help

FLAGS
  -f, --force      Skip graceful shutdown; hard-kill immediately
  -n, --dry-run    Show what would be killed, without killing it
      --timeout D  Grace period before escalating to a hard kill (default 3s)
      --json       Emit machine-readable JSON instead of text
  -q, --quiet      Suppress progress output; report only failures
      --no-color   Disable colored output (also honors NO_COLOR)
  -h, --help       Show this help
  -v, --version    Print version information
```

### Examples

```bash
portkill kill 8080                 # free one port
portkill kill 3000 8080 9090       # free several at once
portkill kill -f 5432              # force-kill, no grace period
portkill kill --dry-run 8080       # preview without killing
portkill list 8080                 # inspect what's on a port
portkill list --json 8080 3000     # scriptable JSON
```

Inspect a port:

```text
$ portkill list 8080
PORT  PROTO  PID    PROCESS  ADDRESS
8080  tcp    54321  node     0.0.0.0:8080
8080  tcp    54322  node     [::]:8080
```

Script against it:

```bash
$ portkill list --json 8080
[
  {
    "port": 8080,
    "in_use": true,
    "processes": [
      { "pid": 54321, "name": "node", "port": 8080, "protocol": "tcp", "address": "0.0.0.0:8080" }
    ]
  }
]
```

### Exit codes

| Code | Meaning                                            |
| ---- | -------------------------------------------------- |
| `0`  | Success (includes "port was not in use").          |
| `1`  | A process could not be killed, or detection failed. |
| `2`  | Usage error (bad flag, missing or invalid port).    |

## How it works

For each port, portkill discovers every bound socket — IPv4 and IPv6, TCP and
UDP, including ports shared by multiple processes — and resolves the owning
process names. Then, unless you pass `--force`, it asks each process to stop
gracefully (`SIGTERM`, or `taskkill` on Windows), waits up to the grace period,
and escalates to a hard kill (`SIGKILL` / `taskkill /F`) only if necessary. It
confirms the process is actually gone before reporting success.

Detection is dependency-free where possible:

| OS              | Method                                                          |
| --------------- | --------------------------------------------------------------- |
| Linux           | Kernel `/proc` tables (no external tools), then `ss`, then `lsof` |
| macOS           | `lsof` (ships with the OS)                                      |
| FreeBSD         | `sockstat`, then `lsof`                                         |
| Windows         | `netstat` + `tasklist` for names                               |

> **Elevated permissions:** killing processes owned by other users or by the
> system requires `sudo` on Unix or an Administrator terminal on Windows.

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the dev
workflow, project layout, and how to add OS support. The codebase is small,
fully tested, and has no third-party dependencies.

## License

[MIT](LICENSE) © Khanal
