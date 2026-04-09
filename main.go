package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/bootstrap"
	"blackdesk/internal/buildinfo"
	"blackdesk/internal/tui"
)

func main() {
	ctx := context.Background()
	if err := runCLI(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func runCLI(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 {
		switch args[0] {
		case "upgrade":
			return runUpgrade(ctx, args[1:], stdout, stderr)
		case "help", "?", "-h", "--help":
			printMainUsage(stdout)
			return nil
		}
	}

	flags := flag.NewFlagSet("blackdesk", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		printMainUsage(flags.Output())
	}

	showVersion := flags.Bool("version", false, "print version information and exit")
	showVersionShort := flags.Bool("v", false, "print version information and exit")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printMainUsage(stdout)
			return nil
		}
		return err
	}

	if flags.NArg() > 0 {
		return fmt.Errorf("unknown command: %s", flags.Arg(0))
	}

	if *showVersion || *showVersionShort {
		fmt.Fprintln(stdout, buildinfo.Detailed("blackdesk"))
		return nil
	}

	deps, err := bootstrap.LoadDependencies(ctx)
	if err != nil {
		return err
	}
	model := tui.NewModel(ctx, deps)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func printMainUsage(w io.Writer) {
	fmt.Fprintln(w, "Blackdesk")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  blackdesk [flags]")
	fmt.Fprintln(w, "  blackdesk upgrade [flags]")
	fmt.Fprintln(w, "  blackdesk help")
	fmt.Fprintln(w, "  blackdesk ?")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -v, --version      Print version information and exit")
	fmt.Fprintln(w, "  -h, --help         Show this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  upgrade            Upgrade Blackdesk to the latest published release")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  blackdesk")
	fmt.Fprintln(w, "  blackdesk --help")
	fmt.Fprintln(w, "  blackdesk ?")
	fmt.Fprintln(w, "  blackdesk --version")
	fmt.Fprintln(w, "  blackdesk upgrade --check")
}
