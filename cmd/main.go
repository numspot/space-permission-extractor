package main

import (
	"fmt"
	"os"

	"permission-extractor-tui/internal/config"
	"permission-extractor-tui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg := config.Config{}

	cfg.LoadFromEnvFile()
	cfg.ApplyDefaults()

	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--clientId":
				if i+1 < len(os.Args) {
					cfg.ClientID = os.Args[i+1]
					i++
				}
			case "--clientSecret":
				if i+1 < len(os.Args) {
					cfg.ClientSecret = os.Args[i+1]
					i++
				}
			case "--space":
				if i+1 < len(os.Args) {
					cfg.SpaceID = os.Args[i+1]
					i++
				}
			case "--baseUrl":
				if i+1 < len(os.Args) {
					cfg.BaseURL = os.Args[i+1]
					i++
				}
			case "--help", "-h":
				fmt.Println("NumSpot Permission Extractor TUI")
				fmt.Println()
				fmt.Println("Usage: permission-extractor-tui [options]")
				fmt.Println()
				fmt.Println("Options:")
				fmt.Println("  --clientId <id>      OAuth2 Client ID")
				fmt.Println("  --clientSecret <secret> OAuth2 Client Secret")
				fmt.Println("  --space <uuid>       Space UUID")
				fmt.Println("  --baseUrl <url>      API Base URL (default: https://api.eu-west-2.numspot.com)")
				fmt.Println("  --help, -h           Show this help")
				fmt.Println()
				fmt.Println("Configuration is loaded from:")
				fmt.Println("  1. ./.env in current directory")
				fmt.Println("  2. Command-line arguments (override env file)")
				fmt.Println("  3. TUI inputs (override all)")
				fmt.Println()
				fmt.Println("You can save configuration from the TUI by pressing Tab to 'Save' button and Enter.")
				os.Exit(0)
			}
		}
	}

	app := tui.NewApp(cfg)

	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
