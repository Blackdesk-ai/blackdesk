package tui

import (
	"fmt"
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
	ReturnDD    float64
	Positive    int
	OK          bool
}

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
	b.WriteString(renderStatusLine(width, section.Render("FORWARD RETURNS (vs ROC/HV)"), renderStatisticsRangeTabs(rangeIdx)) + "\n\n")
	b.WriteString(renderStatisticsRows(buildStatisticsTableRows(series), label, muted, width))
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
	b.WriteString(muted.Render("Current Signal") + "\n")
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("252d"), renderSharpeValue(pos, neg, muted, latest.Sharpe252)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("63d"), renderSharpeValue(pos, neg, muted, latest.Sharpe63)))
	b.WriteString("\n" + muted.Render("Signal Percentile") + "\n")
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("252d"), muted.Render(formatPercentile(statisticsPercentile(points, latest.Sharpe252, func(p statisticsPoint) float64 { return p.Sharpe252 })))))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("63d"), muted.Render(formatPercentile(statisticsPercentile(points, latest.Sharpe63, func(p statisticsPoint) float64 { return p.Sharpe63 })))))

	rows := buildStatisticsRows(series, statisticsHorizon{Label: "3M", Forward: 63})
	if len(rows) > 0 {
		b.WriteString("\n" + muted.Render("3M Baseline") + "\n")
		row := rows[0]
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Avg"), renderSharpeReturn(pos, neg, muted, row.Mean)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Median"), renderSharpeReturn(pos, neg, muted, row.Median)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Win%"), renderSharpePercent(pos, muted, row.Positive)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Avg. DD"), renderSharpeReturn(pos, neg, muted, row.AvgDrawdown)))
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Return/DD"), renderSharpeRatio(pos, neg, muted, row.ReturnDD)))
		if regimes := statisticsCurrentSignalEVs(series, latest, statisticsHorizon{Label: "3M", Forward: 63}); len(regimes) > 0 {
			b.WriteString("\n" + muted.Render("Current Regime EV") + "\n")
			for _, regime := range regimes {
				b.WriteString(fmt.Sprintf("%s %s\n", label.Render("EV "+regime.Label), renderSharpeReturn(pos, neg, muted, regime.EV)))
			}
		}
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderStatisticsRows(rows []statisticsRow, label, muted lipgloss.Style, width int) string {
	if len(rows) == 0 {
		return muted.Render("No matching samples")
	}
	columns := statisticsColumns(width, label)
	headerValues := []string{"Date", "Signal", "N", "Avg", "Median", "Win%", "Avg. DD", "Return/DD"}
	var b strings.Builder
	b.WriteString(renderStatisticsTableLine(columns, headerValues, muted, true) + "\n")
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
			formatMetricSigned(row.ReturnDD),
		}
		b.WriteString(renderStatisticsTableLine(columns, values, lipgloss.NewStyle(), false) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func statisticsColumns(width int, label lipgloss.Style) []insiderTableColumn {
	horizonWidth := 4
	nWidth := 4
	metricWidth := 8
	posWidth := 5
	drawdownWidth := len("Avg. DD")
	ratioWidth := len("Return/DD")
	gaps := 2 * 7
	fixed := horizonWidth + nWidth + metricWidth*2 + posWidth + drawdownWidth + ratioWidth + gaps
	signalWidth := max(10, width-fixed)
	return []insiderTableColumn{
		{title: "Date", width: horizonWidth, align: lipgloss.Left, style: label},
		{title: "Signal", width: signalWidth, align: lipgloss.Left, style: label},
		{title: "N", width: nWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Avg", width: metricWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Median", width: metricWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Pos", width: posWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Avg. DD", width: drawdownWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
		{title: "Return/DD", width: ratioWidth, align: lipgloss.Right, style: lipgloss.NewStyle()},
	}
}

func renderStatisticsTableLine(columns []insiderTableColumn, values []string, fallback lipgloss.Style, header bool) string {
	parts := make([]string, 0, len(columns))
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
	returnDD := 0.0
	if drawdownSum != 0 {
		avgDrawdown := drawdownSum / float64(len(drawdowns))
		if avgDrawdown != 0 {
			returnDD = (sum / float64(len(values))) / -avgDrawdown
		}
	}
	return statisticsRow{
		Horizon:     horizon.Label,
		Signal:      signal.Label,
		Samples:     len(values),
		Mean:        sum / float64(len(values)),
		Median:      median,
		AvgDrawdown: drawdownSum / float64(len(drawdowns)),
		ReturnDD:    returnDD,
		Positive:    int(float64(positive) / float64(len(values)) * 100),
		OK:          true,
	}
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

type statisticsRegimeEV struct {
	Label string
	EV    float64
}

func statisticsCurrentSignalEVs(series domain.PriceSeries, latest statisticsPoint, horizon statisticsHorizon) []statisticsRegimeEV {
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
	out := make([]statisticsRegimeEV, 0, len(labels))
	for _, label := range labels {
		row, ok := bySignal[label]
		if !ok {
			continue
		}
		out = append(out, statisticsRegimeEV{Label: label, EV: row.Mean})
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
