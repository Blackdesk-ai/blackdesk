package rss

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
}
