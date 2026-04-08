package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
)

func TestMarketsMetricLabelsUseAnalystsLabelStyle(t *testing.T) {
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D8C9B8"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3EBDD"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#9F907E"))

	boardCard := renderMarketBoardCard(sectionStyle, labelStyle, muted, Model{}, 28, marketSectionBlock{
		title:      "INDEX FUTURES",
		valueLabel: "Level",
		items:      []marketBoardItem{{label: "S&P Fut", symbol: "ES=F"}},
	})
	if !strings.Contains(boardCard, labelStyle.Render("S&P Fut")) {
		t.Fatal("expected market board row labels to use analysts label style")
	}
	if !strings.Contains(ansi.Strip(boardCard), "Level") {
		t.Fatal("expected market board header to use a contextual value label")
	}

	panel := ansi.Strip(renderMarketTableHeaderWithValueLabel(28, "Price", labelStyle))
	if !strings.Contains(panel, "Name") {
		t.Fatal("expected market table header to render")
	}

	yieldHeader := ansi.Strip(renderMarketBoardHeader(28, "Yield", labelStyle))
	if !strings.Contains(yieldHeader, "Yield") {
		t.Fatal("expected yield market board header to avoid generic price label")
	}

	tablePanel := renderMarketTableRow(marketTableRow{name: "Europe", price: "--", chg: "--"}, 28, labelStyle)
	if !strings.Contains(tablePanel, labelStyle.Render("Europe")) {
		t.Fatal("expected market table row labels to use analysts label style")
	}
}

func TestMarketNewsSourceRowsHidesZeroCounts(t *testing.T) {
	rows := marketNewsSourceRows(
		[]domain.MarketNewsSource{
			{Name: "Reuters"},
			{Name: "AP"},
			{Name: "Bloomberg"},
		},
		[]domain.NewsItem{
			{Publisher: "Bloomberg", Title: "One"},
			{Publisher: "Bloomberg", Title: "Two"},
		},
	)

	if len(rows) != 1 {
		t.Fatalf("expected 1 source row (zero-count hidden), got %d", len(rows))
	}
	if rows[0].name != "Bloomberg" || rows[0].price != "2" {
		t.Fatalf("expected Bloomberg count row, got %+v", rows[0])
	}
}
