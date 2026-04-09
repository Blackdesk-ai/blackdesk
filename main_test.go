package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunCLIHelpAliasesPrintUsage(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"--help"}, {"?"}, {"help"}} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		if err := runCLI(context.Background(), args, &stdout, &stderr); err != nil {
			t.Fatalf("runCLI(%q) returned error: %v", args, err)
		}
		out := stdout.String()
		if !strings.Contains(out, "Usage:") {
			t.Fatalf("runCLI(%q) should print usage, got %q", args, out)
		}
		if !strings.Contains(out, "blackdesk ?") {
			t.Fatalf("runCLI(%q) should mention question-mark alias, got %q", args, out)
		}
	}
}

func TestRunCLIVersionAliasesPrintVersion(t *testing.T) {
	for _, args := range [][]string{{"-v"}, {"--version"}} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		if err := runCLI(context.Background(), args, &stdout, &stderr); err != nil {
			t.Fatalf("runCLI(%q) returned error: %v", args, err)
		}
		out := stdout.String()
		if !strings.Contains(out, "blackdesk ") || !strings.Contains(out, "commit:") {
			t.Fatalf("runCLI(%q) should print version details, got %q", args, out)
		}
	}
}

func TestRunUpgradeHelpPrintsUsage(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"--help"}} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		if err := runUpgrade(context.Background(), args, &stdout, &stderr); err != nil {
			t.Fatalf("runUpgrade(%q) returned error: %v", args, err)
		}
		out := stdout.String()
		if !strings.Contains(out, "blackdesk upgrade [flags]") {
			t.Fatalf("runUpgrade(%q) should print upgrade usage, got %q", args, out)
		}
		if !strings.Contains(out, "--check") {
			t.Fatalf("runUpgrade(%q) should document --check, got %q", args, out)
		}
	}
}
