package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) statusText() string {
	switch m.tabIdx {
	case tabMarkets:
		return strings.Join([]string{
			"Keys: " + renderStatusKeyHint("/", "search"),
			renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("i", "insight"),
			renderStatusInlineKeyHint("r", "refresh"),
			renderStatusKeyHint("?", "help"),
		}, " | ")
	case tabScreener:
		return strings.Join([]string{
			"Keys: " + renderStatusKeyHint("/", "search"),
			renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("↑/↓", "results"),
			renderStatusKeyHint("←/→", "screeners"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("a", "add watchlist"),
			renderStatusInlineKeyHint("r", "refresh"),
			renderStatusInlineKeyHint("Enter", "open quote"),
			renderStatusKeyHint("?", "help"),
		}, " | ")
	case tabNews:
		return strings.Join([]string{
			"Keys: " + renderStatusKeyHint("/", "search"),
			renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("↑/↓", "stories"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("r", "refresh"),
			renderStatusInlineKeyHint("o", "open story"),
			renderStatusKeyHint("?", "help"),
		}, " | ")
	case tabAI:
		return strings.Join([]string{
			"Keys: " + renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("c", "connector/model"),
			renderStatusKeyHint("↑/↓", "scroll"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("f", "fullscreen"),
			renderStatusInlineKeyHint("r", "run"),
			renderStatusInlineKeyHint("x", "clear"),
			renderStatusKeyHint("?", "help"),
		}, " | ")
	default:
		return m.quoteStatusText()
	}
}

func (m Model) quoteStatusText() string {
	parts := []string{
		"Keys: " + renderStatusKeyHint("/", "search"),
		renderStatusKeyHint(".", "ask AI"),
		renderStatusKeyHint("Tab", "tabs"),
	}
	if m.quoteCenterMode != quoteCenterChart {
		parts = append(parts, renderStatusInlineKeyHint("c", "chart"))
	}
	if m.quoteCenterMode != quoteCenterFundamentals {
		parts = append(parts, renderStatusInlineKeyHint("f", "fundamentals"))
	}
	if m.quoteCenterMode != quoteCenterTechnicals {
		parts = append(parts, renderStatusInlineKeyHint("t", "technicals"))
	}
	if m.services.HasStatements() && m.quoteCenterMode != quoteCenterStatements {
		parts = append(parts, renderStatusInlineKeyHint("s", "statements"))
	}
	if m.services.HasInsiders() && m.quoteCenterMode != quoteCenterInsiders {
		parts = append(parts, renderStatusInlineKeyHint("h", "insiders"))
	}
	parts = append(parts, renderStatusInlineKeyHint("r", "refresh"))
	parts = append(parts, renderStatusKeyHint("?", "help"))
	return strings.Join(parts, " | ")
}

func renderStatusKeyHint(key, label string) string {
	hotkeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Bold(true)
	return hotkeyStyle.Render(key) + " " + label
}

func renderStatusInlineKeyHint(key, label string) string {
	hotkeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Bold(true)
	idx := strings.Index(strings.ToLower(label), strings.ToLower(key))
	if idx < 0 || key == "" {
		return renderStatusKeyHint(key, label)
	}
	return label[:idx] + hotkeyStyle.Render(label[idx:idx+len(key)]) + label[idx+len(key):]
}
