package application

import (
	"context"

	"blackdesk/internal/domain"
)

type MarketRiskProvider interface {
	GetMarketRisk(context.Context) (domain.MarketRiskSnapshot, error)
}
