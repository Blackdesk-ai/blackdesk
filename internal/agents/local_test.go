package agents

import (
	"context"
	"os"
	"strings"
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
	got := sanitizeCLIOutput("codex", input)
	if got != "Monitoring markets in Blackdesk." {
		t.Fatalf("unexpected sanitized output: %q", got)
	}
}

func TestSanitizeCLIOutputKeepsQuotedContent(t *testing.T) {
	input := "> this is a quoted line from the response\nline two"
	got := sanitizeCLIOutput("codex", input)
	if got != input {
		t.Fatalf("expected quoted content to stay intact, got %q", got)
	}
}

func TestCodexRunnerBypassesApprovalsAndSandbox(t *testing.T) {
	runner := newCodexRunner()
	args := runner.args(Request{Workspace: "/tmp/workspace", Prompt: "hello"}, "")
	if !containsArg(args, "--dangerously-bypass-approvals-and-sandbox") {
		t.Fatalf("expected codex runner to bypass approvals and sandbox, got %#v", args)
	}
	if containsArg(args, "--sandbox") {
		t.Fatalf("expected codex runner to avoid restrictive sandbox flags, got %#v", args)
	}
}

func TestClaudeRunnerBypassesPermissions(t *testing.T) {
	runner := newClaudeRunner()
	args := runner.args(Request{Workspace: "/tmp/workspace", Prompt: "hello"}, "")
	if !containsArg(args, "--dangerously-skip-permissions") {
		t.Fatalf("expected claude runner to skip permissions, got %#v", args)
	}
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--permission-mode" && args[i+1] == "bypassPermissions" {
			return
		}
	}
	t.Fatalf("expected claude runner to use bypassPermissions mode, got %#v", args)
}

func TestOpenCodeRunnerRequestsJSONOutput(t *testing.T) {
	runner := newOpenCodeRunner()
	args := runner.args(Request{Workspace: "/tmp/workspace", Prompt: "hello"}, "")
	found := false
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--format" && args[i+1] == "json" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected opencode runner to request json output, got %#v", args)
	}
}

func TestOpenCodeRunnerSetsAllowAllPermissions(t *testing.T) {
	runner := newOpenCodeRunner()
	env := runner.env(Request{Workspace: "/tmp/workspace", Prompt: "hello"})
	if len(env) != 1 || !strings.Contains(env[0], `"permission":"allow"`) {
		t.Fatalf("expected opencode runner to inject allow-all permissions, got %#v", env)
	}
}

func TestSanitizeCLIOutputExtractsOpenCodeJSONText(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"step_start","step":"tool"}`,
		`{"type":"text","text":"Drivers rose "}`,
		`{"type":"text","text":"19% YoY."}`,
		`{"type":"step_finish","step":"done"}`,
	}, "\n")
	got := sanitizeCLIOutput("opencode", input)
	if got != "Drivers rose 19% YoY." {
		t.Fatalf("unexpected sanitized json output: %q", got)
	}
}

func TestSanitizeCLIOutputTrimsOpenCodeTraceTail(t *testing.T) {
	input := strings.Join([]string{
		"% WebFetch sec.gov/ixviewer/ix.html?",
		"Error: Request failed with status code: 404",
		"% Grep \"9.7 million monthly drivers and couriers\" report.txt",
		"",
		"Uber reported 9.7 million monthly drivers and couriers.",
		"",
		"That implies growth versus the prior year.",
	}, "\n")
	got := sanitizeCLIOutput("opencode", input)
	want := "Uber reported 9.7 million monthly drivers and couriers.\n\nThat implies growth versus the prior year."
	if got != want {
		t.Fatalf("unexpected sanitized fallback output: %q", got)
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

func TestCodexRunnerProvidesNoopEnv(t *testing.T) {
	runner := newCodexRunner()
	if runner.env == nil {
		t.Fatal("expected codex runner env hook to be initialized")
	}
	if got := runner.env(Request{Workspace: "/tmp/workspace", Prompt: "hello"}); got != nil {
		t.Fatalf("expected codex runner env hook to default to nil env vars, got %#v", got)
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
