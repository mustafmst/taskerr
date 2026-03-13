package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mustafmst/taskerr/internal/cli"
	"github.com/mustafmst/taskerr/internal/config"
	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/tui"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	dataService, err := data.NewService(cfg)
	if err != nil {
		log.Fatalf("Error initializing data service: %v", err)
	}
	defer dataService.Close()

	if len(os.Args) > 1 {
		// cli app startup
		rootCmd := cli.SetupCLIApp(dataService)
		if err := rootCmd.Execute(); err != nil {
			log.Fatalf("Error executing CLI: %v", err)
		}
	} else {
		// tui app startup
		p := tea.NewProgram(tui.NewMainWindowModel(dataService), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error starting TUI: %v", err)
		}
	}
	os.Exit(0)
}
