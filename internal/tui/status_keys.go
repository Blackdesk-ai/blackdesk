package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) statusText() string {
	switch m.tabIdx {
	case tabMarkets:
		parts := []string{
			"Keys: " + renderStatusKeyHint("/", "search"),
			renderStatusKeyHint("Ctrl+K", "palette"),
			renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("i", "insight"),
			renderStatusInlineKeyHint("r", "refresh"),
		}
		parts = append(parts, renderStatusKeyHint("?", "help"))
		parts = m.appendUpdateStatusKey(parts)
		return strings.Join(parts, " | ")
	case tabScreener:
		parts := []string{
			"Keys: " + renderStatusKeyHint("/", "search"),
			renderStatusKeyHint("Ctrl+K", "palette"),
			renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("↑/↓", "results"),
			renderStatusKeyHint("←/→", "screeners"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("a", "add watchlist"),
			renderStatusInlineKeyHint("r", "refresh"),
			renderStatusInlineKeyHint("Enter", "open quote"),
		}
		parts = append(parts, renderStatusKeyHint("?", "help"))
		parts = m.appendUpdateStatusKey(parts)
		return strings.Join(parts, " | ")
	case tabNews:
		parts := []string{
			"Keys: " + renderStatusKeyHint("/", "search"),
			renderStatusKeyHint("Ctrl+K", "palette"),
			renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("↑/↓", "stories"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("r", "refresh"),
			renderStatusInlineKeyHint("o", "open story"),
		}
		parts = append(parts, renderStatusKeyHint("?", "help"))
		parts = m.appendUpdateStatusKey(parts)
		return strings.Join(parts, " | ")
	case tabAI:
		parts := []string{
			"Keys: " + renderStatusKeyHint("Ctrl+K", "palette"),
			renderStatusKeyHint(".", "ask AI"),
			renderStatusKeyHint("c", "connector/model"),
			renderStatusKeyHint("↑/↓", "scroll"),
			renderStatusKeyHint("Tab", "tabs"),
			renderStatusInlineKeyHint("f", "fullscreen"),
			renderStatusInlineKeyHint("r", "run"),
			renderStatusInlineKeyHint("x", "clear"),
		}
		parts = append(parts, renderStatusKeyHint("?", "help"))
		parts = m.appendUpdateStatusKey(parts)
		return strings.Join(parts, " | ")
	default:
		return m.quoteStatusText()
	}
}

func (m Model) quoteStatusText() string {
	parts := []string{
		"Keys: " + renderStatusKeyHint("/", "search"),
		renderStatusKeyHint("Ctrl+K", "palette"),
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
	parts = m.appendUpdateStatusKey(parts)
	return strings.Join(parts, " | ")
}

func (m Model) appendUpdateStatusKey(parts []string) []string {
	if !m.updateAvailable || strings.TrimSpace(m.latestVersion) == "" || m.upgradeRunning {
		return parts
	}
	return append(parts, renderStatusUpdateKeyHint("u", "update app"))
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

func renderStatusUpdateKeyHint(key, label string) string {
	hotkeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394"))
	return hotkeyStyle.Render(key) + " " + labelStyle.Render(label)
}
