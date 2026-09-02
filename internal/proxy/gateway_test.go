package proxy

import (
	"context"
	"fmt"
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

func TestGatewayEndToEndRouting(t *testing.T) {
	// 1. Mock Frontend Server
	frontendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("Frontend Home: " + r.URL.Path))
	}))
	defer frontendServer.Close()

	frontendURL, _ := url.Parse(frontendServer.URL)
	frontendPort, _ := strconv.Atoi(frontendURL.Port())

	// 2. Mock Backend Server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proto := r.Header.Get("X-Forwarded-Proto")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"msg":"Backend API","path":"%s","proto":"%s"}`, r.URL.Path, proto)))
	}))
	defer backendServer.Close()

	backendURL, _ := url.Parse(backendServer.URL)
	backendPort, _ := strconv.Atoi(backendURL.Port())

	// 3. Configure PORTA with Multi-Service (Frontend + Backend)
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test-gateway"},
		Services: map[string]*config.ServiceCfg{
			"frontend": {
				Host:  "127.0.0.1",
				Port:  frontendPort,
				Route: "/",
			},
			"backend": {
				Host:      "127.0.0.1",
				Port:      backendPort,
				Route:     "/api",
				StripPath: true,
			},
			"offline_svc": {
				Host:  "127.0.0.1",
				Port:  64321, // Non-listening port
				Route: "/offline",
			},
		},
		Security: config.SecurityConfig{Mode: "public"},
	}

	reg := registry.BuildFromConfig(cfg)

	// Mark services as appropriate
	fSvc, _ := reg.Get("frontend")
	fSvc.SetStatus(registry.StatusOnline)

	bSvc, _ := reg.Get("backend")
	bSvc.SetStatus(registry.StatusOnline)

	offSvc, _ := reg.Get("offline_svc")
	offSvc.SetStatus(registry.StatusOffline)

	// 4. Start Gateway on ephemeral port
	gateway := NewGateway(cfg, reg)
	addr, err := gateway.Start()
	if err != nil {
		t.Fatalf("failed to start gateway: %v", err)
	}
	defer func() {
		_ = gateway.Stop(context.Background())
	}()

	client := &http.Client{Timeout: 3 * time.Second}
	gatewayBase := fmt.Sprintf("http://%s", addr)

	// TEST 1: Request to Frontend (/)
	resp1, err := client.Get(gatewayBase + "/about")
	if err != nil {
		t.Fatalf("failed to GET frontend: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	_ = resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK || string(body1) != "Frontend Home: /about" {
		t.Fatalf("expected 200 'Frontend Home: /about', got %d '%s'", resp1.StatusCode, string(body1))
	}

	// TEST 2: Request to Backend (/api/v1/users) with strip_path
	resp2, err := client.Get(gatewayBase + "/api/v1/users")
	if err != nil {
		t.Fatalf("failed to GET backend: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from backend, got %d", resp2.StatusCode)
	}
	expectedBackendSubstring := `"path":"/v1/users"`
	if !contains(string(body2), expectedBackendSubstring) {
		t.Fatalf("expected stripped path in backend response, got: %s", string(body2))
	}
	expectedProto := `"proto":"https"`
	if !contains(string(body2), expectedProto) {
		t.Fatalf("expected X-Forwarded-Proto header, got: %s", string(body2))
	}

	// TEST 3: Request to Offline Service -> returns 502 Bad Gateway without crashing
	resp3, err := client.Get(gatewayBase + "/offline")
	if err != nil {
		t.Fatalf("failed to GET offline service: %v", err)
	}
	_ = resp3.Body.Close()

	if resp3.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502 Bad Gateway for offline service, got %d", resp3.StatusCode)
	}

	// TEST 4: Frontend is STILL responding 200 after offline service failure (Service Independence)
	resp4, err := client.Get(gatewayBase + "/")
	if err != nil {
		t.Fatalf("failed to GET frontend after offline test: %v", err)
	}
	_ = resp4.Body.Close()
	if resp4.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from frontend, got %d", resp4.StatusCode)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
