package yahoo

import (
	"context"
	"strings"
	"testing"

	"net/http"
)

func TestGetEarningsUsesQuoteSummaryModules(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-123"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			if req.URL.Path != "/v10/finance/quoteSummary/AAPL" {
				t.Fatalf("unexpected path %q", req.URL.Path)
			}
			if got := req.URL.Query().Get("modules"); got != "price,calendarEvents,earningsHistory,earningsTrend" {
				t.Fatalf("unexpected modules %q", got)
			}
			body := []byte(`{
				"quoteSummary": {
					"result": [{
						"price": {
							"shortName": "Apple Inc."
						},
						"calendarEvents": {
							"earnings": {
								"earningsDate": [1777670400, 1777843200],
								"earningsAverage": { "raw": 1.62 },
								"earningsLow": { "raw": 1.55 },
								"earningsHigh": { "raw": 1.71 },
								"revenueAverage": { "raw": 95400000000 },
								"revenueLow": { "raw": 94100000000 },
								"revenueHigh": { "raw": 97800000000 }
							}
						},
						"earningsHistory": {
							"history": [
								{
									"quarter": { "fmt": "2025-12-31" },
									"epsEstimate": { "raw": 1.48 },
									"epsActual": { "raw": 1.52 },
									"epsDifference": { "raw": 0.04 },
									"surprisePercent": { "raw": 0.027 }
								}
							]
						},
						"earningsTrend": {
							"trend": [
								{
									"period": "0q",
									"earningsEstimate": {
										"avg": { "raw": 1.62 },
										"low": { "raw": 1.55 },
										"high": { "raw": 1.71 },
										"yearAgoEps": { "raw": 1.42 },
										"growth": { "raw": 0.14 },
										"numberOfAnalysts": { "raw": 28 }
									},
									"revenueEstimate": {
										"avg": { "raw": 95400000000 },
										"low": { "raw": 94100000000 },
										"high": { "raw": 97800000000 },
										"yearAgoRevenue": { "raw": 90300000000 },
										"growth": { "raw": 0.056 },
										"numberOfAnalysts": { "raw": 24 }
									}
								}
							]
						}
					}],
					"error": null
				}
			}`)
			return jsonResponse(req, http.StatusOK, body, "")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetEarnings(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if got.Symbol != "AAPL" || got.CompanyName != "Apple Inc." {
		t.Fatalf("unexpected snapshot %+v", got)
	}
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 earnings items, got %+v", got.Items)
	}
	if got.Items[0].Kind != "upcoming" {
		t.Fatalf("expected upcoming item first, got %+v", got.Items)
	}
	if got.Items[1].EPSActual != 1.52 {
		t.Fatalf("expected reported EPS actual, got %+v", got.Items[1])
	}
	if len(got.Estimates) != 1 || got.Estimates[0].AnalystCount != 28 {
		t.Fatalf("expected estimate trend, got %+v", got.Estimates)
	}
}
