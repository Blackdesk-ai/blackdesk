package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textinput"

	"blackdesk/internal/application"
	"blackdesk/internal/domain"
	"blackdesk/internal/storage"
)

type statementRequest struct {
	kind      domain.StatementKind
	frequency domain.StatementFrequency
}

var aiStatementRequests = []statementRequest{
	{kind: domain.StatementKindIncome, frequency: domain.StatementFrequencyAnnual},
	{kind: domain.StatementKindBalanceSheet, frequency: domain.StatementFrequencyAnnual},
	{kind: domain.StatementKindCashFlow, frequency: domain.StatementFrequencyAnnual},
	{kind: domain.StatementKindIncome, frequency: domain.StatementFrequencyQuarterly},
	{kind: domain.StatementKindBalanceSheet, frequency: domain.StatementFrequencyQuarterly},
	{kind: domain.StatementKindCashFlow, frequency: domain.StatementFrequencyQuarterly},
}

type Model struct {
	ctx           context.Context
	services      *application.Services
	config        storage.Config
	workspaceRoot string

	width           int
	height          int
	selectedIdx     int
	watchlistScroll int
	rangeIdx        int
	tabIdx          int
	quoteCenterMode quoteCenterMode
	status          string
	lastUpdated     time.Time
	clock           time.Time
	lastAutoRefresh time.Time
	lastMarketNews  time.Time
	appVersion      string
	latestVersion   string
	updateAvailable bool

	searchInput                 textinput.Model
	searchMode                  bool
	searchItems                 []domain.SymbolRef
	searchIdx                   int
	aiInput                     textinput.Model
	aiFocused                   bool
	aiRunning                   bool
	aiOutput                    string
	aiErr                       error
	aiDuration                  time.Duration
	aiMessages                  []aiMessage
	aiScroll                    int
	aiLastContext               string
	aiLastSymbol                string
	aiFullscreen                bool
	aiPickerOpen                bool
	aiPickerStep                aiPickerStep
	aiModels                    map[string][]string
	aiModelIdx                  int
	aiModelErr                  error
	aiModelBusy                 bool
	aiMarketOpinion             string
	aiMarketOpinionErr          error
	aiMarketOpinionRunning      bool
	aiMarketOpinionUpdated      time.Time
	aiQuoteInsight              string
	aiQuoteInsightErr           error
	aiQuoteInsightRunning       bool
	aiQuoteInsightUpdated       time.Time
	aiQuoteInsightSymbol        string
	marketOpinionHistory        map[string]domain.PriceSeries
	marketOpinionHistoryAt      map[string]time.Time
	pendingMarketOpinionRefresh bool
	helpOpen                    bool

	quote             domain.QuoteSnapshot
	watchQuotes       map[string]domain.QuoteSnapshot
	series            domain.PriceSeries
	technicalCache    map[string]domain.PriceSeries
	statementCache    map[string]domain.FinancialStatement
	insiderCache      map[string]domain.InsiderSnapshot
	news              []domain.NewsItem
	newsSelected      int
	marketNews        []domain.NewsItem
	marketNewsSources []domain.MarketNewsSource
	marketNewsFresh   map[string]struct{}
	marketNewsSeen    map[string]struct{}
	marketNewsSel     int
	marketNewsScroll  int
	marketNewsUpdated time.Time
	screenerDefs      []domain.ScreenerDefinition
	screenerIdx       int
	screenerResult    domain.ScreenerResult
	screenerSel       int
	screenerScroll    int
	screenerLoaded    bool
	profileScroll     int
	fundamentals      domain.FundamentalsSnapshot
	statement         domain.FinancialStatement
	insiders          domain.InsiderSnapshot
	statementKind     domain.StatementKind
	statementFreq     domain.StatementFrequency

	errQuote            error
	errHistory          error
	errTechnicalHistory error
	errNews             error
	errMarketNews       error
	errFundamentals     error
	errStatement        error
	errInsiders         error
	errScreener         error
}
