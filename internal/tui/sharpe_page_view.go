package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func renderQuoteSharpeBoard(section, label, muted, pos, neg lipgloss.Style, width, height, rangeIdx int, series domain.PriceSeries) string {
	var b strings.Builder
	b.WriteString(section.Render("SHARPE") + "\n\n")
	chartSeries := displaySharpeSeriesForRange(buildSharpeChartSeries(series), ranges[rangeIdx].Range)
	if len(chartSeries.Candles) == 0 {
		b.WriteString(renderWrappedTextBlock(muted, "No Sharpe history loaded for the active symbol yet.", width))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	chartHeight := max(9, height-7)

	plotWidth := ui.ChartPlotWidth(width)
	leftPad := strings.Repeat(" ", ui.ChartPlotPad())
	b.WriteString(ui.RenderLineChartWithReference(chartSeries.Candles, width, chartHeight, 0) + "\n")
	b.WriteString(muted.Render(ui.RenderTimeAxis(chartSeries.Candles, plotWidth)) + "\n")
	b.WriteString("\n")
	b.WriteString(leftPad + section.Render("TIMEFRAMES") + " ")
	selectedTimeframe := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1C1712")).Background(lipgloss.Color("#E7B66B")).Padding(0, 1)
	for i, item := range ranges {
		text := item.Label
		if i == rangeIdx {
			b.WriteString(selectedTimeframe.Render(text) + " ")
			continue
		}
		b.WriteString(muted.Render(text) + " ")
	}
	b.WriteString(section.Render("←/→"))
	b.WriteString("\n\n")
	b.WriteString(leftPad + muted.Render(renderSharpeSeriesSummary(chartSeries)))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderQuoteSharpePreview(label, muted, pos, neg lipgloss.Style, width, height int, chartSeries domain.PriceSeries) string {
	var b strings.Builder

	stats, ok := sharpeSeriesStats(chartSeries)
	if !ok {
		b.WriteString(muted.Render("Sharpe preview becomes available after enough history loads."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	latestStyle := lipgloss.NewStyle()
	if stats.Latest > 0 {
		latestStyle = pos
	} else if stats.Latest < 0 {
		latestStyle = neg
	}
	b.WriteString(muted.Render("Latest") + "\n")
	b.WriteString(latestStyle.Render(formatMetricSigned(stats.Latest)) + "\n")
	b.WriteString("\n" + muted.Render("Range") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("Best %s  •  Worst %s", formatMetricSigned(stats.Max), formatMetricSigned(stats.Min)), width))
	b.WriteString("\n\n" + muted.Render("Central Tendency") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("Average %s  •  Median %s", formatMetricSigned(stats.Mean), formatMetricSigned(stats.Median)), width))
	b.WriteString("\n\n" + muted.Render("Hit Rate") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("Positive readings %d%% of observations", stats.PositivePct), width))
	b.WriteString("\n\n" + muted.Render("Window") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("Computed from %d displayed observations using extended history so the Sharpe path covers about 5 years.", len(chartSeries.Candles)), width))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func buildSharpeChartSeries(series domain.PriceSeries) domain.PriceSeries {
	closes := extractCloses(series.Candles)
	if len(closes) <= 252 {
		return domain.PriceSeries{Symbol: series.Symbol, Range: series.Range, Interval: series.Interval}
	}
	candles := make([]domain.Candle, 0, len(series.Candles)-252)
	for i := 252; i < len(series.Candles); i++ {
		window := closes[:i+1]
		value := returnOverVol(lookbackReturn(window, 252), historicalVolatility(window, 252), 1)
		candles = append(candles, domain.Candle{
			Time:  series.Candles[i].Time,
			Open:  value,
			High:  value,
			Low:   value,
			Close: value,
		})
	}
	return domain.PriceSeries{
		Symbol:      series.Symbol,
		Range:       series.Range,
		Interval:    series.Interval,
		Candles:     candles,
		Freshness:   series.Freshness,
		LastUpdated: series.LastUpdated,
	}
}

func displaySharpeSeriesForRange(series domain.PriceSeries, rangeKey string) domain.PriceSeries {
	if len(series.Candles) <= 2 {
		return series
	}
	last := series.Candles[len(series.Candles)-1].Time
	cutoff := last
	switch rangeKey {
	case "1d":
		cutoff = last.AddDate(0, 0, -1)
	case "5d":
		cutoff = last.AddDate(0, 0, -5)
	case "1mo":
		cutoff = last.AddDate(0, -1, 0)
	case "3mo":
		cutoff = last.AddDate(0, -3, 0)
	case "6mo":
		cutoff = last.AddDate(0, -6, 0)
	case "1y":
		cutoff = last.AddDate(-1, 0, 0)
	default:
		cutoff = last.AddDate(-5, 0, 0)
	}
	start := 0
	for i, candle := range series.Candles {
		if !candle.Time.Before(cutoff) {
			start = i
			break
		}
	}
	if len(series.Candles)-start < 2 {
		start = max(0, len(series.Candles)-2)
	}
	trimmed := append([]domain.Candle(nil), series.Candles[start:]...)
	return domain.PriceSeries{
		Symbol:      series.Symbol,
		Range:       rangeKey,
		Interval:    series.Interval,
		Candles:     trimmed,
		Freshness:   series.Freshness,
		LastUpdated: series.LastUpdated,
	}
}

func renderSharpeSeriesSummary(series domain.PriceSeries) string {
	stats, ok := sharpeSeriesStats(series)
	if !ok {
		return "No Sharpe history"
	}
	return fmt.Sprintf("%s %s | latest %s | avg %s | best %s | worst %s", series.Range, series.Interval, formatMetricSigned(stats.Latest), formatMetricSigned(stats.Mean), formatMetricSigned(stats.Max), formatMetricSigned(stats.Min))
}

type sharpeStats struct {
	Latest      float64
	Mean        float64
	Median      float64
	Min         float64
	Max         float64
	PositivePct int
}

func sharpeSeriesStats(series domain.PriceSeries) (sharpeStats, bool) {
	if len(series.Candles) == 0 {
		return sharpeStats{}, false
	}
	values := make([]float64, 0, len(series.Candles))
	positive := 0
	sum := 0.0
	minValue := series.Candles[0].Close
	maxValue := series.Candles[0].Close
	for _, candle := range series.Candles {
		value := candle.Close
		values = append(values, value)
		sum += value
		if value > 0 {
			positive++
		}
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	sort.Float64s(values)
	median := values[len(values)/2]
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	}
	return sharpeStats{
		Latest:      series.Candles[len(series.Candles)-1].Close,
		Mean:        sum / float64(len(series.Candles)),
		Median:      median,
		Min:         minValue,
		Max:         maxValue,
		PositivePct: int(float64(positive) / float64(len(series.Candles)) * 100),
	}, true
}
