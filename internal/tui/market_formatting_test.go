package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorizeQARPScoreUsesThresholdBands(t *testing.T) {
	if got := colorizeQARPScore("1.60", 0.016, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Bold(true).Render("1.60") {
		t.Fatal("expected qarp scores above 1.5 to use the rare opportunity band style")
	}
	if got := colorizeQARPScore("1.30", 0.013, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Render("1.30") {
		t.Fatal("expected qarp scores above 1.2 to use the very good band style")
	}
	if got := colorizeQARPScore("1.00", 0.010, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Render("1.00") {
		t.Fatal("expected qarp scores from 0.8 to 1.2 to use the good band style")
	}
	if got := colorizeQARPScore("0.60", 0.006, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#A6A29D")).Render("0.60") {
		t.Fatal("expected qarp scores from 0.5 to 0.8 to use the fair band style")
	}
	if got := colorizeQARPScore("0.40", 0.004, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73")).Render("0.40") {
		t.Fatal("expected qarp scores below 0.5 to use the weak band style")
	}
	if got := colorizeQARPScore("0.80", 0.008, false); got != "0.80" {
		t.Fatal("expected unstyled qarp score to remain unchanged")
	}
}

func TestColorizeImpliedReturnScoreUsesThresholdBands(t *testing.T) {
	if got := colorizeImpliedReturnScore("6.00%", 0.06, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Render("6.00%") {
		t.Fatal("expected implied return above 5% to use the positive band style")
	}
	if got := colorizeImpliedReturnScore("5.00%", 0.05, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#A6A29D")).Render("5.00%") {
		t.Fatal("expected implied return at 5% to use the neutral band style")
	}
	if got := colorizeImpliedReturnScore("0.00%", 0, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#A6A29D")).Render("0.00%") {
		t.Fatal("expected implied return at 0% to use the neutral band style")
	}
	if got := colorizeImpliedReturnScore("-1.00%", -0.01, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73")).Render("-1.00%") {
		t.Fatal("expected implied return below 0% to use the negative band style")
	}
}

func TestColorizeImpliedSharpeScoreUsesThresholdBands(t *testing.T) {
	if got := colorizeImpliedSharpeScore("1.20", 1.2, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Bold(true).Render("1.20") {
		t.Fatal("expected implied sharpe above 1.0 to use the strong band style")
	}
	if got := colorizeImpliedSharpeScore("0.68", 0.68, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Render("0.68") {
		t.Fatal("expected implied sharpe from 0.5 to 1.0 to use the good band style")
	}
	if got := colorizeImpliedSharpeScore("0.20", 0.2, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#A6A29D")).Render("0.20") {
		t.Fatal("expected implied sharpe from 0.0 to 0.5 to use the neutral band style")
	}
	if got := colorizeImpliedSharpeScore("-0.10", -0.1, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73")).Render("-0.10") {
		t.Fatal("expected implied sharpe below 0.0 to use the weak band style")
	}
}
