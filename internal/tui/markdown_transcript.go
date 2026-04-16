package tui

import "strings"

func renderMarkdownTranscript(input string, width int) []string {
	text := normalizeMarkdownInput(strings.TrimSpace(input))
	if text == "" {
		return []string{""}
	}
	renderWidth := max(24, width)
	lines := splitLines(text)
	rendered := make([]string, 0, len(lines)*2)

	for i := 0; i < len(lines); {
		if block, consumed, ok := renderMarkdownTable(lines[i:], renderWidth); ok {
			rendered = append(rendered, block...)
			i += consumed
			continue
		}
		if block, consumed, ok := renderMarkdownCodeBlock(lines[i:], renderWidth); ok {
			rendered = append(rendered, block...)
			i += consumed
			continue
		}
		rendered = append(rendered, renderMarkdownLine(lines[i]))
		i++
	}

	return wrapMarkdownLines(rendered, renderWidth)
}
