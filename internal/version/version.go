// Package version exposes build metadata that is stamped into the binary at
// link time via -ldflags. See the Makefile and the release workflow.
package version

import "runtime"

// These values are overridden at build time, e.g.:
//
//	go build -ldflags "-X github.com/khanalsaroj/portkill/internal/version.Version=v1.2.3"
var (
	// Version is the semantic version of the build (e.g. "v1.2.3").
	Version = "dev"
	// Commit is the short git SHA the binary was built from.
	Commit = "none"
	// Date is the build timestamp in RFC 3339 form.
	Date = "unknown"
)

// String returns a single-line, human-readable version banner, e.g.
// "portkill v1.2.3 (abc1234, 2026-06-05) go1.25 linux/amd64".
func String() string {
	return "portkill " + Version + " (" + Commit + ", " + Date + ") " +
		runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH
}

// Short returns just the version string (e.g. "v1.2.3").
func Short() string { return Version }
