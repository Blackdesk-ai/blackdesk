package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m Model) needsTechnicalHistory(symbol string) bool {
	series, ok := m.technicalCache[strings.ToUpper(symbol)]
	if !ok {
		return true
	}
	return len(series.Candles) < 252
}

func (m Model) technicalSeries(symbol string) domain.PriceSeries {
	return m.technicalCache[strings.ToUpper(symbol)]
}
