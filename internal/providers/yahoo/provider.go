package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"blackdesk/internal/domain"
	"blackdesk/internal/storage"
)

const (
	defaultTimeout = 10 * time.Second
	defaultUA      = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	maxQuoteBatch  = 50
)

type Config struct {
	BaseURL string
	Client  *http.Client
	Cache   storage.Cache
	Timeout time.Duration
}

type Provider struct {
	quoteBase        string
	chartBase        string
	quoteSummaryBase string
	timeseriesBase   string
	searchBase       string
	cookieURL        string
	crumbURL         string
	client           *http.Client
	cache            storage.Cache
	session          *session
}

func New(cfg Config) *Provider {
	client := cfg.Client
	if client == nil {
		jar, _ := cookiejar.New(nil)
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = defaultTimeout
		}
		client = &http.Client{
			Timeout: timeout,
			Jar:     jar,
		}
	} else if client.Jar == nil {
		jar, _ := cookiejar.New(nil)
		client.Jar = jar
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://query1.finance.yahoo.com"
	}

	cache := cfg.Cache
	if cache == nil {
		cache = storage.NewMemoryCache()
	}

	return &Provider{
		quoteBase:        baseURL + "/v7/finance/quote",
		chartBase:        baseURL + "/v8/finance/chart/",
		quoteSummaryBase: baseURL + "/v10/finance/quoteSummary/",
		timeseriesBase:   baseURL + "/ws/fundamentals-timeseries/v1/finance/timeseries/",
		searchBase:       "https://query2.finance.yahoo.com/v1/finance/search",
		cookieURL:        "https://fc.yahoo.com/",
		crumbURL:         baseURL + "/v1/test/getcrumb",
		client:           client,
		cache:            cache,
		session:          newSession(),
	}
}

func (p *Provider) Name() string {
	return "yahoo"
}

func (p *Provider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{
		Quote:        true,
		History:      true,
		News:         true,
		Fundamentals: true,
		Search:       true,
		Statements:   true,
		Insiders:     true,
		Screeners:    true,
	}
}

func (p *Provider) GetQuote(ctx context.Context, symbol string) (domain.QuoteSnapshot, error) {
	quotes, err := p.GetQuotes(ctx, []string{symbol})
	if err != nil {
		return domain.QuoteSnapshot{}, err
	}
	return quoteBySymbol(quotes, symbol)
}

func (p *Provider) GetQuotes(ctx context.Context, symbols []string) ([]domain.QuoteSnapshot, error) {
	normalized := normalizeSymbols(symbols)
	if len(normalized) == 0 {
		return nil, nil
	}

	outBySymbol := make(map[string]domain.QuoteSnapshot, len(normalized))
	missing := make([]string, 0, len(normalized))
	for _, symbol := range normalized {
		quote, ok, err := p.cachedQuote(symbol)
		if err != nil {
			return nil, err
		}
		if ok {
			outBySymbol[symbol] = quote
			continue
		}
		missing = append(missing, symbol)
	}

	for start := 0; start < len(missing); start += maxQuoteBatch {
		end := min(start+maxQuoteBatch, len(missing))
		batch, err := p.fetchQuotesBatch(ctx, missing[start:end])
		if err != nil {
			return nil, err
		}
		for _, quote := range batch {
			symbol := strings.ToUpper(quote.Symbol)
			outBySymbol[symbol] = quote
			if err := p.cacheQuote(symbol, quote); err != nil {
				return nil, err
			}
		}
	}

	quotes := make([]domain.QuoteSnapshot, 0, len(normalized))
	for _, symbol := range normalized {
		quote, ok := outBySymbol[symbol]
		if !ok {
			continue
		}
		quotes = append(quotes, quote)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("quotes not found")
	}
	return quotes, nil
}

func (p *Provider) GetHistory(ctx context.Context, symbol, rangeKey, interval string) (domain.PriceSeries, error) {
	var resp chartResponse
	params := url.Values{}
	params.Set("range", rangeKey)
	params.Set("interval", interval)
	params.Set("includePrePost", "false")
	params.Set("events", "div,splits")
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.chartBase + url.PathEscape(strings.ToUpper(symbol)),
		Params:   params,
		CacheKey: "chart:" + strings.ToUpper(symbol) + ":" + rangeKey + ":" + interval,
		TTL:      20 * time.Second,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return domain.PriceSeries{}, err
	}
	return normalizeChart(symbol, rangeKey, interval, resp)
}

func (p *Provider) GetNews(ctx context.Context, symbol string) ([]domain.NewsItem, error) {
	var resp searchResponse
	params := defaultSearchParams(strings.ToUpper(symbol))
	params.Set("quotesCount", "6")
	params.Set("newsCount", "8")
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.searchBase,
		Params:   params,
		CacheKey: "search:" + symbol + ":news",
		TTL:      5 * time.Minute,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return normalizeNews(resp), nil
}

func (p *Provider) SearchSymbols(ctx context.Context, query string) ([]domain.SymbolRef, error) {
	var resp searchResponse
	params := defaultSearchParams(strings.TrimSpace(query))
	params.Set("quotesCount", "8")
	params.Set("newsCount", "0")
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.searchBase,
		Params:   params,
		CacheKey: "search:" + query + ":symbols",
		TTL:      5 * time.Minute,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return normalizeSearch(resp), nil
}

func (p *Provider) GetFundamentals(ctx context.Context, symbol string) (domain.FundamentalsSnapshot, error) {
	var resp quoteSummaryResponse
	normalizedSymbol := strings.ToUpper(symbol)
	params := url.Values{}
	params.Set("modules", "price,summaryDetail,defaultKeyStatistics,financialData,assetProfile")
	params.Set("corsDomain", "finance.yahoo.com")
	params.Set("formatted", "false")
	params.Set("symbol", normalizedSymbol)
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.quoteSummaryBase + url.PathEscape(normalizedSymbol),
		Params:   params,
		CacheKey: "fundamentals:" + normalizedSymbol,
		TTL:      30 * time.Minute,
		Auth:     authRequired,
	}, &resp)
	if err != nil {
		return domain.FundamentalsSnapshot{}, err
	}

	fundamentals, err := normalizeFundamentals(normalizedSymbol, resp)
	if err != nil {
		return domain.FundamentalsSnapshot{}, err
	}
	if fundamentals.PEGRatio == 0 {
		if peg, err := p.fetchTrailingPEGRatio(ctx, normalizedSymbol); err == nil {
			fundamentals.PEGRatio = peg
		}
	}
	p.hydrateSupplementalFundamentals(ctx, normalizedSymbol, &fundamentals)
	return fundamentals, nil
}

func (p *Provider) fetchTrailingPEGRatio(ctx context.Context, symbol string) (float64, error) {
	var resp fundamentalsTimeseriesResponse
	now := time.Now().UTC()
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("type", "trailingPegRatio")
	params.Set("period1", fmt.Sprintf("%d", now.AddDate(0, -6, 0).Unix()))
	params.Set("period2", fmt.Sprintf("%d", now.Unix()))
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.timeseriesBase + url.PathEscape(symbol),
		Params:   params,
		CacheKey: "fundamentals-timeseries:trailingPegRatio:" + symbol,
		TTL:      30 * time.Minute,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return 0, err
	}
	for _, result := range resp.Timeseries.Result {
		for i := len(result.TrailingPegRatio) - 1; i >= 0; i-- {
			if v := result.TrailingPegRatio[i].ReportedValue.Raw; v != 0 {
				return v, nil
			}
		}
	}
	return 0, nil
}

type authMode int

const (
	authNone authMode = iota
	authOptional
	authRequired
)

type requestSpec struct {
	URL      string
	Params   url.Values
	CacheKey string
	TTL      time.Duration
	Auth     authMode
}

func (p *Provider) fetchJSON(ctx context.Context, spec requestSpec, dest any) error {
	var data []byte
	if p.cache.Get(spec.CacheKey, &data) {
		return json.Unmarshal(data, dest)
	}

	body, err := p.doRequest(ctx, spec)
	if err != nil {
		return err
	}
	p.cache.Set(spec.CacheKey, body, spec.TTL)
	return json.Unmarshal(body, dest)
}

func (p *Provider) doRequest(ctx context.Context, spec requestSpec) ([]byte, error) {
	for attempt := 0; attempt < 2; attempt++ {
		withCrumb := spec.Auth == authRequired || (spec.Auth == authOptional && attempt == 1)
		if withCrumb {
			if err := p.ensureCredentials(ctx); err != nil {
				return nil, err
			}
		}

		reqURL, err := p.buildURL(spec, withCrumb)
		if err != nil {
			return nil, err
		}
		body, status, err := p.get(ctx, reqURL)
		if err != nil {
			return nil, err
		}

		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			p.clearCrumb()
			if attempt == 0 {
				continue
			}
			return nil, fmt.Errorf("yahoo returned %d", status)
		}
		if status == http.StatusTooManyRequests {
			p.clearCrumb()
			return nil, fmt.Errorf("yahoo rate limited the request (429)")
		}
		if status >= 400 {
			return nil, fmt.Errorf("yahoo returned %d", status)
		}
		if looksLikeInvalidCrumb(body) {
			p.clearCrumb()
			if attempt == 0 {
				continue
			}
			return nil, fmt.Errorf("yahoo returned invalid crumb response")
		}
		return body, nil
	}
	return nil, fmt.Errorf("yahoo request failed after retry")
}

func (p *Provider) get(ctx context.Context, reqURL string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func (p *Provider) buildURL(spec requestSpec, withCrumb bool) (string, error) {
	u, err := url.Parse(spec.URL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, values := range spec.Params {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	if withCrumb {
		crumb, ok := p.session.getCrumb()
		if !ok {
			return "", fmt.Errorf("yahoo crumb unavailable")
		}
		q.Set("crumb", crumb)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (p *Provider) fetchQuotesBatch(ctx context.Context, symbols []string) ([]domain.QuoteSnapshot, error) {
	var resp quoteResponse
	params := url.Values{}
	params.Set("symbols", strings.Join(symbols, ","))
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.quoteBase,
		Params:   params,
		CacheKey: "quote-batch:" + strings.Join(symbols, ","),
		TTL:      20 * time.Second,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return normalizeQuotes(resp), nil
}

func (p *Provider) cachedQuote(symbol string) (domain.QuoteSnapshot, bool, error) {
	var data []byte
	if !p.cache.Get("quote:"+symbol, &data) {
		return domain.QuoteSnapshot{}, false, nil
	}
	var quote domain.QuoteSnapshot
	if err := json.Unmarshal(data, &quote); err != nil {
		return domain.QuoteSnapshot{}, false, err
	}
	return quote, true, nil
}

func (p *Provider) cacheQuote(symbol string, quote domain.QuoteSnapshot) error {
	body, err := json.Marshal(quote)
	if err != nil {
		return err
	}
	p.cache.Set("quote:"+symbol, body, 20*time.Second)
	return nil
}

func normalizeSymbols(symbols []string) []string {
	normalized := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		key := strings.ToUpper(strings.TrimSpace(symbol))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	return normalized
}

func defaultSearchParams(query string) url.Values {
	params := url.Values{}
	params.Set("q", query)
	params.Set("listsCount", "0")
	params.Set("enableFuzzyQuery", "false")
	params.Set("quotesQueryId", "tss_match_phrase_query")
	params.Set("multiQuoteQueryId", "multi_quote_single_token_query")
	params.Set("newsQueryId", "news_cie_vespa")
	params.Set("enableCb", "true")
	params.Set("enableNavLinks", "false")
	params.Set("enableResearchReports", "false")
	params.Set("enableCulturalAssets", "false")
	params.Set("recommendedCount", "0")
	params.Set("lang", "en-US")
	params.Set("region", "US")
	params.Set("corsDomain", "finance.yahoo.com")
	params.Set("formatted", "false")
	return params
}
