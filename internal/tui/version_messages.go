package tui

import "blackdesk/internal/updater"

type versionCheckLoadedMsg struct {
	result updater.CheckResult
	err    error
}
