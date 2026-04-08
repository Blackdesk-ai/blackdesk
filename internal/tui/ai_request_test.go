package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"blackdesk/internal/storage"
)

func TestAIRequestUsesEmbeddedMarketPrompt(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.config.AIConnector = "codex"
	model.config.AIModel = "gpt-5.4-mini"

	req, err := model.buildAIRequest("ping")
	if err != nil {
		t.Fatalf("buildAIRequest failed: %v", err)
	}
	if !strings.Contains(req.SystemPrompt, "market analysis terminal") {
		t.Fatal("expected embedded AI system prompt to be included")
	}
	if !strings.Contains(req.SystemPrompt, "prioritize market and trading help") {
		t.Fatal("expected market-first instructions in system prompt")
	}
	if !strings.Contains(req.SystemPrompt, "avoid meta commentary about the app's internal context") {
		t.Fatal("expected professional phrasing guardrail in system prompt")
	}
}

func TestAIRequestAlwaysIncludesBlackdeskContext(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.aiLastContext = "old"
	model.aiLastSymbol = model.activeSymbol()

	req, err := model.buildAIRequest("what is hv21")
	if err != nil {
		t.Fatalf("buildAIRequest failed: %v", err)
	}
	if !strings.Contains(req.SystemPrompt, "<blackdesk_context_update>") {
		t.Fatal("expected blackdesk context to be included on every request")
	}
}

func TestAIRequestIncludesFullHistoryWithinBudget(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	for i := 1; i <= 14; i++ {
		role := aiMessageUser
		if i%2 == 0 {
			role = aiMessageAssistant
		}
		model.aiMessages = append(model.aiMessages, aiMessage{
			Role: role,
			Body: fmt.Sprintf("msg-%02d", i),
		})
	}

	req, err := model.buildAIRequest("follow-up")
	if err != nil {
		t.Fatalf("buildAIRequest failed: %v", err)
	}
	if !strings.Contains(req.SystemPrompt, "msg-01") || !strings.Contains(req.SystemPrompt, "msg-14") {
		t.Fatal("expected full short history to be included when it fits in budget")
	}
}

func TestAIRequestIncludesContextOnlyOncePerRequest(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})

	req, err := model.buildAIRequest("ping")
	if err != nil {
		t.Fatalf("buildAIRequest failed: %v", err)
	}
	if strings.Count(req.SystemPrompt, "<blackdesk_context_update>") != 1 {
		t.Fatal("expected exactly one context block per request")
	}
}

func TestWriteAILastRequestDump(t *testing.T) {
	root := t.TempDir()
	envelope := RequestEnvelope{
		Prompt:         "what is hv21",
		SystemPrompt:   "system",
		ContextPayload: `{"technical_values":{"HV21":"23.4%"}}`,
		ActiveSymbol:   "AAPL",
	}

	if err := writeAILastRequestDump(root, "codex", "gpt-5.4-mini", envelope); err != nil {
		t.Fatalf("writeAILastRequestDump should be a no-op by default: %v", err)
	}
	if _, err := os.Stat(root + "/.blackdesk/last_ai_request.json"); !os.IsNotExist(err) {
		t.Fatalf("expected no AI dump by default, got err=%v", err)
	}

	t.Setenv(aiLastRequestDumpEnv, "1")
	if err := writeAILastRequestDump(root, "codex", "gpt-5.4-mini", envelope); err != nil {
		t.Fatalf("writeAILastRequestDump failed: %v", err)
	}

	data, err := os.ReadFile(root + "/.blackdesk/last_ai_request.json")
	if err != nil {
		t.Fatalf("read dump failed: %v", err)
	}

	var dump aiLastRequestDump
	if err := json.Unmarshal(data, &dump); err != nil {
		t.Fatalf("unmarshal dump failed: %v", err)
	}
	if dump.UserPrompt != "what is hv21" {
		t.Fatalf("unexpected user prompt in dump: %q", dump.UserPrompt)
	}
	if !strings.Contains(dump.ContextPayload, `"HV21":"23.4%"`) {
		t.Fatal("expected context payload in AI dump")
	}
}
