package tui

import "strings"

func renderMarkdownTable(lines []string, width int) ([]string, int, bool) {
	if len(lines) < 2 {
		return nil, 0, false
	}
	header := strings.TrimSpace(lines[0])
	divider := strings.TrimSpace(lines[1])
	if !isMarkdownTableRow(header) || !isMarkdownTableDivider(divider) {
		return nil, 0, false
	}

	headers := parseMarkdownTableCells(header)
	rows := make([][]string, 0, 8)
	consumed := 2
	for consumed < len(lines) {
		line := strings.TrimSpace(lines[consumed])
		if line == "" || !isMarkdownTableRow(line) {
			break
		}
		rows = append(rows, parseMarkdownTableCells(line))
		consumed++
	}
	if len(rows) == 0 {
		return nil, 0, false
	}

	out := make([]string, 0, len(rows)*(len(headers)+2))
	for rowIdx, row := range rows {
		if rowIdx > 0 {
			out = append(out, "")
		}
		title := firstNonEmptyCell(row)
		if title == "" {
			title = "Row"
		}
		out = append(out, title)
		out = append(out, strings.Repeat("-", min(width, max(3, len([]rune(title))))))
		for i := 0; i < min(len(headers), len(row)); i++ {
			label := strings.TrimSpace(headers[i])
			value := strings.TrimSpace(row[i])
			if label == "" || value == "" {
				continue
			}
			if i == 0 && value == title {
				continue
			}
			out = append(out, label+": "+renderMarkdownInline(value))
		}
	}
	return out, consumed, true
}

func isMarkdownTableRow(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") && strings.Count(line, "|") >= 2
}

func isMarkdownTableDivider(line string) bool {
	if !isMarkdownTableRow(line) {
		return false
	}
	for _, cell := range parseMarkdownTableCells(line) {
		cell = strings.ReplaceAll(cell, "-", "")
		cell = strings.ReplaceAll(cell, ":", "")
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func parseMarkdownTableCells(line string) []string {
	line = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|"))
	parts := strings.Split(line, "|")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.TrimSpace(part))
	}
	return out
}

func firstNonEmptyCell(cells []string) string {
	for _, cell := range cells {
		if strings.TrimSpace(cell) != "" {
			return strings.TrimSpace(cell)
		}
	}
	return ""
}
