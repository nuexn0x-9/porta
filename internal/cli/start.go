package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/health"
	"github.com/porta-dev/porta/internal/proxy"
	"github.com/porta-dev/porta/internal/registry"
	"github.com/porta-dev/porta/internal/tunnel/cloudflare"
	"github.com/porta-dev/porta/internal/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the reverse proxy gateway and establish public tunnel (Foreground)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 1. Load Configuration
		fmt.Printf("[i] Loading configuration from %s...\n", cfgFile)
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		// 2. Initialize Logger
		logger, err := ui.InitLogger(".porta/logs", os.Stdout)
		if err != nil {
			fmt.Printf("[!] Warning: failed to initialize disk logger: %v\n", err)
		}
		defer logger.Close()

		// 3. Build Service Registry
		reg := registry.BuildFromConfig(cfg)
		fmt.Printf("[✓] Registered %d service(s):\n", reg.Count())
		for _, svc := range reg.List() {
			fmt.Printf("    • %s -> %s (%s:%d)\n", svc.Route, svc.Name, svc.Host, svc.Port)
		}

		// 4. Start Health Checker
		healthChecker := health.NewChecker(reg)
		healthChecker.Start(ctx)
		defer healthChecker.Stop()

		// 5. Start Reverse Proxy Gateway on ephemeral loopback port (127.0.0.1:0)
		gateway := proxy.NewGateway(cfg, reg)
		gatewayAddr, err := gateway.Start()
		if err != nil {
			return fmt.Errorf("failed to start gateway: %w", err)
		}
		defer func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer shutdownCancel()
			_ = gateway.Stop(shutdownCtx)
		}()
		fmt.Printf("[✓] Local gateway listening on %s\n", gatewayAddr)

		// 6. Start Cloudflare Tunnel Provider
		fmt.Println("[i] Establishing secure Cloudflare Quick Tunnel...")
		driver := cloudflare.NewDriver()
		session, err := driver.Start(ctx, gatewayAddr)
		if err != nil {
			return fmt.Errorf("tunnel error: %w", err)
		}
		defer func() {
			_ = driver.Stop()
		}()

		fmt.Printf("[✓] Tunnel active! Public URL: %s\n", session.PublicURL)

		// 7. Setup Signal Handling for Graceful Shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			cancel()
		}()

		// 8. Launch Interactive TUI in foreground
		tui := ui.NewTUI(cfg, reg, session.PublicURL)
		tui.Run(ctx)

		fmt.Println("\n[i] Gracefully shutting down PORTA...")
		return nil
	},
}
