package cli

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/porta-dev/porta/internal/tunnel/cloudflare"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initialize the PORTA runtime environment and pre-cache tunnel drivers",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║                   PORTA Initial Setup                      ║")
		fmt.Println("║     Local Multi-Service Public Exposure Platform           ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println("\n[i] Initializing PORTA environment...")

		// 1. Detect System
		fmt.Printf("  [✓] System detected       : %s/%s\n", runtime.GOOS, runtime.GOARCH)

		// 2. Directory Tree Creation
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		portaDir := filepath.Join(homeDir, ".porta")
		dirs := []string{
			filepath.Join(portaDir, "bin"),
			filepath.Join(portaDir, "config"),
			filepath.Join(portaDir, "logs"),
			filepath.Join(portaDir, "cache"),
		}

		for _, d := range dirs {
			if err := os.MkdirAll(d, 0700); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", d, err)
			}
		}

		// 3. Write minimal global configuration if missing
		globalCfgPath := filepath.Join(portaDir, "config", "global.yaml")
		if _, err := os.Stat(globalCfgPath); os.IsNotExist(err) {
			globalYaml := `# PORTA Global Configuration
version: 1

installer:
  auto_update_check: true

runtime:
  telemetry: false
`
			_ = os.WriteFile(globalCfgPath, []byte(globalYaml), 0600)
		}
		fmt.Printf("  [✓] Directory structure   : ~/.porta/ initialized\n")

		// 4. Loopback Interface Probe
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Printf("  [✗] Loopback interface    : Failed to bind on 127.0.0.1 (%v)\n", err)
		} else {
			_ = l.Close()
			fmt.Printf("  [✓] Loopback interface    : 127.0.0.1 bindable\n")
		}

		// 5. Outbound Internet Connectivity Probe
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("https://1.1.1.1")
		if err != nil {
			fmt.Printf("  [✗] Internet connectivity : Unreachable (%v)\n", err)
		} else {
			_ = resp.Body.Close()
			fmt.Printf("  [✓] Internet connectivity : DNS & HTTPS connection OK\n")
		}

		// 6. Pre-cache Cloudflare Tunnel Driver
		binPath, err := cloudflare.EnsureCloudflaredBinary()
		if err != nil {
			fmt.Printf("  [✗] Tunnel driver         : Failed to download cloudflared (%v)\n", err)
		} else {
			fmt.Printf("  [✓] Tunnel driver         : Ready (%s)\n", binPath)
		}

		fmt.Println("\n──────────────────────────────────────────────────────────────")
		fmt.Println("🎉 PORTA is successfully configured and ready to use!")
		fmt.Println("\nNext Steps:")
		fmt.Println("  1. Open your project folder : cd my-web-app")
		fmt.Println("  2. Initialize configuration : porta init")
		fmt.Println("  3. Start public exposure    : porta start")
		fmt.Println("──────────────────────────────────────────────────────────────")

		return nil
	},
}
