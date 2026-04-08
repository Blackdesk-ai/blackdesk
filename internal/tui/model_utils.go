package tui

import (
	"os/exec"
	"runtime"
	"strings"
)

var openURLFunc = openURL

func valueOrDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func openURL(raw string) error {
	if raw == "" {
		return nil
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", raw)
	case "linux":
		cmd = exec.Command("xdg-open", raw)
	default:
		return nil
	}
	return cmd.Start()
}
