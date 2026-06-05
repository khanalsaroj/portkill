// Command portkill finds and terminates the processes bound to network ports.
//
// Usage:
//
//	portkill kill 8080
//	portkill kill 3000 8080 9090
//	portkill list 8080
//
// See `portkill help` for the full set of commands and flags.
package main

import (
	"os"

	"github.com/khanalsaroj/portkill/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
