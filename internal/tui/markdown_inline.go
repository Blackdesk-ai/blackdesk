package tui

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	markdownLinkPattern   = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	markdownBoldPattern   = regexp.MustCompile(`\*\*(.*?)\*\*|__(.*?)__`)
	markdownItalicPattern = regexp.MustCompile(`\*(.*?)\*|_(.*?)_`)
	markdownCodePattern   = regexp.MustCompile("`([^`]+)`")
	markdownURLPattern    = regexp.MustCompile(`https?://[^\s<>()]+`)
	gluedURLPattern       = regexp.MustCompile(`([^\s\[(])((?:https?://))`)
)

func renderMarkdownInline(line string) string {
	line = markdownLinkPattern.ReplaceAllStringFunc(line, func(s string) string {
		matches := markdownLinkPattern.FindStringSubmatch(s)
		if len(matches) < 3 {
			return s
		}
		label := strings.TrimSpace(matches[1])
		url := strings.TrimSpace(matches[2])
		if label == "" {
			label = url
		}
		return label + ": " + terminalHyperlink(url, url)
	})
	line = markdownCodePattern.ReplaceAllString(line, "$1")
	line = markdownBoldPattern.ReplaceAllStringFunc(line, func(s string) string {
		matches := markdownBoldPattern.FindStringSubmatch(s)
		for _, part := range matches[1:] {
			if part != "" {
				return part
			}
		}
		return s
	})
	line = markdownItalicPattern.ReplaceAllStringFunc(line, func(s string) string {
		matches := markdownItalicPattern.FindStringSubmatch(s)
		for _, part := range matches[1:] {
			if part != "" {
				return part
			}
		}
		return s
	})
	line = markdownURLPattern.ReplaceAllStringFunc(line, func(s string) string {
		return terminalHyperlink(compactURLLabel(s), s)
	})
	return strings.TrimSpace(line)
}

func normalizeMarkdownInput(input string) string {
	replacer := strings.NewReplacer(
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"<BR>", "\n",
		"<BR/>", "\n",
		"<BR />", "\n",
	)
	input = replacer.Replace(input)
	return gluedURLPattern.ReplaceAllString(input, "$1 $2")
}

func terminalHyperlink(label, url string) string {
	label = strings.TrimSpace(label)
	url = strings.TrimSpace(url)
	if label == "" {
		label = url
	}
	if url == "" {
		return label
	}
	return "\x1b]8;;" + url + "\x1b\\" + label + "\x1b]8;;\x1b\\"
}

func compactURLLabel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	if host == "" {
		return raw
	}
	path := strings.Trim(u.EscapedPath(), "/")
	if path == "" {
		return host
	}
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		parts = parts[:2]
	}
	label := host + "/" + strings.Join(parts, "/")
	if u.RawQuery != "" {
		label += "?"
	}
	return label
}
