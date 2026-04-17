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

type aiFilingRunState struct {
	symbol        string
	snapshot      domain.FilingsSnapshot
	filing        domain.FilingDocument
	prompt        string
	chunks        []filingTextChunk
	analyses      []filingChunkAnalysisSummary
	nextChunkIdx  int
	synthesizing  bool
	totalDuration time.Duration
	truncation    aiRequestTruncation
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
	ctx                context.Context
	services           *application.Services
	marketRiskProvider application.MarketRiskProvider
	config             storage.Config
	workspaceRoot      string

	width           int
	height          int
	selectedIdx     int
	watchlistScroll int
	rangeIdx        int
	sharpeRangeIdx  int
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
	upgradeRunning  bool
	restartOnQuit   bool

	searchInput                 textinput.Model
	searchMode                  bool
	searchItems                 []domain.SymbolRef
	searchIdx                   int
	searchDebounceID            int
	searchRequestID             int
	searchRequestQuery          string
	commandInput                textinput.Model
	commandPaletteOpen          bool
	commandPaletteItems         []commandPaletteItem
	commandPaletteIdx           int
	commandPaletteSymbolItems   []domain.SymbolRef
	commandPaletteDebounceID    int
	commandPaletteRequestID     int
	commandPaletteRequestQuery  string
	aiInput                     textinput.Model
	aiFocused                   bool
	aiRunning                   bool
	aiOutput                    string
	aiErr                       error
	aiDuration                  time.Duration
	aiMessages                  []aiMessage
	aiConversationSummary       string
	aiCompactedMessages         int
	aiLastRequestTruncation     aiRequestTruncation
	aiContextRevision           int
	aiLastContextRevision       int
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
	aiFilingRun                 aiFilingRunState
	aiFilingRunActive           bool
	marketOpinionHistory        map[string]domain.PriceSeries
	marketOpinionHistoryAt      map[string]time.Time
	pendingMarketOpinionRefresh bool
	helpOpen                    bool
	globalPageOpen              bool
	globalPageKind              globalPageKind

	quote             domain.QuoteSnapshot
	watchQuotes       map[string]domain.QuoteSnapshot
	series            domain.PriceSeries
	technicalCache    map[string]domain.PriceSeries
	sharpeCache       map[string]domain.PriceSeries
	statementCache    map[string]domain.FinancialStatement
	insiderCache      map[string]domain.InsiderSnapshot
	ownersCache       map[string]domain.OwnershipSnapshot
	analystCache      map[string]domain.AnalystRecommendationsSnapshot
	filingsCache      map[string]domain.FilingsSnapshot
	earningsCache     map[string]domain.EarningsSnapshot
	calendarCache     map[calendarFilterMode]domain.EconomicCalendarSnapshot
	news              []domain.NewsItem
	newsSelected      int
	marketNews        []domain.NewsItem
	marketNewsSources []domain.MarketNewsSource
	marketRisk        domain.MarketRiskSnapshot
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
	owners            domain.OwnershipSnapshot
	analyst           domain.AnalystRecommendationsSnapshot
	statementKind     domain.StatementKind
	statementFreq     domain.StatementFrequency
	ownersSel         int
	analystSel        int
	filings           domain.FilingsSnapshot
	filingsSel        int
	filingsFilter     filingsFilterMode
	earnings          domain.EarningsSnapshot
	earningsSel       int
	calendar          domain.EconomicCalendarSnapshot
	calendarSel       int
	calendarFilter    calendarFilterMode

	errQuote            error
	errHistory          error
	errTechnicalHistory error
	errSharpeHistory    error
	errNews             error
	errMarketNews       error
	errFundamentals     error
	errStatement        error
	errInsiders         error
	errOwners           error
	errAnalyst          error
	errFilings          error
	errEarnings         error
	errCalendar         error
	errScreener         error
}
