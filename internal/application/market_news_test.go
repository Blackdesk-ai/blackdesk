package application

import (
	"fmt"
	"testing"
	"time"

	"blackdesk/internal/domain"
)

func TestFilterRecentMarketNewsKeepsLast48HoursAndZeroTimestamp(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	items := []domain.NewsItem{
		{Title: "Recent", Time: time.Date(2026, 4, 6, 9, 15, 0, 0, time.UTC)},
		{Title: "Within 48h", Time: time.Date(2026, 4, 5, 20, 30, 0, 0, time.UTC)},
		{Title: "Within 48h edge", Time: time.Date(2026, 4, 4, 13, 0, 0, 0, time.UTC)},
		{Title: "Old", Time: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)},
		{Title: "No timestamp"},
	}

	got := FilterRecentMarketNews(items, now)
	if len(got) != 4 {
		t.Fatalf("expected 4 items (3 recent + 1 no timestamp), got %d", len(got))
	}
	if got[0].Title != "Recent" || got[1].Title != "Within 48h" || got[2].Title != "Within 48h edge" || got[3].Title != "No timestamp" {
		t.Fatalf("unexpected filter output: %+v", got)
	}
}

func TestMergeMarketNewsTracksFreshItemsAndAppliesSuccessStatus(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	prev := []domain.NewsItem{
		{Title: "Old one", Publisher: "Desk", URL: "https://example.com/1", Time: now.Add(-2 * time.Hour)},
	}
	incoming := []domain.NewsItem{
		prev[0],
		{Title: "Fresh", Publisher: "Desk", URL: "https://example.com/2", Time: now.Add(-30 * time.Minute)},
	}

	got := MergeMarketNews(MarketNewsMergeInput{
		PreviousItems: prev,
		IncomingItems: incoming,
		SelectedIndex: 9,
		Now:           now,
	})

	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}
	if _, ok := got.Fresh["https://example.com/2"]; !ok {
		t.Fatal("expected fresh item to be tracked")
	}
	if got.Status != "Loaded 2 market headlines" || !got.ApplyStatus {
		t.Fatalf("unexpected status result: apply=%v status=%q", got.ApplyStatus, got.Status)
	}
	if got.SelectedIndex != 0 {
		t.Fatalf("expected selection repaired to 0, got %d", got.SelectedIndex)
	}
}

func TestMergeMarketNewsPreservesVisibleItemsOnError(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	incoming := []domain.NewsItem{
		{Title: "Still visible", Publisher: "Desk", URL: "https://example.com/1", Time: now.Add(-30 * time.Minute)},
	}

	got := MergeMarketNews(MarketNewsMergeInput{
		IncomingItems: incoming,
		Now:           now,
		Err:           errFake("feed failed"),
	})

	if len(got.Items) != 1 {
		t.Fatalf("expected items to remain visible, got %d", len(got.Items))
	}
	if got.ApplyStatus {
		t.Fatal("expected no status override when items still exist on error")
	}
}

func TestMergeMarketNewsCapsVisibleItemsAt80(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	incoming := make([]domain.NewsItem, 0, 90)
	for i := 0; i < 90; i++ {
		incoming = append(incoming, domain.NewsItem{
			Title:     fmt.Sprintf("Story %02d", i),
			Publisher: "Desk",
			URL:       fmt.Sprintf("https://example.com/%02d", i),
			Time:      now.Add(-time.Duration(i) * time.Minute),
		})
	}

	got := MergeMarketNews(MarketNewsMergeInput{
		IncomingItems: incoming,
		Now:           now,
	})

	if len(got.Items) != 80 {
		t.Fatalf("expected 80 visible items, got %d", len(got.Items))
	}
	if got.Items[0].Title != "Story 00" || got.Items[79].Title != "Story 79" {
		t.Fatalf("unexpected cap boundaries first=%q last=%q", got.Items[0].Title, got.Items[79].Title)
	}
}

type errFake string

func (e errFake) Error() string { return string(e) }
