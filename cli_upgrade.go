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
	flags.Usage = func() {
		printUpgradeUsage(flags.Output())
	}

	checkOnly := flags.Bool("check", false, "check whether a newer release is available")
	requestedVersion := flags.String("version", "", "upgrade to a specific release version")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			printUpgradeUsage(stdout)
			return nil
		}
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

func printUpgradeUsage(w io.Writer) {
	fmt.Fprintln(w, "Blackdesk upgrade")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  blackdesk upgrade [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --check              Check whether a newer release is available")
	fmt.Fprintln(w, "  --version <version>  Upgrade to a specific release version")
	fmt.Fprintln(w, "  -h, --help           Show this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  blackdesk upgrade --check")
	fmt.Fprintln(w, "  blackdesk upgrade")
	fmt.Fprintln(w, "  blackdesk upgrade --version 0.1.0")
}
