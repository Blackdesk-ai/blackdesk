package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

type sharpeSeriesSpec struct {
	ShortLabel string
	Lookback   int
}

type sharpeChartSeries struct {
	Spec      sharpeSeriesSpec
	Series    domain.PriceSeries
	Forward3M domain.PriceSeries
}

var sharpeSeriesSpecs = []sharpeSeriesSpec{
	{ShortLabel: "252d", Lookback: 252},
	{ShortLabel: "63d", Lookback: 63},
}

func renderQuoteSharpeBoard(section, label, muted, pos, neg lipgloss.Style, width, height, rangeIdx int, series domain.PriceSeries) string {
	var b strings.Builder
	b.WriteString(section.Render("RISK ADJUSTED") + "\n\n")
	chartSeries := displaySharpeSeriesForRange(buildSharpeChartSeries(series), ranges[rangeIdx].Range)
	if len(chartSeries.Candles) == 0 {
		b.WriteString(renderWrappedTextBlock(muted, "No risk-adjusted history loaded for the active symbol yet.", width))
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

func renderQuoteSharpePreview(label, muted, pos, neg lipgloss.Style, width, height int, sourceSeries domain.PriceSeries, chartSeries []sharpeChartSeries) string {
	var b strings.Builder

	stats := sharpeSeriesPreviewStats(chartSeries)
	if len(stats) == 0 {
		b.WriteString(muted.Render("Risk-adjusted preview becomes available after enough history loads."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString(muted.Render("Latest") + "\n")
	for _, stat := range stats {
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render(stat.Label), renderSharpeValue(pos, neg, muted, stat.Latest)))
	}
	if delta, ok := sharpeLatestDelta(stats); ok {
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Δ"), renderSharpeValue(pos, neg, muted, delta)))
	}

	b.WriteString("\n" + muted.Render("Range") + "\n")
	for idx, stat := range stats {
		line := fmt.Sprintf("%s %s %s  •  %s %s", label.Render(stat.Label), muted.Render("Best"), renderSharpeValue(pos, neg, muted, stat.Max), muted.Render("Worst"), renderSharpeValue(pos, neg, muted, stat.Min))
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), line, width))
		if idx < len(stats)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n" + muted.Render("Central Tendency") + "\n")
	for idx, stat := range stats {
		line := fmt.Sprintf("%s %s %s  •  %s %s", label.Render(stat.Label), muted.Render("avg."), renderSharpeValue(pos, neg, muted, stat.Mean), muted.Render("mdn."), renderSharpeValue(pos, neg, muted, stat.Median))
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), line, width))
		if idx < len(stats)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n" + muted.Render("Hit Rate") + "\n")
	for idx, stat := range stats {
		line := fmt.Sprintf("%s %s %s  •  %s %s", label.Render(stat.Label), muted.Render("> 0"), renderSharpePercent(pos, muted, stat.PositivePct), muted.Render("> 1"), renderSharpePercent(pos, muted, stat.AboveOnePct))
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), line, width))
		if idx < len(stats)-1 {
			b.WriteString("\n")
		}
	}

	if forwardStat, ok := sharpeForwardStat(stats); ok {
		b.WriteString("\n\n")
		b.WriteString(muted.Render("Fwd. Return") + "\n")
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), fmt.Sprintf("%s %s", label.Render("3M Avg"), renderSharpeReturn(pos, neg, muted, forwardStat.Forward3MMean)), width))
		b.WriteString("\n")
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), fmt.Sprintf("%s %s", label.Render("3M Median"), renderSharpeReturn(pos, neg, muted, forwardStat.Forward3MMedian)), width))
		b.WriteString("\n")
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), fmt.Sprintf("%s %s", label.Render("3M Win%"), renderSharpePercent(pos, muted, forwardStat.Forward3MPositivePct)), width))
		points := buildStatisticsPoints(sourceSeries)
		if len(points) > 0 {
			latest := points[len(points)-1]
			for _, regime := range statisticsCurrentSignalEVs(sourceSeries, latest, statisticsHorizon{Label: "3M", Forward: 63}) {
				b.WriteString("\n")
				b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), fmt.Sprintf("%s %s", label.Render("EV "+regime.Label), renderSharpeReturn(pos, neg, muted, regime.EV)), width))
			}
		}
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderSharpeValue(pos, neg, muted lipgloss.Style, value float64) string {
	text := formatMetricSigned(value)
	if value > 0 {
		return pos.Render(text)
	}
	if value < 0 {
		return neg.Render(text)
	}
	return muted.Render(text)
}

func renderSharpePercent(pos, muted lipgloss.Style, value int) string {
	text := fmt.Sprintf("%d%%", value)
	if value > 50 {
		return pos.Render(text)
	}
	return muted.Render(text)
}

func renderSharpeReturn(pos, neg, muted lipgloss.Style, value float64) string {
	text := formatSignedPercentRatio(value)
	if value > 0 {
		return pos.Render(text)
	}
	if value < 0 {
		return neg.Render(text)
	}
	return muted.Render(text)
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

func buildSharpePreviewSeries(series domain.PriceSeries, lookback int) domain.PriceSeries {
	closes := extractCloses(series.Candles)
	if len(closes) <= lookback {
		return domain.PriceSeries{Symbol: series.Symbol, Range: series.Range, Interval: series.Interval}
	}
	candles := make([]domain.Candle, 0, len(series.Candles)-lookback)
	for i := lookback; i < len(series.Candles); i++ {
		window := closes[:i+1]
		value := returnOverVol(annualizedLookbackReturn(window, lookback), historicalVolatility(window, lookback), 1)
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

func buildSharpePreviewSeriesSet(series domain.PriceSeries) []sharpeChartSeries {
	out := make([]sharpeChartSeries, 0, len(sharpeSeriesSpecs))
	for _, spec := range sharpeSeriesSpecs {
		out = append(out, sharpeChartSeries{
			Spec:      spec,
			Series:    buildSharpePreviewSeries(series, spec.Lookback),
			Forward3M: buildSharpeForwardReturnSeries(series, spec.Lookback, 63),
		})
	}
	return out
}

func buildSharpeForwardReturnSeries(series domain.PriceSeries, lookback, forward int) domain.PriceSeries {
	closes := extractCloses(series.Candles)
	if lookback <= 0 || forward <= 0 || len(closes) <= lookback+forward {
		return domain.PriceSeries{Symbol: series.Symbol, Range: series.Range, Interval: series.Interval}
	}
	candles := make([]domain.Candle, 0, len(series.Candles)-lookback-forward)
	for i := lookback; i+forward < len(series.Candles); i++ {
		base := closes[i]
		future := closes[i+forward]
		if base <= 0 || future <= 0 {
			continue
		}
		value := future/base - 1
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

func displaySharpeChartSeriesForRange(series []sharpeChartSeries, rangeKey string) []sharpeChartSeries {
	out := make([]sharpeChartSeries, 0, len(series))
	for _, item := range series {
		out = append(out, sharpeChartSeries{
			Spec:      item.Spec,
			Series:    displaySharpeSeriesForRange(item.Series, rangeKey),
			Forward3M: displaySharpeSeriesForRange(item.Forward3M, rangeKey),
		})
	}
	return out
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
		return "No risk-adjusted history"
	}
	return fmt.Sprintf("%s %s | latest %s | avg %s | best %s | worst %s", series.Range, series.Interval, formatMetricSigned(stats.Latest), formatMetricSigned(stats.Mean), formatMetricSigned(stats.Max), formatMetricSigned(stats.Min))
}

type sharpePreviewStat struct {
	Label                string
	Latest               float64
	Mean                 float64
	Median               float64
	Min                  float64
	Max                  float64
	PositivePct          int
	AboveOnePct          int
	Forward3MMean        float64
	Forward3MMedian      float64
	Forward3MPositivePct int
	Forward3MOK          bool
}

type sharpeStats struct {
	Latest      float64
	Mean        float64
	Median      float64
	Min         float64
	Max         float64
	PositivePct int
	AboveOnePct int
}

func sharpeSeriesStats(series domain.PriceSeries) (sharpeStats, bool) {
	if len(series.Candles) == 0 {
		return sharpeStats{}, false
	}
	values := make([]float64, 0, len(series.Candles))
	positive := 0
	aboveOne := 0
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
		if value > 1 {
			aboveOne++
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
		AboveOnePct: int(float64(aboveOne) / float64(len(series.Candles)) * 100),
	}, true
}

func sharpeSeriesPreviewStats(series []sharpeChartSeries) []sharpePreviewStat {
	out := make([]sharpePreviewStat, 0, len(series))
	for _, item := range series {
		stats, ok := sharpeSeriesStats(item.Series)
		if !ok {
			continue
		}
		preview := sharpePreviewStat{
			Label:       item.Spec.ShortLabel,
			Latest:      stats.Latest,
			Mean:        stats.Mean,
			Median:      stats.Median,
			Min:         stats.Min,
			Max:         stats.Max,
			PositivePct: stats.PositivePct,
			AboveOnePct: stats.AboveOnePct,
		}
		if forwardStats, ok := sharpeSeriesStats(item.Forward3M); ok {
			preview.Forward3MMean = forwardStats.Mean
			preview.Forward3MMedian = forwardStats.Median
			preview.Forward3MPositivePct = forwardStats.PositivePct
			preview.Forward3MOK = true
		}
		out = append(out, preview)
	}
	return out
}

func sharpeForwardStat(stats []sharpePreviewStat) (sharpePreviewStat, bool) {
	for _, stat := range stats {
		if stat.Forward3MOK && stat.Label == "252d" {
			return stat, true
		}
	}
	for _, stat := range stats {
		if stat.Forward3MOK {
			return stat, true
		}
	}
	return sharpePreviewStat{}, false
}

func sharpeLatestDelta(stats []sharpePreviewStat) (float64, bool) {
	var latest252 float64
	var latest63 float64
	have252 := false
	have63 := false
	for _, stat := range stats {
		switch stat.Label {
		case "252d":
			latest252 = stat.Latest
			have252 = true
		case "63d":
			latest63 = stat.Latest
			have63 = true
		}
	}
	if !have252 || !have63 {
		return 0, false
	}
	return latest63 - latest252, true
}

func annualizedLookbackReturn(values []float64, periods int) float64 {
	if periods <= 0 || len(values) <= periods {
		return 0
	}
	base := values[len(values)-1-periods]
	last := values[len(values)-1]
	if base <= 0 || last <= 0 {
		return 0
	}
	return math.Pow(last/base, 252.0/float64(periods)) - 1
}
