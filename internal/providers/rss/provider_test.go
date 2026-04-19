package rss

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"testing"
	"time"

	"blackdesk/internal/domain"
	"blackdesk/internal/storage"
)

func TestGetMarketNewsAggregatesAndDeduplicatesSources(t *testing.T) {
	now := time.Now().UTC()
	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
			<rss version="2.0">
			  <channel>
			    <title>Wire</title>
			    <item>
			      <guid>story-1</guid>
			      <title>Futures rise into CPI print</title>
			      <link>https://example.com/story-1</link>
			      <pubDate>%s</pubDate>
			      <description><![CDATA[<p>Risk assets bid <b>higher</b> overnight.</p>]]></description>
			    </item>
			    <item>
			      <guid>story-dup</guid>
			      <title>Duplicate Treasury yields story</title>
			      <link>https://example.com/duplicate</link>
			      <pubDate>%s</pubDate>
			    </item>
			  </channel>
			</rss>`,
			now.Add(-5*time.Minute).Format(time.RFC1123Z),
			now.Add(-15*time.Minute).Format(time.RFC1123Z),
		)
	}))
	defer rssServer.Close()

	atomServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<?xml version="1.0" encoding="utf-8"?>
			<feed xmlns="http://www.w3.org/2005/Atom">
			  <title>Atom Wire</title>
			  <entry>
			    <id>tag:example.com,2026:story-2</id>
			    <title>Fed speakers keep yields elevated</title>
			    <updated>%s</updated>
			    <summary>Rates stay firm after hawkish remarks.</summary>
			    <link href="https://example.com/story-2" rel="alternate"></link>
			  </entry>
			  <entry>
			    <id>tag:example.com,2026:dup</id>
			    <title>Duplicate Treasury yields story</title>
			    <updated>%s</updated>
			    <summary>Second copy should be deduped.</summary>
			    <link href="https://example.com/duplicate" rel="alternate"></link>
			  </entry>
			</feed>`,
			now.Add(-2*time.Minute).Format(time.RFC3339),
			now.Add(-10*time.Minute).Format(time.RFC3339),
		)
	}))
	defer atomServer.Close()

	provider := New(Config{
		Client: rssServer.Client(),
		Cache:  storage.NewMemoryCache(),
		Sources: []FeedSource{
			{Name: "RSS One", URL: rssServer.URL},
			{Name: "Atom Two", URL: atomServer.URL},
		},
		MaxItems: 10,
	})

	items, err := provider.GetMarketNews(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 deduplicated market items, got %d", len(items))
	}
	if items[0].Title != "Fed speakers keep yields elevated" {
		t.Fatalf("expected newest atom story first, got %+v", items[0])
	}
	if items[1].Title != "Futures rise into CPI print" {
		t.Fatalf("expected RSS story second, got %+v", items[1])
	}
	if !strings.Contains(items[1].Summary, "Risk assets bid higher overnight.") {
		t.Fatalf("expected HTML-stripped summary, got %q", items[1].Summary)
	}
}

func TestGetMarketNewsUsesCacheAfterFirstFetch(t *testing.T) {
	var hits int
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
			<rss version="2.0">
			  <channel>
			    <title>Wire</title>
			    <item>
			      <guid>story-1</guid>
			      <title>Treasury yields ease after jobs report</title>
			      <link>https://example.com/story-1</link>
			      <pubDate>%s</pubDate>
			    </item>
			  </channel>
			</rss>`, now.Format(time.RFC1123Z))
	}))
	defer server.Close()

	provider := New(Config{
		Client: server.Client(),
		Cache:  storage.NewMemoryCache(),
		TTL:    time.Minute,
		Sources: []FeedSource{
			{Name: "Wire", URL: server.URL},
		},
	})

	for i := 0; i < 2; i++ {
		items, err := provider.GetMarketNews(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 || items[0].Title != "Treasury yields ease after jobs report" {
			t.Fatalf("unexpected cached items %+v", items)
		}
	}
	if hits != 1 {
		t.Fatalf("expected one network fetch due to cache, got %d", hits)
	}
}

func TestGetMarketNewsHonorsSourceMaxItems(t *testing.T) {
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
			<rss version="2.0">
			  <channel>
			    <title>Wire</title>
			    <item>
			      <guid>story-1</guid>
			      <title>Oil rises as traders watch OPEC output</title>
			      <link>https://example.com/story-1</link>
			      <pubDate>%s</pubDate>
			    </item>
			    <item>
			      <guid>story-2</guid>
			      <title>Gold slips as dollar steadies</title>
			      <link>https://example.com/story-2</link>
			      <pubDate>%s</pubDate>
			    </item>
			    <item>
			      <guid>story-3</guid>
			      <title>Treasury yields ease after inflation data</title>
			      <link>https://example.com/story-3</link>
			      <pubDate>%s</pubDate>
			    </item>
			  </channel>
			</rss>`,
			now.Add(-1*time.Minute).Format(time.RFC1123Z),
			now.Add(-2*time.Minute).Format(time.RFC1123Z),
			now.Add(-3*time.Minute).Format(time.RFC1123Z),
		)
	}))
	defer server.Close()

	provider := New(Config{
		Client: server.Client(),
		Cache:  storage.NewMemoryCache(),
		Sources: []FeedSource{
			{Name: "Yahoo Finance", URL: server.URL, MaxItems: 2},
		},
	})

	items, err := provider.GetMarketNews(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected source max item cap to keep 2 items, got %d", len(items))
	}
	if items[0].Title != "Oil rises as traders watch OPEC output" || items[1].Title != "Gold slips as dollar steadies" {
		t.Fatalf("unexpected capped items %+v", items)
	}
}

func TestGetMarketNewsFiltersNonMarketHeadlinesFromWireFeeds(t *testing.T) {
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
			<rss version="2.0">
			  <channel>
			    <title>Wire</title>
			    <item>
			      <guid>story-1</guid>
			      <title>I'm 56 and only have $60,000 in my IRA. Is it too late for me?</title>
			      <link>https://example.com/personal-finance</link>
			      <pubDate>%s</pubDate>
			    </item>
			    <item>
			      <guid>story-2</guid>
			      <title>Treasury yields steady ahead of CPI report</title>
			      <link>https://example.com/cpi</link>
			      <pubDate>%s</pubDate>
			    </item>
			  </channel>
			</rss>`,
			now.Add(-3*time.Minute).Format(time.RFC1123Z),
			now.Add(-2*time.Minute).Format(time.RFC1123Z),
		)
	}))
	defer server.Close()

	provider := New(Config{
		Client: server.Client(),
		Cache:  storage.NewMemoryCache(),
		Sources: []FeedSource{
			{Name: "Reuters", URL: server.URL},
		},
	})

	items, err := provider.GetMarketNews(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one filtered market story, got %d", len(items))
	}
	if items[0].Title != "Treasury yields steady ahead of CPI report" {
		t.Fatalf("unexpected surviving headline %+v", items[0])
	}
}

func TestGetMarketNewsKeepsMacroAndRegulatorySources(t *testing.T) {
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
			<rss version="2.0">
			  <channel>
			    <title>Fed</title>
			    <item>
			      <guid>statement-1</guid>
			      <title>Federal Reserve issues discount rate meeting minutes</title>
			      <link>https://example.com/fed-minutes</link>
			      <pubDate>%s</pubDate>
			    </item>
			  </channel>
			</rss>`,
			now.Add(-1*time.Minute).Format(time.RFC1123Z),
		)
	}))
	defer server.Close()

	provider := New(Config{
		Client: server.Client(),
		Cache:  storage.NewMemoryCache(),
		Sources: []FeedSource{
			{Name: "Federal Reserve", URL: server.URL},
		},
	})

	items, err := provider.GetMarketNews(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected regulatory item to survive, got %d", len(items))
	}
	if items[0].Title != "Federal Reserve issues discount rate meeting minutes" {
		t.Fatalf("unexpected regulatory headline %+v", items[0])
	}
}

func TestGetMarketNewsKeepsTrustedWireHeadlinesWithoutKeywordMatch(t *testing.T) {
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
			<rss version="2.0">
			  <channel>
			    <title>Reuters</title>
			    <item>
			      <guid>story-1</guid>
			      <title>Trump says decision on tariffs coming this week</title>
			      <link>https://example.com/reuters-tariffs</link>
			      <pubDate>%s</pubDate>
			    </item>
			  </channel>
			</rss>`,
			now.Add(-1*time.Minute).Format(time.RFC1123Z),
		)
	}))
	defer server.Close()

	provider := New(Config{
		Client: server.Client(),
		Cache:  storage.NewMemoryCache(),
		Sources: []FeedSource{
			{Name: "Reuters", URL: server.URL},
		},
	})

	items, err := provider.GetMarketNews(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected trusted wire item to survive, got %d", len(items))
	}
	if items[0].Publisher != "Reuters" {
		t.Fatalf("expected Reuters publisher label, got %+v", items[0])
	}
}

func TestMarketNewsSourcesReturnsConfiguredNames(t *testing.T) {
	provider := New(Config{
		Sources: []FeedSource{
			{Name: "Reuters", URL: "https://example.com/reuters.xml"},
			{Name: "AP", URL: "https://example.com/ap.xml"},
			{Name: "WSJ", URL: "https://example.com/wsj.xml"},
		},
	})

	got := provider.MarketNewsSources()
	if len(got) != 3 {
		t.Fatalf("expected 3 source names, got %d", len(got))
	}
	if got[0].Name != "Reuters" || got[1].Name != "AP" || got[2].Name != "WSJ" {
		t.Fatalf("unexpected source catalog %+v", got)
	}
}

func TestMarketNewsSourcesDeduplicatesAliasesAndSourceNames(t *testing.T) {
	provider := New(Config{
		Sources: []FeedSource{
			{Name: "Google News Wire", URL: "https://example.com/google.xml", Aliases: []string{"Reuters", "AP", "CNBC"}},
			{Name: "Google News Breaking Wire", URL: "https://example.com/google-breaking.xml", Aliases: []string{"Reuters", "AP"}},
			{Name: "CNBC", URL: "https://example.com/cnbc.xml"},
			{Name: "Bloomberg", URL: "https://example.com/bloomberg.xml"},
		},
	})

	got := provider.MarketNewsSources()
	if len(got) != 4 {
		t.Fatalf("expected 4 unique source names, got %d", len(got))
	}
	if got[0].Name != "Reuters" || got[1].Name != "AP" || got[2].Name != "CNBC" || got[3].Name != "Bloomberg" {
		t.Fatalf("unexpected deduplicated source catalog %+v", got)
	}
}

func TestDefaultSourcesIncludeReutersAPBreakingWire(t *testing.T) {
	sources := DefaultSources()

	var breaking FeedSource
	for _, source := range sources {
		if source.Name == "Google News Breaking Wire" {
			breaking = source
			break
		}
	}
	if breaking.Name == "" {
		t.Fatal("expected Google News Breaking Wire default source")
	}
	if len(breaking.Aliases) != 2 || breaking.Aliases[0] != "Reuters" || breaking.Aliases[1] != "AP" {
		t.Fatalf("unexpected breaking aliases %+v", breaking.Aliases)
	}

	parsed, err := neturl.Parse(breaking.URL)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query().Get("q")
	for _, want := range []string{"site:reuters.com", "site:apnews.com", "hormuz", "shipping", "sanctions"} {
		if !strings.Contains(query, want) {
			t.Fatalf("expected breaking query %q to contain %q", query, want)
		}
	}
}

func TestDefaultSourcesIncludeGlobalAlertsWire(t *testing.T) {
	sources := DefaultSources()

	var global FeedSource
	for _, source := range sources {
		if source.Name == "Google News Global Alerts" {
			global = source
			break
		}
	}
	if global.Name == "" {
		t.Fatal("expected Google News Global Alerts default source")
	}

	parsed, err := neturl.Parse(global.URL)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query().Get("q")
	for _, want := range []string{"site:reuters.com", "site:apnews.com", "site:bbc.com", "war", "sanctions", "default"} {
		if !strings.Contains(query, want) {
			t.Fatalf("expected global alerts query %q to contain %q", query, want)
		}
	}
}

func TestDefaultSourcesIncludeYahooFinanceDirectWire(t *testing.T) {
	sources := DefaultSources()

	var yahoo FeedSource
	for _, source := range sources {
		if source.Name == "Yahoo Finance" {
			yahoo = source
			break
		}
	}
	if yahoo.Name == "" {
		t.Fatal("expected Yahoo Finance default source")
	}
	if yahoo.MaxItems != 24 {
		t.Fatalf("expected Yahoo Finance source cap 24, got %d", yahoo.MaxItems)
	}

	parsed, err := neturl.Parse(yahoo.URL)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query().Get("q")
	for _, want := range []string{"site:finance.yahoo.com", "when:2d"} {
		if !strings.Contains(query, want) {
			t.Fatalf("expected Yahoo query %q to contain %q", query, want)
		}
	}
}

func TestDefaultSourcesIncludeAdvisorPerspectives(t *testing.T) {
	sources := DefaultSources()

	var advisor FeedSource
	for _, source := range sources {
		if source.Name == "Advisor Perspectives" {
			advisor = source
			break
		}
	}
	if advisor.Name == "" {
		t.Fatal("expected Advisor Perspectives default source")
	}
	if advisor.URL != "https://www.advisorperspectives.com/commentaries.rss" {
		t.Fatalf("unexpected Advisor Perspectives URL %q", advisor.URL)
	}
}

func TestDefaultSourcesIncludeGuardianFeeds(t *testing.T) {
	sources := DefaultSources()

	var business FeedSource
	var world FeedSource
	for _, source := range sources {
		switch source.Name {
		case "Guardian Business":
			business = source
		case "Guardian World":
			world = source
		}
	}
	if business.Name == "" || world.Name == "" {
		t.Fatalf("expected Guardian default sources, got business=%q world=%q", business.Name, world.Name)
	}
	if business.URL != "https://www.theguardian.com/uk/business/rss" {
		t.Fatalf("unexpected Guardian Business URL %q", business.URL)
	}
	if world.URL != "https://www.theguardian.com/world/rss" {
		t.Fatalf("unexpected Guardian World URL %q", world.URL)
	}
}

func TestDefaultSourcesIncludeBBCWorld(t *testing.T) {
	sources := DefaultSources()

	var world FeedSource
	for _, source := range sources {
		if source.Name == "BBC World" {
			world = source
			break
		}
	}
	if world.Name == "" {
		t.Fatal("expected BBC World default source")
	}
	if world.URL != "https://feeds.bbci.co.uk/news/world/rss.xml" {
		t.Fatalf("unexpected BBC World URL %q", world.URL)
	}
}

func TestDedupeAndSortNewsDropsExactTitleDuplicatesAcrossSources(t *testing.T) {
	now := time.Now().UTC()
	items := []domain.NewsItem{
		{
			Title:     "Oil climbs as traders watch OPEC output - Reuters",
			Publisher: "Reuters",
			URL:       "https://news.google.com/articles/reuters-1",
			Time:      now,
		},
		{
			Title:     "Oil climbs as traders watch OPEC output",
			Publisher: "Bloomberg",
			URL:       "https://example.com/bloomberg-1",
			Time:      now.Add(-time.Minute),
		},
	}

	got := dedupeAndSortNews(items)
	if len(got) != 1 {
		t.Fatalf("expected duplicate titles to collapse to one item, got %d", len(got))
	}
}

func TestParseRSSFeedGoogleNewsUsesSourceAndStripsPublisherSuffix(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<rss version="2.0">
		  <channel>
		    <title>Google News</title>
		    <item>
		      <guid>story-1</guid>
		      <title>Asian shares mostly gain while European trading stays closed for a holiday - AP News</title>
		      <link>https://news.google.com/rss/articles/example-1?oc=5</link>
		      <pubDate>Mon, 06 Apr 2026 09:04:00 GMT</pubDate>
		      <description><![CDATA[<a href="https://news.google.com/rss/articles/example-1?oc=5">Asian shares mostly gain while European trading stays closed for a holiday</a>&nbsp;&nbsp;<font color="#6f6f6f">AP News</font>]]></description>
		      <source url="https://apnews.com">AP News</source>
		    </item>
		  </channel>
		</rss>`)

	items, err := parseRSSFeed(body, FeedSource{Name: "AP", URL: "https://news.google.com/rss/search?q=ap"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one parsed item, got %d", len(items))
	}
	if items[0].Publisher != "AP" {
		t.Fatalf("expected publisher from source tag, got %+v", items[0])
	}
	if items[0].Title != "Asian shares mostly gain while European trading stays closed for a holiday" {
		t.Fatalf("expected stripped title, got %+v", items[0])
	}
}

func TestParseRSSFeedGoogleNewsCollapsesYahooFinanceLocales(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
		<rss version="2.0">
		  <channel>
		    <title>Google News</title>
		    <item>
		      <guid>story-1</guid>
		      <title>Oil steadies as traders watch OPEC+ output - Yahoo Finance Singapore</title>
		      <link>https://news.google.com/rss/articles/example-yf?oc=5</link>
		      <pubDate>Mon, 06 Apr 2026 09:04:00 GMT</pubDate>
		      <description><![CDATA[<a href="https://news.google.com/rss/articles/example-yf?oc=5">Oil steadies as traders watch OPEC+ output</a>&nbsp;&nbsp;<font color="#6f6f6f">Yahoo Finance Singapore</font>]]></description>
		      <source url="https://finance.yahoo.com">Yahoo Finance Singapore</source>
		    </item>
		  </channel>
		</rss>`)

	items, err := parseRSSFeed(body, FeedSource{Name: "Yahoo Finance", URL: "https://news.google.com/rss/search?q=yahoo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one parsed item, got %d", len(items))
	}
	if items[0].Publisher != "Yahoo Finance" {
		t.Fatalf("expected collapsed Yahoo Finance publisher, got %+v", items[0])
	}
	if items[0].Title != "Oil steadies as traders watch OPEC+ output" {
		t.Fatalf("expected stripped Yahoo title, got %+v", items[0])
	}
}

func TestParseFeedTimeSupportsInvestingTimestampFormat(t *testing.T) {
	got := parseFeedTime("2026-04-06 10:44:11")
	if got.IsZero() {
		t.Fatal("expected investing.com timestamp to parse")
	}
	if got.Year() != 2026 || got.Month() != time.April || got.Day() != 6 {
		t.Fatalf("unexpected parsed time %v", got)
	}
}

func TestNormalizePublisherNameCollapsesAPNews(t *testing.T) {
	if got := normalizePublisherName("AP News"); got != "AP" {
		t.Fatalf("expected AP News to collapse to AP, got %q", got)
	}
	if got := normalizePublisherName("Associated Press"); got != "AP" {
		t.Fatalf("expected Associated Press to collapse to AP, got %q", got)
	}
	if got := normalizePublisherName("BBC News"); got != "BBC" {
		t.Fatalf("expected BBC News to collapse to BBC, got %q", got)
	}
	if got := normalizePublisherName("The Guardian"); got != "Guardian" {
		t.Fatalf("expected The Guardian to collapse to Guardian, got %q", got)
	}
}
