package tui

type marketBoardItem struct {
	label  string
	symbol string
}

var marketUSBoard = []marketBoardItem{
	{label: "S&P 500", symbol: "SPY"},
	{label: "Nasdaq 100", symbol: "QQQ"},
	{label: "Russell", symbol: "IWM"},
	{label: "Dow", symbol: "DIA"},
}

var marketFuturesBoard = []marketBoardItem{
	{label: "S&P Fut", symbol: "ES=F"},
	{label: "Nasdaq Fut", symbol: "NQ=F"},
	{label: "Dow Fut", symbol: "YM=F"},
	{label: "Russell Fut", symbol: "RTY=F"},
}

var marketRatesBoard = []marketBoardItem{
	{label: "IG Credit", symbol: "LQD"},
	{label: "HY Credit", symbol: "HYG"},
}

var marketMacroBoard = []marketBoardItem{
	{label: "Gold", symbol: "GC=F"},
	{label: "Silver", symbol: "SI=F"},
	{label: "Oil", symbol: "CL=F"},
	{label: "Nat Gas", symbol: "NG=F"},
	{label: "Bitcoin", symbol: "BTC-USD"},
}

var marketFXBoard = []marketBoardItem{
	{label: "Dollar", symbol: "DX-Y.NYB"},
	{label: "Euro", symbol: "EURUSD=X"},
	{label: "Yen", symbol: "JPYUSD=X"},
	{label: "Pound", symbol: "GBPUSD=X"},
	{label: "Yuan", symbol: "CNYUSD=X"},
	{label: "Aussie", symbol: "AUDUSD=X"},
	{label: "Loonie", symbol: "CADUSD=X"},
	{label: "Swiss", symbol: "CHFUSD=X"},
}

var marketVolBoard = []marketBoardItem{
	{label: "VIX 9D", symbol: "^VIX9D"},
	{label: "VIX", symbol: "^VIX"},
	{label: "VIX 3M", symbol: "^VIX3M"},
}

var marketYieldBoard = []marketBoardItem{
	{label: "2Y", symbol: "2YY=F"},
	{label: "10Y", symbol: "^TNX"},
	{label: "30Y", symbol: "^TYX"},
}

var marketRegionBoard = []marketBoardItem{
	{label: "All World", symbol: "ACWI"},
	{label: "Developed", symbol: "EFA"},
	{label: "Emerging", symbol: "EEM"},
	{label: "Europe", symbol: "EZU"},
	{label: "Asia", symbol: "AAXJ"},
	{label: "Pacific", symbol: "IPAC"},
	{label: "Latin America", symbol: "ILF"},
	{label: "Frontier", symbol: "FM"},
}

var marketCountryBoard = []marketBoardItem{
	{label: "UK", symbol: "EWU"},
	{label: "Germany", symbol: "EWG"},
	{label: "France", symbol: "EWQ"},
	{label: "Italy", symbol: "EWI"},
	{label: "Japan", symbol: "EWJ"},
	{label: "China", symbol: "MCHI"},
	{label: "India", symbol: "INDA"},
	{label: "Canada", symbol: "EWC"},
}

var marketSectorBoard = []marketBoardItem{
	{label: "Comm Svcs", symbol: "XLC"},
	{label: "Cons Disc", symbol: "XLY"},
	{label: "Cons Stap", symbol: "XLP"},
	{label: "Energy", symbol: "XLE"},
	{label: "Financials", symbol: "XLF"},
	{label: "Health Care", symbol: "XLV"},
	{label: "Industrials", symbol: "XLI"},
	{label: "Materials", symbol: "XLB"},
	{label: "Real Estate", symbol: "XLRE"},
	{label: "Technology", symbol: "XLK"},
	{label: "Utilities", symbol: "XLU"},
}
