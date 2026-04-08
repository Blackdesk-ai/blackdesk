package application

import (
	"testing"
	"time"

	"blackdesk/internal/domain"
)

func TestFilterRecentMarketNewsKeepsLast24HoursAndZeroTimestamp(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	items := []domain.NewsItem{
		{Title: "Recent", Time: time.Date(2026, 4, 6, 9, 15, 0, 0, time.UTC)},
		{Title: "Within 24h", Time: time.Date(2026, 4, 5, 20, 30, 0, 0, time.UTC)},
		{Title: "Old", Time: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)},
		{Title: "No timestamp"},
	}

	got := FilterRecentMarketNews(items, now)
	if len(got) != 3 {
		t.Fatalf("expected 3 items (2 recent + 1 no timestamp), got %d", len(got))
	}
	if got[0].Title != "Recent" || got[1].Title != "Within 24h" || got[2].Title != "No timestamp" {
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

type errFake string

func (e errFake) Error() string { return string(e) }
