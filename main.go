package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mustafmst/taskerr/internal/cli"
	"github.com/mustafmst/taskerr/internal/tui"
)

func main() {
	if len(os.Args) > 1 {
		// cli app startup
		rootCmd := cli.SetupCliApp()
		if err := rootCmd.Execute(); err != nil {
			log.Fatalf("Error executing CLI: %v", err)
		}
	} else {
		// tui app startup
		p := tea.NewProgram(tui.Model{})
		if err := p.Start(); err != nil {
			log.Fatalf("Error starting TUI: %v", err)
		}
	}
	os.Exit(0)
}
