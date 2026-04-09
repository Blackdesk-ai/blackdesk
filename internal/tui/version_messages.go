package tui

import "blackdesk/internal/updater"

type versionCheckLoadedMsg struct {
	result updater.CheckResult
	err    error
}

type versionUpgradeLoadedMsg struct {
	result updater.UpgradeResult
	err    error
}
