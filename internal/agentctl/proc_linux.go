//go:build linux

package agentctl

import (
	"os/exec"
	"syscall"
)

func DetachCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func signalTerm(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

func signalKill(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

func statfsDisk(path string) (totalBytes, freeBytes uint64, ok bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0, false
	}
	bsize := uint64(st.Bsize)
	return st.Blocks * bsize, st.Bavail * bsize, true
}
