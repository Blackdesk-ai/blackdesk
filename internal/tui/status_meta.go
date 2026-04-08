package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/buildinfo"
)

func (m Model) statusMetaText() string {
	provider := "unknown"
	if name := strings.TrimSpace(m.services.ActiveProviderName()); name != "" {
		switch strings.ToLower(name) {
		case "yahoo":
			provider = "Yahoo Finance"
		default:
			provider = name
		}
	}
	ai := m.activeAIStatusLabel()
	return "Source: " + provider + " | AI: " + ai + " | " + m.versionStatusLabel()
}

func (m Model) renderStatusMeta(muted, accent lipgloss.Style) string {
	provider := "unknown"
	if name := strings.TrimSpace(m.services.ActiveProviderName()); name != "" {
		switch strings.ToLower(name) {
		case "yahoo":
			provider = "Yahoo Finance"
		default:
			provider = name
		}
	}

	parts := []string{
		muted.Render("Source: " + provider),
		muted.Render("AI: " + m.activeAIStatusLabel()),
	}
	if m.updateAvailable && strings.TrimSpace(m.latestVersion) != "" {
		parts = append(parts, accent.Render(m.versionStatusLabel()))
	} else {
		parts = append(parts, muted.Render(m.versionStatusLabel()))
	}
	return strings.Join(parts, muted.Render(" | "))
}

func (m Model) hasSufficientMarketOpinionData() bool {
	loaded := 0
	for _, symbol := range marketDashboardSymbols {
		if _, ok := m.watchQuotes[strings.ToUpper(symbol)]; ok {
			loaded++
		}
	}
	return loaded >= 12
}

func (m Model) versionStatusLabel() string {
	current := buildinfo.VersionLabel(m.appVersion)
	if m.updateAvailable && strings.TrimSpace(m.latestVersion) != "" {
		return current + " -> " + buildinfo.VersionLabel(m.latestVersion)
	}
	return current
}
