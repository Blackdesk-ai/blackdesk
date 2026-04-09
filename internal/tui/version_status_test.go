package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestStatusMetaTextIncludesVersionLabel(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.appVersion = "0.1.0"

	got := model.statusMetaText()
	if !strings.Contains(got, "| v0.1.0") {
		t.Fatalf("expected version label in status meta text, got %q", got)
	}
}

func TestStatusMetaTextShowsAvailableUpdate(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.appVersion = "0.1.0"
	model.latestVersion = "0.2.0"
	model.updateAvailable = true

	got := model.statusMetaText()
	if !strings.Contains(got, "v0.1.0 -> v0.2.0") {
		t.Fatalf("expected update indicator in status meta text, got %q", got)
	}
}

func TestShouldCheckForUpdatesIncludesDevBuilds(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.appVersion = "dev"
	if !model.shouldCheckForUpdates() {
		t.Fatal("expected dev builds to check for the latest published release")
	}

	model.appVersion = "0.1.0"
	if !model.shouldCheckForUpdates() {
		t.Fatal("expected released builds to check for updates")
	}
}

func TestStatusMetaTextShowsDevUpgradeTarget(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.appVersion = "dev"
	model.latestVersion = "0.2.0"
	model.updateAvailable = true

	got := model.statusMetaText()
	if !strings.Contains(got, "dev -> v0.2.0") {
		t.Fatalf("expected dev build to advertise latest release, got %q", got)
	}
}

func TestRenderStatusMetaHighlightsAvailableUpdate(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.appVersion = "0.1.0"
	model.latestVersion = "0.2.0"
	model.updateAvailable = true

	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#9F907E"))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Bold(true)
	got := model.renderStatusMeta(muted, accent)

	if !strings.Contains(got, "v0.1.0 -> v0.2.0") {
		t.Fatalf("expected rendered update label, got %q", got)
	}
}

func TestStatusTextShowsUpgradeKeyOnlyWhenUpdateAvailable(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabMarkets

	if got := model.statusText(); strings.Contains(got, "u update app") {
		t.Fatalf("expected no upgrade key without update, got %q", got)
	}

	model.latestVersion = "0.2.0"
	model.updateAvailable = true
	got := model.statusText()
	if !strings.Contains(got, "u update app") {
		t.Fatalf("expected upgrade key when update is available, got %q", got)
	}
	if strings.Index(got, "? help") > strings.Index(got, "u update app") {
		t.Fatalf("expected update key after help, got %q", got)
	}

	model.upgradeRunning = true
	if got := model.statusText(); strings.Contains(got, "u update app") {
		t.Fatalf("expected upgrade key to hide while upgrading, got %q", got)
	}
}

func TestUStartsUpgradeOnlyWhenUpdateIsAvailable(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabMarkets

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m := updated.(Model)
	if m.upgradeRunning {
		t.Fatal("expected u to do nothing without an available update")
	}
	if cmd != nil {
		t.Fatal("expected no upgrade command without an available update")
	}

	model.latestVersion = "0.2.0"
	model.updateAvailable = true
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(Model)
	if !m.upgradeRunning {
		t.Fatal("expected u to start upgrade when an update is available")
	}
	if cmd == nil {
		t.Fatal("expected upgrade command when update is available")
	}
}
