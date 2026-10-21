package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ricardoruwer/tt-cli/internal/api"
	"github.com/ricardoruwer/tt-cli/internal/config"
	"github.com/ricardoruwer/tt-cli/internal/ui"
)

func main() {
	cfg, warning, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "tt:", err)
		os.Exit(1)
	}
	if warning != "" {
		fmt.Fprintln(os.Stderr, "tt: warning:", warning)
	}
	if cfg.Token == "" {
		path, _ := config.Path()
		fmt.Fprintf(os.Stderr, "tt: no token found.\n\nCreate %s with:\n\n  token = \"<your token>\"\n\nor export TT_TOKEN.\n", path)
		os.Exit(1)
	}

	client := api.New(cfg.BaseURL, cfg.Token)
	if _, err := tea.NewProgram(ui.New(client, time.Now())).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tt:", err)
		os.Exit(1)
	}
}
