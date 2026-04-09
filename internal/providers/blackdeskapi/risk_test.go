package blackdeskapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"blackdesk/internal/storage"
)

func TestRiskProviderGetMarketRisk(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second,
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/risk" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body: io.NopCloser(strings.NewReader(`{
			"score": 1,
			"label": "NEUTRAL",
			"scale": {"min": -4, "max": 4},
			"thresholds": {"sma_buffer_pct": 1, "breadth_50_buffer": 2},
			"components": {"s5th_vs_sma200": -1, "spy_vs_sma200": 1},
			"inputs": {
				"s5th": {"name": "S&P 500 Stocks Above 200-Day Average", "symbol": "$S5TH", "current": 54.98, "sma200": 59},
				"spy": {"name": "SPY", "symbol": "SPY", "current": 676.01, "sma200": 663.60}
			},
			"market_now": "2026-04-09T08:30:00-04:00",
			"market_timezone": "America/New_York",
			"market_calendar": "XNYS",
			"generated_at_utc": "2026-04-09T12:30:34Z"
		}`)),
				Request: r,
			}, nil
		}),
	}

	provider := NewRiskProvider(RiskConfig{
		Endpoint: "https://api.blackdesk.test/risk",
		Client:   client,
		Cache:    storage.NewMemoryCache(),
		TTL:      time.Minute,
	})

	got, err := provider.GetMarketRisk(context.Background())
	if err != nil {
		t.Fatalf("GetMarketRisk returned error: %v", err)
	}
	if !got.Available {
		t.Fatal("expected available market risk snapshot")
	}
	if got.Score != 1 || got.Label != "NEUTRAL" {
		t.Fatalf("unexpected risk snapshot %+v", got)
	}
	if got.Min != -4 || got.Max != 4 {
		t.Fatalf("unexpected risk scale %+v", got)
	}
	if got.Thresholds.SMABufferPct != 1 || got.Thresholds.Breadth50Buffer != 2 {
		t.Fatalf("unexpected risk thresholds %+v", got.Thresholds)
	}
	if got.Components["s5th_vs_sma200"] != -1 || got.Components["spy_vs_sma200"] != 1 {
		t.Fatalf("unexpected risk components %+v", got.Components)
	}
	if got.Inputs["spy"].Symbol != "SPY" || got.Inputs["s5th"].SMA200 != 59 {
		t.Fatalf("unexpected risk inputs %+v", got.Inputs)
	}
	if got.MarketZone != "America/New_York" || got.MarketCalendar != "XNYS" {
		t.Fatalf("unexpected market metadata %+v", got)
	}
	if got.GeneratedAt.IsZero() {
		t.Fatal("expected generated_at_utc to parse")
	}
}

func TestRiskProviderGetMarketRiskReturnsErrorOnBadStatus(t *testing.T) {
	provider := NewRiskProvider(RiskConfig{
		Endpoint: "https://api.blackdesk.test/risk",
		Client: &http.Client{
			Timeout: time.Second,
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadGateway,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("nope")),
					Request:    r,
				}, nil
			}),
		},
	})

	if _, err := provider.GetMarketRisk(context.Background()); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
