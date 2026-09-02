package cloudflare

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnsureCloudflaredBinary finds an existing cloudflared binary or downloads one to ~/.porta/bin/
func EnsureCloudflaredBinary() (string, error) {
	// 1. Check if cloudflared is already in system PATH
	if path, err := exec.LookPath("cloudflared"); err == nil {
		return path, nil
	}

	// 2. Check local user cache directory ~/.porta/bin/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	portaBinDir := filepath.Join(homeDir, ".porta", "bin")
	binName := "cloudflared"
	if runtime.GOOS == "windows" {
		binName = "cloudflared.exe"
	}
	cachedBinaryPath := filepath.Join(portaBinDir, binName)

	if info, err := os.Stat(cachedBinaryPath); err == nil && !info.IsDir() {
		return cachedBinaryPath, nil
	}

	// 3. Download on-demand
	if err := os.MkdirAll(portaBinDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory %s: %w", portaBinDir, err)
	}

	downloadURL, err := getDownloadURL()
	if err != nil {
		return "", err
	}

	fmt.Printf("[i] Downloading Cloudflare tunnel driver to %s...\n", cachedBinaryPath)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download cloudflared from %s: %w", downloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download cloudflared: HTTP %d from %s", resp.StatusCode, downloadURL)
	}

	tmpFile := cachedBinaryPath + ".tmp"
	out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary binary file: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	_ = out.Close()
	if err != nil {
		_ = os.Remove(tmpFile)
		return "", fmt.Errorf("failed to save downloaded binary: %w", err)
	}

	if err := os.Rename(tmpFile, cachedBinaryPath); err != nil {
		return "", fmt.Errorf("failed to move downloaded binary to %s: %w", cachedBinaryPath, err)
	}

	fmt.Println("[✓] Cloudflare tunnel driver downloaded and cached successfully.")
	return cachedBinaryPath, nil
}

func getDownloadURL() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "arm64" {
			return "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-arm64.exe", nil
		}
		return "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe", nil
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-arm64.tgz", nil
		}
		return "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-amd64.tgz", nil
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64", nil
		}
		return "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64", nil
	default:
		return "", fmt.Errorf("unsupported operating system for automatic driver download: %s", runtime.GOOS)
	}
}
