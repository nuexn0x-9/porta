package cli

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var forceInit bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new porta.yaml configuration in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := "porta.yaml"

		if _, err := os.Stat(targetPath); err == nil && !forceInit {
			return fmt.Errorf("porta.yaml already exists. Use --force to overwrite")
		}

		fmt.Println("[i] Scanning local listening ports for active web applications...")

		// Common web development ports to probe (strictly excludes DBs like 5432, 3306, 6379)
		candidatePorts := []int{3000, 5173, 8000, 8080, 9000, 4000, 4200, 8081}
		detectedPorts := make([]int, 0)

		for _, port := range candidatePorts {
			if isPortOpen("127.0.0.1", port) {
				detectedPorts = append(detectedPorts, port)
			}
		}

		dirName := filepath.Base(getWorkingDir())
		dirName = strings.ToLower(strings.ReplaceAll(dirName, " ", "-"))
		if dirName == "." || dirName == "/" || dirName == "" {
			dirName = "my-app"
		}

		var yamlContent string

		if len(detectedPorts) == 1 {
			// Single service detected (DECISION SERVICE-001)
			p := detectedPorts[0]
			fmt.Printf("[✓] Detected 1 active service on port :%d\n", p)
			yamlContent = fmt.Sprintf(`# PORTA Configuration
version: "1"

project:
  name: %s

services:
  app:
    port: %d
    route: /

tunnel:
  provider: cloudflare

security:
  mode: public
`, dirName, p)
		} else if len(detectedPorts) > 1 {
			// Multiple services detected
			fmt.Printf("[✓] Detected %d active services on ports: %v\n", len(detectedPorts), detectedPorts)
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf(`# PORTA Configuration
version: "1"

project:
  name: %s

services:
`, dirName))

			for idx, p := range detectedPorts {
				svcName := fmt.Sprintf("service-%d", idx+1)
				route := fmt.Sprintf("/service-%d", idx+1)
				if idx == 0 {
					svcName = "frontend"
					route = "/"
				} else if idx == 1 {
					svcName = "backend"
					route = "/api"
				}

				sb.WriteString(fmt.Sprintf(`  %s:
    port: %d
    route: %s
`, svcName, p, route))
			}

			sb.WriteString(`
tunnel:
  provider: cloudflare

security:
  mode: public
`)
			yamlContent = sb.String()
		} else {
			// No active ports found -> Starter template
			fmt.Println("[!] No active web ports detected. Generating default starter template...")
			yamlContent = fmt.Sprintf(`# PORTA Configuration
version: "1"

project:
  name: %s

services:
  # Example frontend web app
  frontend:
    port: 3000
    route: /

  # Example backend API server
  backend:
    port: 8000
    route: /api
    strip_path: false

tunnel:
  provider: cloudflare

security:
  mode: public
`, dirName)
		}

		if err := os.WriteFile(targetPath, []byte(yamlContent), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", targetPath, err)
		}

		fmt.Printf("[✓] Successfully initialized %s\n", targetPath)
		fmt.Println("[i] Run 'porta start' to launch your public environment!")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "overwrite existing porta.yaml without prompting")
}

func isPortOpen(host string, port int) bool {
	timeout := 200 * time.Millisecond
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return wd
}
