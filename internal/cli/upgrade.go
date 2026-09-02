package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check for and install the latest version of PORTA",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("[i] Current PORTA version: %s\n", Version)
		fmt.Println("[i] Checking GitHub Releases for updates...")

		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/nuexn0x-9/porta/releases/latest", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", "PORTA-CLI-Upgrader")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to GitHub: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			fmt.Printf("[✓] You are running the latest development/baseline version (%s). No newer releases published on GitHub.\n", Version)
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
		}

		type Release struct {
			TagName string `json:"tag_name"`
			Name    string `json:"name"`
			HTMLURL string `json:"html_url"`
		}

		var rel Release
		if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
			return fmt.Errorf("failed to parse release metadata: %w", err)
		}

		if rel.TagName == Version || rel.TagName == "" {
			fmt.Printf("[✓] You are already on the latest version (%s)!\n", Version)
			return nil
		}

		fmt.Printf("[✓] New version available: %s (Release: %s)\n", rel.TagName, rel.HTMLURL)

		// Self-update binary
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine executable path: %w", err)
		}

		assetName := fmt.Sprintf("porta-%s-%s", runtime.GOOS, runtime.GOARCH)
		if runtime.GOOS == "windows" {
			assetName += ".exe"
		}

		downloadURL := fmt.Sprintf("https://github.com/nuexn0x-9/porta/releases/download/%s/%s", rel.TagName, assetName)
		checksumURL := fmt.Sprintf("https://github.com/nuexn0x-9/porta/releases/download/%s/checksums.txt", rel.TagName)

		fmt.Printf("[i] Downloading update from %s...\n", downloadURL)
		tmpPath := execPath + ".tmp"
		out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to create temporary binary: %w", err)
		}

		binResp, err := client.Get(downloadURL)
		if err != nil {
			_ = out.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("download failed: %w", err)
		}
		defer binResp.Body.Close()

		hasher := sha256.New()
		multiOut := io.MultiWriter(out, hasher)
		if _, err := io.Copy(multiOut, binResp.Body); err != nil {
			_ = out.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to save binary: %w", err)
		}
		_ = out.Close()

		actualHash := hex.EncodeToString(hasher.Sum(nil))

		// Checksum verification
		chkResp, err := client.Get(checksumURL)
		if err == nil && chkResp.StatusCode == http.StatusOK {
			chkData, _ := io.ReadAll(chkResp.Body)
			_ = chkResp.Body.Close()
			for _, line := range strings.Split(string(chkData), "\n") {
				if strings.Contains(line, assetName) {
					expectedHash := strings.Fields(line)[0]
					if strings.EqualFold(actualHash, expectedHash) {
						fmt.Printf("[✓] SHA-256 integrity verified (%s)\n", actualHash)
					} else {
						_ = os.Remove(tmpPath)
						return fmt.Errorf("SHA-256 mismatch! Expected %s, got %s", expectedHash, actualHash)
					}
					break
				}
			}
		}

		// Replace executable
		if runtime.GOOS == "windows" {
			oldPath := execPath + ".old"
			_ = os.Remove(oldPath)
			if err := os.Rename(execPath, oldPath); err != nil {
				_ = os.Remove(tmpPath)
				return fmt.Errorf("failed to move old executable on Windows: %w", err)
			}
		}

		if err := os.Rename(tmpPath, execPath); err != nil {
			return fmt.Errorf("failed to replace executable: %w", err)
		}

		fmt.Printf("[✓] Successfully upgraded PORTA to %s at %s!\n", rel.TagName, filepath.Clean(execPath))
		return nil
	},
}
