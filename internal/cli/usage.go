package cli

const usage = `portkill — find and free the processes holding your ports

USAGE:
  portkill <command> [flags] <port> [port...]

COMMANDS:
  kill        Terminate the process(es) bound to one or more ports
  list, ls    Show what is listening on one or more ports
  version     Print version information
  help        Show this help

FLAGS:
  -f, --force       Skip graceful shutdown; hard-kill immediately
  -n, --dry-run     Show what would be killed, without killing it
      --timeout D    Grace period before escalating to a hard kill (default 3s)
      --json        Emit machine-readable JSON instead of text
  -q, --quiet       Suppress progress output; report only failures
      --no-color    Disable colored output (also honors NO_COLOR)
  -h, --help        Show this help
  -v, --version     Print version information

EXAMPLES:
  portkill kill 8080                 Free port 8080
  portkill kill 3000 8080 9090       Free several ports at once
  portkill kill -f 5432              Force-kill without a grace period
  portkill kill --dry-run 8080       Preview what would be killed
  portkill list 8080                 See what is using port 8080
  portkill list --json 8080 3000     Scriptable JSON output

NOTES:
  • Killing processes owned by other users may require sudo/Administrator.
  • By default portkill asks a process to stop, then force-kills it if it has
    not exited within the grace period — so the port is always freed.
`
