package yahoo

import (
	"context"
	"net/http"
	"testing"
)

func TestScreenersReturnsCatalog(t *testing.T) {
	p := New(Config{})

	got := p.Screeners()
	if len(got) < 8 {
		t.Fatalf("expected screener catalog, got %d entries", len(got))
	}
	if got[0].ID == "" || got[0].Category == "" {
		t.Fatalf("unexpected first screener %+v", got[0])
	}
}

func TestGetScreenerParsesEquityResults(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "query2.finance.yahoo.com" || req.URL.Path != "/v1/finance/screener/predefined/saved" {
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		if got := req.URL.Query().Get("scrIds"); got != "most_actives" {
			t.Fatalf("unexpected screener id %q", got)
		}
		if got := req.URL.Query().Get("count"); got != "25" {
			t.Fatalf("expected default screener count 25, got %q", got)
		}
		body := []byte(`{
			"finance": {
				"result": [{
					"title": "Most Actives",
					"description": "Discover the most traded equities in the trading day.",
					"canonicalName": "MOST_ACTIVES",
					"total": 213,
					"criteriaMeta": {
						"sortField": "dayvolume",
						"sortType": "DESC",
						"quoteType": "EQUITY"
					},
					"quotes": [{
						"symbol": "NVDA",
						"shortName": "NVIDIA Corporation",
						"fullExchangeName": "NasdaqGS",
						"exchange": "NMS",
						"quoteType": "EQUITY",
						"typeDisp": "Equity",
						"currency": "USD",
						"marketState": "POST",
						"regularMarketPrice": 177.64,
						"regularMarketChange": 0.25,
						"regularMarketChangePercent": 0.1409,
						"regularMarketVolume": 103113676,
						"averageDailyVolume3Month": 181124446,
						"marketCap": 4317540253696,
						"regularMarketTime": 1775505600,
						"trailingPE": 36.25,
						"forwardPE": 15.98,
						"priceToBook": 27.44,
						"fiftyTwoWeekChangePercent": 81.67,
						"dividendYield": 0.02,
						"averageAnalystRating": "1.3 - Strong Buy"
					}]
				}],
				"error": null
			}
		}`)
		return jsonResponse(req, http.StatusOK, body, "")
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetScreener(ctx, "most_actives", 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Definition.ID != "most_actives" || got.Definition.Category != "Market Movers" {
		t.Fatalf("unexpected screener definition %+v", got.Definition)
	}
	if got.Total != 213 || got.SortField != "dayvolume" {
		t.Fatalf("unexpected screener metadata %+v", got)
	}
	if len(got.Criteria) == 0 {
		t.Fatal("expected normalized screener criteria")
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 screener item, got %d", len(got.Items))
	}
	if got.Items[0].Symbol != "NVDA" || got.Items[0].Name != "NVIDIA Corporation" {
		t.Fatalf("unexpected screener item %+v", got.Items[0])
	}
	if len(got.Items[0].Metrics) == 0 || got.Items[0].Metrics[0].Key != "avg_3m_volume" {
		t.Fatalf("expected derived metrics, got %+v", got.Items[0].Metrics)
	}
	foundRV := false
	for _, metric := range got.Items[0].Metrics {
		if metric.Key == "relative_volume" {
			foundRV = true
			if metric.Value != "0.57x" {
				t.Fatalf("expected RV metric to be 0.57x, got %+v", metric)
			}
		}
	}
	if !foundRV {
		t.Fatalf("expected RV metric in screener item, got %+v", got.Items[0].Metrics)
	}
}

func TestGetScreenerRejectsUnsupportedFundCatalogEntries(t *testing.T) {
	ctx := context.Background()

	p := newTestProvider("https://query1.finance.yahoo.test", newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return textResponse(req, http.StatusNotFound, "not found"), nil
	}))
	_, err := p.GetScreener(ctx, "top_mutual_funds", 25)
	if err == nil {
		t.Fatal("expected removed fund screener to be unsupported")
	}
}
