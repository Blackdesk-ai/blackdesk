package tui

import (
	"context"
	"strings"
	"testing"

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
