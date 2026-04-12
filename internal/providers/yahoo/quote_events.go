package yahoo

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

type quoteEventsResponse struct {
	QuoteSummary struct {
		Result []struct {
			Price struct {
				ShortName string `json:"shortName"`
				LongName  string `json:"longName"`
			} `json:"price"`
			CalendarEvents struct {
				DividendDate   timestampField `json:"dividendDate"`
				ExDividendDate timestampField `json:"exDividendDate"`
				Earnings       struct {
					EarningsDate    []timestampField `json:"earningsDate"`
					EarningsAverage numberField      `json:"earningsAverage"`
					EarningsLow     numberField      `json:"earningsLow"`
					EarningsHigh    numberField      `json:"earningsHigh"`
					RevenueAverage  numberField      `json:"revenueAverage"`
					RevenueLow      numberField      `json:"revenueLow"`
					RevenueHigh     numberField      `json:"revenueHigh"`
				} `json:"earnings"`
			} `json:"calendarEvents"`
			EarningsHistory struct {
				History []struct {
					Quarter struct {
						Fmt string `json:"fmt"`
					} `json:"quarter"`
					EPSEstimate     numberField `json:"epsEstimate"`
					EPSActual       numberField `json:"epsActual"`
					EPSDifference   numberField `json:"epsDifference"`
					SurprisePercent numberField `json:"surprisePercent"`
				} `json:"history"`
			} `json:"earningsHistory"`
			EarningsTrend struct {
				Trend []struct {
					Period           string                  `json:"period"`
					EarningsEstimate quoteEventEstimateBlock `json:"earningsEstimate"`
					RevenueEstimate  quoteEventEstimateBlock `json:"revenueEstimate"`
				} `json:"trend"`
			} `json:"earningsTrend"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

type quoteEventEstimateBlock struct {
	Avg              numberField `json:"avg"`
	Low              numberField `json:"low"`
	High             numberField `json:"high"`
	YearAgoEps       numberField `json:"yearAgoEps"`
	YearAgoRevenue   numberField `json:"yearAgoRevenue"`
	Growth           numberField `json:"growth"`
	NumberOfAnalysts numberField `json:"numberOfAnalysts"`
}

func (p *Provider) GetEarnings(ctx context.Context, symbol string) (domain.EarningsSnapshot, error) {
	resp, err := p.fetchQuoteEvents(ctx, symbol, "price,calendarEvents,earningsHistory,earningsTrend", "earnings")
	if err != nil {
		return domain.EarningsSnapshot{}, err
	}
	return normalizeEarningsSnapshot(symbol, resp)
}

func (p *Provider) fetchQuoteEvents(ctx context.Context, symbol, modules, cachePrefix string) (quoteEventsResponse, error) {
	var resp quoteEventsResponse
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	params := url.Values{}
	params.Set("modules", modules)
	params.Set("corsDomain", "finance.yahoo.com")
	params.Set("formatted", "false")
	params.Set("symbol", normalizedSymbol)
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.quoteSummaryBase + url.PathEscape(normalizedSymbol),
		Params:   params,
		CacheKey: cachePrefix + ":" + normalizedSymbol,
		TTL:      30 * time.Minute,
		Auth:     authRequired,
	}, &resp)
	if err != nil {
		return quoteEventsResponse{}, err
	}
	return resp, nil
}

func normalizeEarningsSnapshot(symbol string, resp quoteEventsResponse) (domain.EarningsSnapshot, error) {
	result, err := firstQuoteEventsResult(resp)
	if err != nil {
		return domain.EarningsSnapshot{}, err
	}

	snapshot := domain.EarningsSnapshot{
		Symbol:      strings.ToUpper(strings.TrimSpace(symbol)),
		CompanyName: firstNonEmptyString(result.Price.ShortName, result.Price.LongName),
		Freshness:   domain.FreshnessLive,
		Provider:    "yahoo",
		UpdatedAt:   time.Now(),
	}

	upcomingDates := normalizeTimestampList(result.CalendarEvents.Earnings.EarningsDate)
	if len(upcomingDates) > 0 {
		item := domain.EarningsItem{
			Kind:           "upcoming",
			Title:          "Next earnings",
			WindowStart:    upcomingDates[0],
			WindowEnd:      upcomingDates[len(upcomingDates)-1],
			EPSEstimate:    result.CalendarEvents.Earnings.EarningsAverage.Raw,
			EPSLow:         result.CalendarEvents.Earnings.EarningsLow.Raw,
			EPSHigh:        result.CalendarEvents.Earnings.EarningsHigh.Raw,
			RevenueAverage: result.CalendarEvents.Earnings.RevenueAverage.Raw,
			RevenueLow:     result.CalendarEvents.Earnings.RevenueLow.Raw,
			RevenueHigh:    result.CalendarEvents.Earnings.RevenueHigh.Raw,
		}
		snapshot.Items = append(snapshot.Items, item)
	}

	for _, item := range result.EarningsHistory.History {
		quarterEnd := parseYahooDate(item.Quarter.Fmt)
		snapshot.Items = append(snapshot.Items, domain.EarningsItem{
			Kind:            "reported",
			Title:           "Reported quarter",
			QuarterEnd:      quarterEnd,
			EPSEstimate:     item.EPSEstimate.Raw,
			EPSActual:       item.EPSActual.Raw,
			EPSDifference:   item.EPSDifference.Raw,
			SurprisePercent: item.SurprisePercent.Raw,
		})
	}

	sort.SliceStable(snapshot.Items, func(i, j int) bool {
		left := earningsItemSortTime(snapshot.Items[i])
		right := earningsItemSortTime(snapshot.Items[j])
		if left.Equal(right) {
			return snapshot.Items[i].Kind < snapshot.Items[j].Kind
		}
		return left.After(right)
	})

	for _, item := range result.EarningsTrend.Trend {
		snapshot.Estimates = append(snapshot.Estimates, domain.EarningsEstimate{
			Period:          strings.TrimSpace(item.Period),
			EPSAverage:      item.EarningsEstimate.Avg.Raw,
			EPSLow:          item.EarningsEstimate.Low.Raw,
			EPSHigh:         item.EarningsEstimate.High.Raw,
			EPSYearAgo:      item.EarningsEstimate.YearAgoEps.Raw,
			EPSGrowth:       item.EarningsEstimate.Growth.Raw,
			RevenueAverage:  item.RevenueEstimate.Avg.Raw,
			RevenueLow:      item.RevenueEstimate.Low.Raw,
			RevenueHigh:     item.RevenueEstimate.High.Raw,
			RevenueYearAgo:  item.RevenueEstimate.YearAgoRevenue.Raw,
			RevenueGrowth:   item.RevenueEstimate.Growth.Raw,
			AnalystCount:    int(math.Round(item.EarningsEstimate.NumberOfAnalysts.Raw)),
			RevenueAnalysts: int(math.Round(item.RevenueEstimate.NumberOfAnalysts.Raw)),
		})
	}
	sort.SliceStable(snapshot.Estimates, func(i, j int) bool {
		return earningsPeriodOrder(snapshot.Estimates[i].Period) < earningsPeriodOrder(snapshot.Estimates[j].Period)
	})

	if snapshot.CompanyName == "" {
		snapshot.CompanyName = snapshot.Symbol
	}
	return snapshot, nil
}

func firstQuoteEventsResult(resp quoteEventsResponse) (*struct {
	Price struct {
		ShortName string `json:"shortName"`
		LongName  string `json:"longName"`
	} `json:"price"`
	CalendarEvents struct {
		DividendDate   timestampField `json:"dividendDate"`
		ExDividendDate timestampField `json:"exDividendDate"`
		Earnings       struct {
			EarningsDate    []timestampField `json:"earningsDate"`
			EarningsAverage numberField      `json:"earningsAverage"`
			EarningsLow     numberField      `json:"earningsLow"`
			EarningsHigh    numberField      `json:"earningsHigh"`
			RevenueAverage  numberField      `json:"revenueAverage"`
			RevenueLow      numberField      `json:"revenueLow"`
			RevenueHigh     numberField      `json:"revenueHigh"`
		} `json:"earnings"`
	} `json:"calendarEvents"`
	EarningsHistory struct {
		History []struct {
			Quarter struct {
				Fmt string `json:"fmt"`
			} `json:"quarter"`
			EPSEstimate     numberField `json:"epsEstimate"`
			EPSActual       numberField `json:"epsActual"`
			EPSDifference   numberField `json:"epsDifference"`
			SurprisePercent numberField `json:"surprisePercent"`
		} `json:"history"`
	} `json:"earningsHistory"`
	EarningsTrend struct {
		Trend []struct {
			Period           string                  `json:"period"`
			EarningsEstimate quoteEventEstimateBlock `json:"earningsEstimate"`
			RevenueEstimate  quoteEventEstimateBlock `json:"revenueEstimate"`
		} `json:"trend"`
	} `json:"earningsTrend"`
}, error) {
	if len(resp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("yahoo quote events unavailable")
	}
	return &resp.QuoteSummary.Result[0], nil
}

func normalizeTimestampList(items []timestampField) []time.Time {
	out := make([]time.Time, 0, len(items))
	for _, item := range items {
		if item.Raw <= 0 {
			continue
		}
		out = append(out, time.Unix(item.Raw, 0).UTC())
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Before(out[j])
	})
	return out
}

func earningsItemSortTime(item domain.EarningsItem) time.Time {
	if !item.WindowStart.IsZero() {
		return item.WindowStart
	}
	return item.QuarterEnd
}

func earningsPeriodOrder(period string) int {
	switch strings.TrimSpace(period) {
	case "0q":
		return 0
	case "+1q":
		return 1
	case "0y":
		return 2
	case "+1y":
		return 3
	default:
		return 10
	}
}

func parseYahooDate(raw string) time.Time {
	ts, _ := time.Parse("2006-01-02", strings.TrimSpace(raw))
	return ts
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
