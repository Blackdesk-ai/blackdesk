package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"blackdesk/internal/domain"
)

type filingAnalysisContext struct {
	Symbol      string            `json:"symbol"`
	CompanyName string            `json:"company_name,omitempty"`
	CIK         string            `json:"cik,omitempty"`
	Form        string            `json:"form"`
	FilingDate  string            `json:"filing_date,omitempty"`
	ReportDate  string            `json:"report_date,omitempty"`
	AcceptedAt  string            `json:"accepted_at,omitempty"`
	Document    string            `json:"document,omitempty"`
	Description string            `json:"description,omitempty"`
	URL         string            `json:"url,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Truncated   bool              `json:"truncated"`
	Text        string            `json:"text"`
	Guide       map[string]string `json:"guide"`
}

func filingAnalysisPrompt(symbol string, item domain.FilingItem) string {
	form := strings.TrimSpace(item.Form)
	if form == "" {
		form = "SEC filing"
	}
	return fmt.Sprintf("Analyze the selected %s filing for %s and produce an investor-focused report on what matters.", form, strings.ToUpper(strings.TrimSpace(symbol)))
}

func (m Model) buildAIFilingAnalysisRequest(symbol string, snapshot domain.FilingsSnapshot, filing domain.FilingDocument, prompt string) (RequestEnvelope, error) {
	ctxPayload, err := json.MarshalIndent(m.aiContextSnapshot(), "", "  ")
	if err != nil {
		return RequestEnvelope{}, err
	}
	rawPayload := string(ctxPayload)
	payload, payloadTruncated := truncateRunesFlag(rawPayload, aiFilingContextChars)
	truncation := aiRequestTruncation{ContextPayload: payloadTruncated}
	filingText, filingTextTruncated := truncateRunesFlag(strings.TrimSpace(filing.Text), aiFilingDocumentChars)
	truncation.FilingText = filingTextTruncated
	filingPayload, err := json.MarshalIndent(filingAnalysisContext{
		Symbol:      strings.ToUpper(strings.TrimSpace(symbol)),
		CompanyName: strings.TrimSpace(snapshot.CompanyName),
		CIK:         strings.TrimSpace(snapshot.CIK),
		Form:        strings.TrimSpace(filing.Item.Form),
		FilingDate:  filingDateLabel(filing.Item),
		ReportDate:  filingReportDateLabel(filing.Item),
		AcceptedAt:  filingAcceptedAtLabel(filing.Item),
		Document:    strings.TrimSpace(filing.Item.PrimaryDocument),
		Description: strings.TrimSpace(filing.Item.PrimaryDocDescription),
		URL:         strings.TrimSpace(filing.Item.URL),
		ContentType: strings.TrimSpace(filing.ContentType),
		Truncated:   filing.Truncated || filingTextTruncated,
		Text:        filingText,
		Guide: map[string]string{
			"form":        "SEC form code for the selected filing. Adapt the analysis to the actual filing type.",
			"text":        "Extracted plain text from the selected SEC filing. This is the primary source to analyze.",
			"truncated":   "True when the filing body had to be clipped for transport. If true, note that the source text may be incomplete.",
			"report_date": "Underlying reporting period when the filing exposes one.",
			"accepted_at": "EDGAR acceptance timestamp when available.",
		},
	}, "", "  ")
	if err != nil {
		return RequestEnvelope{}, err
	}

	var b strings.Builder
	b.WriteString(strings.TrimSpace(aiSystemPromptTemplate))
	b.WriteString("\n\n")
	b.WriteString("You are analyzing a selected SEC filing for the active company.\n")
	b.WriteString("Prioritize the filing text over generic heuristics, and pull out the facts that matter most to an investor or analyst.\n")
	b.WriteString("Write a structured report with these sections in order:\n")
	b.WriteString("1. What Was Filed\n")
	b.WriteString("2. Key Takeaways\n")
	b.WriteString("3. Important Positives\n")
	b.WriteString("4. Important Negatives Or Risks\n")
	b.WriteString("5. Numbers, Dates, And Disclosures To Note\n")
	b.WriteString("6. Follow-Up Questions\n")
	b.WriteString("7. Bottom Line\n")
	b.WriteString("Cite exact figures, dates, and concrete disclosures from the filing when possible.\n")
	b.WriteString("If the filing is an insider, ownership, or governance filing instead of a full report, adapt the analysis to that filing type instead of forcing an earnings-style summary.\n")
	b.WriteString("If the filing text appears incomplete or truncated, say so clearly.\n\n")
	b.WriteString("<blackdesk_context_update>\n")
	b.WriteString(payload)
	b.WriteString("\n</blackdesk_context_update>\n\n")
	b.WriteString("<selected_filing>\n")
	b.WriteString(string(filingPayload))
	b.WriteString("\n</selected_filing>\n")
	systemPrompt, promptTruncated := truncateRunesFlag(b.String(), aiFilingPromptChars)
	truncation.FinalPrompt = promptTruncated

	return RequestEnvelope{
		Prompt:          strings.TrimSpace(prompt),
		SystemPrompt:    systemPrompt,
		ContextPayload:  payload + "\n\n<selected_filing>\n" + string(filingPayload) + "\n</selected_filing>",
		ActiveSymbol:    strings.ToUpper(strings.TrimSpace(symbol)),
		ContextRevision: m.aiContextRevision,
		Truncation:      truncation,
	}, nil
}

func filingReportDateLabel(item domain.FilingItem) string {
	if item.ReportDate.IsZero() {
		return ""
	}
	return item.ReportDate.Format("2006-01-02")
}

func filingAcceptedAtLabel(item domain.FilingItem) string {
	if item.AcceptedAt.IsZero() {
		return ""
	}
	return item.AcceptedAt.Format("2006-01-02 15:04:05")
}
