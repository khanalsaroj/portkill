package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/khanalsaroj/portkill/internal/killer"
)

const usage = `portkill — Terminate processes by port

USAGE:
  portkill kill <port> [<port>...]

DESCRIPTION:
  Finds and terminates processes currently bound to one or more network ports.

COMMANDS:
  kill        Kill process(es) using the specified port(s)

EXAMPLES:
  portkill kill 8080
      Kill the process using port 8080

  portkill kill 3000 8080 9090
      Kill processes using multiple ports

NOTES:
  • You may need elevated privileges (sudo/Administrator) to terminate certain processes.
  • If a port is not in use, portkill will notify you.
`

func main() {

	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "kill":
		runKill(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}

func runKill(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "error: 'kill' requires at least one port number")
		fmt.Fprintln(os.Stderr, "usage: portkill kill <port> [port2] ...")
		os.Exit(1)
	}

	ports := make([]int, 0, len(args))
	for _, arg := range args {
		port, err := strconv.Atoi(arg)
		if err != nil || port < 1 || port > 65535 {
			fmt.Fprintf(os.Stderr, "error: %q is not a valid port number (1–65535)\n", arg)
			os.Exit(1)
		}
		ports = append(ports, port)
	}

	anyFailed := false

	for _, port := range ports {
		if err := killer.KillPort(port); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗  port %d: %v\n", port, err)
			anyFailed = true
		}
	}

	if anyFailed {
		os.Exit(1)
	}
}
