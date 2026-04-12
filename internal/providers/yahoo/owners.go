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

type ownersQuoteSummaryResponse struct {
	QuoteSummary struct {
		Result []ownersQuoteSummaryResult `json:"result"`
	} `json:"quoteSummary"`
}

type ownersQuoteSummaryResult struct {
	Price struct {
		ShortName string `json:"shortName"`
		LongName  string `json:"longName"`
	} `json:"price"`
	MajorHoldersBreakdown struct {
		InsidersPercentHeld          numberField `json:"insidersPercentHeld"`
		InstitutionsPercentHeld      numberField `json:"institutionsPercentHeld"`
		InstitutionsFloatPercentHeld numberField `json:"institutionsFloatPercentHeld"`
		InstitutionsCount            numberField `json:"institutionsCount"`
	} `json:"majorHoldersBreakdown"`
	InstitutionOwnership struct {
		OwnershipList []ownerHolderItem `json:"ownershipList"`
	} `json:"institutionOwnership"`
	FundOwnership struct {
		OwnershipList []ownerHolderItem `json:"ownershipList"`
	} `json:"fundOwnership"`
}

type ownerHolderItem struct {
	Organization string         `json:"organization"`
	ReportDate   timestampField `json:"reportDate"`
	Position     numberField    `json:"position"`
	Value        numberField    `json:"value"`
	PctHeld      numberField    `json:"pctHeld"`
}

func (p *Provider) GetOwners(ctx context.Context, symbol string) (domain.OwnershipSnapshot, error) {
	var resp ownersQuoteSummaryResponse
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	params := url.Values{}
	params.Set("modules", "price,majorHoldersBreakdown,institutionOwnership,fundOwnership")
	params.Set("corsDomain", "finance.yahoo.com")
	params.Set("formatted", "false")
	params.Set("symbol", normalizedSymbol)
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.quoteSummaryBase + url.PathEscape(normalizedSymbol),
		Params:   params,
		CacheKey: "owners:" + normalizedSymbol,
		TTL:      30 * time.Minute,
		Auth:     authRequired,
	}, &resp)
	if err != nil {
		return domain.OwnershipSnapshot{}, err
	}
	return normalizeOwners(normalizedSymbol, resp)
}

func normalizeOwners(symbol string, resp ownersQuoteSummaryResponse) (domain.OwnershipSnapshot, error) {
	if len(resp.QuoteSummary.Result) == 0 {
		return domain.OwnershipSnapshot{}, fmt.Errorf("yahoo owners unavailable")
	}
	result := resp.QuoteSummary.Result[0]
	snapshot := domain.OwnershipSnapshot{
		Symbol:      symbol,
		CompanyName: firstNonEmptyString(result.Price.ShortName, result.Price.LongName),
		Summary: domain.OwnershipSummary{
			InsidersPercentHeld:          result.MajorHoldersBreakdown.InsidersPercentHeld.Raw,
			InstitutionsPercentHeld:      result.MajorHoldersBreakdown.InstitutionsPercentHeld.Raw,
			InstitutionsFloatPercentHeld: result.MajorHoldersBreakdown.InstitutionsFloatPercentHeld.Raw,
			InstitutionsHoldingCount:     int(math.Round(result.MajorHoldersBreakdown.InstitutionsCount.Raw)),
		},
		Freshness: domain.FreshnessLive,
		Provider:  "yahoo",
		UpdatedAt: time.Now(),
	}

	for _, item := range result.InstitutionOwnership.OwnershipList {
		holder := normalizeOwnershipHolder(item)
		if holder.Name == "" {
			continue
		}
		snapshot.Institutions = append(snapshot.Institutions, holder)
	}
	sortOwnershipHolders(snapshot.Institutions)

	for _, item := range result.FundOwnership.OwnershipList {
		holder := normalizeOwnershipHolder(item)
		if holder.Name == "" {
			continue
		}
		snapshot.Funds = append(snapshot.Funds, holder)
	}
	sortOwnershipHolders(snapshot.Funds)

	if snapshot.CompanyName == "" {
		snapshot.CompanyName = snapshot.Symbol
	}
	return snapshot, nil
}

func normalizeOwnershipHolder(item ownerHolderItem) domain.OwnershipHolder {
	return domain.OwnershipHolder{
		Name:        strings.TrimSpace(item.Organization),
		Shares:      roundedInt64(item.Position.Raw),
		Value:       roundedInt64(item.Value.Raw),
		PercentHeld: item.PctHeld.Raw,
		ReportDate:  timestampToTime(item.ReportDate),
	}
}

func sortOwnershipHolders(items []domain.OwnershipHolder) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].PercentHeld != items[j].PercentHeld {
			return items[i].PercentHeld > items[j].PercentHeld
		}
		if items[i].Value != items[j].Value {
			return items[i].Value > items[j].Value
		}
		return items[i].Name < items[j].Name
	})
}
