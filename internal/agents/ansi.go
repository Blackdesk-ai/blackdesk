package agents

import "regexp"

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func stripANSI(input string) string {
	return ansiPattern.ReplaceAllString(input, "")
}
