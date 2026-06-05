# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

A near-complete rewrite focused on correctness, robustness, and ergonomics.

### Added

- `list` (alias `ls`) command to inspect what is bound to a port without killing
  it, including the process name, protocol, PID, and local address.
- `version` command plus `--version` / `-v` flags.
- `--json` output for both `kill` and `list`, for scripting.
- `--force` / `-f` to skip the grace period and hard-kill immediately.
- `--dry-run` / `-n` to preview what would be killed.
- `--timeout` to configure the grace period before escalating to a hard kill.
- `--quiet` / `-q` and `--no-color` output controls (honors `NO_COLOR`).
- Colored, TTY-aware output (with virtual-terminal support on Windows 10+).
- Detection and termination of **all** processes on a port (IPv4 + IPv6, and
  ports shared by multiple processes), not just the first one found.
- UDP support in addition to TCP.
- Native, dependency-free port detection on Linux via `/proc`, falling back to
  `ss` and then `lsof` — so it works even in minimal containers.
- Process-name resolution on every platform.
- Checksum verification in both installers.
- Comprehensive unit tests, a `Makefile`, `golangci-lint` config, a CI workflow
  (test + vet + lint + cross-compile on Linux/macOS/Windows), and contributor docs.

### Changed

- **Graceful-by-default termination with automatic force-escalation:** portkill
  now asks a process to stop and force-kills it only if it does not exit within
  the grace period — and verifies the process is actually gone. The port is
  always freed.
- Rearchitected the `killer` package: OS-agnostic orchestration, build-tagged
  per-OS discovery/termination, and pure, testable output parsers.
- Linux detection now prefers `ss` (almost always present) over `lsof` (often
  absent in containers).
- Modernized the release workflow (current action versions; replaced the
  archived `actions/create-release`); binaries are now stripped and stamped with
  commit and build date.

### Fixed

- **`--version` now works.** Previously the binary errored on `--version`, which
  made both installers report failure on every successful install.
- **Linux installer no longer requires Docker or root.** It installed to a system
  directory only when permitted and otherwise falls back to `~/.local/bin`.
- Ports are matched exactly (`:80` no longer matches `:8080`).
- Established connections are no longer mistaken for listeners.
- Corrected the "ProtKill" typo and other installer rough edges.
