package domain

import "time"

type MarketState string

const (
	MarketStateRegular    MarketState = "regular"
	MarketStatePre        MarketState = "pre"
	MarketStatePost       MarketState = "post"
	MarketStateClosed     MarketState = "closed"
	MarketStateUnknown    MarketState = "unknown"
	FreshnessLive         string      = "live"
	FreshnessCached       string      = "cached"
	FreshnessStale        string      = "stale"
	FreshnessError        string      = "error"
	DefaultRefreshSeconds int         = 60
)

type SymbolRef struct {
	Symbol   string
	Name     string
	Exchange string
	Type     string
}

type QuoteSnapshot struct {
	Symbol               string
	ShortName            string
	Currency             string
	Price                float64
	Change               float64
	ChangePercent        float64
	TrailingPEGRatio     float64
	PreviousClose        float64
	Open                 float64
	DayHigh              float64
	DayLow               float64
	Volume               int64
	AverageVolume        int64
	MarketCap            int64
	RegularMarketTime    time.Time
	MarketState          MarketState
	Exchange             string
	Freshness            string
	Provider             string
	PriceHint            int
	PreMarketPrice       float64
	PreMarketChange      float64
	PreMarketChangePerc  float64
	PostMarketPrice      float64
	PostMarketChange     float64
	PostMarketChangePerc float64
}

type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

type NewsItem struct {
	UUID      string
	Title     string
	Summary   string
	Publisher string
	URL       string
	Time      time.Time
}

type MarketNewsSource struct {
	Name string
}

type MarketRiskThresholds struct {
	SMABufferPct    float64
	Breadth50Buffer float64
}

type MarketRiskInput struct {
	Name    string
	Symbol  string
	Current float64
	SMA200  float64
}

type MarketRiskSnapshot struct {
	Score          int
	Label          string
	Min            int
	Max            int
	Thresholds     MarketRiskThresholds
	Components     map[string]int
	Inputs         map[string]MarketRiskInput
	MarketNow      time.Time
	MarketZone     string
	MarketCalendar string
	GeneratedAt    time.Time
	Available      bool
}

type ScreenerDefinition struct {
	ID          string
	Name        string
	Description string
	Category    string
	Kind        string
}

type ScreenerCriterion struct {
	Field     string
	Operator  string
	Value     string
	Statement string
}

type ScreenerMetric struct {
	Key    string
	Label  string
	Value  string
	Signal string
}

type ScreenerItem struct {
	Symbol        string
	Name          string
	Exchange      string
	Type          string
	Currency      string
	Price         float64
	Change        float64
	ChangePercent float64
	Volume        int64
	AverageVolume int64
	MarketCap     int64
	MarketState   MarketState
	UpdatedAt     time.Time
	Metrics       []ScreenerMetric
}

type ScreenerResult struct {
	Definition ScreenerDefinition
	SortField  string
	SortOrder  string
	Total      int
	Items      []ScreenerItem
	Criteria   []ScreenerCriterion
	Freshness  string
	Provider   string
	UpdatedAt  time.Time
}

type FundamentalsSnapshot struct {
	Symbol                  string
	Sector                  string
	Industry                string
	Description             string
	MarketCap               int64
	EnterpriseValue         int64
	TrailingPE              float64
	ForwardPE               float64
	PEGRatio                float64
	PriceToSales            float64
	EnterpriseToRevenue     float64
	EnterpriseToEBITDA      float64
	BookValue               float64
	TrailingEPS             float64
	EPS                     float64
	RevenuePerShare         float64
	DividendYield           float64
	FiftyTwoWeekLow         float64
	FiftyTwoWeekHigh        float64
	AverageVolume           int64
	Beta                    float64
	PriceToBook             float64
	Revenue                 int64
	GrossProfits            int64
	EBITDA                  int64
	OperatingCashflow       int64
	FreeCashflow            int64
	TotalCash               int64
	TotalDebt               int64
	InvestedCapital         int64
	GrossMargins            float64
	ProfitMargins           float64
	OperatingMargins        float64
	ReturnOnAssets          float64
	ReturnOnEquity          float64
	ReturnOnInvestedCapital float64
	RevenueGrowth           float64
	EarningsGrowth          float64
	CurrentRatio            float64
	QuickRatio              float64
	DebtToEquity            float64
	RecommendationMean      float64
	RecommendationKey       string
	AnalystOpinions         int
	TargetLowPrice          float64
	TargetMeanPrice         float64
	TargetHighPrice         float64
	Freshness               string
}

type StatementKind string

const (
	StatementKindIncome       StatementKind = "income"
	StatementKindBalanceSheet StatementKind = "balance_sheet"
	StatementKindCashFlow     StatementKind = "cash_flow"
)

type StatementFrequency string

const (
	StatementFrequencyAnnual    StatementFrequency = "annual"
	StatementFrequencyQuarterly StatementFrequency = "quarterly"
)

type StatementPeriod struct {
	Label         string
	FiscalYear    int
	FiscalQuarter int
	EndDate       time.Time
}

type StatementValue struct {
	Value   float64
	Present bool
}

type StatementRow struct {
	Key    string
	Label  string
	Values []StatementValue
}

type FinancialStatement struct {
	Symbol    string
	Kind      StatementKind
	Frequency StatementFrequency
	Currency  string
	Periods   []StatementPeriod
	Rows      []StatementRow
	Freshness string
	Provider  string
	UpdatedAt time.Time
}

type InsiderPurchaseActivity struct {
	Period                  string
	BuyShares               int64
	BuyTransactions         int
	SellShares              int64
	SellTransactions        int
	NetShares               int64
	NetTransactions         int
	TotalInsiderShares      int64
	NetPercentInsiderShares float64
}

type FilingItem struct {
	AccessionNumber       string
	Form                  string
	FilingDate            time.Time
	ReportDate            time.Time
	AcceptedAt            time.Time
	PrimaryDocument       string
	PrimaryDocDescription string
	URL                   string
	IsXBRL                bool
	IsInlineXBRL          bool
}

type FilingsSnapshot struct {
	Symbol      string
	CompanyName string
	CIK         string
	Items       []FilingItem
	Freshness   string
	Provider    string
	UpdatedAt   time.Time
}

type FilingDocument struct {
	Item        FilingItem
	ContentType string
	Text        string
	Truncated   bool
	Provider    string
	RetrievedAt time.Time
}

type InsiderTransaction struct {
	Insider   string
	Relation  string
	Ownership string
	Action    string
	Date      time.Time
	Shares    int64
	Value     int64
	Text      string
}

type InsiderRosterMember struct {
	Name                  string
	Relation              string
	LatestTransaction     string
	LatestTransactionAt   time.Time
	SharesOwnedDirectly   int64
	PositionDirectDate    time.Time
	SharesOwnedIndirectly int64
	PositionIndirectDate  time.Time
}

type InsiderSnapshot struct {
	Symbol           string
	PurchaseActivity InsiderPurchaseActivity
	Transactions     []InsiderTransaction
	Roster           []InsiderRosterMember
	Freshness        string
	Provider         string
	UpdatedAt        time.Time
}

type PriceSeries struct {
	Symbol      string
	Range       string
	Interval    string
	Candles     []Candle
	Freshness   string
	LastUpdated time.Time
}

type ProviderCapabilities struct {
	Quote        bool
	History      bool
	News         bool
	MarketNews   bool
	Fundamentals bool
	Search       bool
	Statements   bool
	Insiders     bool
	Screeners    bool
}
