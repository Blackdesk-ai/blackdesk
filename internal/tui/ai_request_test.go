package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"blackdesk/internal/domain"
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

func TestAIRequestIncludesConversationSummaryBeforeRecentHistory(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.aiConversationSummary = "- User: asked for prior filing context\n- AI: highlighted debt rollover risk"
	model.aiMessages = []aiMessage{
		{Role: aiMessageUser, Body: "latest follow-up"},
		{Role: aiMessageAssistant, Body: "recent answer"},
	}

	req, err := model.buildAIRequest("follow-up")
	if err != nil {
		t.Fatalf("buildAIRequest failed: %v", err)
	}
	if !strings.Contains(req.SystemPrompt, "<conversation_summary>") {
		t.Fatal("expected conversation summary block to be included")
	}
	if !strings.Contains(req.SystemPrompt, "debt rollover risk") {
		t.Fatal("expected compacted summary content to be included")
	}
	if strings.Index(req.SystemPrompt, "<conversation_summary>") > strings.Index(req.SystemPrompt, "<conversation>") {
		t.Fatal("expected conversation summary to appear before recent conversation")
	}
}

func TestAITranscriptCompactionRollsOldMessagesIntoSummary(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	chunk := strings.Repeat("AAPL filing context and management commentary. ", 700)
	for i := 0; i < 18; i++ {
		role := aiMessageUser
		if i%2 == 1 {
			role = aiMessageAssistant
		}
		model.aiMessages = append(model.aiMessages, aiMessage{
			Role: role,
			Body: fmt.Sprintf("msg-%02d %s", i, chunk),
		})
	}

	model.maintainAITranscriptBudget()

	if model.aiConversationSummary == "" {
		t.Fatal("expected old transcript to be compacted into a summary")
	}
	if model.aiCompactedMessages == 0 {
		t.Fatal("expected compacted message count to be tracked")
	}
	if len(model.aiMessages) >= 18 {
		t.Fatal("expected only recent raw messages to remain after compaction")
	}
	if !strings.Contains(model.aiConversationSummary, "msg-00") {
		t.Fatal("expected oldest content to survive in compacted summary")
	}
	if strings.Contains(model.aiConversationSummary, "msg-17") {
		t.Fatal("expected newest content to remain in raw conversation, not summary")
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

func TestAIFilingAnalysisRequestIncludesSelectedFilingBlock(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.config.AIConnector = "codex"
	model.config.AIModel = "gpt-5.4"
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.quote.Symbol = "AAPL"
	model.fundamentals.Symbol = "AAPL"

	snapshot := domain.FilingsSnapshot{
		Symbol:      "AAPL",
		CompanyName: "Apple Inc.",
		CIK:         "0000320193",
	}
	doc := domain.FilingDocument{
		Item: domain.FilingItem{
			Form:                  "10-K",
			FilingDate:            model.clock,
			PrimaryDocument:       "aapl-10k.htm",
			PrimaryDocDescription: "Annual report",
			URL:                   "https://www.sec.gov/Archives/example",
		},
		ContentType: "text/html",
		Text:        "Revenue grew 12%. Services margin expanded. Risk factors include China concentration.",
	}

	req, err := model.buildAIFilingAnalysisRequest("AAPL", snapshot, doc, "Analyze the selected 10-K filing for AAPL.")
	if err != nil {
		t.Fatalf("buildAIFilingAnalysisRequest failed: %v", err)
	}
	if !strings.Contains(req.SystemPrompt, "<selected_filing>") {
		t.Fatal("expected filing analysis request to include selected filing block")
	}
	if !strings.Contains(req.SystemPrompt, "Revenue grew 12%.") {
		t.Fatal("expected filing text to be included in filing analysis request")
	}
	if !strings.Contains(req.SystemPrompt, "What Was Filed") || !strings.Contains(req.SystemPrompt, "Bottom Line") {
		t.Fatal("expected filing analysis sections in system prompt")
	}
}
