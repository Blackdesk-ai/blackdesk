package yahoo

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNormalizeInsiders(t *testing.T) {
	resp := insiderQuoteSummaryResponse{}
	resp.QuoteSummary.Result = append(resp.QuoteSummary.Result, insiderQuoteSummaryResult{})
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity = insiderPurchaseActivityItem{
		Period: "6m",
	}
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.BuyInfoShares.Raw = 305864
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.BuyInfoCount.Raw = 3
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.SellInfoShares.Raw = 1606142
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.SellInfoCount.Raw = 12
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.NetInfoShares.Raw = -1300278
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.NetInfoCount.Raw = 15
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.TotalInsiderShares.Raw = 2100000
	resp.QuoteSummary.Result[0].NetSharePurchaseActivity.NetPercentInsiderShares.Raw = -0.382
	resp.QuoteSummary.Result[0].InsiderTransactions.Transactions = []insiderTransactionItem{
		{
			FilerName:       "SIEFFERT KRISTEN N",
			FilerRelation:   "President",
			TransactionText: "Sale at price 23.32 per share.",
			Ownership:       "D",
			StartDate:       timestampField{Raw: 1769904000},
			Shares:          numberField{Raw: 750},
			Value:           numberField{Raw: 17490},
		},
		{
			FilerName:     "THORNOCK TAI A",
			FilerRelation: "Officer",
			MoneyText:     "Purchase at price 24.58 per share.",
			Ownership:     "I",
			StartDate:     timestampField{Raw: 1736985600},
			Shares:        numberField{Raw: 1100},
			Value:         numberField{Raw: 27038},
		},
		{
			FilerName:       "DOE JANE",
			FilerRelation:   "Director",
			TransactionText: "Option exercise at price 10.00 per share.",
			Ownership:       "D",
			StartDate:       timestampField{Raw: 1734307200},
			Shares:          numberField{Raw: 500},
			Value:           numberField{Raw: 5000},
		},
	}
	resp.QuoteSummary.Result[0].InsiderHolders.Holders = []insiderHolderItem{
		{
			Name:               "SIEFFERT KRISTEN N",
			Relation:           "President",
			TransactionDesc:    "Sale",
			LatestTransDate:    timestampField{Raw: 1769904000},
			PositionDirect:     numberField{Raw: 2100000},
			PositionDirectDate: timestampField{Raw: 1769904000},
		},
	}

	got, err := normalizeInsiders("foa", resp)
	if err != nil {
		t.Fatal(err)
	}
	if got.Symbol != "FOA" {
		t.Fatalf("expected FOA symbol, got %+v", got)
	}
	if got.PurchaseActivity.BuyTransactions != 3 || got.PurchaseActivity.NetShares != -1300278 {
		t.Fatalf("unexpected purchase activity %+v", got.PurchaseActivity)
	}
	if len(got.Transactions) != 2 {
		t.Fatalf("expected only buy or sale insider transactions, got %+v", got.Transactions)
	}
	if got.Transactions[0].Insider != "SIEFFERT KRISTEN N" || got.Transactions[0].Ownership != "Direct" || got.Transactions[0].Action != "Sale" {
		t.Fatalf("unexpected normalized transaction %+v", got.Transactions[0])
	}
	if got.Transactions[1].Ownership != "Indirect" || got.Transactions[1].Action != "Buy" || !strings.Contains(got.Transactions[1].Text, "24.58") {
		t.Fatalf("expected fallback to moneyText and indirect ownership, got %+v", got.Transactions[1])
	}
	if len(got.Roster) != 1 || got.Roster[0].SharesOwnedDirectly != 2100000 {
		t.Fatalf("unexpected insider roster %+v", got.Roster)
	}
}

func TestGetInsidersUsesQuoteSummaryModulesAndCrumb(t *testing.T) {
	ctx := context.Background()
	var sawCrumb atomic.Bool
	var sawModules atomic.Bool

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-insiders"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			if req.URL.Query().Get("crumb") == "crumb-insiders" {
				sawCrumb.Store(true)
			}
			if strings.Contains(req.URL.Query().Get("modules"), "insiderTransactions") &&
				strings.Contains(req.URL.Query().Get("modules"), "insiderHolders") &&
				strings.Contains(req.URL.Query().Get("modules"), "netSharePurchaseActivity") {
				sawModules.Store(true)
			}
			body := []byte(`{
				"quoteSummary": {
					"result": [
						{
							"netSharePurchaseActivity": {
								"period": "6m",
								"buyInfoShares": {"raw": 305864},
								"buyInfoCount": {"raw": 3},
								"sellInfoShares": {"raw": 1606142},
								"sellInfoCount": {"raw": 12},
								"netInfoShares": {"raw": -1300278},
								"netInfoCount": {"raw": 15},
								"totalInsiderShares": {"raw": 2100000},
								"netPercentInsiderShares": {"raw": -0.382}
							},
							"insiderTransactions": {
								"transactions": [
									{
										"filerName": "SIEFFERT KRISTEN N",
										"filerRelation": "President",
										"text": "Sale at price 23.32 per share.",
										"ownership": "D",
										"startDate": {"raw": 1769904000},
										"shares": {"raw": 750},
										"value": {"raw": 17490}
									}
								]
							},
							"insiderHolders": {
								"holders": [
									{
										"name": "SIEFFERT KRISTEN N",
										"relation": "President",
										"transactionDescription": "Sale",
										"latestTransDate": {"raw": 1769904000},
										"positionDirect": {"raw": 2100000},
										"positionDirectDate": {"raw": 1769904000},
										"positionIndirect": {"raw": 0},
										"positionIndirectDate": {"raw": 0}
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
	got, err := p.GetInsiders(ctx, "FOA")
	if err != nil {
		t.Fatal(err)
	}
	if !sawCrumb.Load() {
		t.Fatal("expected quoteSummary insiders request with crumb")
	}
	if !sawModules.Load() {
		t.Fatal("expected insider modules in quoteSummary request")
	}
	if got.Symbol != "FOA" || len(got.Transactions) != 1 || len(got.Roster) != 1 {
		t.Fatalf("unexpected insiders snapshot %+v", got)
	}
}
