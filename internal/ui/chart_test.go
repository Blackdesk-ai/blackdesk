package ui

import (
	"strings"
	"testing"
	"time"

	"blackdesk/internal/domain"
	"github.com/charmbracelet/lipgloss"
)

func TestRenderVolumeStripUsesRequestedWidth(t *testing.T) {
	candles := []domain.Candle{
		{Time: time.Now().AddDate(0, 0, -4), Volume: 10},
		{Time: time.Now().AddDate(0, 0, -3), Volume: 20},
		{Time: time.Now().AddDate(0, 0, -2), Volume: 30},
		{Time: time.Now().AddDate(0, 0, -1), Volume: 40},
	}

	got := RenderVolumeStrip(candles, 20)
	if gotWidth := len([]rune(strings.TrimPrefix(got, strings.Repeat(" ", chartPlotPad)))); gotWidth != 20 {
		t.Fatalf("expected width 20, got %d", gotWidth)
	}
}

func TestRenderTimeAxisUsesRequestedWidth(t *testing.T) {
	candles := []domain.Candle{
		{Time: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)},
		{Time: time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC)},
		{Time: time.Date(2026, time.February, 2, 0, 0, 0, 0, time.UTC)},
		{Time: time.Date(2026, time.February, 16, 0, 0, 0, 0, time.UTC)},
		{Time: time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)},
		{Time: time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC)},
		{Time: time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)},
	}

	got := RenderTimeAxis(candles, 48)
	if gotWidth := len([]rune(strings.TrimPrefix(got, strings.Repeat(" ", chartPlotPad)))); gotWidth != 48 {
		t.Fatalf("expected width 48, got %d", gotWidth)
	}
	if !strings.Contains(got, "Jan 2") || !strings.Contains(got, "Apr 2") {
		t.Fatal("expected axis labels to include timeframe endpoints")
	}
	if !strings.Contains(got, "Feb 2") || !strings.Contains(got, "Mar 2") {
		t.Fatal("expected axis labels to include intermediate time markers")
	}
}

func TestDownsampleClosesPreservesLastPrice(t *testing.T) {
	candles := []domain.Candle{
		{Close: 10},
		{Close: 12},
		{Close: 14},
		{Close: 16},
		{Close: 18},
		{Close: 21},
	}

	got := downsampleCloses(candles, 3)
	if got[len(got)-1] != 21 {
		t.Fatalf("expected last close 21, got %v", got[len(got)-1])
	}
}

func TestRenderLineChartUsesRequestedWidthAndRightAxis(t *testing.T) {
	candles := []domain.Candle{
		{Time: time.Now().AddDate(0, 0, -4), Close: 100},
		{Time: time.Now().AddDate(0, 0, -3), Close: 105},
		{Time: time.Now().AddDate(0, 0, -2), Close: 103},
		{Time: time.Now().AddDate(0, 0, -1), Close: 110},
	}

	got := RenderLineChart(candles, 48, 8)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < 8 {
		t.Fatalf("expected at least 8 chart rows, got %d", len(lines))
	}
	for _, line := range lines {
		if width := lipgloss.Width(line); width != 48 {
			t.Fatalf("expected row width 48, got %d: %q", width, line)
		}
	}
	for _, line := range lines {
		runes := []rune(line)
		left := string(runes[:chartOffset])
		right := string(runes[len(runes)-chartAxisWidth:])
		if left != right && right != priceAxisLabel(candles[len(candles)-1].Close) {
			t.Fatalf("expected mirrored axes, got left=%q right=%q in line %q", left, right, line)
		}
	}
	if !strings.Contains(got, priceAxisLabel(candles[len(candles)-1].Close)) {
		t.Fatalf("expected chart to include last price label %q", priceAxisLabel(candles[len(candles)-1].Close))
	}
}

func TestRenderLineChartWithReferenceAddsHorizontalBaseline(t *testing.T) {
	candles := []domain.Candle{
		{Time: time.Now().AddDate(0, 0, -4), Close: -2},
		{Time: time.Now().AddDate(0, 0, -3), Close: 1},
		{Time: time.Now().AddDate(0, 0, -2), Close: -1},
		{Time: time.Now().AddDate(0, 0, -1), Close: 2},
	}

	got := RenderLineChartWithReference(candles, 48, 8, 0)
	if !strings.Contains(got, "┈") {
		t.Fatal("expected chart to include a horizontal reference line")
	}
}

func TestChartYStepShowsMorePriceLevels(t *testing.T) {
	if got := chartYStep(9); got != 2 {
		t.Fatalf("expected y step 2 for height 9, got %d", got)
	}
	if got := chartYStep(12); got != 3 {
		t.Fatalf("expected y step 3 for height 12, got %d", got)
	}
}
