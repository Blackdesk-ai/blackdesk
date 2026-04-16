package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

func renderMarkdownTranscriptSafe(input string, width int) (lines []string) {
	defer func() {
		if recover() != nil {
			fallbackWidth := max(24, width)
			plain := lipgloss.NewStyle().Width(fallbackWidth).Render(strings.TrimSpace(input))
			lines = splitLines(plain)
			if len(lines) == 0 {
				lines = []string{""}
			}
		}
	}()
	return renderMarkdownTranscript(input, width)
}
