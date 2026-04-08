package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type aiLastRequestDump struct {
	GeneratedAt    string `json:"generated_at"`
	ConnectorID    string `json:"connector_id"`
	Model          string `json:"model"`
	ActiveSymbol   string `json:"active_symbol"`
	UserPrompt     string `json:"user_prompt"`
	SystemPrompt   string `json:"system_prompt"`
	ContextPayload string `json:"context_payload"`
}

const aiLastRequestDumpEnv = "BLACKDESK_WRITE_AI_DEBUG_DUMP"

func shouldWriteAILastRequestDump() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(aiLastRequestDumpEnv))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func writeAILastRequestDump(workspaceRoot, connectorID, model string, envelope RequestEnvelope) error {
	if strings.TrimSpace(workspaceRoot) == "" || !shouldWriteAILastRequestDump() {
		return nil
	}
	dir := filepath.Join(workspaceRoot, ".blackdesk")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(aiLastRequestDump{
		GeneratedAt:    time.Now().Format(time.RFC3339),
		ConnectorID:    connectorID,
		Model:          strings.TrimSpace(model),
		ActiveSymbol:   envelope.ActiveSymbol,
		UserPrompt:     envelope.Prompt,
		SystemPrompt:   envelope.SystemPrompt,
		ContextPayload: envelope.ContextPayload,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "last_ai_request.json"), payload, 0o600)
}

func compactAIInsight(input string, limit int) string {
	text := strings.TrimSpace(input)
	if text == "" {
		return ""
	}
	lines := splitLines(text)
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimLeft(line, "-*•0123456789. "))
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return ""
	}
	text = strings.Join(parts, " ")
	text = strings.Join(strings.Fields(text), " ")
	return truncateRunes(text, limit)
}
