package tui

import "blackdesk/internal/domain"

func filingDateLabel(item domain.FilingItem) string {
	if !item.FilingDate.IsZero() {
		return item.FilingDate.Format("2006-01-02")
	}
	if !item.AcceptedAt.IsZero() {
		return item.AcceptedAt.Format("2006-01-02")
	}
	return "-"
}
