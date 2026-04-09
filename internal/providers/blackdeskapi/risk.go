package blackdeskapi

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
	defaultRiskEndpoint = "https://api.blackdesk.ai/risk"
	defaultRiskTimeout  = 4 * time.Second
	defaultRiskTTL      = 30 * time.Second
	maxRiskBodySize     = 64 << 10
	riskCacheKey        = "blackdesk:risk"
)

type RiskConfig struct {
	Endpoint string
	Client   *http.Client
	Cache    storage.Cache
	Timeout  time.Duration
	TTL      time.Duration
}

type RiskProvider struct {
	endpoint string
	client   *http.Client
	cache    storage.Cache
	ttl      time.Duration
}

func NewRiskProvider(cfg RiskConfig) *RiskProvider {
	client := cfg.Client
	if client == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultRiskTimeout
		}
		client = &http.Client{Timeout: timeout}
	}

	cache := cfg.Cache
	if cache == nil {
		cache = storage.NewMemoryCache()
	}

	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = defaultRiskTTL
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultRiskEndpoint
	}

	return &RiskProvider{
		endpoint: endpoint,
		client:   client,
		cache:    cache,
		ttl:      ttl,
	}
}

func (p *RiskProvider) GetMarketRisk(ctx context.Context) (domain.MarketRiskSnapshot, error) {
	raw, err := p.fetch(ctx)
	if err != nil {
		return domain.MarketRiskSnapshot{}, err
	}

	var payload riskResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.MarketRiskSnapshot{}, fmt.Errorf("decode risk response: %w", err)
	}
	return normalizeRiskResponse(payload)
}

func (p *RiskProvider) fetch(ctx context.Context) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("risk provider unavailable")
	}
	var cached []byte
	if p.cache != nil && p.cache.Get(riskCacheKey, &cached) {
		return cached, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build risk request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch risk data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch risk data: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRiskBodySize))
	if err != nil {
		return nil, fmt.Errorf("read risk response: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("read risk response: empty body")
	}
	if p.cache != nil {
		p.cache.Set(riskCacheKey, body, p.ttl)
	}
	return body, nil
}

type riskResponse struct {
	Score int    `json:"score"`
	Label string `json:"label"`
	Scale struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"scale"`
	Thresholds struct {
		SMABufferPct    float64 `json:"sma_buffer_pct"`
		Breadth50Buffer float64 `json:"breadth_50_buffer"`
	} `json:"thresholds"`
	Components map[string]int `json:"components"`
	Inputs     map[string]struct {
		Name    string  `json:"name"`
		Symbol  string  `json:"symbol"`
		Current float64 `json:"current"`
		SMA200  float64 `json:"sma200"`
	} `json:"inputs"`
	MarketNow      string `json:"market_now"`
	MarketTimezone string `json:"market_timezone"`
	MarketCalendar string `json:"market_calendar"`
	GeneratedAtUTC string `json:"generated_at_utc"`
}

func normalizeRiskResponse(payload riskResponse) (domain.MarketRiskSnapshot, error) {
	minVal := payload.Scale.Min
	maxVal := payload.Scale.Max
	if minVal >= maxVal {
		minVal = -4
		maxVal = 4
	}
	score := payload.Score
	if score < minVal {
		score = minVal
	}
	if score > maxVal {
		score = maxVal
	}

	marketNow, err := parseOptionalRFC3339(payload.MarketNow)
	if err != nil {
		return domain.MarketRiskSnapshot{}, fmt.Errorf("decode risk market_now: %w", err)
	}
	generatedAt, err := parseOptionalRFC3339(payload.GeneratedAtUTC)
	if err != nil {
		return domain.MarketRiskSnapshot{}, fmt.Errorf("decode risk generated_at_utc: %w", err)
	}

	return domain.MarketRiskSnapshot{
		Score: score,
		Label: strings.TrimSpace(payload.Label),
		Min:   minVal,
		Max:   maxVal,
		Thresholds: domain.MarketRiskThresholds{
			SMABufferPct:    payload.Thresholds.SMABufferPct,
			Breadth50Buffer: payload.Thresholds.Breadth50Buffer,
		},
		Components:     cloneRiskComponents(payload.Components),
		Inputs:         normalizeRiskInputs(payload.Inputs),
		MarketNow:      marketNow,
		MarketZone:     strings.TrimSpace(payload.MarketTimezone),
		MarketCalendar: strings.TrimSpace(payload.MarketCalendar),
		GeneratedAt:    generatedAt,
		Available:      true,
	}, nil
}

func parseOptionalRFC3339(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}

func cloneRiskComponents(src map[string]int) map[string]int {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]int, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func normalizeRiskInputs(src map[string]struct {
	Name    string  `json:"name"`
	Symbol  string  `json:"symbol"`
	Current float64 `json:"current"`
	SMA200  float64 `json:"sma200"`
}) map[string]domain.MarketRiskInput {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]domain.MarketRiskInput, len(src))
	for key, value := range src {
		out[key] = domain.MarketRiskInput{
			Name:    strings.TrimSpace(value.Name),
			Symbol:  strings.TrimSpace(value.Symbol),
			Current: value.Current,
			SMA200:  value.SMA200,
		}
	}
	return out
}
