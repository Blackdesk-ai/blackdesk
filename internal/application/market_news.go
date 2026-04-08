package application

import (
	"fmt"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

type MarketNewsMergeInput struct {
	PreviousItems   []domain.NewsItem
	PreviousSeen    map[string]struct{}
	PreviousSources []domain.MarketNewsSource
	IncomingItems   []domain.NewsItem
	IncomingSources []domain.MarketNewsSource
	SelectedIndex   int
	Now             time.Time
	Err             error
}

type MarketNewsMergeResult struct {
	Items         []domain.NewsItem
	Sources       []domain.MarketNewsSource
	Seen          map[string]struct{}
	Fresh         map[string]struct{}
	SelectedIndex int
	UpdatedAt     time.Time
	LastRefresh   time.Time
	Status        string
	ApplyStatus   bool
}

func FilterRecentMarketNews(items []domain.NewsItem, now time.Time) []domain.NewsItem {
	if len(items) == 0 {
		return nil
	}
	cutoff := now.Add(-24 * time.Hour)
	filtered := make([]domain.NewsItem, 0, len(items))
	for _, item := range items {
		if item.Time.IsZero() || item.Time.After(cutoff) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func MarketNewsIdentity(item domain.NewsItem) string {
	if key := strings.TrimSpace(item.URL); key != "" {
		return key
	}
	return strings.TrimSpace(item.Publisher) + "|" + strings.TrimSpace(item.Title)
}

func MergeMarketNews(input MarketNewsMergeInput) MarketNewsMergeResult {
	nextItems := FilterRecentMarketNews(input.IncomingItems, input.Now)
	if len(nextItems) > 50 {
		nextItems = nextItems[:50]
	}

	result := MarketNewsMergeResult{
		Items:         nextItems,
		Seen:          cloneStringSet(input.PreviousSeen),
		SelectedIndex: input.SelectedIndex,
	}

	if len(input.PreviousItems) > 0 {
		if result.Seen == nil {
			result.Seen = make(map[string]struct{}, len(input.PreviousItems))
		}
		for _, item := range input.PreviousItems {
			result.Seen[MarketNewsIdentity(item)] = struct{}{}
		}
		result.Fresh = make(map[string]struct{})
		for _, item := range nextItems {
			key := MarketNewsIdentity(item)
			if key == "" {
				continue
			}
			if _, ok := result.Seen[key]; !ok {
				result.Fresh[key] = struct{}{}
				result.Seen[key] = struct{}{}
			}
		}
		if len(result.Seen) > 2000 {
			result.Seen = make(map[string]struct{}, len(nextItems))
			for _, item := range nextItems {
				result.Seen[MarketNewsIdentity(item)] = struct{}{}
			}
		}
	}

	if len(input.IncomingSources) > 0 {
		result.Sources = append([]domain.MarketNewsSource(nil), input.IncomingSources...)
	} else if len(input.PreviousSources) > 0 {
		result.Sources = append([]domain.MarketNewsSource(nil), input.PreviousSources...)
	}

	if input.Err == nil {
		result.UpdatedAt = input.Now
		result.LastRefresh = input.Now
		if len(result.Items) > 0 {
			result.Status = fmt.Sprintf("Loaded %d market headlines", len(result.Items))
		} else {
			result.Status = "No recent market headlines"
		}
		result.ApplyStatus = true
	} else if len(result.Items) == 0 {
		result.Status = "Market news unavailable"
		result.ApplyStatus = true
	}

	if result.SelectedIndex >= len(result.Items) {
		result.SelectedIndex = 0
	}
	if result.SelectedIndex < 0 {
		result.SelectedIndex = 0
	}

	return result
}

func cloneStringSet(items map[string]struct{}) map[string]struct{} {
	if len(items) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(items))
	for key := range items {
		out[key] = struct{}{}
	}
	return out
}
