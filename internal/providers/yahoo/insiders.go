package yahoo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

type insiderQuoteSummaryResponse struct {
	QuoteSummary struct {
		Result []insiderQuoteSummaryResult `json:"result"`
	} `json:"quoteSummary"`
}

type insiderQuoteSummaryResult struct {
	InsiderTransactions struct {
		Transactions []insiderTransactionItem `json:"transactions"`
	} `json:"insiderTransactions"`
	InsiderHolders struct {
		Holders []insiderHolderItem `json:"holders"`
	} `json:"insiderHolders"`
	NetSharePurchaseActivity insiderPurchaseActivityItem `json:"netSharePurchaseActivity"`
}

type insiderTransactionItem struct {
	FilerName       string         `json:"filerName"`
	FilerRelation   string         `json:"filerRelation"`
	MoneyText       string         `json:"moneyText"`
	Text            string         `json:"text"`
	TransactionText string         `json:"transactionText"`
	Ownership       string         `json:"ownership"`
	StartDate       timestampField `json:"startDate"`
	Shares          numberField    `json:"shares"`
	Value           numberField    `json:"value"`
}

type insiderHolderItem struct {
	Name                 string         `json:"name"`
	Relation             string         `json:"relation"`
	TransactionDesc      string         `json:"transactionDescription"`
	LatestTransDate      timestampField `json:"latestTransDate"`
	PositionDirect       numberField    `json:"positionDirect"`
	PositionDirectDate   timestampField `json:"positionDirectDate"`
	PositionIndirect     numberField    `json:"positionIndirect"`
	PositionIndirectDate timestampField `json:"positionIndirectDate"`
}

type insiderPurchaseActivityItem struct {
	Period                  string      `json:"period"`
	BuyInfoShares           numberField `json:"buyInfoShares"`
	BuyInfoCount            numberField `json:"buyInfoCount"`
	SellInfoShares          numberField `json:"sellInfoShares"`
	SellInfoCount           numberField `json:"sellInfoCount"`
	NetInfoShares           numberField `json:"netInfoShares"`
	NetInfoCount            numberField `json:"netInfoCount"`
	TotalInsiderShares      numberField `json:"totalInsiderShares"`
	NetPercentInsiderShares numberField `json:"netPercentInsiderShares"`
}

type timestampField struct {
	Raw int64
}

func (t *timestampField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Raw = 0
		return nil
	}

	var direct int64
	if err := json.Unmarshal(data, &direct); err == nil {
		t.Raw = direct
		return nil
	}

	var decimal float64
	if err := json.Unmarshal(data, &decimal); err == nil {
		t.Raw = int64(math.Round(decimal))
		return nil
	}

	var wrapped struct {
		Raw float64 `json:"raw"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	t.Raw = int64(math.Round(wrapped.Raw))
	return nil
}

func (p *Provider) GetInsiders(ctx context.Context, symbol string) (domain.InsiderSnapshot, error) {
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if normalizedSymbol == "" {
		return domain.InsiderSnapshot{}, fmt.Errorf("insider symbol is required")
	}

	var resp insiderQuoteSummaryResponse
	params := url.Values{}
	params.Set("modules", "insiderTransactions,insiderHolders,netSharePurchaseActivity")
	params.Set("corsDomain", "finance.yahoo.com")
	params.Set("formatted", "false")
	params.Set("symbol", normalizedSymbol)
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.quoteSummaryBase + url.PathEscape(normalizedSymbol),
		Params:   params,
		CacheKey: "insiders:" + normalizedSymbol,
		TTL:      30 * time.Minute,
		Auth:     authRequired,
	}, &resp)
	if err != nil {
		return domain.InsiderSnapshot{}, err
	}
	return normalizeInsiders(normalizedSymbol, resp)
}

func normalizeInsiders(symbol string, resp insiderQuoteSummaryResponse) (domain.InsiderSnapshot, error) {
	if len(resp.QuoteSummary.Result) == 0 {
		return domain.InsiderSnapshot{}, errors.New("insiders not found")
	}

	result := resp.QuoteSummary.Result[0]
	snapshot := domain.InsiderSnapshot{
		Symbol: strings.ToUpper(symbol),
		PurchaseActivity: domain.InsiderPurchaseActivity{
			Period:                  strings.TrimSpace(result.NetSharePurchaseActivity.Period),
			BuyShares:               roundedInt64(result.NetSharePurchaseActivity.BuyInfoShares.Raw),
			BuyTransactions:         roundedInt(result.NetSharePurchaseActivity.BuyInfoCount.Raw),
			SellShares:              roundedInt64(result.NetSharePurchaseActivity.SellInfoShares.Raw),
			SellTransactions:        roundedInt(result.NetSharePurchaseActivity.SellInfoCount.Raw),
			NetShares:               roundedInt64(result.NetSharePurchaseActivity.NetInfoShares.Raw),
			NetTransactions:         roundedInt(result.NetSharePurchaseActivity.NetInfoCount.Raw),
			TotalInsiderShares:      roundedInt64(result.NetSharePurchaseActivity.TotalInsiderShares.Raw),
			NetPercentInsiderShares: result.NetSharePurchaseActivity.NetPercentInsiderShares.Raw,
		},
		Freshness: domain.FreshnessLive,
		Provider:  "yahoo",
		UpdatedAt: time.Now(),
	}

	for _, item := range result.InsiderTransactions.Transactions {
		name := strings.TrimSpace(item.FilerName)
		text := strings.TrimSpace(item.TransactionText)
		if text == "" {
			text = strings.TrimSpace(item.Text)
		}
		if text == "" {
			text = strings.TrimSpace(item.MoneyText)
		}
		action := normalizeInsiderAction(text)
		if name == "" && text == "" {
			continue
		}
		if action != "Buy" && action != "Sale" {
			continue
		}
		snapshot.Transactions = append(snapshot.Transactions, domain.InsiderTransaction{
			Insider:   name,
			Relation:  strings.TrimSpace(item.FilerRelation),
			Ownership: normalizeOwnership(item.Ownership),
			Action:    action,
			Date:      timestampToTime(item.StartDate),
			Shares:    roundedInt64(item.Shares.Raw),
			Value:     roundedInt64(item.Value.Raw),
			Text:      text,
		})
	}
	sort.Slice(snapshot.Transactions, func(i, j int) bool {
		return snapshot.Transactions[i].Date.After(snapshot.Transactions[j].Date)
	})

	for _, item := range result.InsiderHolders.Holders {
		name := strings.TrimSpace(item.Name)
		relation := strings.TrimSpace(item.Relation)
		if name == "" && relation == "" {
			continue
		}
		snapshot.Roster = append(snapshot.Roster, domain.InsiderRosterMember{
			Name:                  name,
			Relation:              relation,
			LatestTransaction:     strings.TrimSpace(item.TransactionDesc),
			LatestTransactionAt:   timestampToTime(item.LatestTransDate),
			SharesOwnedDirectly:   roundedInt64(item.PositionDirect.Raw),
			PositionDirectDate:    timestampToTime(item.PositionDirectDate),
			SharesOwnedIndirectly: roundedInt64(item.PositionIndirect.Raw),
			PositionIndirectDate:  timestampToTime(item.PositionIndirectDate),
		})
	}
	sort.Slice(snapshot.Roster, func(i, j int) bool {
		left := snapshot.Roster[i]
		right := snapshot.Roster[j]
		if !left.LatestTransactionAt.Equal(right.LatestTransactionAt) {
			return left.LatestTransactionAt.After(right.LatestTransactionAt)
		}
		return left.SharesOwnedDirectly+left.SharesOwnedIndirectly > right.SharesOwnedDirectly+right.SharesOwnedIndirectly
	})

	return snapshot, nil
}

func roundedInt64(v float64) int64 {
	return int64(math.Round(v))
}

func roundedInt(v float64) int {
	return int(math.Round(v))
}

func normalizeOwnership(v string) string {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "D":
		return "Direct"
	case "I":
		return "Indirect"
	default:
		return strings.TrimSpace(v)
	}
}

func normalizeInsiderAction(text string) string {
	normalized := strings.ToLower(strings.TrimSpace(text))
	switch {
	case strings.Contains(normalized, "purchase"), strings.Contains(normalized, "buy"):
		return "Buy"
	case strings.Contains(normalized, "sale"), strings.Contains(normalized, "sell"):
		return "Sale"
	case strings.Contains(normalized, "exercise"):
		return "Exercise"
	case strings.Contains(normalized, "award"):
		return "Award"
	case strings.Contains(normalized, "gift"):
		return "Gift"
	case strings.Contains(normalized, "convert"):
		return "Conversion"
	default:
		return ""
	}
}

func timestampToTime(v timestampField) time.Time {
	if v.Raw <= 0 {
		return time.Time{}
	}
	return time.Unix(v.Raw, 0)
}
