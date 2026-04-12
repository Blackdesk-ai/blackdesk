package sec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"blackdesk/internal/domain"
	"blackdesk/internal/storage"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "Blackdesk/0.1 (https://blackdesk.ai; support@blackdesk.ai)"
)

type Config struct {
	Client    *http.Client
	Cache     storage.Cache
	UserAgent string
}

type Provider struct {
	client     *http.Client
	cache      storage.Cache
	userAgent  string
	tickersURL string
	dataBase   string
	wwwBase    string
}

type tickerMapEntry struct {
	CIK    int    `json:"cik_str"`
	Ticker string `json:"ticker"`
	Title  string `json:"title"`
}

type submissionsResponse struct {
	Name    string   `json:"name"`
	Tickers []string `json:"tickers"`
	Filings struct {
		Recent struct {
			AccessionNumber       []string `json:"accessionNumber"`
			FilingDate            []string `json:"filingDate"`
			ReportDate            []string `json:"reportDate"`
			AcceptanceDateTime    []string `json:"acceptanceDateTime"`
			Form                  []string `json:"form"`
			IsXBRL                []int    `json:"isXBRL"`
			IsInlineXBRL          []int    `json:"isInlineXBRL"`
			PrimaryDocument       []string `json:"primaryDocument"`
			PrimaryDocDescription []string `json:"primaryDocDescription"`
		} `json:"recent"`
	} `json:"filings"`
}

func New(cfg Config) *Provider {
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	cache := cfg.Cache
	if cache == nil {
		cache = storage.NewMemoryCache()
	}
	userAgent := strings.TrimSpace(cfg.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	return &Provider{
		client:     client,
		cache:      cache,
		userAgent:  userAgent,
		tickersURL: "https://www.sec.gov/files/company_tickers.json",
		dataBase:   "https://data.sec.gov",
		wwwBase:    "https://www.sec.gov",
	}
}

func (p *Provider) GetFilings(ctx context.Context, symbol string) (domain.FilingsSnapshot, error) {
	needle := strings.ToUpper(strings.TrimSpace(symbol))
	if needle == "" {
		return domain.FilingsSnapshot{}, fmt.Errorf("filings symbol is required")
	}
	cik, title, err := p.lookupCIK(ctx, needle)
	if err != nil {
		return domain.FilingsSnapshot{}, err
	}

	var resp submissionsResponse
	err = p.fetchJSON(ctx, "sec:submissions:"+cik, 30*time.Minute, p.dataBase+"/submissions/CIK"+cik+".json", &resp)
	if err != nil {
		return domain.FilingsSnapshot{}, err
	}

	snapshot := domain.FilingsSnapshot{
		Symbol:      needle,
		CompanyName: firstNonEmpty(strings.TrimSpace(resp.Name), title),
		CIK:         cik,
		Freshness:   domain.FreshnessLive,
		Provider:    "sec",
		UpdatedAt:   time.Now(),
	}

	count := len(resp.Filings.Recent.Form)
	for i := 0; i < count; i++ {
		item := domain.FilingItem{
			AccessionNumber:       sliceAt(resp.Filings.Recent.AccessionNumber, i),
			Form:                  sliceAt(resp.Filings.Recent.Form, i),
			FilingDate:            parseDate(sliceAt(resp.Filings.Recent.FilingDate, i)),
			ReportDate:            parseDate(sliceAt(resp.Filings.Recent.ReportDate, i)),
			AcceptedAt:            parseAcceptedAt(sliceAt(resp.Filings.Recent.AcceptanceDateTime, i)),
			PrimaryDocument:       sliceAt(resp.Filings.Recent.PrimaryDocument, i),
			PrimaryDocDescription: sliceAt(resp.Filings.Recent.PrimaryDocDescription, i),
			IsXBRL:                sliceAtInt(resp.Filings.Recent.IsXBRL, i) == 1,
			IsInlineXBRL:          sliceAtInt(resp.Filings.Recent.IsInlineXBRL, i) == 1,
		}
		if item.AccessionNumber == "" || item.Form == "" {
			continue
		}
		item.URL = filingURL(p.wwwBase, cik, item.AccessionNumber, item.PrimaryDocument)
		snapshot.Items = append(snapshot.Items, item)
	}

	if snapshot.CompanyName == "" {
		snapshot.CompanyName = title
	}
	return snapshot, nil
}

func (p *Provider) lookupCIK(ctx context.Context, symbol string) (string, string, error) {
	var entries map[string]tickerMapEntry
	if err := p.fetchJSON(ctx, "sec:company_tickers", 24*time.Hour, p.tickersURL, &entries); err != nil {
		return "", "", err
	}
	for _, entry := range entries {
		if strings.EqualFold(strings.TrimSpace(entry.Ticker), symbol) {
			return fmt.Sprintf("%010d", entry.CIK), strings.TrimSpace(entry.Title), nil
		}
	}
	return "", "", fmt.Errorf("SEC filings unavailable for %s", symbol)
}

func (p *Provider) fetchJSON(ctx context.Context, cacheKey string, ttl time.Duration, rawURL string, dest any) error {
	var body []byte
	if p.cache.Get(cacheKey, &body) {
		return json.Unmarshal(body, dest)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SEC request failed: %s", resp.Status)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	p.cache.Set(cacheKey, body, ttl)
	return json.Unmarshal(body, dest)
}

func filingURL(wwwBase, cik, accession, primaryDocument string) string {
	cik = strings.TrimLeft(cik, "0")
	accessionDigits := strings.ReplaceAll(strings.TrimSpace(accession), "-", "")
	document := strings.TrimSpace(primaryDocument)
	if document == "" {
		document = "index.htm"
	}
	return fmt.Sprintf("%s/Archives/edgar/data/%s/%s/%s", strings.TrimRight(wwwBase, "/"), cik, accessionDigits, document)
}

func parseDate(raw string) time.Time {
	ts, _ := time.Parse("2006-01-02", strings.TrimSpace(raw))
	return ts
}

func parseAcceptedAt(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	ts, _ := time.Parse("20060102150405", raw)
	return ts
}

func sliceAt(items []string, idx int) string {
	if idx < 0 || idx >= len(items) {
		return ""
	}
	return strings.TrimSpace(items[idx])
}

func sliceAtInt(items []int, idx int) int {
	if idx < 0 || idx >= len(items) {
		return 0
	}
	return items[idx]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
