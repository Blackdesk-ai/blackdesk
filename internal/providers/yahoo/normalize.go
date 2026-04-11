package yahoo

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

func normalizeQuote(resp quoteResponse) (domain.QuoteSnapshot, error) {
	quotes := normalizeQuotes(resp)
	if len(quotes) == 0 {
		return domain.QuoteSnapshot{}, errors.New("quote not found")
	}
	return quotes[0], nil
}

func normalizeQuotes(resp quoteResponse) []domain.QuoteSnapshot {
	quotes := make([]domain.QuoteSnapshot, 0, len(resp.QuoteResponse.Result))
	for _, q := range resp.QuoteResponse.Result {
		quotes = append(quotes, domain.QuoteSnapshot{
			Symbol:               q.Symbol,
			ShortName:            q.ShortName,
			Currency:             q.Currency,
			Price:                q.RegularMarketPrice,
			Change:               q.RegularMarketChange,
			ChangePercent:        q.RegularMarketChangePercent,
			TrailingPEGRatio:     q.TrailingPegRatio,
			PreviousClose:        q.RegularMarketPreviousClose,
			Open:                 q.RegularMarketOpen,
			DayHigh:              q.RegularMarketDayHigh,
			DayLow:               q.RegularMarketDayLow,
			Volume:               q.RegularMarketVolume,
			AverageVolume:        q.AverageDailyVolume3Month,
			MarketCap:            q.MarketCap,
			RegularMarketTime:    time.Unix(q.RegularMarketTime, 0),
			MarketState:          normalizeMarketState(q.MarketState),
			Exchange:             q.FullExchangeName,
			Freshness:            domain.FreshnessLive,
			Provider:             "yahoo",
			PriceHint:            q.PriceHint,
			PreMarketPrice:       q.PreMarketPrice,
			PreMarketChange:      q.PreMarketChange,
			PreMarketChangePerc:  q.PreMarketChangePercent,
			PostMarketPrice:      q.PostMarketPrice,
			PostMarketChange:     q.PostMarketChange,
			PostMarketChangePerc: q.PostMarketChangePercent,
		})
	}
	return quotes
}

func quoteBySymbol(quotes []domain.QuoteSnapshot, symbol string) (domain.QuoteSnapshot, error) {
	needle := strings.ToUpper(strings.TrimSpace(symbol))
	for _, quote := range quotes {
		if strings.EqualFold(quote.Symbol, needle) {
			return quote, nil
		}
	}
	return domain.QuoteSnapshot{}, fmt.Errorf("quote not found for %s", needle)
}

func normalizeChart(symbol, rangeKey, interval string, resp chartResponse) (domain.PriceSeries, error) {
	if len(resp.Chart.Result) == 0 || len(resp.Chart.Result[0].Indicators.Quote) == 0 {
		return domain.PriceSeries{}, errors.New("chart data not found")
	}
	result := resp.Chart.Result[0]
	quote := result.Indicators.Quote[0]

	candles := make([]domain.Candle, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		if i >= len(quote.Close) || quote.Close[i] == nil {
			continue
		}
		candle := domain.Candle{
			Time:  time.Unix(ts, 0),
			Close: *quote.Close[i],
		}
		if i < len(quote.Open) && quote.Open[i] != nil {
			candle.Open = *quote.Open[i]
		}
		if i < len(quote.High) && quote.High[i] != nil {
			candle.High = *quote.High[i]
		}
		if i < len(quote.Low) && quote.Low[i] != nil {
			candle.Low = *quote.Low[i]
		}
		if i < len(quote.Volume) && quote.Volume[i] != nil {
			candle.Volume = *quote.Volume[i]
		}
		candles = append(candles, candle)
	}
	if len(candles) == 0 {
		return domain.PriceSeries{}, errors.New("chart candles empty")
	}
	return domain.PriceSeries{
		Symbol:      strings.ToUpper(symbol),
		Range:       rangeKey,
		Interval:    interval,
		Candles:     candles,
		Freshness:   domain.FreshnessLive,
		LastUpdated: time.Now(),
	}, nil
}

func normalizeNews(resp searchResponse) []domain.NewsItem {
	items := make([]domain.NewsItem, 0, len(resp.News))
	for _, item := range resp.News {
		items = append(items, domain.NewsItem{
			UUID:      item.UUID,
			Title:     item.Title,
			Publisher: item.Publisher,
			URL:       item.Link,
			Time:      time.Unix(item.ProviderPublishTime, 0),
		})
	}
	return items
}

func normalizeSearch(resp searchResponse) []domain.SymbolRef {
	out := make([]domain.SymbolRef, 0, len(resp.Quotes))
	for _, q := range resp.Quotes {
		name := q.ShortName
		if name == "" {
			name = q.LongName
		}
		out = append(out, domain.SymbolRef{
			Symbol:   q.Symbol,
			Name:     name,
			Exchange: q.Exchange,
			Type:     q.QuoteType,
		})
	}
	return out
}

func normalizeFundamentals(symbol string, resp quoteSummaryResponse) (domain.FundamentalsSnapshot, error) {
	if len(resp.QuoteSummary.Result) == 0 {
		return domain.FundamentalsSnapshot{}, errors.New("fundamentals not found")
	}
	r := resp.QuoteSummary.Result[0]
	marketCap := int64(r.Price.MarketCap.Raw)
	revenue := int64(r.FinancialData.TotalRevenue.Raw)
	ebitda := int64(r.FinancialData.Ebitda.Raw)
	totalCash := int64(r.FinancialData.TotalCash.Raw)
	totalDebt := int64(r.FinancialData.TotalDebt.Raw)
	enterpriseValue := int64(r.DefaultKeyStatistics.EnterpriseValue.Raw)
	if enterpriseValue == 0 {
		enterpriseValue = deriveEnterpriseValue(marketCap, totalDebt, totalCash)
	}
	eps := r.DefaultKeyStatistics.TrailingEps.Raw
	if eps == 0 {
		eps = r.FinancialData.EpsCurrentYear.Raw
	}
	priceToSales := r.DefaultKeyStatistics.PriceToSalesTrailing12Months.Raw
	if priceToSales == 0 {
		priceToSales = ratioInt64(marketCap, revenue)
	}
	enterpriseToRevenue := r.DefaultKeyStatistics.EnterpriseToRevenue.Raw
	if enterpriseToRevenue == 0 {
		enterpriseToRevenue = ratioInt64(enterpriseValue, revenue)
	}
	enterpriseToEBITDA := r.DefaultKeyStatistics.EnterpriseToEbitda.Raw
	if enterpriseToEBITDA == 0 {
		enterpriseToEBITDA = ratioInt64(enterpriseValue, ebitda)
	}
	earningsGrowth := firstNonZero(r.FinancialData.EarningsGrowth.Raw, r.DefaultKeyStatistics.EarningsQuarterlyGrowth.Raw)
	return domain.FundamentalsSnapshot{
		Symbol:              strings.ToUpper(symbol),
		Sector:              r.AssetProfile.Sector,
		Industry:            r.AssetProfile.Industry,
		Description:         strings.TrimSpace(r.AssetProfile.LongBusinessSummary),
		MarketCap:           marketCap,
		EnterpriseValue:     enterpriseValue,
		TrailingPE:          r.SummaryDetail.TrailingPE.Raw,
		ForwardPE:           r.SummaryDetail.ForwardPE.Raw,
		PEGRatio:            r.DefaultKeyStatistics.PegRatio.Raw,
		PriceToSales:        priceToSales,
		EnterpriseToRevenue: enterpriseToRevenue,
		EnterpriseToEBITDA:  enterpriseToEBITDA,
		BookValue:           r.DefaultKeyStatistics.BookValue.Raw,
		TrailingEPS:         r.DefaultKeyStatistics.TrailingEps.Raw,
		EPS:                 eps,
		RevenuePerShare:     r.FinancialData.RevenuePerShare.Raw,
		DividendYield:       r.SummaryDetail.DividendYield.Raw,
		FiftyTwoWeekLow:     r.SummaryDetail.FiftyTwoWeekLow.Raw,
		FiftyTwoWeekHigh:    r.SummaryDetail.FiftyTwoWeekHigh.Raw,
		AverageVolume:       int64(r.SummaryDetail.AverageVolume.Raw),
		Beta:                r.SummaryDetail.Beta.Raw,
		PriceToBook:         r.DefaultKeyStatistics.PriceToBook.Raw,
		Revenue:             revenue,
		GrossProfits:        int64(r.FinancialData.GrossProfits.Raw),
		EBITDA:              ebitda,
		OperatingCashflow:   int64(r.FinancialData.OperatingCashflow.Raw),
		FreeCashflow:        int64(r.FinancialData.FreeCashflow.Raw),
		TotalCash:           totalCash,
		TotalDebt:           totalDebt,
		GrossMargins:        r.FinancialData.GrossMargins.Raw,
		ProfitMargins:       r.FinancialData.ProfitMargins.Raw,
		OperatingMargins:    r.FinancialData.OperatingMargins.Raw,
		ReturnOnAssets:      r.FinancialData.ReturnOnAssets.Raw,
		ReturnOnEquity:      r.FinancialData.ReturnOnEquity.Raw,
		RevenueGrowth:       r.FinancialData.RevenueGrowth.Raw,
		EarningsGrowth:      earningsGrowth,
		CurrentRatio:        r.FinancialData.CurrentRatio.Raw,
		QuickRatio:          r.FinancialData.QuickRatio.Raw,
		// Yahoo returns debtToEquity on a percent-like base (e.g. 145 for 1.45x).
		DebtToEquity:       r.FinancialData.DebtToEquity.Raw / 100,
		RecommendationMean: r.FinancialData.RecommendationMean.Raw,
		RecommendationKey:  r.FinancialData.RecommendationKey,
		AnalystOpinions:    int(r.FinancialData.NumberOfAnalystOpinions.Raw),
		TargetLowPrice:     r.FinancialData.TargetLowPrice.Raw,
		TargetMeanPrice:    r.FinancialData.TargetMeanPrice.Raw,
		TargetHighPrice:    r.FinancialData.TargetHighPrice.Raw,
		Freshness:          domain.FreshnessLive,
	}, nil
}

func firstNonZero(values ...float64) float64 {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

func ratioInt64(numerator, denominator int64) float64 {
	if numerator == 0 || denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func deriveEnterpriseValue(marketCap, totalDebt, totalCash int64) int64 {
	if marketCap == 0 {
		return 0
	}
	return marketCap + totalDebt - totalCash
}

func normalizeMarketState(v string) domain.MarketState {
	switch strings.ToUpper(v) {
	case "REGULAR":
		return domain.MarketStateRegular
	case "PRE", "PREPRE":
		return domain.MarketStatePre
	case "POST", "POSTPOST":
		return domain.MarketStatePost
	case "CLOSED":
		return domain.MarketStateClosed
	default:
		return domain.MarketStateUnknown
	}
}
