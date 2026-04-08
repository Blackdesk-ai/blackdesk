package tui

import "github.com/charmbracelet/lipgloss"

func marketMoveStyle(move float64) lipgloss.Style {
	style := lipgloss.NewStyle()
	switch {
	case move > 0:
		return style.Foreground(lipgloss.Color("#62D394"))
	case move < 0:
		return style.Foreground(lipgloss.Color("#FF7A73"))
	default:
		return style.Foreground(lipgloss.Color("#A6A29D"))
	}
}

func colorizeMarketPrice(text string, move float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return marketMoveStyle(move).Render(text)
}

func colorizeMarketChange(text string, move float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return marketMoveStyle(move).Render(text)
}
