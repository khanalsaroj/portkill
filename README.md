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

### From source

```bash
git clone https://github.com/khanalsaroj/portkill.git
cd portkill
go build -o portkill ./cmd/portkill
```

Then move the binary somewhere on your `$PATH`:

```bash
# macOS / Linux
sudo mv portkill /usr/local/bin/

# Windows — move portkill.exe to a directory in %PATH%
```

> **Elevated permissions** — killing processes owned by other users or system
> processes requires `sudo` on Unix or an Administrator prompt on Windows.

## Platform support

| OS      | Port resolution | Process termination |
|---------|----------------|---------------------|
| macOS   | `lsof`         | `SIGTERM`           |
| Linux   | `lsof`         | `SIGTERM`           |
| Windows | `netstat -ano` | `taskkill /F /T`    |

## Project layout

```
portkill/
├── cmd/
│   └── portkill/
│       └── main.go           # CLI entry point & argument parsing
├── internal/
│   └── killer/
│       ├── killer.go         # Public API + user-facing output
│       ├── killer_unix.go    # Linux & macOS implementation
│       └── killer_windows.go # Windows implementation
├── go.mod
└── README.md
```

The root directory is intentionally kept clean for installer scripts
(`.sh`, `.ps1`) and a GitHub Actions release workflow to be added later.

## Exit codes

| Code | Meaning                                              |
|------|------------------------------------------------------|
| `0`  | All requested ports handled successfully             |
| `1`  | One or more ports could not be killed / bad argument |

This makes `portkill` safe to use in shell scripts with `set -e`.

## Contributing

PRs welcome. Please keep OS-specific logic inside `internal/killer/` and gated
behind the appropriate build tags (`//go:build linux || darwin` / `//go:build windows`).
