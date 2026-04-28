package tui

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
)

type statisticsHorizon struct {
	Label   string
	Forward int
}

type statisticsRangeSpec struct {
	Label  string
	Ranges []string
}

type statisticsSignal struct {
	Label string
	Match func(statisticsPoint) bool
}

type statisticsPoint struct {
	Index     int
	Sharpe252 float64
	Sharpe63  float64
}

type statisticsRow struct {
	Horizon     string
	Signal      string
	Samples     int
	Mean        float64
	Median      float64
	AvgDrawdown float64
	AvgLoss     float64
	FwdRetVol   float64
	FwdRetVolOK bool
	Positive    int
	OK          bool
}

const statisticsMinForwardRetVolSamples = 10

var statisticsHorizons = []statisticsHorizon{
	{Label: "1M", Forward: 21},
	{Label: "3M", Forward: 63},
	{Label: "6M", Forward: 126},
	{Label: "12M", Forward: 252},
}

var statisticsRangeSpecs = []statisticsRangeSpec{
	{Label: "5Y", Ranges: statisticsHistoryRanges},
	{Label: "10Y", Ranges: statistics10YHistoryRanges},
	{Label: "Max", Ranges: statisticsMaxHistoryRanges},
}

var statisticsSignals = []statisticsSignal{
	{Label: "All periods", Match: func(statisticsPoint) bool { return true }},
	{Label: "12M < 0", Match: func(p statisticsPoint) bool { return p.Sharpe252 < 0 }},
	{Label: "12M > 0", Match: func(p statisticsPoint) bool { return p.Sharpe252 > 0 }},
	{Label: "12M > 1", Match: func(p statisticsPoint) bool { return p.Sharpe252 > 1 }},
	{Label: "3M < 0", Match: func(p statisticsPoint) bool { return p.Sharpe63 < 0 }},
	{Label: "3M > 0", Match: func(p statisticsPoint) bool { return p.Sharpe63 > 0 }},
	{Label: "3M > 1", Match: func(p statisticsPoint) bool { return p.Sharpe63 > 1 }},
}

func (m Model) statisticsRangeSpec() statisticsRangeSpec {
	if m.statisticsRangeIdx < 0 || m.statisticsRangeIdx >= len(statisticsRangeSpecs) {
		return statisticsRangeSpecs[0]
	}
	return statisticsRangeSpecs[m.statisticsRangeIdx]
}

func renderQuoteStatisticsBoard(section, label, muted lipgloss.Style, series domain.PriceSeries, width, height, rangeIdx int) string {
	var b strings.Builder
	b.WriteString(section.Render("STATISTICS") + "\n\n")
	if len(series.Candles) <= 252 {
		b.WriteString(muted.Render("Statistics need at least one year of daily history."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	points := buildStatisticsPoints(series)
	activeSignals := activeStatisticsSignals(points)
	b.WriteString(renderStatusLine(width, section.Render("FORWARD RETURNS (vs ROC/HV)"), renderStatisticsRangeTabs(rangeIdx)) + "\n\n")
	b.WriteString(renderStatisticsRows(buildStatisticsTableRows(series), label, muted, width, activeSignals))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderStatisticsRangeTabs(activeIdx int) string {
	active := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#14110D")).
		Background(lipgloss.Color("#E7B66B")).
		Bold(true).
		Padding(0, 1)
	inactive := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D8C9B8")).
		Background(lipgloss.Color("#2A2520")).
		Padding(0, 1)
	parts := make([]string, 0, len(statisticsRangeSpecs)+1)
	for idx, spec := range statisticsRangeSpecs {
		if idx == activeIdx {
			parts = append(parts, active.Render(spec.Label))
			continue
		}
		parts = append(parts, inactive.Render(spec.Label))
	}
	parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("#9F907E")).Render("←/→"))
	return strings.Join(parts, " ")
}

func renderQuoteStatisticsPreview(section, label, muted, pos, neg lipgloss.Style, series domain.PriceSeries, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("STATISTICS") + "\n\n")
	if len(series.Candles) <= 252 {
		b.WriteString(muted.Render("Historical stats need more daily history."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	points := buildStatisticsPoints(series)
	if len(points) == 0 {
		b.WriteString(muted.Render("No signal samples available."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	latest := points[len(points)-1]

	rows := buildStatisticsRows(series, statisticsHorizon{Label: "3M", Forward: 63})
	if len(rows) > 0 {
		b.WriteString(muted.Render("3M Baseline") + "\n\n")
		row := rows[0]
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Avg Fwd"), renderSharpeReturn(pos, neg, muted, row.Mean)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Med Fwd"), renderSharpeReturn(pos, neg, muted, row.Median)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Win%"), renderSharpePercent(pos, muted, row.Positive)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Avg MaxDD"), renderSharpeReturn(pos, neg, muted, row.AvgDrawdown)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Avg Loss"), renderSharpeReturn(pos, neg, muted, row.AvgLoss)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Fwd R/Vol"), renderStatisticsRatioValue(pos, neg, muted, row.FwdRetVol, row.FwdRetVolOK)))
		if regimes := statisticsCurrentRegimeStats(series, latest, statisticsHorizon{Label: "3M", Forward: 63}); len(regimes) > 0 {
			b.WriteString("\n" + muted.Render("Current Regime") + "\n\n")
			for idx, regime := range regimes {
				if idx > 0 {
					b.WriteString("\n")
				}
				b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Avg Fwd "+regime.Label), renderSharpeReturn(pos, neg, muted, regime.Avg)))
				b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Win% "+regime.Label), renderSharpePercent(pos, muted, regime.Win)))
				b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Avg MaxDD "+regime.Label), renderSharpeReturn(pos, neg, muted, regime.AvgDrawdown)))
				b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Fwd R/Vol "+regime.Label), renderStatisticsRatioValue(pos, neg, muted, regime.FwdRetVol, regime.FwdRetVolOK)))
			}
		}
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderStatisticsRows(rows []statisticsRow, label, muted lipgloss.Style, width int, activeSignals map[string]struct{}) string {
	if len(rows) == 0 {
		return muted.Render("No matching samples")
	}
	columns := statisticsColumns(width, label)
	headerValues := []string{"Date", "Signal", "N", "Avg Fwd", "Med Fwd", "Win%", "Avg MaxDD", "Avg Loss", "Fwd R/Vol"}
	var b strings.Builder
	b.WriteString(renderStatisticsTableLine(columns, headerValues, muted, true, false) + "\n")
	prevHorizon := ""
	for _, row := range rows {
		if prevHorizon != "" && row.Horizon != prevHorizon {
			b.WriteString("\n")
		}
		prevHorizon = row.Horizon
		values := []string{
			row.Horizon,
			row.Signal,
			fmt.Sprintf("%d", row.Samples),
			formatSignedPercentRatio(row.Mean),
			formatSignedPercentRatio(row.Median),
			fmt.Sprintf("%d%%", row.Positive),
			formatSignedPercentRatio(row.AvgDrawdown),
			formatSignedPercentRatio(row.AvgLoss),
			formatStatisticsRatio(row.FwdRetVol, row.FwdRetVolOK),
		}
		_, active := activeSignals[row.Signal]
		b.WriteString(renderStatisticsTableLine(columns, values, lipgloss.NewStyle(), false, active) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func statisticsColumns(width int, label lipgloss.Style) []insiderTableColumn {
	horizonWidth := 4
	nWidth := 4
	metricWidth := 8
	posWidth := 5
	drawdownWidth := len("Avg MaxDD")
	lossWidth := len("Avg Loss")
	retVolWidth := len("Fwd R/Vol")
	gaps := 2 * 8
	fixed := horizonWidth + nWidth + metricWidth*2 + posWidth + drawdownWidth + lossWidth + retVolWidth + gaps
	signalWidth := max(10, width-fixed)
	return []insiderTableColumn{
		{title: "Date", width: horizonWidth, align: lipgloss.Left, style: label},
		{title: "Signal", width: signalWidth, align: lipgloss.Left, style: label},
		{title: "N", width: nWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Avg Fwd", width: metricWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Med Fwd", width: metricWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Pos", width: posWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Avg MaxDD", width: drawdownWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Avg Loss", width: lossWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Fwd R/Vol", width: retVolWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
	}
}

func renderStatisticsTableLine(columns []insiderTableColumn, values []string, fallback lipgloss.Style, header bool, active bool) string {
	parts := make([]string, 0, len(columns))
	activeDateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Bold(true)
	activeSignalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Bold(true)
	for i, col := range columns {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		style := col.style
		if header {
			style = fallback
		}
		if !header && i >= 3 && value != "-" {
			style = marketMoveStyle(statisticsMoveValue(i, value))
		}
		if !header && active {
			switch i {
			case 0:
				style = activeDateStyle
			case 1:
				style = activeSignalStyle
			}
		}
		parts = append(parts, style.Width(col.width).Align(col.align).Render(ansi.Truncate(value, col.width, "")))
	}
	return strings.Join(parts, "  ")
}

func statisticsMoveValue(column int, value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(value, "+"), "%"), 64)
	if err != nil {
		return 0
	}
	if column == 5 {
		if parsed > 50 {
			return 1
		}
		return 0
	}
	if parsed > 0 {
		return 1
	}
	if parsed < 0 {
		return -1
	}
	return 0
}

func formatStatisticsRatio(value float64, ok bool) string {
	if !ok {
		return "-"
	}
	return formatMetricSigned(value)
}

func renderStatisticsRatioValue(pos, neg, muted lipgloss.Style, value float64, ok bool) string {
	if !ok {
		return muted.Render("-")
	}
	return renderSharpeRatio(pos, neg, muted, value)
}

func buildStatisticsTableRows(series domain.PriceSeries) []statisticsRow {
	out := make([]statisticsRow, 0, len(statisticsHorizons)*len(statisticsSignals))
	for _, horizon := range statisticsHorizons {
		rows := buildStatisticsRows(series, horizon)
		out = append(out, rows...)
	}
	return out
}

func buildStatisticsRows(series domain.PriceSeries, horizon statisticsHorizon) []statisticsRow {
	points := buildStatisticsPoints(series)
	out := make([]statisticsRow, 0, len(statisticsSignals))
	for _, signal := range statisticsSignals {
		row := buildStatisticsRow(series, points, signal, horizon)
		if row.OK {
			out = append(out, row)
		}
	}
	return out
}

func buildStatisticsPoints(series domain.PriceSeries) []statisticsPoint {
	closes := extractCloses(series.Candles)
	if len(closes) <= 252 {
		return nil
	}
	points := make([]statisticsPoint, 0, len(closes)-252)
	for i := 252; i < len(closes); i++ {
		window := closes[:i+1]
		points = append(points, statisticsPoint{
			Index:     i,
			Sharpe252: returnOverVol(annualizedLookbackReturn(window, 252), historicalVolatility(window, 252), 1),
			Sharpe63:  returnOverVol(annualizedLookbackReturn(window, 63), historicalVolatility(window, 63), 1),
		})
	}
	return points
}

func buildStatisticsRow(series domain.PriceSeries, points []statisticsPoint, signal statisticsSignal, horizon statisticsHorizon) statisticsRow {
	closes := extractCloses(series.Candles)
	values := make([]float64, 0, len(points))
	drawdowns := make([]float64, 0, len(points))
	positive := 0
	lossSum := 0.0
	lossCount := 0
	for _, point := range points {
		if point.Index+horizon.Forward >= len(closes) || !signal.Match(point) {
			continue
		}
		base := closes[point.Index]
		future := closes[point.Index+horizon.Forward]
		if base <= 0 || future <= 0 {
			continue
		}
		value := future/base - 1
		values = append(values, value)
		drawdowns = append(drawdowns, statisticsForwardDrawdown(closes, point.Index, horizon.Forward))
		if value > 0 {
			positive++
		} else if value < 0 {
			lossSum += value
			lossCount++
		}
	}
	if len(values) == 0 {
		return statisticsRow{Horizon: horizon.Label, Signal: signal.Label}
	}
	sort.Float64s(values)
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	drawdownSum := 0.0
	for _, drawdown := range drawdowns {
		drawdownSum += drawdown
	}
	median := values[len(values)/2]
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	}
	mean := sum / float64(len(values))
	avgDrawdown := drawdownSum / float64(len(drawdowns))
	avgLoss := 0.0
	if lossCount > 0 {
		avgLoss = lossSum / float64(lossCount)
	}
	fwdRetVol, fwdRetVolOK := forwardReturnVolRatio(values, horizon.Forward)
	return statisticsRow{
		Horizon:     horizon.Label,
		Signal:      signal.Label,
		Samples:     len(values),
		Mean:        mean,
		Median:      median,
		AvgDrawdown: avgDrawdown,
		AvgLoss:     avgLoss,
		FwdRetVol:   fwdRetVol,
		FwdRetVolOK: fwdRetVolOK,
		Positive:    int(float64(positive) / float64(len(values)) * 100),
		OK:          true,
	}
}

func annualizedForwardReturn(totalReturn float64, forward int) float64 {
	if forward <= 0 {
		return 0
	}
	gross := 1 + totalReturn
	if gross <= 0 {
		return -1
	}
	return math.Pow(gross, 252.0/float64(forward)) - 1
}

func annualizedForwardReturnVol(forwardReturns []float64, forward int) float64 {
	if forward <= 0 || len(forwardReturns) < 2 {
		return 0
	}
	mean := 0.0
	for _, value := range forwardReturns {
		mean += value
	}
	mean /= float64(len(forwardReturns))
	variance := 0.0
	for _, value := range forwardReturns {
		diff := value - mean
		variance += diff * diff
	}
	variance /= float64(len(forwardReturns) - 1)
	return math.Sqrt(variance) * math.Sqrt(252.0/float64(forward))
}

func forwardReturnVolRatio(forwardReturns []float64, forward int) (float64, bool) {
	if len(forwardReturns) < statisticsMinForwardRetVolSamples {
		return 0, false
	}
	mean := 0.0
	for _, value := range forwardReturns {
		mean += value
	}
	mean /= float64(len(forwardReturns))
	vol := annualizedForwardReturnVol(forwardReturns, forward)
	if vol == 0 {
		return 0, false
	}
	return annualizedForwardReturn(mean, forward) / vol, true
}

func statisticsForwardDrawdown(closes []float64, start, forward int) float64 {
	base := closes[start]
	minReturn := 0.0
	for i := start + 1; i <= start+forward; i++ {
		if closes[i] <= 0 {
			continue
		}
		move := closes[i]/base - 1
		if move < minReturn {
			minReturn = move
		}
	}
	return minReturn
}

func statisticsPercentile(points []statisticsPoint, latest float64, valueFn func(statisticsPoint) float64) int {
	if len(points) == 0 {
		return 0
	}
	count := 0
	for _, point := range points {
		if valueFn(point) <= latest {
			count++
		}
	}
	return int(float64(count) / float64(len(points)) * 100)
}

func formatPercentile(value int) string {
	if value <= 0 {
		return "0th"
	}
	if value >= 100 {
		return "100th"
	}
	lastTwo := value % 100
	suffix := "th"
	if lastTwo < 11 || lastTwo > 13 {
		switch value % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", value, suffix)
}

type statisticsRegimeStats struct {
	Label       string
	Avg         float64
	AvgDrawdown float64
	FwdRetVol   float64
	FwdRetVolOK bool
	Win         int
}

func statisticsCurrentRegimeStats(series domain.PriceSeries, latest statisticsPoint, horizon statisticsHorizon) []statisticsRegimeStats {
	rows := buildStatisticsRows(series, horizon)
	if len(rows) == 0 {
		return nil
	}
	bySignal := make(map[string]statisticsRow, len(rows))
	for _, row := range rows {
		bySignal[row.Signal] = row
	}
	labels := make([]string, 0, 2)
	if label := statisticsSignalLabelForValue(latest.Sharpe252, "12M"); label != "" {
		labels = append(labels, label)
	}
	if label := statisticsSignalLabelForValue(latest.Sharpe63, "3M"); label != "" {
		labels = append(labels, label)
	}
	out := make([]statisticsRegimeStats, 0, len(labels))
	for _, label := range labels {
		row, ok := bySignal[label]
		if !ok {
			continue
		}
		out = append(out, statisticsRegimeStats{
			Label:       label,
			Avg:         row.Mean,
			AvgDrawdown: row.AvgDrawdown,
			FwdRetVol:   row.FwdRetVol,
			FwdRetVolOK: row.FwdRetVolOK,
			Win:         row.Positive,
		})
	}
	return out
}

func statisticsSignalLabelForValue(value float64, prefix string) string {
	if value > 1 {
		return prefix + " > 1"
	}
	if value > 0 {
		return prefix + " > 0"
	}
	if value < 0 {
		return prefix + " < 0"
	}
	return ""
}

func activeStatisticsSignals(points []statisticsPoint) map[string]struct{} {
	if len(points) == 0 {
		return nil
	}
	latest := points[len(points)-1]
	active := make(map[string]struct{}, 2)
	if label := statisticsSignalLabelForValue(latest.Sharpe252, "12M"); label != "" {
		active[label] = struct{}{}
	}
	if label := statisticsSignalLabelForValue(latest.Sharpe63, "3M"); label != "" {
		active[label] = struct{}{}
	}
	return active
}
