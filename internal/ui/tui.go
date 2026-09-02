package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
)

// TUI renders the interactive foreground terminal interface
type TUI struct {
	cfg       *config.Config
	reg       *registry.Registry
	publicURL string
	logger    *Logger
	logLines  []string
	mu        sync.Mutex
	maxLogs   int
}

// NewTUI creates a new TUI instance
func NewTUI(cfg *config.Config, reg *registry.Registry, publicURL string) *TUI {
	return &TUI{
		cfg:       cfg,
		reg:       reg,
		publicURL: publicURL,
		logger:    GetLogger(),
		logLines:  make([]string, 0),
		maxLogs:   8,
	}
}

// Run renders and continuously updates the TUI until context cancellation
func (t *TUI) Run(ctx context.Context) {
	logCh := t.logger.Subscribe()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Initial render
	t.Render()

	for {
		select {
		case <-ctx.Done():
			return
		case line := <-logCh:
			t.mu.Lock()
			t.logLines = append(t.logLines, line)
			if len(t.logLines) > t.maxLogs {
				t.logLines = t.logLines[len(t.logLines)-t.maxLogs:]
			}
			t.mu.Unlock()
			t.Render()
		case <-ticker.C:
			t.Render()
		}
	}
}

// Render prints the formatted ANSI status screen
func (t *TUI) Render() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Clear screen and move cursor to top-left
	fmt.Print("\033[H\033[2J")

	border := "─────────────────────────────────────────────────────────────"
	fmt.Printf("\033[1;34m┌%s┐\033[0m\n", border)
	fmt.Printf("\033[1;34m│\033[0m  \033[1;36mPORTA v1.0.0\033[0m — Local Multi-Service Public Exposure Engine  \033[1;34m│\033[0m\n")
	fmt.Printf("\033[1;34m├%s┤\033[0m\n", border)

	// Project & URL Info
	secLabel := "Public / Unrestricted"
	if t.cfg.Security.Mode == "password" {
		secLabel = "Protected (Basic Auth)"
	} else if t.cfg.Security.Mode == "token" {
		secLabel = "Protected (Bearer/Query Token)"
	}

	fmt.Printf("\033[1;34m│\033[0m   \033[1mProject\033[0m     : %-46s\033[1;34m│\033[0m\n", t.cfg.Project.Name)
	fmt.Printf("\033[1;34m│\033[0m   \033[1mPublic URL\033[0m  : \033[1;32m%-46s\033[0m\033[1;34m│\033[0m\n", t.publicURL)
	fmt.Printf("\033[1;34m│\033[0m   \033[1mSecurity\033[0m    : %-46s\033[1;34m│\033[0m\n", secLabel)
	fmt.Printf("\033[1;34m├%s┤\033[0m\n", border)

	// Services & Routes
	fmt.Printf("\033[1;34m│\033[0m   \033[1;33mSERVICES & ROUTES:\033[0m                                        \033[1;34m│\033[0m\n")
	services := t.reg.List()
	for _, svc := range services {
		statusStr := formatStatus(svc.GetStatus())
		routeDesc := fmt.Sprintf("%s (%s:%d) -> %s", svc.Name, svc.Host, svc.Port, svc.Route)
		fmt.Printf("\033[1;34m│\033[0m   • %-40s %s   \033[1;34m│\033[0m\n", truncate(routeDesc, 40), statusStr)
	}

	fmt.Printf("\033[1;34m├%s┤\033[0m\n", border)
	fmt.Printf("\033[1;34m│\033[0m   \033[1mLIVE LOGS\033[0m (Press \033[1;31mCtrl+C\033[0m to stop):                             \033[1;34m│\033[0m\n")

	if len(t.logLines) == 0 {
		fmt.Printf("\033[1;34m│\033[0m   \033[90mWaiting for incoming public requests...\033[0m                   \033[1;34m│\033[0m\n")
	} else {
		for _, line := range t.logLines {
			fmt.Printf("   %s\n", line)
		}
	}

	fmt.Printf("\033[1;34m└%s┘\033[0m\n", border)
}

func formatStatus(st registry.HealthStatus) string {
	switch st {
	case registry.StatusOnline:
		return "\033[1;32m[ONLINE]\033[0m"
	case registry.StatusStarting:
		return "\033[1;33m[STARTING]\033[0m"
	case registry.StatusOffline:
		return "\033[1;31m[OFFLINE]\033[0m"
	default:
		return "\033[90m[UNKNOWN]\033[0m"
	}
}

func truncate(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}
