//go:build !windows

package utils

import (
	"os/exec"
	"syscall"
)

// PrepareChildProcess configures POSIX process group for clean group termination
func PrepareChildProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
