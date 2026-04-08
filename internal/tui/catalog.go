package tui

import "time"

var ranges = []struct {
	Label    string
	Range    string
	Interval string
}{
	{Label: "1D", Range: "1d", Interval: "5m"},
	{Label: "5D", Range: "5d", Interval: "15m"},
	{Label: "1M", Range: "1mo", Interval: "1d"},
	{Label: "3M", Range: "3mo", Interval: "1d"},
	{Label: "6M", Range: "6mo", Interval: "1d"},
	{Label: "1Y", Range: "1y", Interval: "1wk"},
	{Label: "5Y", Range: "5y", Interval: "1mo"},
}

var headerTabs = []string{
	"Markets",
	"Quote",
	"News",
	"Screeners",
	"AI",
}

const (
	brandHeaderWordmark = "|- BLACKDESK"
)

const (
	tabMarkets = iota
	tabQuote
	tabNews
	tabScreener
	tabAI
)

const (
	marketNewsRefreshInterval = 20 * time.Second
	screenerResultCount       = 25
)

var marketOpinionHistorySymbols = []string{
	"SPY",
	"HYG",
	"^VIX",
	"2YY=F",
	"^TNX",
	"GC=F",
	"DX-Y.NYB",
}

var marketDashboardSymbols = []string{
	"SPY", "QQQ", "IWM", "DIA",
	"ES=F", "NQ=F", "YM=F", "RTY=F",
	"HYG", "LQD",
	"GC=F", "SI=F", "CL=F", "NG=F",
	"DX-Y.NYB", "EURUSD=X", "JPYUSD=X", "GBPUSD=X",
	"CNYUSD=X", "AUDUSD=X", "CADUSD=X", "CHFUSD=X",
	"2YY=F", "^TNX", "^TYX",
	"^VIX9D", "^VIX3M",
	"BTC-USD",
	"ACWI", "EFA",
	"EEM", "EZU", "AAXJ", "IPAC",
	"ILF", "FM",
	"EWU", "EWG", "EWQ",
	"EWJ", "EWI", "MCHI", "INDA",
	"EWC",
	"XLC", "XLY", "XLP", "XLE",
	"XLF", "XLV", "XLI", "XLB",
	"XLRE", "XLK", "XLU",
	"^VIX",
}
