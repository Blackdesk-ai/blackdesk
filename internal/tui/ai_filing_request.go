package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"blackdesk/internal/domain"
)

type filingTextChunk struct {
	Index int
	Total int
	Start int
	End   int
	Text  string
}

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

type filingChunkAnalysisContext struct {
	Symbol          string            `json:"symbol"`
	CompanyName     string            `json:"company_name,omitempty"`
	CIK             string            `json:"cik,omitempty"`
	Form            string            `json:"form"`
	FilingDate      string            `json:"filing_date,omitempty"`
	ReportDate      string            `json:"report_date,omitempty"`
	AcceptedAt      string            `json:"accepted_at,omitempty"`
	Document        string            `json:"document,omitempty"`
	Description     string            `json:"description,omitempty"`
	URL             string            `json:"url,omitempty"`
	ContentType     string            `json:"content_type,omitempty"`
	SourceTruncated bool              `json:"source_truncated"`
	ChunkIndex      int               `json:"chunk_index"`
	ChunkCount      int               `json:"chunk_count"`
	ChunkStartRune  int               `json:"chunk_start_rune"`
	ChunkEndRune    int               `json:"chunk_end_rune"`
	Text            string            `json:"text"`
	Guide           map[string]string `json:"guide"`
}

type filingChunkAnalysisSummary struct {
	ChunkIndex int    `json:"chunk_index"`
	ChunkRange string `json:"chunk_range"`
	Analysis   string `json:"analysis"`
}

type filingChunkSynthesisContext struct {
	Symbol          string                       `json:"symbol"`
	CompanyName     string                       `json:"company_name,omitempty"`
	CIK             string                       `json:"cik,omitempty"`
	Form            string                       `json:"form"`
	FilingDate      string                       `json:"filing_date,omitempty"`
	ReportDate      string                       `json:"report_date,omitempty"`
	AcceptedAt      string                       `json:"accepted_at,omitempty"`
	Document        string                       `json:"document,omitempty"`
	Description     string                       `json:"description,omitempty"`
	URL             string                       `json:"url,omitempty"`
	ContentType     string                       `json:"content_type,omitempty"`
	SourceTruncated bool                         `json:"source_truncated"`
	ChunkCount      int                          `json:"chunk_count"`
	ChunkAnalyses   []filingChunkAnalysisSummary `json:"chunk_analyses"`
	Guide           map[string]string            `json:"guide"`
}

func filingAnalysisPrompt(symbol string, item domain.FilingItem) string {
	form := strings.TrimSpace(item.Form)
	if form == "" {
		form = "SEC filing"
	}
	return fmt.Sprintf("Analyze the selected %s filing for %s and produce an investor-focused report on what matters.", form, strings.ToUpper(strings.TrimSpace(symbol)))
}

func splitFilingTextChunks(text string, size, overlap int) []filingTextChunk {
	text = strings.TrimSpace(text)
	if text == "" || size <= 0 {
		return nil
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = max(0, size/10)
	}
	step := size - overlap
	if step <= 0 {
		step = size
	}

	runes := []rune(text)
	ranges := make([][2]int, 0, max(1, (len(runes)+step-1)/step))
	for start := 0; start < len(runes); start += step {
		end := min(len(runes), start+size)
		ranges = append(ranges, [2]int{start, end})
		if end == len(runes) {
			break
		}
	}

	chunks := make([]filingTextChunk, 0, len(ranges))
	for i, bounds := range ranges {
		chunks = append(chunks, filingTextChunk{
			Index: i + 1,
			Total: len(ranges),
			Start: bounds[0],
			End:   bounds[1],
			Text:  string(runes[bounds[0]:bounds[1]]),
		})
	}
	return chunks
}

func (m Model) buildAIFilingAnalysisRequest(symbol string, snapshot domain.FilingsSnapshot, filing domain.FilingDocument, prompt string) (RequestEnvelope, error) {
	truncation := aiRequestTruncation{}
	filingText, filingTextTruncated := truncateRunesFlag(strings.TrimSpace(filing.Text), aiFilingDocumentChars)
	truncation.FilingText = filing.Truncated || filingTextTruncated
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
	b.WriteString("If the filing text appears incomplete or truncated, say so clearly.\n")
	b.WriteString("Base this first response only on the selected filing payload below, not on the broader app context.\n\n")
	b.WriteString("<selected_filing>\n")
	b.WriteString(string(filingPayload))
	b.WriteString("\n</selected_filing>\n")
	systemPrompt, promptTruncated := truncateRunesFlag(b.String(), aiFilingPromptChars)
	truncation.FinalPrompt = promptTruncated

	return RequestEnvelope{
		Prompt:          strings.TrimSpace(prompt),
		SystemPrompt:    systemPrompt,
		ContextPayload:  "",
		ActiveSymbol:    strings.ToUpper(strings.TrimSpace(symbol)),
		ContextRevision: 0,
		Truncation:      truncation,
	}, nil
}

func (m Model) buildAIFilingChunkAnalysisRequest(symbol string, snapshot domain.FilingsSnapshot, filing domain.FilingDocument, chunk filingTextChunk) (RequestEnvelope, error) {
	truncation := aiRequestTruncation{FilingText: filing.Truncated}
	payload, err := json.MarshalIndent(filingChunkAnalysisContext{
		Symbol:          strings.ToUpper(strings.TrimSpace(symbol)),
		CompanyName:     strings.TrimSpace(snapshot.CompanyName),
		CIK:             strings.TrimSpace(snapshot.CIK),
		Form:            strings.TrimSpace(filing.Item.Form),
		FilingDate:      filingDateLabel(filing.Item),
		ReportDate:      filingReportDateLabel(filing.Item),
		AcceptedAt:      filingAcceptedAtLabel(filing.Item),
		Document:        strings.TrimSpace(filing.Item.PrimaryDocument),
		Description:     strings.TrimSpace(filing.Item.PrimaryDocDescription),
		URL:             strings.TrimSpace(filing.Item.URL),
		ContentType:     strings.TrimSpace(filing.ContentType),
		SourceTruncated: filing.Truncated,
		ChunkIndex:      chunk.Index,
		ChunkCount:      chunk.Total,
		ChunkStartRune:  chunk.Start,
		ChunkEndRune:    chunk.End,
		Text:            strings.TrimSpace(chunk.Text),
		Guide: map[string]string{
			"text":             "Sequential chunk from the selected SEC filing. Adjacent chunks may overlap slightly for continuity.",
			"chunk_index":      "1-based position of this chunk inside the filing.",
			"chunk_count":      "Total number of chunks extracted from the filing text.",
			"chunk_start_rune": "Approximate inclusive rune offset where this chunk begins in the filing text.",
			"chunk_end_rune":   "Approximate exclusive rune offset where this chunk ends in the filing text.",
			"source_truncated": "True when the SEC filing body itself was already clipped before chunking.",
		},
	}, "", "  ")
	if err != nil {
		return RequestEnvelope{}, err
	}

	var b strings.Builder
	b.WriteString(strings.TrimSpace(aiSystemPromptTemplate))
	b.WriteString("\n\n")
	b.WriteString("You are reviewing one chunk from a selected SEC filing.\n")
	b.WriteString("This is not the final report. Extract the investor-relevant facts from this chunk only.\n")
	b.WriteString("Adjacent chunks may overlap slightly. Do not treat repeated overlap lines as separate new facts.\n")
	b.WriteString("Keep the response compact and structured with these sections in order:\n")
	b.WriteString("1. What This Chunk Covers\n")
	b.WriteString("2. Material Facts And Disclosures\n")
	b.WriteString("3. Numbers, Dates, And Thresholds\n")
	b.WriteString("4. Positives\n")
	b.WriteString("5. Negatives Or Risks\n")
	b.WriteString("6. Open Threads For Later Chunks\n")
	b.WriteString("Use bullets, quote exact figures when present, and stay concise enough to fit comfortably into a later synthesis step.\n")
	if filing.Truncated {
		b.WriteString("The source filing text was already clipped before chunking, so note incomplete areas if relevant.\n")
	}
	b.WriteString("\n<selected_filing_chunk>\n")
	b.WriteString(string(payload))
	b.WriteString("\n</selected_filing_chunk>\n")
	systemPrompt, promptTruncated := truncateRunesFlag(b.String(), aiFilingPromptChars)
	truncation.FinalPrompt = promptTruncated

	return RequestEnvelope{
		Prompt:          fmt.Sprintf("Review filing chunk %d of %d for %s and extract the material investor-relevant facts from this chunk.", chunk.Index, chunk.Total, strings.ToUpper(strings.TrimSpace(symbol))),
		SystemPrompt:    systemPrompt,
		ContextPayload:  "",
		ActiveSymbol:    strings.ToUpper(strings.TrimSpace(symbol)),
		ContextRevision: 0,
		Truncation:      truncation,
	}, nil
}

func (m Model) buildAIFilingSynthesisRequest(symbol string, snapshot domain.FilingsSnapshot, filing domain.FilingDocument, prompt string, analyses []filingChunkAnalysisSummary) (RequestEnvelope, error) {
	truncation := aiRequestTruncation{FilingText: filing.Truncated}
	collapsed := make([]filingChunkAnalysisSummary, 0, len(analyses))
	for _, item := range analyses {
		item.Analysis = truncateRunes(strings.TrimSpace(item.Analysis), aiFilingChunkSummaryChars)
		if item.Analysis == "" {
			continue
		}
		collapsed = append(collapsed, item)
	}
	payload, err := json.MarshalIndent(filingChunkSynthesisContext{
		Symbol:          strings.ToUpper(strings.TrimSpace(symbol)),
		CompanyName:     strings.TrimSpace(snapshot.CompanyName),
		CIK:             strings.TrimSpace(snapshot.CIK),
		Form:            strings.TrimSpace(filing.Item.Form),
		FilingDate:      filingDateLabel(filing.Item),
		ReportDate:      filingReportDateLabel(filing.Item),
		AcceptedAt:      filingAcceptedAtLabel(filing.Item),
		Document:        strings.TrimSpace(filing.Item.PrimaryDocument),
		Description:     strings.TrimSpace(filing.Item.PrimaryDocDescription),
		URL:             strings.TrimSpace(filing.Item.URL),
		ContentType:     strings.TrimSpace(filing.ContentType),
		SourceTruncated: filing.Truncated,
		ChunkCount:      len(analyses),
		ChunkAnalyses:   collapsed,
		Guide: map[string]string{
			"chunk_analyses":   "Ordered chunk-level analyses covering the selected filing from start to finish.",
			"source_truncated": "True when the SEC filing body itself was already clipped before chunking.",
			"chunk_count":      "Total number of analyzed filing chunks included below.",
		},
	}, "", "  ")
	if err != nil {
		return RequestEnvelope{}, err
	}

	var b strings.Builder
	b.WriteString(strings.TrimSpace(aiSystemPromptTemplate))
	b.WriteString("\n\n")
	b.WriteString("You are synthesizing the full selected SEC filing from ordered chunk analyses.\n")
	b.WriteString("Treat the chunk analyses as a compressed digest of the full filing text.\n")
	b.WriteString("Deduplicate repeated overlap facts, reconcile repeated disclosures, and produce one unified investor-focused report.\n")
	b.WriteString("Write a structured report with these sections in order:\n")
	b.WriteString("1. What Was Filed\n")
	b.WriteString("2. Key Takeaways\n")
	b.WriteString("3. Important Positives\n")
	b.WriteString("4. Important Negatives Or Risks\n")
	b.WriteString("5. Numbers, Dates, And Disclosures To Note\n")
	b.WriteString("6. Follow-Up Questions\n")
	b.WriteString("7. Bottom Line\n")
	b.WriteString("Cite exact figures, dates, and concrete disclosures from the chunk analyses when possible.\n")
	b.WriteString("If the filing is an insider, ownership, or governance filing instead of a full report, adapt the analysis to that filing type instead of forcing an earnings-style summary.\n")
	if filing.Truncated {
		b.WriteString("The source filing text was clipped before analysis, so say clearly where coverage may be incomplete.\n")
	}
	b.WriteString("\n<selected_filing_synthesis>\n")
	b.WriteString(string(payload))
	b.WriteString("\n</selected_filing_synthesis>\n")
	systemPrompt, promptTruncated := truncateRunesFlag(b.String(), aiFilingPromptChars)
	truncation.FinalPrompt = promptTruncated

	return RequestEnvelope{
		Prompt:          strings.TrimSpace(prompt),
		SystemPrompt:    systemPrompt,
		ContextPayload:  "",
		ActiveSymbol:    strings.ToUpper(strings.TrimSpace(symbol)),
		ContextRevision: 0,
		Truncation:      truncation,
	}, nil
}

func filingChunkRangeLabel(chunk filingTextChunk) string {
	return fmt.Sprintf("chars %d-%d", chunk.Start+1, chunk.End)
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
