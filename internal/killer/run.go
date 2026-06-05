package killer

import (
	"bytes"
	"os/exec"
)

// runCommand executes name with args, returning captured stdout and stderr
// separately along with any run error. Callers interpret exit codes themselves
// (e.g. lsof exits 1 to mean "no match", which is not a failure).
func runCommand(name string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(name, args...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return out.String(), errBuf.String(), err
}
