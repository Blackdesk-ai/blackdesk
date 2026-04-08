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
	ctx := context.Background()
	if len(os.Args) > 1 && os.Args[1] == "upgrade" {
		if err := runUpgrade(ctx, os.Args[2:], os.Stdout, os.Stderr); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		return
	}

	clean := flag.Bool("clean", false, "reset config to defaults before starting")
	showVersion := flag.Bool("version", false, "print version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(buildinfo.Detailed("blackdesk"))
		return
	}

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
