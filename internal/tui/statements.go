package tui

import (
	"math"
	"time"

	"blackdesk/internal/domain"
)

const quarterlyStatementComparisonTolerance = 45 * 24 * time.Hour

func statementGrowthPercent(stmt domain.FinancialStatement, values []domain.StatementValue, idx int) (float64, bool) {
	if idx < 0 || idx >= len(values) || idx >= len(stmt.Periods) || !values[idx].Present {
		return 0, false
	}

	compareIdx := statementComparisonIndex(stmt.Periods, stmt.Frequency, idx)
	if compareIdx < 0 || compareIdx >= len(values) || !values[compareIdx].Present {
		return 0, false
	}

	previous := values[compareIdx].Value
	if previous == 0 {
		return 0, false
	}

	changePct := ((values[idx].Value - previous) / math.Abs(previous)) * 100
	return changePct, true
}

func statementComparisonIndex(periods []domain.StatementPeriod, frequency domain.StatementFrequency, idx int) int {
	if idx < 0 || idx >= len(periods) {
		return -1
	}

	if frequency != domain.StatementFrequencyQuarterly {
		if idx+1 < len(periods) {
			return idx + 1
		}
		return -1
	}

	current := periods[idx].EndDate
	if !current.IsZero() {
		target := current.AddDate(-1, 0, 0)
		bestIdx := -1
		bestDelta := time.Duration(1<<63 - 1)
		for candidate := idx + 1; candidate < len(periods); candidate++ {
			end := periods[candidate].EndDate
			if end.IsZero() {
				continue
			}
			delta := end.Sub(target)
			if delta < 0 {
				delta = -delta
			}
			if delta <= quarterlyStatementComparisonTolerance && delta < bestDelta {
				bestIdx = candidate
				bestDelta = delta
			}
		}
		if bestIdx >= 0 {
			return bestIdx
		}
	}

	if idx+4 < len(periods) {
		return idx + 4
	}
	return -1
}
