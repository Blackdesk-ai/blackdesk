//go:build darwin || linux

package agents

import (
	"os/exec"
	"syscall"
)

func configureIsolatedSubprocess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
