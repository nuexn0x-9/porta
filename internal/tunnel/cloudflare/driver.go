package cloudflare

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/porta-dev/porta/internal/tunnel"
	"github.com/porta-dev/porta/internal/utils"
)

var (
	tryCloudflareRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
)

// Driver implements the TunnelProvider interface for Cloudflare Quick Tunnel
type Driver struct {
	mu         sync.Mutex
	binPath    string
	cmd        *exec.Cmd
	session    *tunnel.TunnelSession
	cancelFunc context.CancelFunc
	stopped    bool
}

// NewDriver creates a new Cloudflare tunnel provider driver
func NewDriver() *Driver {
	return &Driver{}
}

// Name returns "cloudflare"
func (d *Driver) Name() string {
	return "cloudflare"
}

// Start spawns cloudflared tunnel pointing to local gateway address
func (d *Driver) Start(ctx context.Context, localGatewayAddr string) (*tunnel.TunnelSession, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	binPath, err := EnsureCloudflaredBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to locate or download cloudflared binary: %w", err)
	}
	d.binPath = binPath

	tunnelCtx, cancel := context.WithCancel(ctx)
	d.cancelFunc = cancel
	d.stopped = false

	gatewayURL := fmt.Sprintf("http://%s", localGatewayAddr)
	cmd := exec.CommandContext(tunnelCtx, binPath, "tunnel", "--url", gatewayURL, "--no-autoupdate")
	utils.PrepareChildProcess(cmd)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to pipe cloudflared stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start cloudflared process: %w", err)
	}
	d.cmd = cmd

	// Channel to receive the allocated public URL
	urlChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stderr)
		urlFound := false

		for scanner.Scan() {
			line := scanner.Text()

			if !urlFound {
				match := tryCloudflareRegex.FindString(line)
				if match != "" {
					urlFound = true
					urlChan <- match
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	// Wait for URL or timeout
	select {
	case pubURL := <-urlChan:
		d.session = &tunnel.TunnelSession{
			PublicURL: pubURL,
			Provider:  "cloudflare",
			StartedAt: time.Now(),
		}
		return d.session, nil
	case err := <-errChan:
		_ = d.Stop()
		return nil, fmt.Errorf("error reading cloudflared output: %w", err)
	case <-time.After(25 * time.Second):
		_ = d.Stop()
		return nil, fmt.Errorf("timeout waiting for Cloudflare tunnel URL to be assigned")
	case <-ctx.Done():
		_ = d.Stop()
		return nil, ctx.Err()
	}
}

// HealthCheck checks if the child process is still running
func (d *Driver) HealthCheck(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cmd == nil || d.cmd.Process == nil {
		return fmt.Errorf("tunnel process is not running")
	}

	if d.stopped {
		return fmt.Errorf("tunnel has been stopped")
	}

	return nil
}

// Stop terminates the cloudflared process gracefully
func (d *Driver) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.stopped = true
	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	if d.cmd != nil && d.cmd.Process != nil {
		_ = d.cmd.Process.Kill()
		_ = d.cmd.Wait()
		d.cmd = nil
	}
	return nil
}
