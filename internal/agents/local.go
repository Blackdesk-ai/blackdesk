package agents

import (
	"context"
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
	models   func(context.Context, Descriptor) ([]string, error)
	trimANSI bool
}

func newCodexRunner() localRunner {
	return newLocalRunner(
		Descriptor{ID: "codex", Label: "Codex", Binary: "codex", Description: "OpenAI local Codex CLI"},
		func(req Request, outputFile string) []string {
			args := []string{"exec", "--skip-git-repo-check", "--sandbox", "read-only", "-C", req.Workspace}
			if model := strings.TrimSpace(req.Model); model != "" {
				args = append(args, "--model", model)
			}
			if strings.TrimSpace(outputFile) != "" {
				args = append(args, "--output-last-message", outputFile)
			}
			return append(args, buildPrompt(req))
		},
		codexModelList,
	)
}

func newClaudeRunner() localRunner {
	return newLocalRunner(
		Descriptor{ID: "claude", Label: "Claude Code", Binary: "claude", Description: "Anthropic Claude Code CLI"},
		func(req Request, outputFile string) []string {
			args := []string{"-p", "--permission-mode", "plan", "--add-dir", req.Workspace}
			if model := strings.TrimSpace(req.Model); model != "" {
				args = append(args, "--model", model)
			}
			return append(args, buildPrompt(req))
		},
		claudeCodeModelList,
	)
}

func newOpenCodeRunner() localRunner {
	return newLocalRunner(
		Descriptor{ID: "opencode", Label: "OpenCode", Binary: "opencode", Description: "OpenCode local agent CLI"},
		func(req Request, outputFile string) []string {
			args := []string{"run", "--dir", req.Workspace}
			if model := strings.TrimSpace(req.Model); model != "" {
				args = append(args, "--model", model)
			}
			return append(args, buildPrompt(req))
		},
		openCodeModelList,
	)
}

func newLocalRunner(desc Descriptor, args func(Request, string) []string, models func(context.Context, Descriptor) ([]string, error)) localRunner {
	path, err := exec.LookPath(desc.Binary)
	if err == nil {
		desc.Path = path
		desc.Available = true
	}
	return localRunner{desc: desc, args: args, models: models, trimANSI: true}
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
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if outputFile != "" {
		if data, readErr := os.ReadFile(outputFile); readErr == nil && strings.TrimSpace(string(data)) != "" {
			text = strings.TrimSpace(string(data))
		}
	}
	if r.trimANSI {
		text = stripANSI(text)
	}
	text = sanitizeCLIOutput(text)
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

func sanitizeCLIOutput(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
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
	return strings.TrimSpace(strings.Join(lines[start:], "\n"))
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
