package yahoo

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNormalizeOwnersSortsByPercentHeld(t *testing.T) {
	resp := ownersQuoteSummaryResponse{}
	resp.QuoteSummary.Result = append(resp.QuoteSummary.Result, ownersQuoteSummaryResult{})
	resp.QuoteSummary.Result[0].Price.ShortName = "Apple Inc."
	resp.QuoteSummary.Result[0].MajorHoldersBreakdown.InsidersPercentHeld.Raw = 0.021
	resp.QuoteSummary.Result[0].MajorHoldersBreakdown.InstitutionsPercentHeld.Raw = 0.634
	resp.QuoteSummary.Result[0].MajorHoldersBreakdown.InstitutionsFloatPercentHeld.Raw = 0.652
	resp.QuoteSummary.Result[0].MajorHoldersBreakdown.InstitutionsCount.Raw = 5123
	resp.QuoteSummary.Result[0].InstitutionOwnership.OwnershipList = []ownerHolderItem{
		{
			Organization: "Vanguard Group",
			ReportDate:   timestampField{Raw: 1761955200},
			Position:     numberField{Raw: 1_401_231_221},
			Value:        numberField{Raw: 312_000_000_000},
			PctHeld:      numberField{Raw: 0.091},
		},
		{
			Organization: "BlackRock",
			ReportDate:   timestampField{Raw: 1761955200},
			Position:     numberField{Raw: 1_111_000_000},
			Value:        numberField{Raw: 248_000_000_000},
			PctHeld:      numberField{Raw: 0.072},
		},
	}
	resp.QuoteSummary.Result[0].FundOwnership.OwnershipList = []ownerHolderItem{
		{
			Organization: "Vanguard Total Stock Market Index Fund",
			ReportDate:   timestampField{Raw: 1761955200},
			Position:     numberField{Raw: 410_000_000},
			Value:        numberField{Raw: 91_000_000_000},
			PctHeld:      numberField{Raw: 0.027},
		},
	}

	got, err := normalizeOwners("AAPL", resp)
	if err != nil {
		t.Fatalf("normalizeOwners error: %v", err)
	}
	if got.CompanyName != "Apple Inc." {
		t.Fatalf("expected company name, got %+v", got)
	}
	if got.Summary.InstitutionsHoldingCount != 5123 || got.Summary.InstitutionsPercentHeld != 0.634 {
		t.Fatalf("unexpected owners summary %+v", got.Summary)
	}
	if len(got.Institutions) != 2 || got.Institutions[0].Name != "Vanguard Group" {
		t.Fatalf("expected sorted institutions, got %+v", got.Institutions)
	}
	if len(got.Funds) != 1 || got.Funds[0].Shares != 410_000_000 {
		t.Fatalf("unexpected funds list %+v", got.Funds)
	}
}

func TestGetOwnersUsesQuoteSummaryModulesAndCrumb(t *testing.T) {
	ctx := context.Background()
	var sawCrumb atomic.Bool
	var sawModules atomic.Bool

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-owners"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			if req.URL.Query().Get("crumb") == "crumb-owners" {
				sawCrumb.Store(true)
			}
			modules := req.URL.Query().Get("modules")
			if strings.Contains(modules, "majorHoldersBreakdown") &&
				strings.Contains(modules, "institutionOwnership") &&
				strings.Contains(modules, "fundOwnership") {
				sawModules.Store(true)
			}
			body := []byte(`{
				"quoteSummary": {
					"result": [
						{
							"price": {
								"shortName": "Apple Inc."
							},
							"majorHoldersBreakdown": {
								"insidersPercentHeld": {"raw": 0.021},
								"institutionsPercentHeld": {"raw": 0.634},
								"institutionsFloatPercentHeld": {"raw": 0.652},
								"institutionsCount": {"raw": 5123}
							},
							"institutionOwnership": {
								"ownershipList": [
									{
										"organization": "Vanguard Group",
										"reportDate": {"raw": 1761955200},
										"position": {"raw": 1401231221},
										"value": {"raw": 312000000000},
										"pctHeld": {"raw": 0.091}
									}
								]
							},
							"fundOwnership": {
								"ownershipList": [
									{
										"organization": "Vanguard Total Stock Market Index Fund",
										"reportDate": {"raw": 1761955200},
										"position": {"raw": 410000000},
										"value": {"raw": 91000000000},
										"pctHeld": {"raw": 0.027}
									}
								]
							}
						}
					]
				}
			}`)
			return jsonResponse(req, http.StatusOK, body, "")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetOwners(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if !sawCrumb.Load() {
		t.Fatal("expected quoteSummary owners request with crumb")
	}
	if !sawModules.Load() {
		t.Fatal("expected owners modules in quoteSummary request")
	}
	if got.Symbol != "AAPL" || len(got.Institutions) != 1 || len(got.Funds) != 1 {
		t.Fatalf("unexpected owners snapshot %+v", got)
	}
}
