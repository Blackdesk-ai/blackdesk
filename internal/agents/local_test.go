package agents

import (
	"context"
	"os"
	"testing"
)

func TestRegistryIncludesDefaultConnectors(t *testing.T) {
	registry := NewRegistry()
	items := registry.List()
	if len(items) != 3 {
		t.Fatalf("expected 3 default connectors, got %d", len(items))
	}
	if items[0].ID != "codex" || items[1].ID != "claude" || items[2].ID != "opencode" {
		t.Fatalf("unexpected connector order: %#v", items)
	}
}

func TestRegistryRejectsUnknownConnector(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.Run(context.Background(), "missing", Request{Prompt: "hello", Workspace: "."})
	if err == nil {
		t.Fatal("expected error for unknown connector")
	}
}

func TestStripANSI(t *testing.T) {
	if got := stripANSI("\x1b[31mred\x1b[0m"); got != "red" {
		t.Fatalf("expected red, got %q", got)
	}
}

func TestSanitizeCLIOutputRemovesLeadingStatusLine(t *testing.T) {
	input := "> build · gpt-5.4-mini\n\nMonitoring markets in Blackdesk."
	got := sanitizeCLIOutput(input)
	if got != "Monitoring markets in Blackdesk." {
		t.Fatalf("unexpected sanitized output: %q", got)
	}
}

func TestSanitizeCLIOutputKeepsQuotedContent(t *testing.T) {
	input := "> this is a quoted line from the response\nline two"
	got := sanitizeCLIOutput(input)
	if got != input {
		t.Fatalf("expected quoted content to stay intact, got %q", got)
	}
}

func TestCodexModelList(t *testing.T) {
	models, err := codexModelList(context.Background(), Descriptor{})
	if err != nil {
		t.Fatalf("expected codex model list, got error: %v", err)
	}
	want := []string{"gpt-5.4", "gpt-5.4-mini", "gpt-5.4-nano", "gpt-5.3-codex"}
	if len(models) != len(want) {
		t.Fatalf("unexpected model count: got %d want %d", len(models), len(want))
	}
	for i := range want {
		if models[i] != want[i] {
			t.Fatalf("unexpected model at %d: got %q want %q", i, models[i], want[i])
		}
	}
}

func TestClaudeCodeModelList(t *testing.T) {
	models, err := claudeCodeModelList(context.Background(), Descriptor{})
	if err != nil {
		t.Fatalf("expected claude code model list, got error: %v", err)
	}
	want := []string{"default", "sonnet", "opus", "haiku", "opusplan"}
	if len(models) != len(want) {
		t.Fatalf("unexpected model count: got %d want %d", len(models), len(want))
	}
	for i := range want {
		if models[i] != want[i] {
			t.Fatalf("unexpected model at %d: got %q want %q", i, models[i], want[i])
		}
	}
}

func TestSanitizeCLIOutputPrefersFinalMessageFileStyleOutput(t *testing.T) {
	tmp, err := os.CreateTemp("", "blackdesk-agent-last-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.WriteString("pong"); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}
	if got := string(data); got != "pong" {
		t.Fatalf("unexpected temp output: %q", got)
	}
}
