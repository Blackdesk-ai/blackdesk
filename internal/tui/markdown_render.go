package tui

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

func renderMarkdownLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(trimmed, "### "):
		return strings.ToUpper(renderMarkdownInline(strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))))
	case strings.HasPrefix(trimmed, "## "):
		return strings.ToUpper(renderMarkdownInline(strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))))
	case strings.HasPrefix(trimmed, "# "):
		return strings.ToUpper(renderMarkdownInline(strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))))
	case strings.HasPrefix(trimmed, "> "):
		return "│ " + renderMarkdownInline(strings.TrimSpace(strings.TrimPrefix(trimmed, "> ")))
	case isMarkdownBullet(trimmed):
		return "• " + renderMarkdownInline(strings.TrimSpace(trimmed[2:]))
	case isMarkdownNumbered(trimmed):
		idx := strings.Index(trimmed, ". ")
		return trimmed[:idx+2] + renderMarkdownInline(strings.TrimSpace(trimmed[idx+2:]))
	default:
		return renderMarkdownInline(trimmed)
	}
}

func wrapMarkdownLines(lines []string, width int) []string {
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			wrapped = append(wrapped, "")
			continue
		}
		if xansi.StringWidth(line) <= width {
			wrapped = append(wrapped, line)
			continue
		}
		wrapped = append(wrapped, splitLines(xansi.Wrap(line, width, ""))...)
	}
	return wrapped
}

func isMarkdownBullet(line string) bool {
	return strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ")
}

func isMarkdownNumbered(line string) bool {
	if len(line) < 3 {
		return false
	}
	idx := strings.Index(line, ". ")
	if idx <= 0 {
		return false
	}
	for _, r := range line[:idx] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
