package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mustafmst/taskerr/internal/cli"
	"github.com/mustafmst/taskerr/internal/config"
	"github.com/mustafmst/taskerr/internal/tui"
)

func main() {
	configProvider := config.NewConfigProvider()
	_, err := configProvider.Get()
	if err != nil {
		log.Fatalf("Error initializing config provider: %v", err)
	}
	if len(os.Args) > 1 {
		// cli app startup
		rootCmd := cli.SetupCliApp()
		if err := rootCmd.Execute(); err != nil {
			log.Fatalf("Error executing CLI: %v", err)
		}
	} else {
		// tui app startup
		p := tea.NewProgram(tui.Model{})
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error starting TUI: %v", err)
		}
	}
	os.Exit(0)
}
