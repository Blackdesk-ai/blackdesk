package tui

import "strings"

func renderMarkdownCodeBlock(lines []string, width int) ([]string, int, bool) {
	_ = width
	open := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(open, "```") {
		return nil, 0, false
	}
	lang := strings.TrimSpace(strings.TrimPrefix(open, "```"))
	out := make([]string, 0, len(lines))
	title := "Code"
	if lang != "" {
		title += " (" + lang + ")"
	}
	out = append(out, title)
	consumed := 1
	for consumed < len(lines) {
		line := lines[consumed]
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			consumed++
			return out, consumed, true
		}
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
		} else {
			out = append(out, "    "+line)
		}
		consumed++
	}
	return out, consumed, true
}
