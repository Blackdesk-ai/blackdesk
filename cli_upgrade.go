package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"blackdesk/internal/buildinfo"
	"blackdesk/internal/updater"
)

func runUpgrade(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	flags.SetOutput(stderr)

	checkOnly := flags.Bool("check", false, "check whether a newer release is available")
	requestedVersion := flags.String("version", "", "upgrade to a specific release version")
	if err := flags.Parse(args); err != nil {
		return err
	}

	client := updater.Default()
	currentVersion := buildinfo.NormalizedVersion()

	if *checkOnly {
		result, err := client.Check(ctx, currentVersion)
		if err != nil {
			return err
		}
		switch {
		case result.UpdateAvailable:
			fmt.Fprintf(stdout, "Update available: %s -> %s\n", buildinfo.VersionLabel(result.CurrentVersion), buildinfo.VersionLabel(result.LatestVersion))
		case !result.Comparable:
			fmt.Fprintf(stdout, "Latest published release: %s\n", buildinfo.VersionLabel(result.LatestVersion))
		default:
			fmt.Fprintf(stdout, "Blackdesk %s is up to date\n", buildinfo.VersionLabel(result.CurrentVersion))
		}
		return nil
	}

	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	result, err := client.Upgrade(ctx, executablePath, currentVersion, *requestedVersion)
	if err != nil {
		return err
	}
	if result.AlreadyCurrent {
		fmt.Fprintf(stdout, "Blackdesk %s is already installed\n", buildinfo.VersionLabel(result.InstalledVersion))
		return nil
	}

	previous := buildinfo.VersionLabel(result.PreviousVersion)
	if strings.TrimSpace(result.PreviousVersion) == "" {
		previous = "unknown"
	}
	fmt.Fprintf(stdout, "Updated Blackdesk %s -> %s\n", previous, buildinfo.VersionLabel(result.InstalledVersion))
	if result.RestartRequired {
		fmt.Fprintln(stdout, "Restart Blackdesk to begin using the staged update.")
	}
	return nil
}
