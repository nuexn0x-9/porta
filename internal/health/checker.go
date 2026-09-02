package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/porta-dev/porta/internal/registry"
	"github.com/porta-dev/porta/internal/security"
)

// Checker continuously monitors health of all services in the registry
type Checker struct {
	reg        *registry.Registry
	client     *http.Client
	stopChan   chan struct{}
	wg         sync.WaitGroup
	onStatusCh chan *registry.Service
}

// NewChecker initializes the health check monitor
func NewChecker(reg *registry.Registry) *Checker {
	transport := &http.Transport{
		DialContext:         security.SafeDialContext(),
		DisableKeepAlives:   true,
		MaxIdleConnsPerHost: -1,
	}

	return &Checker{
		reg: reg,
		client: &http.Client{
			Transport: transport,
			Timeout:   2 * time.Second,
		},
		stopChan:   make(chan struct{}),
		onStatusCh: make(chan *registry.Service, 20),
	}
}

// Start launches the background health monitoring loop
func (c *Checker) Start(ctx context.Context) {
	// 1. Initial pre-flight check immediately
	c.ProbeAll()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stopChan:
				return
			case <-ticker.C:
				c.ProbeAll()
			}
		}
	}()
}

// ProbeAll checks every registered service concurrently
func (c *Checker) ProbeAll() {
	services := c.reg.List()
	for _, svc := range services {
		go c.probeService(svc)
	}
}

// probeService checks an individual service
func (c *Checker) probeService(svc *registry.Service) {
	addr := svc.TargetAddr()

	// 1. If HTTP health check path is configured
	if svc.HealthCheck.Path != "" {
		targetURL := fmt.Sprintf("http://%s%s", addr, svc.HealthCheck.Path)
		req, err := http.NewRequest(http.MethodGet, targetURL, nil)
		if err == nil {
			req.Header.Set("User-Agent", "PORTA-HealthChecker/1.0")
			resp, err := c.client.Do(req)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					svc.SetStatus(registry.StatusOnline)
					return
				}
			}
		}
	}

	// 2. Default TCP socket ping
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		svc.SetStatus(registry.StatusOffline)
	} else {
		_ = conn.Close()
		svc.SetStatus(registry.StatusOnline)
	}
}

// Stop stops the background health check worker
func (c *Checker) Stop() {
	close(c.stopChan)
	c.wg.Wait()
}
