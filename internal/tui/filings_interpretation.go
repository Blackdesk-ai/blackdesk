package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func filingDisplayTitle(item domain.FilingItem) string {
	switch strings.ToUpper(strings.TrimSpace(item.Form)) {
	case "10-K":
		return "Annual report"
	case "10-Q":
		return "Quarterly report"
	case "8-K":
		return "Current report"
	case "3":
		return "Initial insider ownership"
	case "4":
		return "Insider trade update"
	case "5":
		return "Annual insider ownership"
	case "DEF 14A":
		return "Proxy statement"
	case "13D":
		return "Activist / beneficial ownership"
	case "13G":
		return "Passive beneficial ownership"
	default:
		if desc := strings.TrimSpace(item.PrimaryDocDescription); desc != "" && !strings.HasSuffix(strings.ToLower(desc), ".xml") {
			return normalizeFilingText(desc)
		}
		return "SEC filing"
	}
}

func filingMeaning(item domain.FilingItem) string {
	switch strings.ToUpper(strings.TrimSpace(item.Form)) {
	case "10-K":
		return "Annual company report with audited financials, business overview, risks, and management discussion."
	case "10-Q":
		return "Quarterly company report with updated financials and management discussion."
	case "8-K":
		return "Current report used to disclose material events such as earnings, deals, leadership changes, or other important updates."
	case "3":
		return "Initial ownership filing submitted when an insider first becomes subject to reporting."
	case "4":
		return "Insider transaction filing showing buys, sales, awards, exercises, or other ownership changes."
	case "5":
		return "Annual insider filing for transactions or holdings that were not reported earlier on Form 4."
	case "DEF 14A":
		return "Proxy statement covering shareholder voting items, executive pay, and governance matters."
	case "13D":
		return "Beneficial ownership filing typically used when an investor takes an active stake."
	case "13G":
		return "Beneficial ownership filing typically used for passive holdings."
	default:
		return "Recent SEC filing available through EDGAR."
	}
}

func filingDocumentLabel(item domain.FilingItem) string {
	if desc := normalizeFilingText(item.PrimaryDocDescription); desc != "" && !strings.HasSuffix(strings.ToLower(desc), ".xml") {
		return desc
	}
	if doc := strings.TrimSpace(item.PrimaryDocument); doc != "" {
		return doc
	}
	return "-"
}

func normalizeFilingText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "_", " ")
	return strings.Join(strings.Fields(text), " ")
}
