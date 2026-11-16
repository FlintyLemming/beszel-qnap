//go:build !linux

package agentctl

import (
	"os"
	"os/exec"
)

func DetachCmd(_ *exec.Cmd) {}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	_, err := os.FindProcess(pid)
	return err == nil
}

func signalTerm(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}

func signalKill(pid int) error {
	return signalTerm(pid)
}

func statfsDisk(_ string) (totalBytes, freeBytes uint64, ok bool) {
	return 0, 0, false
}
