package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"blackdesk/internal/domain"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/lipgloss"
)

const chartOffset = 8
const chartAxisGap = 1
const chartAxisWidth = chartOffset
const chartPlotPad = chartOffset + 1

func RenderChartSummary(series domain.PriceSeries) string {
	if len(series.Candles) == 0 {
		return "No price history"
	}
	first := series.Candles[0].Close
	last := series.Candles[len(series.Candles)-1].Close
	change := last - first
	changePct := 0.0
	if first != 0 {
		changePct = change / first * 100
	}
	return fmt.Sprintf("%s %s | %d candles | %.2f -> %.2f (%+.2f%%)", series.Range, series.Interval, len(series.Candles), first, last, changePct)
}

func ChartPlotWidth(totalWidth int) int {
	return max(8, totalWidth-chartOffset-chartAxisGap-chartAxisWidth)
}

func ChartPlotPad() int {
	return chartPlotPad
}

func RenderLineChart(candles []domain.Candle, width, height int) string {
	if len(candles) == 0 || width < 8 || height < 4 {
		return "no chart data"
	}
	bodyWidth := max(8, width-chartAxisGap-chartAxisWidth)
	plotWidth := max(1, bodyWidth-chartPlotPad)
	reduced := downsampleCandles(candles, max(2, plotWidth*2))
	if len(reduced) == 0 {
		return "no chart data"
	}
	series := make([]float64, 0, len(reduced))
	points := make([]timeserieslinechart.TimePoint, 0, len(reduced))
	for _, candle := range reduced {
		series = append(series, candle.Close)
		points = append(points, timeserieslinechart.TimePoint{
			Time:  candle.Time.UTC(),
			Value: candle.Close,
		})
	}
	minTime := points[0].Time
	maxTime := points[len(points)-1].Time
	if !maxTime.After(minTime) {
		maxTime = minTime.Add(time.Second)
	}
	minY, maxY := chartYRange(series)
	yStep := chartYStep(height)

	chart := timeserieslinechart.New(
		bodyWidth,
		height,
		timeserieslinechart.WithXYSteps(0, yStep),
		timeserieslinechart.WithTimeRange(minTime, maxTime),
		timeserieslinechart.WithYRange(minY, maxY),
		timeserieslinechart.WithYLabelFormatter(compactPriceLabelFormatter()),
		timeserieslinechart.WithLineStyle(runes.ArcLineStyle),
		timeserieslinechart.WithStyle(lipgloss.NewStyle()),
		timeserieslinechart.WithAxesStyles(lipgloss.NewStyle(), lipgloss.NewStyle()),
	)
	for _, point := range points {
		chart.Push(point)
	}
	chart.DrawBraille()
	return colorizePlot(addRightAxis(chart.View(), width, series[len(series)-1], minY, maxY, height), series, width)
}

func chartYStep(height int) int {
	return max(1, height/4)
}

func RenderVolumeStrip(candles []domain.Candle, width int) string {
	if len(candles) == 0 || width < 8 {
		return "no volume"
	}
	series := downsampleVolumes(candles, width)
	if len(series) == 0 {
		return "no volume"
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	hi := 0.0
	for _, v := range series {
		if v > hi {
			hi = v
		}
	}
	if hi == 0 {
		return strings.Repeat(" ", chartPlotPad) + strings.Repeat(".", width)
	}
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", chartPlotPad))
	for _, v := range series {
		idx := int(v / hi * float64(len(levels)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(levels) {
			idx = len(levels) - 1
		}
		b.WriteRune(levels[idx])
	}
	return b.String()
}

func RenderTimeAxis(candles []domain.Candle, width int) string {
	if len(candles) == 0 || width < 8 {
		return ""
	}
	series := downsampleCandles(candles, width)
	if len(series) == 0 {
		return ""
	}
	line := make([]rune, width)
	for i := range line {
		line[i] = ' '
	}

	for _, tick := range buildTimeAxisTicks(series, width) {
		writeAxisLabel(line, tick.anchor, tick.label, tick.align)
	}
	return strings.Repeat(" ", chartPlotPad) + string(line)
}

type axisTick struct {
	anchor int
	label  string
	align  string
}

func buildTimeAxisTicks(series []domain.Candle, width int) []axisTick {
	if len(series) == 0 || width <= 0 {
		return nil
	}
	lastIdx := len(series) - 1
	span := series[lastIdx].Time.Sub(series[0].Time)
	sampleLabel := formatAxisTime(series[min(lastIdx, max(0, lastIdx/2))].Time, span)
	labelWidth := max(4, len([]rune(sampleLabel)))
	maxTicks := max(3, min(6, width/(labelWidth+2)))

	ticks := make([]axisTick, 0, maxTicks)
	usedUntil := -1
	for i := 0; i < maxTicks; i++ {
		idx := 0
		switch {
		case i == 0:
			idx = 0
		case i == maxTicks-1:
			idx = lastIdx
		default:
			idx = int(math.Round(float64(i) * float64(lastIdx) / float64(maxTicks-1)))
		}

		label := formatAxisTime(series[idx].Time, span)
		align := "center"
		anchor := int(math.Round(float64(i) * float64(width-1) / float64(maxTicks-1)))
		switch {
		case i == 0:
			align = "left"
			anchor = 0
		case i == maxTicks-1:
			align = "right"
			anchor = width - 1
		}

		start, end := axisLabelBounds(anchor, label, align, width)
		if start <= usedUntil {
			continue
		}
		usedUntil = end
		ticks = append(ticks, axisTick{
			anchor: anchor,
			label:  label,
			align:  align,
		})
	}
	return ticks
}

func writeAxisLabel(line []rune, anchor int, label, align string) {
	start, _ := axisLabelBounds(anchor, label, align, len(line))
	for i, r := range []rune(label) {
		pos := start + i
		if pos >= 0 && pos < len(line) {
			line[pos] = r
		}
	}
}

func axisLabelBounds(anchor int, label, align string, width int) (int, int) {
	labelRunes := []rune(label)
	start := anchor
	switch align {
	case "center":
		start -= len(labelRunes) / 2
	case "right":
		start -= len(labelRunes) - 1
	}
	if start < 0 {
		start = 0
	}
	if start+len(labelRunes) > width {
		start = max(0, width-len(labelRunes))
	}
	return start, start + len(labelRunes) - 1
}

func addRightAxis(plot string, totalWidth int, lastPrice, minY, maxY float64, height int) string {
	lines := strings.Split(plot, "\n")
	if len(lines) == 0 {
		return plot
	}
	lastPriceLine := chartPriceRow(lastPrice, minY, maxY, height, len(lines))

	out := make([]string, 0, len(lines))
	for i, line := range lines {
		label := mirroredAxisLabel(line)
		if i == lastPriceLine {
			label = priceAxisLabel(lastPrice)
		}
		suffix := strings.Repeat(" ", chartAxisGap) + label
		bodyWidth := max(1, totalWidth-len([]rune(suffix)))
		out = append(out, fitWidth(line, bodyWidth)+suffix)
	}
	return strings.Join(out, "\n")
}

func chartPriceRow(price, minY, maxY float64, height, lineCount int) int {
	if lineCount == 0 {
		return 0
	}
	if maxY <= minY || height <= 1 {
		return min(lineCount-1, max(0, lineCount/2))
	}
	position := (maxY - price) / (maxY - minY)
	row := int(math.Round(position * float64(height-1)))
	return min(lineCount-1, max(0, row))
}

func colorizePlot(plot string, series []float64, totalWidth int) string {
	if len(series) == 0 {
		return plot
	}

	bodyWidth := max(1, totalWidth-chartAxisGap-chartAxisWidth)
	baseline := series[0]
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394"))
	if series[len(series)-1] < baseline {
		lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73"))
	}
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89"))

	lines := strings.Split(plot, "\n")
	for i, line := range lines {
		runes := []rune(line)
		var b strings.Builder
		for x, r := range runes {
			if x >= chartPlotPad && x < bodyWidth && isPlotRune(r) {
				seriesIdx := min(len(series)-1, max(0, x-chartPlotPad))
				if series[seriesIdx] == baseline {
					b.WriteString(axisStyle.Render(string(r)))
					continue
				}
				b.WriteString(lineStyle.Render(string(r)))
				continue
			}
			if x < chartOffset || x >= bodyWidth+chartAxisGap {
				if !unicodeIsSpace(r) {
					b.WriteString(axisStyle.Render(string(r)))
					continue
				}
			}
			if x == chartOffset && (r == '│' || r == '┃') {
				b.WriteString(axisStyle.Render(string(r)))
				continue
			}
			b.WriteRune(r)
		}
		lines[i] = b.String()
	}
	return strings.Join(lines, "\n")
}

func isPlotRune(r rune) bool {
	switch r {
	case '╭', '╮', '╯', '╰', '─', '│':
		return true
	default:
		return r >= '\u2800' && r <= '\u28ff'
	}
}

func mirroredAxisLabel(line string) string {
	r := []rune(line)
	if len(r) == 0 {
		return strings.Repeat(" ", chartAxisWidth)
	}
	if len(r) < chartAxisWidth {
		return padRight(string(r), chartAxisWidth)
	}
	return string(r[:chartAxisWidth])
}

func downsampleCloses(candles []domain.Candle, width int) []float64 {
	if width <= 0 || len(candles) == 0 {
		return nil
	}
	if len(candles) == 1 {
		out := make([]float64, width)
		for i := range out {
			out[i] = candles[0].Close
		}
		return out
	}
	if len(candles) < width {
		reduced := downsampleCandles(candles, width)
		out := make([]float64, 0, len(reduced))
		for _, candle := range reduced {
			out = append(out, candle.Close)
		}
		return out
	}

	out := make([]float64, 0, width)
	step := float64(len(candles)) / float64(width)
	for i := range width {
		start := int(float64(i) * step)
		end := int(float64(i+1) * step)
		if end <= start {
			end = start + 1
		}
		if end > len(candles) {
			end = len(candles)
		}
		out = append(out, candles[end-1].Close)
	}
	return out
}

func chartYRange(series []float64) (float64, float64) {
	if len(series) == 0 {
		return 0, 1
	}
	minY := series[0]
	maxY := series[0]
	for _, v := range series[1:] {
		if v < minY {
			minY = v
		}
		if v > maxY {
			maxY = v
		}
	}
	span := maxY - minY
	if span == 0 {
		pad := math.Abs(minY) * 0.01
		if pad < 1 {
			pad = 1
		}
		return minY - pad, maxY + pad
	}
	pad := span * 0.08
	return minY - pad, maxY + pad
}

func compactPriceLabelFormatter() linechart.LabelFormatter {
	return func(_ int, v float64) string {
		return priceAxisLabel(v)
	}
}

func priceAxisLabel(v float64) string {
	label := fmt.Sprintf("%.2f", v)
	switch {
	case math.Abs(v) >= 10000:
		label = fmt.Sprintf("%.0f", v)
	case math.Abs(v) >= 1000:
		label = fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%8s", fitWidth(label, chartAxisWidth))
}

func unicodeIsSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func downsampleVolumes(candles []domain.Candle, width int) []float64 {
	reduced := downsampleCandles(candles, width)
	out := make([]float64, 0, len(reduced))
	for _, candle := range reduced {
		out = append(out, float64(candle.Volume))
	}
	return out
}

func downsampleCandles(candles []domain.Candle, width int) []domain.Candle {
	if width <= 0 || len(candles) == 0 {
		return nil
	}
	if len(candles) == 1 {
		out := make([]domain.Candle, width)
		for i := range out {
			out[i] = candles[0]
		}
		return out
	}
	if len(candles) < width {
		out := make([]domain.Candle, 0, width)
		last := float64(len(candles) - 1)
		for i := 0; i < width; i++ {
			pos := float64(i) * last / float64(width-1)
			loIdx := int(pos)
			hiIdx := loIdx + 1
			if hiIdx >= len(candles) {
				hiIdx = len(candles) - 1
			}
			ratio := pos - float64(loIdx)
			lo := candles[loIdx]
			hi := candles[hiIdx]
			out = append(out, domain.Candle{
				Time:   lo.Time,
				Close:  lo.Close + (hi.Close-lo.Close)*ratio,
				Volume: int64(float64(lo.Volume) + float64(hi.Volume-lo.Volume)*ratio),
			})
		}
		return out
	}

	out := make([]domain.Candle, 0, width)
	step := float64(len(candles)) / float64(width)
	for i := range width {
		start := int(float64(i) * step)
		end := int(float64(i+1) * step)
		if end <= start {
			end = start + 1
		}
		if end > len(candles) {
			end = len(candles)
		}
		sumClose := 0.0
		sumVolume := int64(0)
		count := 0.0
		for _, candle := range candles[start:end] {
			sumClose += candle.Close
			sumVolume += candle.Volume
			count++
		}
		mid := start + (end-start)/2
		if mid >= len(candles) {
			mid = len(candles) - 1
		}
		out = append(out, domain.Candle{
			Time:   candles[mid].Time,
			Close:  sumClose / count,
			Volume: sumVolume,
		})
	}
	return out
}

func formatAxisTime(ts time.Time, span time.Duration) string {
	switch {
	case span >= 365*24*time.Hour:
		return ts.Format("Jan 06")
	case span >= 72*time.Hour:
		return ts.Format("Jan 2")
	default:
		return ts.Format("15:04")
	}
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return s + strings.Repeat(" ", width-len(r))
}

func fitWidth(s string, width int) string {
	r := []rune(s)
	if len(r) > width {
		return string(r[:width])
	}
	if len(r) < width {
		return string(r) + strings.Repeat(" ", width-len(r))
	}
	return string(r)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
