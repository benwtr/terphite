// Command terphite is a terminal browser for Graphite metrics, loosely
// modeled on Graphite's web Composer.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/benwtr/terphite/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("terphite", version)
		return
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s https://user:pass@yourgraphite.net:4321\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(os.Stderr, "       (credentials may also be supplied via GRAPHITE_USER / GRAPHITE_PASS env vars)")
		os.Exit(1)
	}

	cfg := tui.Config{
		GraphiteURI: os.Args[1],
		Username:    os.Getenv("GRAPHITE_USER"),
		Password:    os.Getenv("GRAPHITE_PASS"),
	}

	m, err := tui.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "terphite:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "terphite:", err)
		os.Exit(1)
	}
}
