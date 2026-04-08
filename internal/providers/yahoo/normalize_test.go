package yahoo

import "testing"

func TestNormalizeQuote(t *testing.T) {
	resp := quoteResponse{}
	resp.QuoteResponse.Result = append(resp.QuoteResponse.Result, struct {
		Symbol                     string  `json:"symbol"`
		ShortName                  string  `json:"shortName"`
		Currency                   string  `json:"currency"`
		RegularMarketPrice         float64 `json:"regularMarketPrice"`
		RegularMarketChange        float64 `json:"regularMarketChange"`
		RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
		TrailingPegRatio           float64 `json:"trailingPegRatio"`
		RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
		RegularMarketOpen          float64 `json:"regularMarketOpen"`
		RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
		RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
		RegularMarketVolume        int64   `json:"regularMarketVolume"`
		AverageDailyVolume3Month   int64   `json:"averageDailyVolume3Month"`
		MarketCap                  int64   `json:"marketCap"`
		RegularMarketTime          int64   `json:"regularMarketTime"`
		MarketState                string  `json:"marketState"`
		FullExchangeName           string  `json:"fullExchangeName"`
		PriceHint                  int     `json:"priceHint"`
		PreMarketPrice             float64 `json:"preMarketPrice"`
		PreMarketChange            float64 `json:"preMarketChange"`
		PreMarketChangePercent     float64 `json:"preMarketChangePercent"`
		PostMarketPrice            float64 `json:"postMarketPrice"`
		PostMarketChange           float64 `json:"postMarketChange"`
		PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
	}{
		Symbol:                     "AAPL",
		ShortName:                  "Apple",
		Currency:                   "USD",
		RegularMarketPrice:         200.12,
		RegularMarketChange:        2.11,
		RegularMarketChangePercent: 1.06,
		TrailingPegRatio:           2.35,
		RegularMarketPreviousClose: 198.01,
		RegularMarketOpen:          199.50,
		RegularMarketDayHigh:       201.00,
		RegularMarketDayLow:        198.22,
		RegularMarketVolume:        100,
		AverageDailyVolume3Month:   200,
		MarketCap:                  123,
		RegularMarketTime:          1710000000,
		MarketState:                "REGULAR",
		FullExchangeName:           "NasdaqGS",
		PriceHint:                  2,
		PreMarketPrice:             201.25,
		PreMarketChange:            1.13,
		PreMarketChangePercent:     0.57,
		PostMarketPrice:            202.10,
		PostMarketChange:           1.98,
		PostMarketChangePercent:    0.99,
	})

	quote, err := normalizeQuote(resp)
	if err != nil {
		t.Fatal(err)
	}
	if quote.Symbol != "AAPL" || quote.Price != 200.12 {
		t.Fatalf("unexpected quote %+v", quote)
	}
	if quote.MarketState != "regular" {
		t.Fatalf("unexpected market state %s", quote.MarketState)
	}
	if quote.PreMarketPrice != 201.25 || quote.PostMarketPrice != 202.10 {
		t.Fatalf("unexpected extended-hours quote %+v", quote)
	}
	if quote.TrailingPEGRatio != 2.35 {
		t.Fatalf("unexpected peg ratio %+v", quote)
	}
}

func TestNormalizeQuotes(t *testing.T) {
	resp := quoteResponse{}
	resp.QuoteResponse.Result = append(resp.QuoteResponse.Result,
		struct {
			Symbol                     string  `json:"symbol"`
			ShortName                  string  `json:"shortName"`
			Currency                   string  `json:"currency"`
			RegularMarketPrice         float64 `json:"regularMarketPrice"`
			RegularMarketChange        float64 `json:"regularMarketChange"`
			RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
			TrailingPegRatio           float64 `json:"trailingPegRatio"`
			RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
			RegularMarketOpen          float64 `json:"regularMarketOpen"`
			RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
			RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
			RegularMarketVolume        int64   `json:"regularMarketVolume"`
			AverageDailyVolume3Month   int64   `json:"averageDailyVolume3Month"`
			MarketCap                  int64   `json:"marketCap"`
			RegularMarketTime          int64   `json:"regularMarketTime"`
			MarketState                string  `json:"marketState"`
			FullExchangeName           string  `json:"fullExchangeName"`
			PriceHint                  int     `json:"priceHint"`
			PreMarketPrice             float64 `json:"preMarketPrice"`
			PreMarketChange            float64 `json:"preMarketChange"`
			PreMarketChangePercent     float64 `json:"preMarketChangePercent"`
			PostMarketPrice            float64 `json:"postMarketPrice"`
			PostMarketChange           float64 `json:"postMarketChange"`
			PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
		}{Symbol: "AAPL", RegularMarketPrice: 200.12},
		struct {
			Symbol                     string  `json:"symbol"`
			ShortName                  string  `json:"shortName"`
			Currency                   string  `json:"currency"`
			RegularMarketPrice         float64 `json:"regularMarketPrice"`
			RegularMarketChange        float64 `json:"regularMarketChange"`
			RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
			TrailingPegRatio           float64 `json:"trailingPegRatio"`
			RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
			RegularMarketOpen          float64 `json:"regularMarketOpen"`
			RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
			RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
			RegularMarketVolume        int64   `json:"regularMarketVolume"`
			AverageDailyVolume3Month   int64   `json:"averageDailyVolume3Month"`
			MarketCap                  int64   `json:"marketCap"`
			RegularMarketTime          int64   `json:"regularMarketTime"`
			MarketState                string  `json:"marketState"`
			FullExchangeName           string  `json:"fullExchangeName"`
			PriceHint                  int     `json:"priceHint"`
			PreMarketPrice             float64 `json:"preMarketPrice"`
			PreMarketChange            float64 `json:"preMarketChange"`
			PreMarketChangePercent     float64 `json:"preMarketChangePercent"`
			PostMarketPrice            float64 `json:"postMarketPrice"`
			PostMarketChange           float64 `json:"postMarketChange"`
			PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
		}{Symbol: "MSFT", RegularMarketPrice: 420.50},
	)

	quotes := normalizeQuotes(resp)
	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}
	if quotes[1].Symbol != "MSFT" || quotes[1].Price != 420.50 {
		t.Fatalf("unexpected quote %+v", quotes[1])
	}
}

func TestNormalizeChart(t *testing.T) {
	open1, high1, low1, close1 := 10.0, 12.0, 9.0, 11.0
	open2, high2, low2, close2 := 11.0, 13.0, 10.0, 12.0
	vol1, vol2 := int64(100), int64(200)
	resp := chartResponse{}
	resp.Chart.Result = append(resp.Chart.Result, struct {
		Timestamp  []int64 `json:"timestamp"`
		Indicators struct {
			Quote []struct {
				Open   []*float64 `json:"open"`
				High   []*float64 `json:"high"`
				Low    []*float64 `json:"low"`
				Close  []*float64 `json:"close"`
				Volume []*int64   `json:"volume"`
			} `json:"quote"`
		} `json:"indicators"`
	}{
		Timestamp: []int64{1710000000, 1710086400},
	})
	resp.Chart.Result[0].Indicators.Quote = append(resp.Chart.Result[0].Indicators.Quote, struct {
		Open   []*float64 `json:"open"`
		High   []*float64 `json:"high"`
		Low    []*float64 `json:"low"`
		Close  []*float64 `json:"close"`
		Volume []*int64   `json:"volume"`
	}{
		Open:   []*float64{&open1, &open2},
		High:   []*float64{&high1, &high2},
		Low:    []*float64{&low1, &low2},
		Close:  []*float64{&close1, &close2},
		Volume: []*int64{&vol1, &vol2},
	})

	series, err := normalizeChart("AAPL", "1mo", "1d", resp)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(series.Candles); got != 2 {
		t.Fatalf("expected 2 candles, got %d", got)
	}
	if series.Candles[1].Close != 12.0 {
		t.Fatalf("unexpected close %.2f", series.Candles[1].Close)
	}
}

func TestNormalizeFundamentals(t *testing.T) {
	resp := quoteSummaryResponse{}
	resp.QuoteSummary.Result = append(resp.QuoteSummary.Result, struct {
		Price struct {
			MarketCap numberField `json:"marketCap"`
		} `json:"price"`
		SummaryDetail struct {
			DividendYield    numberField `json:"dividendYield"`
			FiftyTwoWeekLow  numberField `json:"fiftyTwoWeekLow"`
			FiftyTwoWeekHigh numberField `json:"fiftyTwoWeekHigh"`
			TrailingPE       numberField `json:"trailingPE"`
			ForwardPE        numberField `json:"forwardPE"`
			AverageVolume    numberField `json:"averageVolume"`
			Beta             numberField `json:"beta"`
		} `json:"summaryDetail"`
		DefaultKeyStatistics struct {
			EarningsQuarterlyGrowth      numberField `json:"earningsQuarterlyGrowth"`
			PegRatio                     numberField `json:"pegRatio"`
			PriceToBook                  numberField `json:"priceToBook"`
			PriceToSalesTrailing12Months numberField `json:"priceToSalesTrailing12Months"`
			EnterpriseValue              numberField `json:"enterpriseValue"`
			EnterpriseToRevenue          numberField `json:"enterpriseToRevenue"`
			EnterpriseToEbitda           numberField `json:"enterpriseToEbitda"`
			BookValue                    numberField `json:"bookValue"`
			TrailingEps                  numberField `json:"trailingEps"`
		} `json:"defaultKeyStatistics"`
		FinancialData struct {
			EpsCurrentYear          numberField `json:"epsCurrentYear"`
			RevenuePerShare         numberField `json:"revenuePerShare"`
			RecommendationKey       string      `json:"recommendationKey"`
			RecommendationMean      numberField `json:"recommendationMean"`
			NumberOfAnalystOpinions numberField `json:"numberOfAnalystOpinions"`
			TargetLowPrice          numberField `json:"targetLowPrice"`
			TargetMeanPrice         numberField `json:"targetMeanPrice"`
			TargetHighPrice         numberField `json:"targetHighPrice"`
			TotalRevenue            numberField `json:"totalRevenue"`
			GrossProfits            numberField `json:"grossProfits"`
			Ebitda                  numberField `json:"ebitda"`
			OperatingCashflow       numberField `json:"operatingCashflow"`
			FreeCashflow            numberField `json:"freeCashflow"`
			TotalCash               numberField `json:"totalCash"`
			TotalDebt               numberField `json:"totalDebt"`
			CurrentRatio            numberField `json:"currentRatio"`
			QuickRatio              numberField `json:"quickRatio"`
			DebtToEquity            numberField `json:"debtToEquity"`
			ReturnOnAssets          numberField `json:"returnOnAssets"`
			ReturnOnEquity          numberField `json:"returnOnEquity"`
			RevenueGrowth           numberField `json:"revenueGrowth"`
			EarningsGrowth          numberField `json:"earningsGrowth"`
			GrossMargins            numberField `json:"grossMargins"`
			ProfitMargins           numberField `json:"profitMargins"`
			OperatingMargins        numberField `json:"operatingMargins"`
		} `json:"financialData"`
		AssetProfile struct {
			Sector              string `json:"sector"`
			Industry            string `json:"industry"`
			LongBusinessSummary string `json:"longBusinessSummary"`
		} `json:"assetProfile"`
	}{
		AssetProfile: struct {
			Sector              string `json:"sector"`
			Industry            string `json:"industry"`
			LongBusinessSummary string `json:"longBusinessSummary"`
		}{Sector: "Technology", Industry: "Consumer Electronics", LongBusinessSummary: "Apple designs consumer electronics."},
	})
	resp.QuoteSummary.Result[0].Price.MarketCap.Raw = 1000
	resp.QuoteSummary.Result[0].SummaryDetail.TrailingPE.Raw = 25
	resp.QuoteSummary.Result[0].SummaryDetail.ForwardPE.Raw = 20
	resp.QuoteSummary.Result[0].SummaryDetail.DividendYield.Raw = 0.005
	resp.QuoteSummary.Result[0].DefaultKeyStatistics.PegRatio.Raw = 1.75
	resp.QuoteSummary.Result[0].DefaultKeyStatistics.PriceToSalesTrailing12Months.Raw = 6.8
	resp.QuoteSummary.Result[0].DefaultKeyStatistics.EnterpriseValue.Raw = 1200
	resp.QuoteSummary.Result[0].DefaultKeyStatistics.EnterpriseToRevenue.Raw = 7.2
	resp.QuoteSummary.Result[0].DefaultKeyStatistics.EnterpriseToEbitda.Raw = 18.4
	resp.QuoteSummary.Result[0].DefaultKeyStatistics.BookValue.Raw = 4.2
	resp.QuoteSummary.Result[0].DefaultKeyStatistics.TrailingEps.Raw = 4.9
	resp.QuoteSummary.Result[0].FinancialData.EpsCurrentYear.Raw = 6.5
	resp.QuoteSummary.Result[0].FinancialData.RevenuePerShare.Raw = 24.5
	resp.QuoteSummary.Result[0].FinancialData.RecommendationKey = "buy"
	resp.QuoteSummary.Result[0].FinancialData.RecommendationMean.Raw = 1.9
	resp.QuoteSummary.Result[0].FinancialData.NumberOfAnalystOpinions.Raw = 37
	resp.QuoteSummary.Result[0].FinancialData.TargetLowPrice.Raw = 180
	resp.QuoteSummary.Result[0].FinancialData.TargetMeanPrice.Raw = 225
	resp.QuoteSummary.Result[0].FinancialData.TargetHighPrice.Raw = 250
	resp.QuoteSummary.Result[0].FinancialData.TotalRevenue.Raw = 400
	resp.QuoteSummary.Result[0].FinancialData.GrossProfits.Raw = 175
	resp.QuoteSummary.Result[0].FinancialData.Ebitda.Raw = 150
	resp.QuoteSummary.Result[0].FinancialData.OperatingCashflow.Raw = 120
	resp.QuoteSummary.Result[0].FinancialData.FreeCashflow.Raw = 95
	resp.QuoteSummary.Result[0].FinancialData.TotalCash.Raw = 65
	resp.QuoteSummary.Result[0].FinancialData.TotalDebt.Raw = 98
	resp.QuoteSummary.Result[0].FinancialData.CurrentRatio.Raw = 1.12
	resp.QuoteSummary.Result[0].FinancialData.QuickRatio.Raw = 0.94
	resp.QuoteSummary.Result[0].FinancialData.DebtToEquity.Raw = 145
	resp.QuoteSummary.Result[0].FinancialData.ReturnOnAssets.Raw = 0.21
	resp.QuoteSummary.Result[0].FinancialData.ReturnOnEquity.Raw = 0.44
	resp.QuoteSummary.Result[0].FinancialData.RevenueGrowth.Raw = 0.07
	resp.QuoteSummary.Result[0].FinancialData.EarningsGrowth.Raw = 0.11
	resp.QuoteSummary.Result[0].FinancialData.GrossMargins.Raw = 0.456
	resp.QuoteSummary.Result[0].FinancialData.ProfitMargins.Raw = 0.212

	f, err := normalizeFundamentals("AAPL", resp)
	if err != nil {
		t.Fatal(err)
	}
	if f.Sector != "Technology" || f.TrailingPE != 25 {
		t.Fatalf("unexpected fundamentals %+v", f)
	}
	if f.Description != "Apple designs consumer electronics." {
		t.Fatalf("unexpected description %q", f.Description)
	}
	if f.RecommendationMean != 1.9 || f.AnalystOpinions != 37 {
		t.Fatalf("unexpected analyst data %+v", f)
	}
	if f.TargetMeanPrice != 225 || f.TargetLowPrice != 180 || f.TargetHighPrice != 250 {
		t.Fatalf("unexpected price targets %+v", f)
	}
	if f.EPS != 4.9 || f.TrailingEPS != 4.9 {
		t.Fatalf("unexpected eps %+v", f)
	}
	if f.PEGRatio != 1.75 {
		t.Fatalf("unexpected peg ratio %+v", f)
	}
	if f.EnterpriseValue != 1200 || f.PriceToSales != 6.8 || f.EnterpriseToEBITDA != 18.4 {
		t.Fatalf("unexpected valuation metrics %+v", f)
	}
	if f.GrossMargins != 0.456 || f.ProfitMargins != 0.212 {
		t.Fatalf("unexpected margins %+v", f)
	}
	if f.Revenue != 400 || f.FreeCashflow != 95 || f.TotalDebt != 98 {
		t.Fatalf("unexpected financial data %+v", f)
	}
}

func TestNormalizeFundamentalsDerivesMissingValuationRatios(t *testing.T) {
	resp := quoteSummaryResponse{}
	resp.QuoteSummary.Result = append(resp.QuoteSummary.Result, struct {
		Price struct {
			MarketCap numberField `json:"marketCap"`
		} `json:"price"`
		SummaryDetail struct {
			DividendYield    numberField `json:"dividendYield"`
			FiftyTwoWeekLow  numberField `json:"fiftyTwoWeekLow"`
			FiftyTwoWeekHigh numberField `json:"fiftyTwoWeekHigh"`
			TrailingPE       numberField `json:"trailingPE"`
			ForwardPE        numberField `json:"forwardPE"`
			AverageVolume    numberField `json:"averageVolume"`
			Beta             numberField `json:"beta"`
		} `json:"summaryDetail"`
		DefaultKeyStatistics struct {
			EarningsQuarterlyGrowth      numberField `json:"earningsQuarterlyGrowth"`
			PegRatio                     numberField `json:"pegRatio"`
			PriceToBook                  numberField `json:"priceToBook"`
			PriceToSalesTrailing12Months numberField `json:"priceToSalesTrailing12Months"`
			EnterpriseValue              numberField `json:"enterpriseValue"`
			EnterpriseToRevenue          numberField `json:"enterpriseToRevenue"`
			EnterpriseToEbitda           numberField `json:"enterpriseToEbitda"`
			BookValue                    numberField `json:"bookValue"`
			TrailingEps                  numberField `json:"trailingEps"`
		} `json:"defaultKeyStatistics"`
		FinancialData struct {
			EpsCurrentYear          numberField `json:"epsCurrentYear"`
			RevenuePerShare         numberField `json:"revenuePerShare"`
			RecommendationKey       string      `json:"recommendationKey"`
			RecommendationMean      numberField `json:"recommendationMean"`
			NumberOfAnalystOpinions numberField `json:"numberOfAnalystOpinions"`
			TargetLowPrice          numberField `json:"targetLowPrice"`
			TargetMeanPrice         numberField `json:"targetMeanPrice"`
			TargetHighPrice         numberField `json:"targetHighPrice"`
			TotalRevenue            numberField `json:"totalRevenue"`
			GrossProfits            numberField `json:"grossProfits"`
			Ebitda                  numberField `json:"ebitda"`
			OperatingCashflow       numberField `json:"operatingCashflow"`
			FreeCashflow            numberField `json:"freeCashflow"`
			TotalCash               numberField `json:"totalCash"`
			TotalDebt               numberField `json:"totalDebt"`
			CurrentRatio            numberField `json:"currentRatio"`
			QuickRatio              numberField `json:"quickRatio"`
			DebtToEquity            numberField `json:"debtToEquity"`
			ReturnOnAssets          numberField `json:"returnOnAssets"`
			ReturnOnEquity          numberField `json:"returnOnEquity"`
			RevenueGrowth           numberField `json:"revenueGrowth"`
			EarningsGrowth          numberField `json:"earningsGrowth"`
			GrossMargins            numberField `json:"grossMargins"`
			ProfitMargins           numberField `json:"profitMargins"`
			OperatingMargins        numberField `json:"operatingMargins"`
		} `json:"financialData"`
		AssetProfile struct {
			Sector              string `json:"sector"`
			Industry            string `json:"industry"`
			LongBusinessSummary string `json:"longBusinessSummary"`
		} `json:"assetProfile"`
	}{})
	resp.QuoteSummary.Result[0].Price.MarketCap.Raw = 4000
	resp.QuoteSummary.Result[0].FinancialData.TotalRevenue.Raw = 500
	resp.QuoteSummary.Result[0].FinancialData.Ebitda.Raw = 125
	resp.QuoteSummary.Result[0].FinancialData.TotalCash.Raw = 250
	resp.QuoteSummary.Result[0].FinancialData.TotalDebt.Raw = 600

	f, err := normalizeFundamentals("AAPL", resp)
	if err != nil {
		t.Fatal(err)
	}
	if f.EnterpriseValue != 4350 {
		t.Fatalf("expected derived enterprise value, got %+v", f)
	}
	if f.PriceToSales != 8 {
		t.Fatalf("expected derived P/S, got %+v", f)
	}
	if f.EnterpriseToRevenue != 8.7 {
		t.Fatalf("expected derived EV/revenue, got %+v", f)
	}
	if f.EnterpriseToEBITDA != 34.8 {
		t.Fatalf("expected derived EV/EBITDA, got %+v", f)
	}
}
