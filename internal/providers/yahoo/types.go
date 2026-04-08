package yahoo

import "encoding/json"

type quoteResponse struct {
	QuoteResponse struct {
		Result []struct {
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
		} `json:"result"`
	} `json:"quoteResponse"`
}

type chartResponse struct {
	Chart struct {
		Result []struct {
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
		} `json:"result"`
		Error any `json:"error"`
	} `json:"chart"`
}

type searchResponse struct {
	Quotes []struct {
		Symbol         string `json:"symbol"`
		ShortName      string `json:"shortname"`
		LongName       string `json:"longname"`
		Exchange       string `json:"exchange"`
		QuoteType      string `json:"quoteType"`
		IsYahooFinance bool   `json:"isYahooFinance"`
	} `json:"quotes"`
	News []struct {
		UUID                string `json:"uuid"`
		Title               string `json:"title"`
		Publisher           string `json:"publisher"`
		Link                string `json:"link"`
		ProviderPublishTime int64  `json:"providerPublishTime"`
	} `json:"news"`
}

type quoteSummaryResponse struct {
	QuoteSummary struct {
		Result []struct {
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
		} `json:"result"`
	} `json:"quoteSummary"`
}

type fundamentalsTimeseriesResponse struct {
	Timeseries struct {
		Result []struct {
			TrailingPegRatio []struct {
				ReportedValue numberField `json:"reportedValue"`
			} `json:"trailingPegRatio"`
		} `json:"result"`
		Error any `json:"error"`
	} `json:"timeseries"`
}

type numberField struct {
	Raw float64 `json:"raw"`
}

func (n *numberField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Raw = 0
		return nil
	}

	var direct float64
	if err := json.Unmarshal(data, &direct); err == nil {
		n.Raw = direct
		return nil
	}

	type alias numberField
	var wrapped alias
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	n.Raw = wrapped.Raw
	return nil
}
