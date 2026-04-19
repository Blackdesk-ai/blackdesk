package rss

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"blackdesk/internal/domain"
	"blackdesk/internal/storage"
)

const (
	defaultTimeout  = 12 * time.Second
	defaultTTL      = 20 * time.Second
	defaultMaxItems = 120
	defaultUA       = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"
	maxFeedBodySize = 2 << 20
)

var tagStripper = regexp.MustCompile(`(?s)<[^>]*>`)

var marketNewsIncludeTerms = []string{
	"bond",
	"bonds",
	"central bank",
	"china",
	"commodities",
	"commodity",
	"conflict",
	"cpi",
	"crude",
	"currency",
	"currencies",
	"dollar",
	"euro",
	"energy",
	"earnings",
	"economy",
	"economic",
	"equities",
	"equity",
	"etf",
	"fed",
	"fiscal",
	"fomc",
	"forex",
	"futures",
	"gdp",
	"geopolitical",
	"geopolitics",
	"gold",
	"hormuz",
	"inflation",
	"interest rate",
	"ipo",
	"jobs",
	"jobs report",
	"labor market",
	"lng",
	"macro",
	"market",
	"markets",
	"middle east",
	"missile",
	"natural gas",
	"nasdaq",
	"opec",
	"oil",
	"output",
	"payrolls",
	"pce",
	"policy rate",
	"powell",
	"rates",
	"red sea",
	"recession",
	"shipping lanes",
	"s&p",
	"sanction",
	"sanctions",
	"shares",
	"shipping",
	"sterling",
	"stocks",
	"strait",
	"strait of hormuz",
	"supply chain",
	"tariff",
	"tankers",
	"trade",
	"trade war",
	"treasury",
	"treasuries",
	"unemployment",
	"yuan",
	"vix",
	"wall street",
	"war",
	"yield",
	"yields",
}

var defaultGoogleNewsWireSites = []string{
	"reuters.com",
	"apnews.com",
	"ft.com",
	"cnbc.com",
	"wsj.com",
}

var defaultGoogleNewsWireTopics = []string{
	"markets",
	"economy",
	"stocks",
	"bonds",
	"yields",
	"oil",
	"currencies",
	"fed",
	"tariffs",
}

var breakingGoogleNewsWireTopics = []string{
	"breaking",
	"oil",
	"energy",
	"shipping",
	"strait",
	"strait of hormuz",
	"hormuz",
	"red sea",
	"middle east",
	"sanctions",
	"war",
	"conflict",
	"attack",
	"missile",
	"military",
	"opec",
	"supply chain",
	"tariffs",
}

var globalAlertGoogleNewsSites = []string{
	"reuters.com",
	"apnews.com",
	"ft.com",
	"wsj.com",
	"cnbc.com",
	"bbc.com",
	"theguardian.com",
}

var globalAlertGoogleNewsTopics = []string{
	"breaking",
	"war",
	"conflict",
	"sanctions",
	"tariffs",
	"central bank",
	"fed",
	"inflation",
	"rates",
	"yields",
	"oil",
	"shipping",
	"strait",
	"hormuz",
	"debt",
	"default",
}

var marketNewsExcludeTerms = []string{
	"advertorial",
	"best ",
	"black friday",
	"celebrity",
	"coupon",
	"coupons",
	"deal of the day",
	"divorce",
	"gift guide",
	"how to ",
	"husband",
	"i'm ",
	"ira",
	"is it too late",
	"lifestyle",
	"lottery",
	"mortgage rates for homeowners",
	"opinion",
	"podcast",
	"recipe",
	"retirement",
	"shopping",
	"sponsored",
	"subscriber exclusive",
	"tips for",
	"travel",
	"wife",
}

type FeedSource struct {
	Name     string
	URL      string
	Aliases  []string
	MaxItems int
}

type Config struct {
	Client   *http.Client
	Cache    storage.Cache
	Timeout  time.Duration
	TTL      time.Duration
	MaxItems int
	Sources  []FeedSource
}

type Provider struct {
	client   *http.Client
	cache    storage.Cache
	ttl      time.Duration
	maxItems int
	sources  []FeedSource
}

func New(cfg Config) *Provider {
	client := cfg.Client
	if client == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		client = &http.Client{Timeout: timeout}
	}

	cache := cfg.Cache
	if cache == nil {
		cache = storage.NewMemoryCache()
	}

	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = defaultTTL
	}

	maxItems := cfg.MaxItems
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	sources := append([]FeedSource(nil), cfg.Sources...)
	if len(sources) == 0 {
		sources = DefaultSources()
	}

	return &Provider{
		client:   client,
		cache:    cache,
		ttl:      ttl,
		maxItems: maxItems,
		sources:  sources,
	}
}

func DefaultSources() []FeedSource {
	return []FeedSource{
		{
			Name: "Google News Wire",
			URL:  googleNewsSearchFeed(googleNewsSiteQuery(defaultGoogleNewsWireSites, defaultGoogleNewsWireTopics, "2d")),
			Aliases: []string{
				"Reuters",
				"AP",
				"Financial Times",
				"CNBC",
				"WSJ",
			},
		},
		{
			Name: "Google News Breaking Wire",
			URL:  googleNewsSearchFeed(googleNewsSiteQuery([]string{"reuters.com", "apnews.com"}, breakingGoogleNewsWireTopics, "2d")),
			Aliases: []string{
				"Reuters",
				"AP",
			},
		},
		{
			Name: "Google News Global Alerts",
			URL:  googleNewsSearchFeed(googleNewsSiteQuery(globalAlertGoogleNewsSites, globalAlertGoogleNewsTopics, "2d")),
			Aliases: []string{
				"Reuters",
				"AP",
				"BBC",
				"Financial Times",
				"CNBC",
				"WSJ",
			},
		},
		{
			Name:     "Yahoo Finance",
			URL:      googleNewsSearchFeed(googleNewsSiteQuery([]string{"finance.yahoo.com"}, nil, "2d")),
			MaxItems: 24,
		},
		{Name: "Advisor Perspectives", URL: "https://www.advisorperspectives.com/commentaries.rss"},
		{Name: "Guardian Business", URL: "https://www.theguardian.com/uk/business/rss"},
		{Name: "Guardian World", URL: "https://www.theguardian.com/world/rss"},
		{Name: "Bloomberg", URL: "https://feeds.bloomberg.com/markets/news.rss"},
		{Name: "CNBC", URL: "https://search.cnbc.com/rs/search/combinedcms/view.xml?partnerId=wrss01&id=100003114"},
		{Name: "MarketWatch", URL: "https://feeds.content.dowjones.io/public/rss/mw_topstories"},
		{Name: "Investing.com", URL: "https://www.investing.com/rss/news.rss"},
		{Name: "BBC", URL: "https://feeds.bbci.co.uk/news/business/rss.xml"},
		{Name: "BBC World", URL: "https://feeds.bbci.co.uk/news/world/rss.xml"},
		{Name: "Federal Reserve", URL: "https://www.federalreserve.gov/feeds/press_all.xml"},
		{Name: "European Central Bank", URL: "https://www.ecb.europa.eu/rss/press.html"},
		{Name: "Bank of England", URL: "https://www.bankofengland.co.uk/rss/news"},
		{Name: "Bureau of Economic Analysis", URL: "https://apps.bea.gov/rss/rss.xml"},
		{Name: "SEC", URL: "https://www.sec.gov/news/pressreleases.rss"},
	}
}

func googleNewsSearchFeed(query string) string {
	values := neturl.Values{}
	values.Set("q", query)
	values.Set("hl", "en-US")
	values.Set("gl", "US")
	values.Set("ceid", "US:en")
	return "https://news.google.com/rss/search?" + values.Encode()
}

func googleNewsSiteQuery(sites, topics []string, window string) string {
	parts := make([]string, 0, 3)
	if len(sites) > 0 {
		siteTerms := make([]string, 0, len(sites))
		for _, site := range sites {
			site = strings.TrimSpace(site)
			if site == "" {
				continue
			}
			siteTerms = append(siteTerms, "site:"+site)
		}
		if len(siteTerms) > 0 {
			parts = append(parts, "("+strings.Join(siteTerms, " OR ")+")")
		}
	}
	if len(topics) > 0 {
		topicTerms := make([]string, 0, len(topics))
		for _, topic := range topics {
			topic = strings.TrimSpace(topic)
			if topic == "" {
				continue
			}
			topicTerms = append(topicTerms, topic)
		}
		if len(topicTerms) > 0 {
			parts = append(parts, "("+strings.Join(topicTerms, " OR ")+")")
		}
	}
	window = strings.TrimSpace(window)
	if window != "" {
		parts = append(parts, "when:"+window)
	}
	return strings.Join(parts, " ")
}

func (p *Provider) GetMarketNews(ctx context.Context) ([]domain.NewsItem, error) {
	if len(p.sources) == 0 {
		return nil, fmt.Errorf("no RSS sources configured")
	}

	type sourceResult struct {
		items []domain.NewsItem
		err   error
	}

	results := make([]sourceResult, len(p.sources))
	var wg sync.WaitGroup
	for i, source := range p.sources {
		wg.Add(1)
		go func(idx int, source FeedSource) {
			defer wg.Done()
			items, err := p.fetchSource(ctx, source)
			results[idx] = sourceResult{items: items, err: err}
		}(i, source)
	}
	wg.Wait()

	var firstErr error
	merged := make([]domain.NewsItem, 0, len(p.sources)*8)
	for _, result := range results {
		if result.err != nil && firstErr == nil {
			firstErr = result.err
		}
		merged = append(merged, result.items...)
	}
	merged = dedupeAndSortNews(merged)
	if len(merged) > p.maxItems {
		merged = merged[:p.maxItems]
	}
	if len(merged) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return merged, nil
}

func (p *Provider) MarketNewsSources() []domain.MarketNewsSource {
	if p == nil || len(p.sources) == 0 {
		return nil
	}
	out := make([]domain.MarketNewsSource, 0, len(p.sources))
	seen := make(map[string]struct{}, len(p.sources))
	for _, source := range p.sources {
		if len(source.Aliases) > 0 {
			for _, alias := range source.Aliases {
				name := strings.TrimSpace(alias)
				if name == "" {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				out = append(out, domain.MarketNewsSource{Name: name})
			}
			continue
		}
		name := strings.TrimSpace(source.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, domain.MarketNewsSource{Name: name})
	}
	return out
}

func (p *Provider) fetchSource(ctx context.Context, source FeedSource) ([]domain.NewsItem, error) {
	key := "rss-market-news:" + source.URL
	if p.cache != nil {
		var raw []byte
		if p.cache.Get(key, &raw) {
			var cached []domain.NewsItem
			if err := json.Unmarshal(raw, &cached); err == nil {
				return cached, nil
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml;q=0.9, */*;q=0.8")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s returned %s", source.Name, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBodySize))
	if err != nil {
		return nil, err
	}
	items, err := parseFeed(body, source)
	if err != nil {
		return nil, err
	}
	if source.MaxItems > 0 && len(items) > source.MaxItems {
		items = items[:source.MaxItems]
	}
	if p.cache != nil {
		if raw, err := json.Marshal(items); err == nil {
			p.cache.Set(key, raw, p.ttl)
		}
	}
	return items, nil
}

type rssDocument struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	GUID        string    `xml:"guid"`
	PubDate     string    `xml:"pubDate"`
	Description string    `xml:"description"`
	Source      rssSource `xml:"source"`
}

type rssSource struct {
	Name string `xml:",chardata"`
}

type atomDocument struct {
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string     `xml:"id"`
	Title     string     `xml:"title"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Updated   string     `xml:"updated"`
	Published string     `xml:"published"`
	Links     []atomLink `xml:"link"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func parseFeed(body []byte, source FeedSource) ([]domain.NewsItem, error) {
	root, err := feedRootName(body)
	if err != nil {
		return nil, err
	}
	switch root {
	case "feed":
		return parseAtomFeed(body, source)
	default:
		return parseRSSFeed(body, source)
	}
}

func feedRootName(body []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(body))
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		if start, ok := token.(xml.StartElement); ok {
			return start.Name.Local, nil
		}
	}
}

func parseRSSFeed(body []byte, source FeedSource) ([]domain.NewsItem, error) {
	var doc rssDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	items := doc.Channel.Items
	if len(items) == 0 {
		items = doc.Items
	}
	out := make([]domain.NewsItem, 0, len(items))
	for _, item := range items {
		link := strings.TrimSpace(item.Link)
		title := cleanFeedText(item.Title)
		if title == "" || link == "" {
			continue
		}
		rawPublisher := strings.TrimSpace(item.Source.Name)
		if rawPublisher == "" {
			rawPublisher = source.Name
		}
		title = normalizeRSSHeadline(title, rawPublisher)
		if title == "" {
			continue
		}
		publisher := normalizePublisherName(rawPublisher)
		summary := cleanFeedText(item.Description)
		if !includeFeedItem(source, publisher, title, summary, link) {
			continue
		}
		out = append(out, domain.NewsItem{
			UUID:      strings.TrimSpace(item.GUID),
			Title:     title,
			Summary:   summary,
			Publisher: publisher,
			URL:       normalizeLink(link),
			Time:      parseFeedTime(item.PubDate),
		})
	}
	return out, nil
}

func parseAtomFeed(body []byte, source FeedSource) ([]domain.NewsItem, error) {
	var doc atomDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	out := make([]domain.NewsItem, 0, len(doc.Entries))
	for _, entry := range doc.Entries {
		link := atomEntryLink(entry)
		title := cleanFeedText(entry.Title)
		if title == "" || link == "" {
			continue
		}
		published := strings.TrimSpace(entry.Published)
		if published == "" {
			published = strings.TrimSpace(entry.Updated)
		}
		summary := entry.Summary
		if strings.TrimSpace(summary) == "" {
			summary = entry.Content
		}
		summary = cleanFeedText(summary)
		if !includeFeedItem(source, normalizePublisherName(source.Name), title, summary, link) {
			continue
		}
		out = append(out, domain.NewsItem{
			UUID:      strings.TrimSpace(entry.ID),
			Title:     title,
			Summary:   summary,
			Publisher: normalizePublisherName(source.Name),
			URL:       normalizeLink(link),
			Time:      parseFeedTime(published),
		})
	}
	return out, nil
}

func atomEntryLink(entry atomEntry) string {
	for _, link := range entry.Links {
		if strings.EqualFold(strings.TrimSpace(link.Rel), "alternate") && strings.TrimSpace(link.Href) != "" {
			return link.Href
		}
	}
	for _, link := range entry.Links {
		if strings.TrimSpace(link.Href) != "" {
			return link.Href
		}
	}
	return ""
}

func cleanFeedText(input string) string {
	text := strings.TrimSpace(input)
	if text == "" {
		return ""
	}
	text = html.UnescapeString(text)
	text = tagStripper.ReplaceAllString(text, " ")
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func parseFeedTime(raw string) time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC850,
		time.ANSIC,
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05Z0700",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts
		}
	}
	return time.Time{}
}

func normalizeLink(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	parsed, err := neturl.Parse(link)
	if err != nil {
		return link
	}
	parsed.Fragment = ""
	return parsed.String()
}

func includeFeedItem(source FeedSource, publisher, title, summary, link string) bool {
	textOnly := strings.ToLower(strings.Join([]string{
		strings.TrimSpace(title),
		strings.TrimSpace(summary),
	}, " "))
	if textOnly == "" {
		return false
	}
	for _, term := range marketNewsExcludeTerms {
		if strings.Contains(textOnly, term) {
			return false
		}
	}
	if isLikelyNonArticleHeadline(title) {
		return false
	}
	if isTrustedWireSource(publisher) {
		return true
	}
	if isTrustedWireSource(source.Name) {
		return true
	}
	if isMacroOrRegulatorySource(source.Name) {
		return true
	}
	combined := textOnly + " " + strings.ToLower(strings.TrimSpace(link))
	for _, term := range marketNewsIncludeTerms {
		if strings.Contains(combined, term) {
			return true
		}
	}
	return false
}

func isTrustedWireSource(name string) bool {
	switch strings.TrimSpace(name) {
	case "Bloomberg", "WSJ",
		"Reuters", "AP", "AP News", "Associated Press",
		"CNBC", "Financial Times",
		"MarketWatch", "Investing.com",
		"Yahoo Finance", "Yahoo! Finance Canada",
		"BBC":
		return true
	default:
		return false
	}
}

func isMacroOrRegulatorySource(name string) bool {
	switch strings.TrimSpace(name) {
	case "Federal Reserve", "European Central Bank", "Bank of England", "Bureau of Economic Analysis", "SEC":
		return true
	default:
		return false
	}
}

func dedupeAndSortNews(items []domain.NewsItem) []domain.NewsItem {
	seenURL := make(map[string]struct{}, len(items))
	seenTitle := make(map[string]struct{}, len(items))
	out := make([]domain.NewsItem, 0, len(items))
	for _, item := range items {
		urlKey := strings.TrimSpace(item.URL)
		titleKey := normalizedNewsTitle(item.Title)
		if urlKey == "" && titleKey == "" {
			continue
		}
		if urlKey != "" {
			if _, ok := seenURL[urlKey]; ok {
				continue
			}
		}
		if titleKey != "" {
			if _, ok := seenTitle[titleKey]; ok {
				continue
			}
		}
		if urlKey != "" {
			seenURL[urlKey] = struct{}{}
		}
		if titleKey != "" {
			seenTitle[titleKey] = struct{}{}
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti := out[i].Time
		tj := out[j].Time
		switch {
		case ti.IsZero() && tj.IsZero():
			return out[i].Title < out[j].Title
		case ti.IsZero():
			return false
		case tj.IsZero():
			return true
		default:
			return ti.After(tj)
		}
	})
	return out
}

func normalizedNewsTitle(title string) string {
	key := strings.ToLower(strings.Join(strings.Fields(title), " "))
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	for _, suffix := range []string{" - reuters", " - ap", " - associated press", " - financial times", " - bloomberg", " - wsj"} {
		key = strings.TrimSuffix(key, suffix)
	}
	for _, suffix := range []string{" - yahoo finance", " - yahoo finance australia", " - yahoo finance singapore", " - yahoo finance uk", " - yahoo! finance canada"} {
		key = strings.TrimSuffix(key, suffix)
	}
	return strings.TrimSpace(key)
}

func normalizeRSSHeadline(title, publisher string) string {
	title = strings.TrimSpace(strings.Join(strings.Fields(title), " "))
	publisher = strings.TrimSpace(strings.Join(strings.Fields(publisher), " "))
	if title == "" {
		return ""
	}
	if publisher == "" {
		return title
	}
	lowerTitle := strings.ToLower(title)
	lowerPublisher := strings.ToLower(publisher)
	for _, sep := range []string{" - ", " | "} {
		suffix := sep + lowerPublisher
		if strings.HasSuffix(lowerTitle, suffix) {
			cut := len(title) - len(suffix)
			return strings.TrimSpace(title[:cut])
		}
	}
	return title
}

func normalizePublisherName(publisher string) string {
	publisher = strings.TrimSpace(strings.Join(strings.Fields(publisher), " "))
	lower := strings.ToLower(publisher)
	switch {
	case lower == "ap news":
		return "AP"
	case lower == "associated press":
		return "AP"
	case lower == "bbc news":
		return "BBC"
	case lower == "the guardian":
		return "Guardian"
	case strings.HasPrefix(lower, "yahoo finance"):
		return "Yahoo Finance"
	case strings.HasPrefix(lower, "yahoo! finance"):
		return "Yahoo Finance"
	default:
		return publisher
	}
}

func isLikelyNonArticleHeadline(title string) bool {
	lower := strings.ToLower(strings.TrimSpace(title))
	if lower == "" {
		return true
	}
	for _, bad := range []string{
		"breaking news",
		"latest news today",
		"associated press news",
		"ap news",
		"yahoo finance - stock market live",
	} {
		if strings.Contains(lower, bad) {
			return true
		}
	}
	return false
}
