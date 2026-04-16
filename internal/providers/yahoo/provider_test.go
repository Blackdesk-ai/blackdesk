package yahoo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"blackdesk/internal/domain"
)

func TestGetFundamentalsUsesQuoteSummaryWithoutCrumbWhenAvailable(t *testing.T) {
	ctx := context.Background()
	var sawCrumb atomic.Bool

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-123"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			if req.URL.Query().Get("crumb") == "crumb-123" {
				sawCrumb.Store(true)
			}
			return jsonFixtureResponse(t, req, "quote_summary_aapl.json")
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			types := req.URL.Query().Get("type")
			switch {
			case types == "trailingPegRatio":
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[{"trailingPegRatio":[{"reportedValue":{"raw":1.88}}]}],"error":null}}`), "")
			case strings.Contains(types, "quarterlyEBIT"), strings.Contains(types, "annualEBIT"), strings.Contains(types, "quarterlyInvestedCapital"), strings.Contains(types, "annualInvestedCapital"), strings.Contains(types, "quarterlyOperatingCashFlow"), strings.Contains(types, "annualOperatingCashFlow"):
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			default:
				t.Fatalf("unexpected timeseries type %q", types)
			}
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		return nil, nil
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetFundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if sawCrumb.Load() {
		t.Fatal("expected quoteSummary fundamentals request to avoid crumb when not required")
	}
	if got.Symbol != "AAPL" || got.Sector != "Technology" {
		t.Fatalf("unexpected fundamentals %+v", got)
	}
	if got.PEGRatio != 1.88 {
		t.Fatalf("expected supplemental trailing peg ratio, got %+v", got)
	}
}

func TestGetFundamentalsRetriesWithCrumbAfterForbidden(t *testing.T) {
	ctx := context.Background()
	var quoteSummaryCalls atomic.Int32
	var sawCrumb atomic.Bool

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-123"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			call := quoteSummaryCalls.Add(1)
			if call == 1 && req.URL.Query().Get("crumb") == "" {
				return textResponse(req, http.StatusForbidden, "forbidden"), nil
			}
			if req.URL.Query().Get("crumb") == "crumb-123" {
				sawCrumb.Store(true)
			}
			return jsonFixtureResponse(t, req, "quote_summary_aapl.json")
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		return nil, nil
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetFundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if quoteSummaryCalls.Load() != 2 {
		t.Fatalf("expected 2 quoteSummary calls, got %d", quoteSummaryCalls.Load())
	}
	if !sawCrumb.Load() {
		t.Fatal("expected fundamentals retry with crumb after forbidden")
	}
	if got.Symbol != "AAPL" {
		t.Fatalf("unexpected fundamentals %+v", got)
	}
}

func TestGetFundamentalsFallsBackToTimeseriesPEG(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-123"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			return jsonFixtureResponse(t, req, "quote_summary_aapl.json")
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			types := req.URL.Query().Get("type")
			switch {
			case types == "trailingPegRatio":
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[{"trailingPegRatio":[{"reportedValue":{"raw":2.41}}]}],"error":null}}`), "")
			case strings.Contains(types, "quarterlyEBIT"), strings.Contains(types, "annualEBIT"), strings.Contains(types, "quarterlyInvestedCapital"), strings.Contains(types, "annualInvestedCapital"), strings.Contains(types, "quarterlyOperatingCashFlow"), strings.Contains(types, "annualOperatingCashFlow"):
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			default:
				t.Fatalf("unexpected timeseries type %q", types)
			}
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		return nil, nil
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetFundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if got.PEGRatio != 2.41 {
		t.Fatalf("expected timeseries peg ratio, got %+v", got)
	}
}

func TestGetFundamentalsDerivesROICFromSupplementalTimeseries(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-123"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			return jsonFixtureResponse(t, req, "quote_summary_aapl.json")
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			types := req.URL.Query().Get("type")
			switch {
			case types == "trailingPegRatio":
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[{"trailingPegRatio":[{"reportedValue":{"raw":2.41}}]}],"error":null}}`), "")
			case strings.Contains(types, "quarterlyEBIT"):
				body := []byte(`{
					"timeseries": {
						"result": [
							{
								"meta": { "type": ["quarterlyTotalRevenue"] },
								"quarterlyTotalRevenue": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 55.0 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 54.0 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 53.0 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 54.0 } }
								]
							},
							{
								"meta": { "type": ["quarterlyGrossProfit"] },
								"quarterlyGrossProfit": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 39.0 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 38.2 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 37.3 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 39.0 } }
								]
							},
							{
								"meta": { "type": ["quarterlyEBIT"] },
								"quarterlyEBIT": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 32.5 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 33.1 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 31.8 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 33.0 } }
								]
							},
							{
								"meta": { "type": ["quarterlyOperatingIncome"] },
								"quarterlyOperatingIncome": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 30.0 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 31.0 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 29.5 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 30.0 } }
								]
							},
							{
								"meta": { "type": ["quarterlyPretaxIncome"] },
								"quarterlyPretaxIncome": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 35.0 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 35.2 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 35.8 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 35.45 } }
								]
							},
							{
								"meta": { "type": ["quarterlyTaxProvision"] },
								"quarterlyTaxProvision": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 5.3 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 5.4 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 5.2 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 5.48 } }
								]
							},
							{
								"meta": { "type": ["quarterlyNetIncome"] },
								"quarterlyNetIncome": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 28.0 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 27.2 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 26.4 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 26.4 } }
								]
							}
						],
						"error": null
					}
				}`)
				return jsonResponse(req, http.StatusOK, body, "")
			case strings.Contains(types, "annualEBIT"):
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			case strings.Contains(types, "quarterlyInvestedCapital"):
				body := []byte(`{
					"timeseries": {
						"result": [
							{
								"meta": { "type": ["quarterlyInvestedCapital"] },
								"quarterlyInvestedCapital": [
									{ "asOfDate": "2024-12-31", "reportedValue": { "raw": 87.79 } },
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 106.14 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 102.0 } }
								]
							},
							{
								"meta": { "type": ["quarterlyStockholdersEquity"] },
								"quarterlyStockholdersEquity": [
									{ "asOfDate": "2024-12-31", "reportedValue": { "raw": 79.33 } },
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 157.29 } }
								]
							},
							{
								"meta": { "type": ["quarterlyTotalAssets"] },
								"quarterlyTotalAssets": [
									{ "asOfDate": "2024-12-31", "reportedValue": { "raw": 175.0 } },
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 249.0 } }
								]
							}
						],
						"error": null
					}
				}`)
				return jsonResponse(req, http.StatusOK, body, "")
			case strings.Contains(types, "annualInvestedCapital"):
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			case strings.Contains(types, "quarterlyOperatingCashFlow"), strings.Contains(types, "annualOperatingCashFlow"):
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			default:
				t.Fatalf("unexpected timeseries type %q", types)
			}
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		return nil, nil
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetFundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if got.InvestedCapital != 106 {
		t.Fatalf("expected invested capital from supplemental timeseries, got %+v", got)
	}
	wantROIC := ((30.0 + 31.0 + 29.5 + 30.0) * (1 - ((5.3 + 5.4 + 5.2 + 5.48) / (35.0 + 35.2 + 35.8 + 35.45)))) / 87.79
	if math.Abs(got.ReturnOnInvestedCapital-wantROIC) > 0.0001 {
		t.Fatalf("expected derived roic %.6f, got %+v", wantROIC, got)
	}
	wantGrossMargin := (39.0 + 38.2 + 37.3 + 39.0) / (55.0 + 54.0 + 53.0 + 54.0)
	if math.Abs(got.GrossMargins-wantGrossMargin) > 0.0001 {
		t.Fatalf("expected derived gross margin %.6f, got %+v", wantGrossMargin, got)
	}
	wantOperatingMargin := (30.0 + 31.0 + 29.5 + 30.0) / (55.0 + 54.0 + 53.0 + 54.0)
	if math.Abs(got.OperatingMargins-wantOperatingMargin) > 0.0001 {
		t.Fatalf("expected derived operating margin %.6f, got %+v", wantOperatingMargin, got)
	}
	wantProfitMargin := (28.0 + 27.2 + 26.4 + 26.4) / (55.0 + 54.0 + 53.0 + 54.0)
	if math.Abs(got.ProfitMargins-wantProfitMargin) > 0.0001 {
		t.Fatalf("expected derived profit margin %.6f, got %+v", wantProfitMargin, got)
	}
	wantROE := (28.0 + 27.2 + 26.4 + 26.4) / ((157.29 + 79.33) / 2)
	if math.Abs(got.ReturnOnEquity-wantROE) > 0.0001 {
		t.Fatalf("expected derived roe %.6f, got %+v", wantROE, got)
	}
	wantROA := (28.0 + 27.2 + 26.4 + 26.4) / ((249.0 + 175.0) / 2)
	if math.Abs(got.ReturnOnAssets-wantROA) > 0.0001 {
		t.Fatalf("expected derived roa %.6f, got %+v", wantROA, got)
	}
}

func TestGetFundamentalsDerivesCashFlowFromSupplementalTimeseries(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-123"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			body := readFixture(t, "quote_summary_aapl.json")
			var parsed map[string]any
			if err := json.Unmarshal(body, &parsed); err != nil {
				t.Fatal(err)
			}
			result := parsed["quoteSummary"].(map[string]any)["result"].([]any)[0].(map[string]any)
			financialData := result["financialData"].(map[string]any)
			financialData["operatingCashflow"] = map[string]any{"raw": 999000000}
			financialData["freeCashflow"] = map[string]any{"raw": 894000000}
			body, err := json.Marshal(parsed)
			if err != nil {
				t.Fatal(err)
			}
			return jsonResponse(req, http.StatusOK, body, "")
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			types := req.URL.Query().Get("type")
			switch {
			case types == "trailingPegRatio":
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			case strings.Contains(types, "quarterlyOperatingCashFlow"), strings.Contains(types, "quarterlyCashFlowFromContinuingOperatingActivities"), strings.Contains(types, "quarterlyCapitalExpenditure"), strings.Contains(types, "quarterlyFreeCashFlow"):
				body := []byte(`{
					"timeseries": {
						"result": [
							{
								"meta": { "type": ["quarterlyOperatingCashFlow"] },
								"quarterlyOperatingCashFlow": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": 2.18 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 2.00 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 1.95 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 2.05 } }
								]
							},
							{
								"meta": { "type": ["quarterlyCapitalExpenditure"] },
								"quarterlyCapitalExpenditure": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": -1.01 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": -0.90 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": -0.92 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": -0.98 } }
								]
							}
						],
						"error": null
					}
				}`)
				return jsonResponse(req, http.StatusOK, body, "")
			default:
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			}
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetFundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if got.OperatingCashflow != 8 {
		t.Fatalf("expected supplemental operating cash flow 8, got %+v", got)
	}
	if got.FreeCashflow != 4 {
		t.Fatalf("expected supplemental free cash flow 4, got %+v", got)
	}
}

func TestGetFundamentalsPrefersAnnualCashFlowWhenQuarterlyMatchesYearEnd(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-123"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			return jsonFixtureResponse(t, req, "quote_summary_aapl.json")
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			types := req.URL.Query().Get("type")
			switch {
			case types == "trailingPegRatio":
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			case strings.Contains(types, "quarterlyOperatingCashFlow"), strings.Contains(types, "quarterlyCashFlowFromContinuingOperatingActivities"), strings.Contains(types, "quarterlyCapitalExpenditure"), strings.Contains(types, "quarterlyFreeCashFlow"):
				body := []byte(`{
					"timeseries": {
						"result": [
							{
								"meta": { "type": ["quarterlyOperatingCashFlow"] },
								"quarterlyOperatingCashFlow": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": -739.0 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": 452.0 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": 311.0 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": 1015.0 } }
								]
							},
							{
								"meta": { "type": ["quarterlyCapitalExpenditure"] },
								"quarterlyCapitalExpenditure": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": -14.0 } },
									{ "asOfDate": "2025-09-30", "reportedValue": { "raw": -12.0 } },
									{ "asOfDate": "2025-06-30", "reportedValue": { "raw": -11.0 } },
									{ "asOfDate": "2025-03-31", "reportedValue": { "raw": -10.0 } }
								]
							}
						],
						"error": null
					}
				}`)
				return jsonResponse(req, http.StatusOK, body, "")
			case strings.Contains(types, "annualOperatingCashFlow"), strings.Contains(types, "annualCapitalExpenditure"), strings.Contains(types, "annualFreeCashFlow"):
				body := []byte(`{
					"timeseries": {
						"result": [
							{
								"meta": { "type": ["annualOperatingCashFlow"] },
								"annualOperatingCashFlow": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": -739.0 } }
								]
							},
							{
								"meta": { "type": ["annualCapitalExpenditure"] },
								"annualCapitalExpenditure": [
									{ "asOfDate": "2025-12-31", "reportedValue": { "raw": -23.0 } }
								]
							}
						],
						"error": null
					}
				}`)
				return jsonResponse(req, http.StatusOK, body, "")
			default:
				return jsonResponse(req, http.StatusOK, []byte(`{"timeseries":{"result":[],"error":null}}`), "")
			}
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetFundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if got.OperatingCashflow != -739 {
		t.Fatalf("expected annual operating cash flow to win over summed quarterlies, got %+v", got)
	}
	if got.FreeCashflow != -762 {
		t.Fatalf("expected annual-derived free cash flow -762, got %+v", got)
	}
}

func TestGetStatementParsesAnnualIncome(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			types := req.URL.Query().Get("type")
			if !strings.Contains(types, "annualTotalRevenue") {
				t.Fatalf("expected annualTotalRevenue in type query, got %q", types)
			}
			if !strings.Contains(types, "annualNetIncome") {
				t.Fatalf("expected annualNetIncome in type query, got %q", types)
			}
			body := []byte(`{
				"timeseries": {
					"result": [
						{
							"meta": { "type": ["annualTotalRevenue"] },
							"annualTotalRevenue": [
								{
									"asOfDate": "2024-09-30",
									"periodType": "12M",
									"currencyCode": "USD",
									"reportedValue": { "raw": 391035000000 }
								},
								{
									"asOfDate": "2023-09-30",
									"periodType": "12M",
									"currencyCode": "USD",
									"reportedValue": { "raw": { "parsedValue": 383285000000 } }
								}
							]
						},
						{
							"meta": { "type": ["annualNetIncome"] },
							"annualNetIncome": [
								{
									"asOfDate": "2024-09-30",
									"periodType": "12M",
									"currencyCode": "USD",
									"reportedValue": { "raw": 100913000000 }
								},
								{
									"asOfDate": "2023-09-30",
									"periodType": "12M",
									"currencyCode": "USD",
									"reportedValue": { "raw": 96995000000 }
								}
							]
						},
						{
							"meta": { "type": ["annualDilutedEPS"] },
							"annualDilutedEPS": [
								{
									"asOfDate": "2024-09-30",
									"periodType": "12M",
									"currencyCode": "USD",
									"reportedValue": { "raw": 6.11 }
								}
							]
						}
					],
					"error": null
				}
			}`)
			return jsonResponse(req, http.StatusOK, body, "")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetStatement(ctx, "AAPL", domain.StatementKindIncome, domain.StatementFrequencyAnnual)
	if err != nil {
		t.Fatal(err)
	}
	if got.Symbol != "AAPL" || got.Kind != domain.StatementKindIncome || got.Frequency != domain.StatementFrequencyAnnual {
		t.Fatalf("unexpected statement header %+v", got)
	}
	if got.Currency != "USD" {
		t.Fatalf("expected USD currency, got %+v", got)
	}
	if len(got.Periods) != 2 {
		t.Fatalf("expected 2 periods, got %+v", got.Periods)
	}
	if got.Periods[0].Label != "FY 2024" || got.Periods[1].Label != "FY 2023" {
		t.Fatalf("unexpected period labels %+v", got.Periods)
	}
	if len(got.Rows) != 3 {
		t.Fatalf("expected 3 rows, got %+v", got.Rows)
	}
	if got.Rows[0].Key != "TotalRevenue" || got.Rows[0].Values[0].Value != 391035000000 || got.Rows[0].Values[1].Value != 383285000000 {
		t.Fatalf("unexpected revenue row %+v", got.Rows[0])
	}
	if !got.Rows[2].Values[0].Present || got.Rows[2].Values[1].Present {
		t.Fatalf("expected DilutedEPS to be missing for older period, got %+v", got.Rows[2].Values)
	}
}

func TestGetStatementRetriesWithCrumbAfterForbidden(t *testing.T) {
	ctx := context.Background()
	var statementCalls atomic.Int32
	var sawCrumb atomic.Bool

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-statements"), nil
		case strings.HasPrefix(req.URL.Path, "/ws/fundamentals-timeseries/v1/finance/timeseries/"):
			call := statementCalls.Add(1)
			if call == 1 && req.URL.Query().Get("crumb") == "" {
				return textResponse(req, http.StatusForbidden, "forbidden"), nil
			}
			if req.URL.Query().Get("crumb") == "crumb-statements" {
				sawCrumb.Store(true)
			}
			body := []byte(`{
				"timeseries": {
					"result": [
						{
							"meta": { "type": ["annualTotalRevenue"] },
							"annualTotalRevenue": [
								{
									"asOfDate": "2024-09-30",
									"periodType": "12M",
									"currencyCode": "USD",
									"reportedValue": { "raw": 391035000000 }
								}
							]
						}
					],
					"error": null
				}
			}`)
			return jsonResponse(req, http.StatusOK, body, "")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetStatement(ctx, "AAPL", domain.StatementKindIncome, domain.StatementFrequencyAnnual)
	if err != nil {
		t.Fatal(err)
	}
	if statementCalls.Load() != 2 {
		t.Fatalf("expected 2 statement calls, got %d", statementCalls.Load())
	}
	if !sawCrumb.Load() {
		t.Fatal("expected retry with crumb for statements")
	}
	if got.Symbol != "AAPL" || len(got.Rows) != 1 {
		t.Fatalf("unexpected statement %+v", got)
	}
}

func TestGetQuoteRetriesWithCrumbAfterForbidden(t *testing.T) {
	ctx := context.Background()
	var quoteCalls atomic.Int32
	var sawCrumb atomic.Bool

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "crumb-quote"), nil
		case req.URL.Path == "/v7/finance/quote":
			call := quoteCalls.Add(1)
			if call == 1 && req.URL.Query().Get("crumb") == "" {
				return textResponse(req, http.StatusForbidden, "forbidden"), nil
			}
			if req.URL.Query().Get("crumb") == "crumb-quote" {
				sawCrumb.Store(true)
			}
			return jsonFixtureResponse(t, req, "quote_aapl.json")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetQuote(ctx, "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if quoteCalls.Load() != 2 {
		t.Fatalf("expected 2 quote calls, got %d", quoteCalls.Load())
	}
	if !sawCrumb.Load() {
		t.Fatal("expected retry with crumb")
	}
	if got.Symbol != "AAPL" || got.Price <= 0 {
		t.Fatalf("unexpected quote %+v", got)
	}
}

func TestGetQuotesUsesBulkEndpoint(t *testing.T) {
	ctx := context.Background()
	var quoteCalls atomic.Int32

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v7/finance/quote":
			quoteCalls.Add(1)
			if got := req.URL.Query().Get("symbols"); got != "AAPL,MSFT" {
				t.Fatalf("unexpected symbols query %q", got)
			}
			return jsonResponse(req, http.StatusOK, []byte(`{"quoteResponse":{"result":[{"symbol":"AAPL","regularMarketPrice":200.12},{"symbol":"MSFT","regularMarketPrice":420.5}]}}`), "")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	got, err := p.GetQuotes(ctx, []string{"aapl", "MSFT"})
	if err != nil {
		t.Fatal(err)
	}
	if quoteCalls.Load() != 1 {
		t.Fatalf("expected one bulk quote call, got %d", quoteCalls.Load())
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(got))
	}
	if got[0].Symbol != "AAPL" || got[1].Symbol != "MSFT" {
		t.Fatalf("unexpected quotes %+v", got)
	}
}

func TestSearchSymbolsParsesFixture(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v1/finance/search":
			return jsonFixtureResponse(t, req, "search_aapl.json")
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	results, err := p.SearchSymbols(ctx, "apple")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
	if results[0].Symbol != "AAPL" {
		t.Fatalf("unexpected first symbol %+v", results[0])
	}
}

func TestInvalidCrumbResponseIsRejected(t *testing.T) {
	ctx := context.Background()

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.URL.Path == "/":
			return jsonResponse(req, http.StatusOK, []byte(`{}`), "A=B; Path=/")
		case req.URL.Path == "/v1/test/getcrumb":
			return textResponse(req, http.StatusOK, "Too Many Requests"), nil
		case strings.HasPrefix(req.URL.Path, "/v10/finance/quoteSummary/"):
			return textResponse(req, http.StatusForbidden, "forbidden"), nil
		default:
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	_, err := p.GetFundamentals(ctx, "AAPL")
	if err == nil {
		t.Fatal("expected invalid crumb error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid yahoo crumb") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestGetQuoteUsesCache(t *testing.T) {
	ctx := context.Background()
	var quoteCalls atomic.Int32

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v7/finance/quote" {
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		quoteCalls.Add(1)
		return jsonFixtureResponse(t, req, "quote_aapl.json")
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	if _, err := p.GetQuote(ctx, "AAPL"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.GetQuote(ctx, "AAPL"); err != nil {
		t.Fatal(err)
	}
	if quoteCalls.Load() != 1 {
		t.Fatalf("expected one network call, got %d", quoteCalls.Load())
	}
}

func TestGetQuotesUsesPerSymbolCache(t *testing.T) {
	ctx := context.Background()
	var quoteCalls atomic.Int32

	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v7/finance/quote" {
			return textResponse(req, http.StatusNotFound, "not found"), nil
		}
		quoteCalls.Add(1)
		return jsonResponse(req, http.StatusOK, []byte(`{"quoteResponse":{"result":[{"symbol":"AAPL","regularMarketPrice":200.12},{"symbol":"MSFT","regularMarketPrice":420.5}]}}`), "")
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	if _, err := p.GetQuotes(ctx, []string{"AAPL", "MSFT"}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.GetQuote(ctx, "AAPL"); err != nil {
		t.Fatal(err)
	}
	if quoteCalls.Load() != 1 {
		t.Fatalf("expected cached single quote after bulk fetch, got %d calls", quoteCalls.Load())
	}
}

func TestDefaultSearchParamsIncludeYahooFields(t *testing.T) {
	params := defaultSearchParams("aapl")
	raw := params.Encode()
	mustContain := []string{
		url.QueryEscape("corsDomain") + "=finance.yahoo.com",
		url.QueryEscape("formatted") + "=false",
		url.QueryEscape("lang") + "=en-US",
		url.QueryEscape("region") + "=US",
	}
	for _, expected := range mustContain {
		if !strings.Contains(raw, expected) {
			t.Fatalf("expected %q in %q", expected, raw)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestHTTPClient(fn roundTripFunc) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Timeout:   2 * time.Second,
		Jar:       jar,
		Transport: fn,
	}
}

func newTestProvider(baseURL string, client *http.Client) *Provider {
	p := New(Config{
		BaseURL: baseURL,
		Client:  client,
		Timeout: 2 * time.Second,
	})
	p.searchBase = "https://query2.finance.yahoo.test/v1/finance/search"
	p.cookieURL = "https://fc.yahoo.test/"
	p.crumbURL = "https://query1.finance.yahoo.test/v1/test/getcrumb"
	return p
}

func TestFetchCookieAllows404WithCookie(t *testing.T) {
	ctx := context.Background()
	client := newTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "fc.yahoo.test" {
			return jsonResponse(req, http.StatusNotFound, []byte(`{}`), "A=B; Path=/")
		}
		return textResponse(req, http.StatusNotFound, "not found"), nil
	})

	p := newTestProvider("https://query1.finance.yahoo.test", client)
	if err := p.fetchCookie(ctx); err != nil {
		t.Fatal(err)
	}
}

func jsonFixtureResponse(t *testing.T, req *http.Request, name string) (*http.Response, error) {
	t.Helper()
	body := readFixture(t, name)
	return jsonResponse(req, http.StatusOK, body, "")
}

func jsonResponse(req *http.Request, status int, body []byte, setCookie string) (*http.Response, error) {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	if setCookie != "" {
		header.Set("Set-Cookie", setCookie)
	}
	return &http.Response{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    req,
	}, nil
}

func textResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(body) {
		t.Fatalf("invalid json fixture %s", name)
	}
	return body
}
