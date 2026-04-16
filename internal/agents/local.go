package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type Descriptor struct {
	ID          string
	Label       string
	Binary      string
	Path        string
	Available   bool
	Description string
}

type Request struct {
	Prompt       string
	Workspace    string
	SystemPrompt string
	Model        string
}

type Response struct {
	ConnectorID string
	Output      string
	Duration    time.Duration
}

type runner interface {
	Descriptor() Descriptor
	Run(context.Context, Request) (Response, error)
	Models(context.Context) ([]string, error)
}

type Registry struct {
	runners []runner
}

func NewRegistry() *Registry {
	return &Registry{runners: []runner{
		newCodexRunner(),
		newClaudeRunner(),
		newOpenCodeRunner(),
	}}
}

func (r *Registry) List() []Descriptor {
	items := make([]Descriptor, 0, len(r.runners))
	for _, item := range r.runners {
		items = append(items, item.Descriptor())
	}
	return items
}

func (r *Registry) Lookup(id string) (Descriptor, bool) {
	for _, item := range r.runners {
		d := item.Descriptor()
		if strings.EqualFold(d.ID, id) {
			return d, true
		}
	}
	return Descriptor{}, false
}

func (r *Registry) Run(ctx context.Context, connectorID string, req Request) (Response, error) {
	for _, item := range r.runners {
		d := item.Descriptor()
		if strings.EqualFold(d.ID, connectorID) {
			if !d.Available {
				return Response{}, errors.New(d.Label + " is not installed on this machine")
			}
			return item.Run(ctx, req)
		}
	}
	return Response{}, errors.New("unknown AI connector: " + connectorID)
}

func (r *Registry) Models(ctx context.Context, connectorID string) ([]string, error) {
	for _, item := range r.runners {
		d := item.Descriptor()
		if strings.EqualFold(d.ID, connectorID) {
			if !d.Available {
				return nil, errors.New(d.Label + " is not installed on this machine")
			}
			return item.Models(ctx)
		}
	}
	return nil, errors.New("unknown AI connector: " + connectorID)
}

type localRunner struct {
	desc     Descriptor
	args     func(Request, string) []string
	env      func(Request) []string
	models   func(context.Context, Descriptor) ([]string, error)
	trimANSI bool
}

func newCodexRunner() localRunner {
	return newLocalRunner(
		Descriptor{ID: "codex", Label: "Codex", Binary: "codex", Description: "OpenAI local Codex CLI"},
		func(req Request, outputFile string) []string {
			args := []string{"exec", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox", "-C", req.Workspace}
			if model := strings.TrimSpace(req.Model); model != "" {
				args = append(args, "--model", model)
			}
			if strings.TrimSpace(outputFile) != "" {
				args = append(args, "--output-last-message", outputFile)
			}
			return append(args, buildPrompt(req))
		},
		nil,
		codexModelList,
	)
}

func newClaudeRunner() localRunner {
	return newLocalRunner(
		Descriptor{ID: "claude", Label: "Claude Code", Binary: "claude", Description: "Anthropic Claude Code CLI"},
		func(req Request, outputFile string) []string {
			args := []string{"-p", "--permission-mode", "bypassPermissions", "--dangerously-skip-permissions", "--add-dir", req.Workspace}
			if model := strings.TrimSpace(req.Model); model != "" {
				args = append(args, "--model", model)
			}
			return append(args, buildPrompt(req))
		},
		nil,
		claudeCodeModelList,
	)
}

func newOpenCodeRunner() localRunner {
	return newLocalRunner(
		Descriptor{ID: "opencode", Label: "OpenCode", Binary: "opencode", Description: "OpenCode local agent CLI"},
		func(req Request, outputFile string) []string {
			args := []string{"run", "--dir", req.Workspace, "--format", "json"}
			if model := strings.TrimSpace(req.Model); model != "" {
				args = append(args, "--model", model)
			}
			return append(args, buildPrompt(req))
		},
		func(req Request) []string {
			return []string{
				"OPENCODE_CONFIG_CONTENT={\"$schema\":\"https://opencode.ai/config.json\",\"permission\":\"allow\"}",
			}
		},
		openCodeModelList,
	)
}

func newLocalRunner(desc Descriptor, args func(Request, string) []string, env func(Request) []string, models func(context.Context, Descriptor) ([]string, error)) localRunner {
	path, err := exec.LookPath(desc.Binary)
	if err == nil {
		desc.Path = path
		desc.Available = true
	}
	if env == nil {
		env = func(Request) []string { return nil }
	}
	return localRunner{desc: desc, args: args, env: env, models: models, trimANSI: true}
}

func (r localRunner) Descriptor() Descriptor {
	return r.desc
}

func (r localRunner) Run(ctx context.Context, req Request) (Response, error) {
	if !r.desc.Available {
		return Response{}, errors.New(r.desc.Label + " is not installed on this machine")
	}
	started := time.Now()
	outputFile := ""
	if r.desc.ID == "codex" {
		f, err := os.CreateTemp("", "blackdesk-codex-last-*.txt")
		if err == nil {
			outputFile = f.Name()
			_ = f.Close()
			defer os.Remove(outputFile)
		}
	}
	cmd := exec.CommandContext(ctx, r.desc.Path, r.args(req, outputFile)...)
	cmd.Dir = req.Workspace
	cmd.Env = append(os.Environ(), r.env(req)...)
	cmd.Stdin = strings.NewReader("")
	configureIsolatedSubprocess(cmd)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	text := strings.TrimSpace(stdout.String())
	errText := strings.TrimSpace(stderr.String())
	if outputFile != "" {
		if data, readErr := os.ReadFile(outputFile); readErr == nil && strings.TrimSpace(string(data)) != "" {
			text = strings.TrimSpace(string(data))
		}
	}
	if r.trimANSI {
		text = stripANSI(text)
		errText = stripANSI(errText)
	}
	text = sanitizeCLIOutput(r.desc.ID, text)
	errText = sanitizeCLIOutput(r.desc.ID, errText)
	if text == "" {
		text = errText
	}
	resp := Response{
		ConnectorID: r.desc.ID,
		Output:      text,
		Duration:    time.Since(started),
	}
	if err != nil {
		if text == "" {
			text = err.Error()
		}
		resp.Output = text
		return resp, err
	}
	return resp, nil
}

func sanitizeCLIOutput(connectorID, text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if structured := extractStructuredCLIOutput(connectorID, text); structured != "" {
		return structured
	}
	lines := strings.Split(text, "\n")
	start := 0
	for start < len(lines) {
		line := strings.TrimSpace(lines[start])
		if line == "" {
			start++
			continue
		}
		if isCLIStatusLine(line) {
			start++
			for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
				start++
			}
		}
		break
	}
	text = strings.TrimSpace(strings.Join(lines[start:], "\n"))
	if trimmed := extractTrailingAssistantReply(connectorID, text); trimmed != "" {
		return trimmed
	}
	return text
}

func isCLIStatusLine(line string) bool {
	if !strings.HasPrefix(line, "> ") {
		return false
	}
	body := strings.TrimSpace(strings.TrimPrefix(line, "> "))
	if body == "" {
		return false
	}
	return strings.Contains(body, " · ") || strings.Contains(body, " • ")
}

func extractStructuredCLIOutput(connectorID, text string) string {
	if !strings.EqualFold(connectorID, "opencode") {
		return ""
	}
	var b strings.Builder
	parsed := false
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		parsed = true
		kind := strings.ToLower(strings.TrimSpace(jsonString(event["type"])))
		if !isStructuredTextEvent(kind, event) {
			continue
		}
		b.WriteString(strings.Join(collectStructuredText(event), ""))
	}
	if !parsed {
		return ""
	}
	return strings.TrimSpace(b.String())
}

func isStructuredTextEvent(kind string, event map[string]any) bool {
	if kind == "" {
		return false
	}
	switch {
	case strings.Contains(kind, "tool"), strings.Contains(kind, "step"), strings.Contains(kind, "think"), strings.Contains(kind, "error"):
		return false
	case kind == "text", strings.Contains(kind, "text"):
		return true
	case kind == "message", strings.Contains(kind, "message"), kind == "assistant", strings.Contains(kind, "assistant"):
		role := strings.ToLower(strings.TrimSpace(jsonString(event["role"])))
		return role == "" || role == "assistant"
	default:
		return false
	}
}

func collectStructuredText(value any) []string {
	orderedKeys := []string{"text", "delta", "value", "body", "content", "contents", "part", "parts", "message", "messages"}
	switch v := value.(type) {
	case map[string]any:
		parts := make([]string, 0, len(v))
		for _, key := range orderedKeys {
			for rawKey, nested := range v {
				if !strings.EqualFold(strings.TrimSpace(rawKey), key) {
					continue
				}
				parts = append(parts, collectStructuredText(nested)...)
			}
		}
		return parts
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, collectStructuredText(item)...)
		}
		return parts
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{v}
	default:
		return nil
	}
}

func jsonString(value any) string {
	text, _ := value.(string)
	return text
}

func extractTrailingAssistantReply(connectorID, text string) string {
	if !strings.EqualFold(connectorID, "opencode") {
		return ""
	}
	lines := strings.Split(text, "\n")
	lastTrace := -1
	for i, raw := range lines {
		if isOpenCodeTraceLine(strings.TrimSpace(raw)) {
			lastTrace = i
		}
	}
	if lastTrace < 0 || lastTrace >= len(lines)-1 {
		return ""
	}
	return strings.TrimSpace(strings.Join(lines[lastTrace+1:], "\n"))
}

func isOpenCodeTraceLine(line string) bool {
	if line == "" {
		return false
	}
	return strings.HasPrefix(line, "% ") || strings.HasPrefix(line, "$ ") || strings.HasPrefix(line, "Error:")
}

func (r localRunner) Models(ctx context.Context) ([]string, error) {
	if !r.desc.Available {
		return nil, errors.New(r.desc.Label + " is not installed on this machine")
	}
	if r.models == nil {
		return nil, fmt.Errorf("%s does not expose model discovery", r.desc.Label)
	}
	return r.models(ctx, r.desc)
}

func buildPrompt(req Request) string {
	var b strings.Builder
	if system := strings.TrimSpace(req.SystemPrompt); system != "" {
		b.WriteString(system)
		b.WriteString("\n\n")
	}
	b.WriteString(strings.TrimSpace(req.Prompt))
	return b.String()
}

func unsupportedModelList(_ context.Context, desc Descriptor) ([]string, error) {
	return nil, fmt.Errorf("%s does not expose model discovery through its CLI", desc.Label)
}

func codexModelList(_ context.Context, _ Descriptor) ([]string, error) {
	return []string{
		"gpt-5.4",
		"gpt-5.4-mini",
		"gpt-5.4-nano",
		"gpt-5.3-codex",
	}, nil
}

func claudeCodeModelList(_ context.Context, _ Descriptor) ([]string, error) {
	return []string{
		"default",
		"sonnet",
		"opus",
		"haiku",
		"opusplan",
	}, nil
}

func openCodeModelList(ctx context.Context, desc Descriptor) ([]string, error) {
	cmd := exec.CommandContext(ctx, desc.Path, "models")
	output, err := cmd.CombinedOutput()
	text := stripANSI(strings.TrimSpace(string(output)))
	if err != nil {
		if text == "" {
			text = err.Error()
		}
		return nil, errors.New(text)
	}
	if text == "" {
		return nil, errors.New("no models returned")
	}
	seen := make(map[string]struct{})
	models := make([]string, 0, 16)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		models = append(models, line)
	}
	slices.Sort(models)
	return models, nil
}
