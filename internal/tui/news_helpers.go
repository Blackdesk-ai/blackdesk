package tui

import (
	"fmt"
	neturl "net/url"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

func (m Model) currentMarketNews() (domain.NewsItem, bool) {
	if len(m.marketNews) == 0 {
		return domain.NewsItem{}, false
	}
	idx := min(max(0, m.marketNewsSel), len(m.marketNews)-1)
	return m.marketNews[idx], true
}

func newsHost(rawURL string) string {
	if strings.TrimSpace(rawURL) == "" {
		return ""
	}
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(parsed.Hostname(), "www.")
}

func newsTimeLabel(ts time.Time) string {
	if ts.IsZero() {
		return "time n/a"
	}
	age := time.Since(ts)
	switch {
	case age < 0:
		return ts.Local().Format("Mon 15:04")
	case age < time.Minute:
		return "now"
	case age < time.Hour:
		return fmt.Sprintf("%dm ago", int(age.Minutes()))
	case age < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(age.Hours()))
	default:
		return ts.Local().Format("02 Jan")
	}
}
