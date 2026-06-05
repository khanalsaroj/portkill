# Contributing to portkill

Thanks for your interest in improving portkill! This is a small, dependency-free
Go CLI, and contributions of all sizes are welcome.

## Getting started

```bash
git clone https://github.com/khanalsaroj/portkill.git
cd portkill
make build      # builds ./bin/portkill
make test       # runs the test suite
```

If you don't have `make`, the underlying commands work fine on their own:

```bash
go build ./cmd/portkill
go test ./...
```

## Development workflow

Before opening a pull request, please make sure these pass:

```bash
make fmt        # gofmt -w .
make vet        # go vet ./...
make lint       # golangci-lint run  (https://golangci-lint.run)
make test       # go test ./...
```

CI runs the same checks on Linux, macOS, and Windows, plus a cross-compile of
every release target.

## Project layout

```
cmd/portkill/            Tiny main(); delegates to internal/cli
internal/cli/            Argument parsing, dispatch, text/JSON rendering
internal/killer/         Find & terminate processes by port
  killer.go              OS-agnostic orchestration + kill escalation
  find_*.go              Per-OS process discovery (build-tagged)
  terminate_*.go         Per-OS signalling + liveness checks (build-tagged)
  parse_*.go             Pure, unit-tested parsers for tool output
internal/ui/             Terminal color + status symbols (NO_COLOR aware)
internal/version/        Build metadata stamped in via -ldflags
```

### Design principles

- **Zero third-party dependencies.** The standard library only. Please keep it
  that way unless there's a compelling reason.
- **Pure parsing is separate from I/O.** Anything that interprets command output
  (`parse_netstat.go`, `parse_lsof.go`, `parse_ss.go`, `parse_proc.go`,
  `parse_sockstat.go`) is a plain function with no syscalls, so it can be tested
  on any OS. New detection logic should follow the same split.
- **Detection returns structured results; the CLI decides how to print them.**

### Adding or improving OS support

1. Add detection in the appropriate `find_<goos>.go` behind a `//go:build` tag.
2. Put the output-parsing logic in a tag-free `parse_*.go` with table-driven
   tests, so it is covered regardless of the host OS.
3. Run `GOOS=<os> go build ./...` to confirm it cross-compiles.

## Commit messages

Releases are automated from the commit history using
[Conventional Commits](https://www.conventionalcommits.org):

- `feat: ...` → minor version bump
- `fix: ...` → patch version bump
- `feat!: ...` or a `BREAKING CHANGE:` footer → major version bump
- anything else → patch version bump

## Reporting bugs

Please include your OS and architecture, the exact command you ran, and the full
output. A failing `portkill list <port>` is especially helpful for diagnosing
detection issues.
