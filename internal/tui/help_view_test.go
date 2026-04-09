package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderHelpEntryLineAlignsUnicodeAndAsciiKeys(t *testing.T) {
	keyStyle := lipgloss.NewStyle()
	descStyle := lipgloss.NewStyle()

	lineArrow := ansi.Strip(renderHelpEntryLine(keyStyle, descStyle, "↑ / ↓", "Navigate results", 14))
	lineEnter := ansi.Strip(renderHelpEntryLine(keyStyle, descStyle, "Enter", "Submit query", 14))
	lineCtrl := ansi.Strip(renderHelpEntryLine(keyStyle, descStyle, "Ctrl+A", "Add symbol", 14))

	arrowIdx := strings.Index(lineArrow, "Navigate results")
	enterIdx := strings.Index(lineEnter, "Submit query")
	ctrlIdx := strings.Index(lineCtrl, "Add symbol")

	if arrowIdx <= 0 || enterIdx <= 0 || ctrlIdx <= 0 {
		t.Fatalf("expected descriptions to be present: %q | %q | %q", lineArrow, lineEnter, lineCtrl)
	}

	arrowCol := lipgloss.Width(lineArrow[:arrowIdx])
	enterCol := lipgloss.Width(lineEnter[:enterIdx])
	ctrlCol := lipgloss.Width(lineCtrl[:ctrlIdx])
	if arrowCol != enterCol || enterCol != ctrlCol {
		t.Fatalf("expected aligned help descriptions, got columns %d, %d, %d", arrowCol, enterCol, ctrlCol)
	}
}
