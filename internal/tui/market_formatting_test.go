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

func TestColorizeR40ScoreUsesThresholdBands(t *testing.T) {
	if got := colorizeR40Score("65.00%", 0.65, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Bold(true).Render("65.00%") {
		t.Fatal("expected r40 scores above 60% to use the exceptional band style")
	}
	if got := colorizeR40Score("45.00%", 0.45, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Render("45.00%") {
		t.Fatal("expected r40 scores above 40% to use the very good band style")
	}
	if got := colorizeR40Score("30.00%", 0.30, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#D8C9B8")).Render("30.00%") {
		t.Fatal("expected r40 scores from 25% to 40% to use the good band style")
	}
	if got := colorizeR40Score("20.00%", 0.20, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Render("20.00%") {
		t.Fatal("expected r40 scores from 15% to 25% to use the mediocre band style")
	}
	if got := colorizeR40Score("10.00%", 0.10, true); got != lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73")).Render("10.00%") {
		t.Fatal("expected r40 scores below 15% to use the weak band style")
	}
}
