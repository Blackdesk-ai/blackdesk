package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m Model) needsSharpeHistory(symbol string) bool {
	series, ok := m.sharpeCache[strings.ToUpper(symbol)]
	if !ok {
		return true
	}
	if len(series.Candles) < 252 {
		return true
	}
	for _, rangeKey := range sharpeHistoryRanges {
		if strings.EqualFold(series.Range, rangeKey) {
			return false
		}
	}
	return true
}

func (m Model) sharpeSeries(symbol string) domain.PriceSeries {
	return m.sharpeCache[strings.ToUpper(symbol)]
}
