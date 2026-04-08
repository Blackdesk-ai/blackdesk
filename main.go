package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/bootstrap"
	"blackdesk/internal/buildinfo"
	"blackdesk/internal/tui"
)

func main() {
	clean := flag.Bool("clean", false, "reset config to defaults before starting")
	showVersion := flag.Bool("version", false, "print version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(buildinfo.Detailed("blackdesk"))
		return
	}

	ctx := context.Background()
	deps, err := bootstrap.LoadDependencies(ctx, *clean)
	if err != nil {
		log.Fatal(err)
	}
	model := tui.NewModel(ctx, deps)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
