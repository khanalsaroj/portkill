//go:build linux

package killer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// findProcesses lists processes bound to port. It prefers the kernel's /proc
// tables (no external tools, works in minimal containers) and falls back to ss
// and then lsof when /proc can't resolve a PID (e.g. another user's socket).
func findProcesses(port int) ([]Process, error) {
	if procs := findViaProc(port); len(procs) > 0 {
		return procs, nil
	}
	if procs := findViaSS(port); len(procs) > 0 {
		return procs, nil
	}
	if procs := findViaLsof(port); len(procs) > 0 {
		return procs, nil
	}
	return nil, nil
}

// findViaProc reads /proc/net/{tcp,udp}{,6}, then maps the matching socket
// inodes back to PIDs by scanning /proc/<pid>/fd.
func findViaProc(port int) []Process {
	sources := []struct{ file, proto string }{
		{"/proc/net/tcp", ProtoTCP},
		{"/proc/net/tcp6", ProtoTCP},
		{"/proc/net/udp", ProtoUDP},
		{"/proc/net/udp6", ProtoUDP},
	}

	var sockets []procSocket
	for _, s := range sources {
		data, err := os.ReadFile(s.file)
		if err != nil {
			continue // e.g. IPv6 disabled
		}
		sockets = append(sockets, parseProcNet(string(data), s.proto, port)...)
	}
	if len(sockets) == 0 {
		return nil
	}
	return resolveInodes(sockets, port)
}

// resolveInodes walks every process's file descriptors looking for the socket
// inodes we care about, building Process records for the owners it can see.
func resolveInodes(sockets []procSocket, port int) []Process {
	want := make(map[string]procSocket, len(sockets))
	for _, s := range sockets {
		want["socket:["+s.Inode+"]"] = s
	}

	dirs, _ := filepath.Glob("/proc/[0-9]*")
	var procs []Process
	for _, dir := range dirs {
		pid, err := strconv.Atoi(filepath.Base(dir))
		if err != nil {
			continue
		}
		fds, err := os.ReadDir(filepath.Join(dir, "fd"))
		if err != nil {
			continue // not ours to inspect
		}
		for _, fd := range fds {
			target, err := os.Readlink(filepath.Join(dir, "fd", fd.Name()))
			if err != nil {
				continue
			}
			s, ok := want[target]
			if !ok {
				continue
			}
			procs = append(procs, Process{
				PID:      pid,
				Name:     procName(pid),
				Port:     port,
				Protocol: s.Protocol,
				Address:  s.Address,
			})
		}
	}
	return procs
}

func procName(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func findViaSS(port int) []Process {
	if _, err := exec.LookPath("ss"); err != nil {
		return nil
	}
	out, _, _ := runCommand("ss", "-tulnpH")
	return parseSS(out, port)
}

func findViaLsof(port int) []Process {
	if _, err := exec.LookPath("lsof"); err != nil {
		return nil
	}
	out, _, _ := runCommand("lsof", "-nP", "-i", fmt.Sprintf(":%d", port))
	return parseLsof(out, port)
}
