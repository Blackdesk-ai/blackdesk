package sec

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

const maxFilingDocumentBytes = 8 << 20

var (
	secScriptStyleRE = regexp.MustCompile(`(?is)<(script|style)\b[^>]*>.*?</(script|style)>`)
	secCommentRE     = regexp.MustCompile(`(?s)<!--.*?-->`)
	secTagRE         = regexp.MustCompile(`(?s)<[^>]+>`)
	secSpaceRE       = regexp.MustCompile(`[ \t\f\v]+`)
	secBlankLineRE   = regexp.MustCompile(`\n{3,}`)
)

func (p *Provider) GetFilingDocument(ctx context.Context, item domain.FilingItem) (domain.FilingDocument, error) {
	rawURL := strings.TrimSpace(item.URL)
	if rawURL == "" {
		return domain.FilingDocument{}, fmt.Errorf("filing URL is required")
	}

	cacheKey := "sec:filing_document:" + rawURL
	if p.cache != nil {
		var cached string
		if p.cache.Get(cacheKey, &cached) && strings.TrimSpace(cached) != "" {
			return domain.FilingDocument{
				Item:        item,
				Text:        cached,
				Provider:    "sec",
				RetrievedAt: time.Now(),
			}, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return domain.FilingDocument{}, err
	}
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := p.client.Do(req)
	if err != nil {
		return domain.FilingDocument{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return domain.FilingDocument{}, fmt.Errorf("SEC filing request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFilingDocumentBytes+1))
	if err != nil {
		return domain.FilingDocument{}, err
	}
	truncated := len(body) > maxFilingDocumentBytes
	if truncated {
		body = body[:maxFilingDocumentBytes]
	}

	text := extractFilingText(body)
	if strings.TrimSpace(text) == "" {
		return domain.FilingDocument{}, fmt.Errorf("SEC filing text was empty")
	}
	if p.cache != nil {
		p.cache.Set(cacheKey, text, 6*time.Hour)
	}
	return domain.FilingDocument{
		Item:        item,
		ContentType: strings.TrimSpace(resp.Header.Get("Content-Type")),
		Text:        text,
		Truncated:   truncated,
		Provider:    "sec",
		RetrievedAt: time.Now(),
	}, nil
}

func extractFilingText(body []byte) string {
	text := string(bytes.ToValidUTF8(body, []byte(" ")))
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = secScriptStyleRE.ReplaceAllString(text, "\n")
	text = secCommentRE.ReplaceAllString(text, "\n")
	text = secTagRE.ReplaceAllString(text, "\n")
	text = html.UnescapeString(text)

	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = secSpaceRE.ReplaceAllString(strings.TrimSpace(line), " ")
		if line != "" {
			out = append(out, line)
		}
	}
	text = strings.Join(out, "\n")
	text = secBlankLineRE.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}
