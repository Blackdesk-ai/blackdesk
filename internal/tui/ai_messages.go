package tui

import (
	"time"

	"blackdesk/internal/domain"
)

type tickMsg time.Time

type aiResponseLoadedMsg struct {
	connectorID     string
	output          string
	duration        time.Duration
	contextSent     string
	contextRevision int
	truncation      aiRequestTruncation
	symbol          string
	err             error
}

type aiMarketOpinionLoadedMsg struct {
	connectorID string
	output      string
	duration    time.Duration
	histories   map[string]domain.PriceSeries
	err         error
}

type aiQuoteInsightPreparedMsg struct {
	symbol          string
	marketRisk      domain.MarketRiskSnapshot
	marketRiskErr   error
	quote           *domain.QuoteSnapshot
	quoteErr        error
	history         *domain.PriceSeries
	historyErr      error
	technical       *domain.PriceSeries
	technicalErr    error
	statementBundle []domain.FinancialStatement
	statement       *domain.FinancialStatement
	statementLoaded bool
	statementErr    error
	insiders        *domain.InsiderSnapshot
	insidersLoaded  bool
	insidersErr     error
	news            []domain.NewsItem
	newsLoaded      bool
	newsErr         error
	fundamentals    *domain.FundamentalsSnapshot
	fundErr         error
}

type aiQuoteInsightLoadedMsg struct {
	connectorID string
	output      string
	duration    time.Duration
	contextSent string
	symbol      string
	err         error
}

type aiFilingAnalysisPreparedMsg struct {
	prompt          string
	symbol          string
	filing          domain.FilingDocument
	filingErr       error
	marketRisk      domain.MarketRiskSnapshot
	marketRiskErr   error
	quote           *domain.QuoteSnapshot
	quoteErr        error
	history         *domain.PriceSeries
	historyErr      error
	technical       *domain.PriceSeries
	technicalErr    error
	statementBundle []domain.FinancialStatement
	statement       *domain.FinancialStatement
	statementLoaded bool
	statementErr    error
	insiders        *domain.InsiderSnapshot
	insidersLoaded  bool
	insidersErr     error
	news            []domain.NewsItem
	newsLoaded      bool
	newsErr         error
	fundamentals    *domain.FundamentalsSnapshot
	fundErr         error
}

type aiFilingChunkLoadedMsg struct {
	connectorID string
	output      string
	duration    time.Duration
	truncation  aiRequestTruncation
	symbol      string
	err         error
}

type aiFilingSynthesisLoadedMsg struct {
	connectorID string
	output      string
	duration    time.Duration
	truncation  aiRequestTruncation
	symbol      string
	err         error
}

type aiModelsLoadedMsg struct {
	connectorID string
	models      []string
	err         error
}

type aiContextPreparedMsg struct {
	prompt          string
	symbol          string
	marketRisk      domain.MarketRiskSnapshot
	marketRiskErr   error
	quote           *domain.QuoteSnapshot
	quoteErr        error
	quotes          []domain.QuoteSnapshot
	quotesErr       error
	history         *domain.PriceSeries
	historyErr      error
	technical       *domain.PriceSeries
	technicalErr    error
	statementBundle []domain.FinancialStatement
	statement       *domain.FinancialStatement
	statementLoaded bool
	statementErr    error
	insiders        *domain.InsiderSnapshot
	insidersLoaded  bool
	insidersErr     error
	news            []domain.NewsItem
	newsLoaded      bool
	newsErr         error
	fundamentals    *domain.FundamentalsSnapshot
	fundErr         error
}

type aiPickerStep int

const (
	aiPickerStepConnector aiPickerStep = iota
	aiPickerStepModel
)
