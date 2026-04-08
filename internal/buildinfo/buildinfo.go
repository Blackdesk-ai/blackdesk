package buildinfo

import "fmt"

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
