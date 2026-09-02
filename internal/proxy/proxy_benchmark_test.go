package proxy

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
)

func BenchmarkProxyRoutingLatency(b *testing.B) {
	// 1. Direct Backend Server (no proxy)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("benchmark-payload"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	port, _ := strconv.Atoi(u.Port())

	// 2. Gateway setup
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "benchmark-app"},
		Services: map[string]*config.ServiceCfg{
			"app": {
				Host:  "127.0.0.1",
				Port:  port,
				Route: "/",
			},
		},
		Security: config.SecurityConfig{Mode: "public"},
	}

	reg := registry.BuildFromConfig(cfg)
	svc, _ := reg.Get("app")
	svc.SetStatus(registry.StatusOnline)

	gw := NewGateway(cfg, reg)
	addr, err := gw.Start()
	if err != nil {
		b.Fatalf("failed to start gateway: %v", err)
	}
	defer func() {
		_ = gw.Stop(context.Background())
	}()

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
		Timeout: 2 * time.Second,
	}

	gwURL := "http://" + addr + "/test"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(gwURL)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
