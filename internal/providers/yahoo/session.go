package yahoo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type session struct {
	mu    sync.RWMutex
	fetch sync.Mutex
	crumb string
}

func newSession() *session {
	return &session{}
}

func (s *session) getCrumb() (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.crumb, s.crumb != ""
}

func (s *session) setCrumb(crumb string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.crumb = crumb
}

func (p *Provider) clearCrumb() {
	p.session.setCrumb("")
}

func (p *Provider) ensureCredentials(ctx context.Context) error {
	if _, ok := p.session.getCrumb(); ok {
		return nil
	}

	p.session.fetch.Lock()
	defer p.session.fetch.Unlock()

	if _, ok := p.session.getCrumb(); ok {
		return nil
	}

	if err := p.fetchCookie(ctx); err != nil {
		return err
	}
	crumb, err := p.fetchCrumb(ctx)
	if err != nil {
		return err
	}
	p.session.setCrumb(crumb)
	return nil
}

func (p *Provider) fetchCookie(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cookieURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", defaultUA)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if len(resp.Cookies()) == 0 {
		if resp.StatusCode >= 400 {
			return fmt.Errorf("yahoo cookie bootstrap returned %d with no cookies", resp.StatusCode)
		}
		return fmt.Errorf("yahoo cookie bootstrap returned no cookies")
	}
	return nil
}

func (p *Provider) fetchCrumb(ctx context.Context) (string, error) {
	body, status, err := p.get(ctx, p.crumbURL)
	if err != nil {
		return "", err
	}
	if status >= 400 {
		return "", fmt.Errorf("yahoo crumb bootstrap returned %d", status)
	}
	crumb := strings.TrimSpace(string(body))
	if crumb == "" || strings.Contains(crumb, "<") || strings.Contains(crumb, "{") || strings.Contains(strings.ToLower(crumb), "too many requests") {
		return "", fmt.Errorf("invalid yahoo crumb response")
	}
	return crumb, nil
}

func looksLikeInvalidCrumb(body []byte) bool {
	text := strings.ToLower(string(body))
	return strings.Contains(text, "invalid crumb") || strings.Contains(text, "edge: too many requests")
}
