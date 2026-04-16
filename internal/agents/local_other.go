//go:build !darwin && !linux

package agents

import "os/exec"

func configureIsolatedSubprocess(cmd *exec.Cmd) {}
