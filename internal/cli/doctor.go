package cli

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/tunnel/cloudflare"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health, networking, tunnel drivers, and configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("PORTA System Doctor Diagnostics")
		fmt.Println("===============================")

		// 1. OS & Architecture
		fmt.Printf("[✓] OS & Architecture: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		// 2. Loopback Interface check
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Printf("[✗] Loopback Interface: Failed to bind on 127.0.0.1 (%v)\n", err)
		} else {
			_ = l.Close()
			fmt.Println("[✓] Loopback Interface: 127.0.0.1 bindable")
		}

		// 3. Outbound Internet Connectivity
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("https://1.1.1.1")
		if err != nil {
			fmt.Printf("[✗] Outbound Internet: Unreachable (%v)\n", err)
		} else {
			_ = resp.Body.Close()
			fmt.Println("[✓] Outbound Internet: DNS resolution & HTTPS connection OK")
		}

		// 4. Cloudflare Tunnel Driver check
		binPath, err := cloudflare.EnsureCloudflaredBinary()
		if err != nil {
			fmt.Printf("[✗] Tunnel Driver: Cloudflare driver unavailable (%v)\n", err)
		} else {
			fmt.Printf("[✓] Tunnel Driver: Ready (%s)\n", binPath)
		}

		// 5. Configuration File check
		if _, err := os.Stat(cfgFile); err == nil {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				fmt.Printf("[✗] Configuration (%s): Invalid (%v)\n", cfgFile, err)
			} else {
				fmt.Printf("[✓] Configuration (%s): Valid (%d service(s) configured)\n", cfgFile, len(cfg.Services))
			}
		} else {
			fmt.Printf("[i] Configuration (%s): Not found in current directory (Run 'porta init' to create)\n", cfgFile)
		}
	},
}
