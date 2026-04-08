package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestScreenerPrimaryMetricSkipsRelativeVolumeWhenOtherMetricsExist(t *testing.T) {
	result := domain.ScreenerResult{
		Items: []domain.ScreenerItem{
			{
				Metrics: []domain.ScreenerMetric{
					{Key: "relative_volume", Label: "RV", Value: "1.42x"},
					{Key: "forward_pe", Label: "Fwd P/E", Value: "15.20"},
				},
			},
		},
	}

	key, label := screenerPrimaryMetric(result)
	if key != "forward_pe" || label != "Fwd P/E" {
		t.Fatalf("expected primary metric to skip RV when a screener-specific metric exists, got %q / %q", key, label)
	}
}

func TestScreenerMetricValueReturnsRelativeVolume(t *testing.T) {
	item := domain.ScreenerItem{
		Metrics: []domain.ScreenerMetric{
			{Key: "relative_volume", Label: "RV", Value: "1.42x"},
		},
	}

	if got := screenerMetricValue(item, "relative_volume"); got != "1.42x" {
		t.Fatalf("expected RV metric value, got %q", got)
	}
	if got := screenerMetricValue(item, ""); got != "1.42x" {
		t.Fatalf("expected RV fallback metric value, got %q", got)
	}
}

func TestScreenerColumnWidthsStayEven(t *testing.T) {
	prefixWidth, widths := screenerColumnWidths(140)
	if prefixWidth != 2 {
		t.Fatalf("expected fixed prefix width, got %d", prefixWidth)
	}

	minWidth, maxWidth := widths[0], widths[0]
	total := prefixWidth + 6
	for _, width := range widths {
		total += width
		if width < minWidth {
			minWidth = width
		}
		if width > maxWidth {
			maxWidth = width
		}
	}

	if total != 140 {
		t.Fatalf("expected screener columns to fill width exactly, got %d", total)
	}
	if maxWidth-minWidth > 1 {
		t.Fatalf("expected screener columns to stay balanced, got min=%d max=%d", minWidth, maxWidth)
	}
}

func TestNumberKeyFourSelectsScreenerTab(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabMarkets

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(Model)

	if model.tabIdx != tabScreener {
		t.Fatalf("expected key 4 to select Screeners tab, got %d", model.tabIdx)
	}
}

func TestChangingScreenerKeepsResultSelectionIndex(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.screenerDefs = []domain.ScreenerDefinition{
		{ID: "most_actives", Name: "Most Active"},
		{ID: "day_gainers", Name: "Day Gainers"},
	}
	model.screenerIdx = 0
	model.screenerSel = 4
	model.screenerScroll = 2

	model.cycleScreener(1)

	if model.screenerIdx != 1 {
		t.Fatalf("expected screener index to advance, got %d", model.screenerIdx)
	}
	if model.screenerSel != 4 {
		t.Fatalf("expected result selection index to stay stable, got %d", model.screenerSel)
	}
	if model.screenerScroll != 2 {
		t.Fatalf("expected screener scroll to stay stable, got %d", model.screenerScroll)
	}
}
