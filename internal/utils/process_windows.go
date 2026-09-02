//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

// PrepareChildProcess configures Windows child process flags (e.g. HideWindow)
func PrepareChildProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
