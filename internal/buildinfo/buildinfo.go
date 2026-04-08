package buildinfo

import (
	"fmt"
	"strings"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Summary(appName string) string {
	return fmt.Sprintf("%s %s", appName, Version)
}

func Detailed(appName string) string {
	return fmt.Sprintf("%s %s\ncommit: %s\nbuilt: %s", appName, Version, Commit, Date)
}

func NormalizedVersion() string {
	version := strings.TrimSpace(Version)
	return strings.TrimPrefix(version, "v")
}

func VersionLabel(version string) string {
	version = normalizeVersion(version)
	if version == "" {
		return "unknown"
	}
	if version == "dev" {
		return "dev"
	}
	return "v" + version
}

func Label() string {
	return VersionLabel(Version)
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	return strings.TrimPrefix(version, "v")
}
