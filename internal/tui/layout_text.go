package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func clipLines(content string, height int) string {
	lines := splitLines(content)
	if len(lines) > height {
		lines = lines[:height]
	}
	if len(lines) < height {
		lines = append(lines, make([]string, height-len(lines))...)
	}
	return strings.Join(lines, "\n")
}

func splitLines(content string) []string {
	if content == "" {
		return []string{""}
	}
	return strings.Split(content, "\n")
}

func truncateText(content string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(content)
	if len(runes) <= width {
		return content
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func renderHeaderTabs(tabStyle, activeTabStyle lipgloss.Style, activeIdx int) string {
	parts := make([]string, 0, len(headerTabs))
	for i, tab := range headerTabs {
		if i == activeIdx {
			parts = append(parts, activeTabStyle.Render(tab))
			continue
		}
		parts = append(parts, tabStyle.Render(tab))
	}
	return strings.Join(parts, " | ")
}

func renderHeader(titleStyle lipgloss.Style, width int, left, right string) string {
	_ = titleStyle
	if width <= 0 {
		return ""
	}
	if width < 24 {
		return ansi.Truncate(left, width, "...")
	}

	leftWidthVisible := lipgloss.Width(left)
	rightWidthVisible := lipgloss.Width(right)
	if leftWidthVisible+1+rightWidthVisible <= width {
		spacerWidth := max(1, width-leftWidthVisible-rightWidthVisible)
		return left + strings.Repeat(" ", spacerWidth) + right
	}

	minLeftWidth := min(leftWidthVisible, max(12, width/3))
	maxRightWidth := max(0, width-minLeftWidth-1)
	if maxRightWidth <= 0 {
		return ansi.Truncate(left, width, "")
	}
	if rightWidthVisible > maxRightWidth {
		right = ansi.Truncate(right, maxRightWidth, "...")
		rightWidthVisible = lipgloss.Width(right)
	}
	leftWidth := max(0, width-rightWidthVisible-1)
	if leftWidth <= 0 {
		return ansi.Truncate(right, width, "...")
	}
	left = ansi.Truncate(left, leftWidth, "...")
	return left + strings.Repeat(" ", max(1, width-lipgloss.Width(left)-rightWidthVisible)) + right
}

func renderStatusLine(width int, left, right string) string {
	if width <= 0 {
		return ""
	}
	if width < 24 {
		return truncateText(left, width)
	}

	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if right == "" {
		return truncateText(left, width)
	}
	if left == "" {
		return truncateText(right, width)
	}

	rightWidth := lipgloss.Width(right)
	if rightWidth >= width {
		return truncateText(right, width)
	}
	leftWidth := width - rightWidth - 1
	if leftWidth <= 0 {
		return truncateText(right, width)
	}
	leftText := padRight(truncateText(left, leftWidth), leftWidth)
	return leftText + strings.Repeat(" ", max(1, width-lipgloss.Width(leftText)-lipgloss.Width(right))) + right
}
