package buildinfo

import "testing"

func TestDetailedIncludesAllBuildMetadata(t *testing.T) {
	prevVersion := Version
	prevCommit := Commit
	prevDate := Date
	t.Cleanup(func() {
		Version = prevVersion
		Commit = prevCommit
		Date = prevDate
	})

	Version = "0.1.0"
	Commit = "abc1234"
	Date = "2026-04-06T12:00:00Z"

	got := Detailed("blackdesk")
	want := "blackdesk 0.1.0\ncommit: abc1234\nbuilt: 2026-04-06T12:00:00Z"
	if got != want {
		t.Fatalf("unexpected detailed build info:\n got: %q\nwant: %q", got, want)
	}
}
