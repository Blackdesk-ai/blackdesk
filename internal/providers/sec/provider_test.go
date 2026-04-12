package sec

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"blackdesk/internal/domain"
	"blackdesk/internal/storage"
)

func TestGetFilingsBuildsRecentItems(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			var body string
			switch req.URL.Path {
			case "/files/company_tickers.json":
				body = `{"0":{"cik_str":320193,"ticker":"AAPL","title":"Apple Inc."}}`
			case "/submissions/CIK0000320193.json":
				body = `{
				"name":"Apple Inc.",
				"tickers":["AAPL"],
				"filings":{"recent":{
					"accessionNumber":["0000320193-24-000123"],
					"filingDate":["2024-11-01"],
					"reportDate":["2024-09-28"],
					"acceptanceDateTime":["20241101163025"],
					"form":["10-K"],
					"isXBRL":[1],
					"isInlineXBRL":[1],
					"primaryDocument":["aapl-20240928x10k.htm"],
					"primaryDocDescription":["Annual report"]
				}}
			}`
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
					Body:       io.NopCloser(strings.NewReader("not found")),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}

	provider := New(Config{
		Client: client,
		Cache:  storage.NewMemoryCache(),
	})
	provider.tickersURL = "https://example.test/files/company_tickers.json"
	provider.dataBase = "https://example.test"
	provider.wwwBase = "https://example.test"

	got, err := provider.GetFilings(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("GetFilings error: %v", err)
	}
	if got.Symbol != "AAPL" || got.CompanyName != "Apple Inc." || got.CIK != "0000320193" {
		t.Fatalf("unexpected filings snapshot %+v", got)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected one filing, got %d", len(got.Items))
	}
	if got.Items[0].Form != "10-K" {
		t.Fatalf("expected 10-K, got %+v", got.Items[0])
	}
	if got.Items[0].URL == "" {
		t.Fatal("expected filing URL")
	}
}

func TestGetFilingDocumentExtractsReadableText(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body := `<html><body><h1>Annual Report</h1><p>Revenue increased 12%.</p><script>ignored()</script></body></html>`
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
				Request:    req,
			}, nil
		}),
	}

	provider := New(Config{
		Client: client,
		Cache:  storage.NewMemoryCache(),
	})
	doc, err := provider.GetFilingDocument(context.Background(), domain.FilingItem{
		URL: "https://example.test/filing.htm",
	})
	if err != nil {
		t.Fatalf("GetFilingDocument error: %v", err)
	}
	if !strings.Contains(doc.Text, "Annual Report") || !strings.Contains(doc.Text, "Revenue increased 12%.") {
		t.Fatalf("unexpected filing text %q", doc.Text)
	}
	if strings.Contains(doc.Text, "ignored()") {
		t.Fatalf("expected script tags to be removed, got %q", doc.Text)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
