package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
)

// FAT Scenario 2: Single Service Exposure
func TestFAT_Scenario2_SingleServiceExposure(t *testing.T) {
	// 1. Start mock single web app on localhost
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<h1>Welcome to My Single App</h1>"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	port, _ := strconv.Atoi(u.Port())

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "single-fat-app"},
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
		t.Fatalf("gateway failed to start: %v", err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/", addr))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "Welcome to My Single App") {
		t.Fatalf("expected 200 OK with HTML content, got %d: %s", resp.StatusCode, string(body))
	}
}

// FAT Scenario 3 & 4: Multi-Service Application & Ingress Routing (Root, Nested, Query Params)
func TestFAT_Scenario3_And_4_MultiServiceAndRouting(t *testing.T) {
	// 1. Mock Frontend on :3000
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("Frontend: " + r.URL.Path))
	}))
	defer frontend.Close()
	fPort, _ := strconv.Atoi(mustParseURL(frontend.URL).Port())

	// 2. Mock Backend on :8000
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		qID := r.URL.Query().Get("id")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"service":"backend","path":"%s","id":"%s"}`, r.URL.Path, qID)))
	}))
	defer backend.Close()
	bPort, _ := strconv.Atoi(mustParseURL(backend.URL).Port())

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "multiservice-fat-app"},
		Services: map[string]*config.ServiceCfg{
			"frontend": {
				Host:  "127.0.0.1",
				Port:  fPort,
				Route: "/",
			},
			"backend": {
				Host:      "127.0.0.1",
				Port:      bPort,
				Route:     "/api",
				StripPath: false,
			},
		},
		Security: config.SecurityConfig{Mode: "public"},
	}

	reg := registry.BuildFromConfig(cfg)
	fSvc, _ := reg.Get("frontend")
	fSvc.SetStatus(registry.StatusOnline)
	bSvc, _ := reg.Get("backend")
	bSvc.SetStatus(registry.StatusOnline)

	gw := NewGateway(cfg, reg)
	addr, err := gw.Start()
	if err != nil {
		t.Fatalf("gateway failed to start: %v", err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	client := &http.Client{Timeout: 3 * time.Second}
	base := fmt.Sprintf("http://%s", addr)

	// A. Root Route -> Frontend
	resp1, err := client.Get(base + "/")
	if err != nil {
		t.Fatalf("frontend root failed: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	_ = resp1.Body.Close()
	if string(body1) != "Frontend: /" {
		t.Fatalf("expected 'Frontend: /', got '%s'", string(body1))
	}

	// B. Nested Route -> Backend
	resp2, err := client.Get(base + "/api/users")
	if err != nil {
		t.Fatalf("backend nested route failed: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if !strings.Contains(string(body2), `"path":"/api/users"`) {
		t.Fatalf("expected /api/users path preserved, got: %s", string(body2))
	}

	// C. Query Parameters -> Backend
	resp3, err := client.Get(base + "/api/users?id=10")
	if err != nil {
		t.Fatalf("backend query failed: %v", err)
	}
	body3, _ := io.ReadAll(resp3.Body)
	_ = resp3.Body.Close()
	if !strings.Contains(string(body3), `"id":"10"`) {
		t.Fatalf("expected query parameter id=10 preserved, got: %s", string(body3))
	}
}

// FAT Scenario 5: Health Check & Graceful Recovery Behavior
func TestFAT_Scenario5_HealthCheckAndRecovery(t *testing.T) {
	var isServiceLive atomic.Bool
	isServiceLive.Store(true)

	// Mock server that can be toggled offline/online
	dynamicBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isServiceLive.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("SERVICE_ACTIVE"))
	}))
	defer dynamicBackend.Close()

	port, _ := strconv.Atoi(mustParseURL(dynamicBackend.URL).Port())

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "recovery-app"},
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
		t.Fatalf("gateway failed: %v", err)
	}
	defer func() { _ = gw.Stop(context.Background()) }()

	client := &http.Client{Timeout: 2 * time.Second}
	base := fmt.Sprintf("http://%s", addr)

	// 1. Initial request -> 200 OK
	resp1, _ := client.Get(base + "/")
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp1.StatusCode)
	}
	_ = resp1.Body.Close()

	// 2. Simulate service stopping -> marks status OFFLINE
	svc.SetStatus(registry.StatusOffline)
	resp2, _ := client.Get(base + "/")
	if resp2.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502 Bad Gateway when offline, got %d", resp2.StatusCode)
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	if !strings.Contains(string(body2), "502 Bad Gateway") {
		t.Fatalf("expected friendly 502 page, got: %s", string(body2))
	}

	// 3. Service restarts -> marks status ONLINE
	svc.SetStatus(registry.StatusOnline)
	resp3, _ := client.Get(base + "/")
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK after recovery, got %d", resp3.StatusCode)
	}
	_ = resp3.Body.Close()
}

// FAT Scenario 6: Authentication Gatekeeper (Basic Auth & Token Auth)
func TestFAT_Scenario6_Authentication(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("AUTHENTICATED_ACCESS"))
	}))
	defer backend.Close()
	port, _ := strconv.Atoi(mustParseURL(backend.URL).Port())

	// 1. Basic Auth Mode
	cfgBasic := &config.Config{
		Project: config.ProjectConfig{Name: "auth-basic-app"},
		Services: map[string]*config.ServiceCfg{
			"app": {Host: "127.0.0.1", Port: port, Route: "/"},
		},
		Security: config.SecurityConfig{
			Mode:     "password",
			Password: "admin:SecretDemo123",
		},
	}
	regBasic := registry.BuildFromConfig(cfgBasic)
	s1, _ := regBasic.Get("app")
	s1.SetStatus(registry.StatusOnline)

	gwBasic := NewGateway(cfgBasic, regBasic)
	addrBasic, _ := gwBasic.Start()
	defer func() { _ = gwBasic.Stop(context.Background()) }()

	client := &http.Client{Timeout: 2 * time.Second}
	baseBasic := fmt.Sprintf("http://%s", addrBasic)

	// A. Without credentials -> 401
	r1, _ := client.Get(baseBasic + "/")
	if r1.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without credentials, got %d", r1.StatusCode)
	}
	_ = r1.Body.Close()

	// B. Invalid credentials -> 401
	reqInvalid, _ := http.NewRequest(http.MethodGet, baseBasic+"/", nil)
	reqInvalid.SetBasicAuth("admin", "wrongpassword")
	r2, _ := client.Do(reqInvalid)
	if r2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized with invalid credentials, got %d", r2.StatusCode)
	}
	_ = r2.Body.Close()

	// C. Valid credentials -> 200 OK
	reqValid, _ := http.NewRequest(http.MethodGet, baseBasic+"/", nil)
	reqValid.SetBasicAuth("admin", "SecretDemo123")
	r3, _ := client.Do(reqValid)
	if r3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK with valid credentials, got %d", r3.StatusCode)
	}
	_ = r3.Body.Close()

	// 2. Token Auth Mode (Bearer & Query Token)
	cfgToken := &config.Config{
		Project: config.ProjectConfig{Name: "auth-token-app"},
		Services: map[string]*config.ServiceCfg{
			"app": {Host: "127.0.0.1", Port: port, Route: "/"},
		},
		Security: config.SecurityConfig{
			Mode:  "token",
			Token: "my-bearer-secret-token",
		},
	}
	regToken := registry.BuildFromConfig(cfgToken)
	s2, _ := regToken.Get("app")
	s2.SetStatus(registry.StatusOnline)

	gwToken := NewGateway(cfgToken, regToken)
	addrToken, _ := gwToken.Start()
	defer func() { _ = gwToken.Stop(context.Background()) }()

	baseToken := fmt.Sprintf("http://%s", addrToken)

	// A. Without token -> 403 Forbidden
	t1, _ := client.Get(baseToken + "/")
	if t1.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 without token, got %d", t1.StatusCode)
	}
	_ = t1.Body.Close()

	// B. With Authorization: Bearer -> 200 OK
	reqBearer, _ := http.NewRequest(http.MethodGet, baseToken+"/", nil)
	reqBearer.Header.Set("Authorization", "Bearer my-bearer-secret-token")
	t2, _ := client.Do(reqBearer)
	if t2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with Bearer token, got %d", t2.StatusCode)
	}
	_ = t2.Body.Close()

	// C. With ?porta_token= query param -> 200 OK
	t3, _ := client.Get(baseToken + "/?porta_token=my-bearer-secret-token")
	if t3.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with query token, got %d", t3.StatusCode)
	}
	_ = t3.Body.Close()
}

// FAT Scenario 8: WebSocket Echo Handshake & Bidirectional Piping
func TestFAT_Scenario8_WebSocketPiping(t *testing.T) {
	// TCP echo listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listener failed: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil || strings.TrimSpace(line) == "" {
				break
			}
		}
		_, _ = conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"))
		buf := make([]byte, 512)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			_, _ = conn.Write(buf[:n])
		}
	}()

	backendPort := listener.Addr().(*net.TCPAddr).Port

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "ws-app"},
		Services: map[string]*config.ServiceCfg{
			"ws": {Host: "127.0.0.1", Port: backendPort, Route: "/"},
		},
		Security: config.SecurityConfig{Mode: "public"},
	}

	reg := registry.BuildFromConfig(cfg)
	svc, _ := reg.Get("ws")
	svc.SetStatus(registry.StatusOnline)

	gw := NewGateway(cfg, reg)
	addr, _ := gw.Start()
	defer func() { _ = gw.Stop(context.Background()) }()

	// Connect client TCP to gateway
	clientConn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to dial gateway: %v", err)
	}
	defer clientConn.Close()

	handshake := "GET / HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"
	_, _ = clientConn.Write([]byte(handshake))

	r := bufio.NewReader(clientConn)
	status, _ := r.ReadString('\n')
	if !strings.Contains(status, "101") {
		t.Fatalf("expected 101 Switching Protocols, got: %s", status)
	}
}

// FAT Scenario 9: Unbuffered SSE Streaming
func TestFAT_Scenario9_SSEStreaming(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if ok {
			_, _ = fmt.Fprintf(w, "event: message\ndata: Chunk-1\n\n")
			flusher.Flush()
			_, _ = fmt.Fprintf(w, "event: message\ndata: Chunk-2\n\n")
			flusher.Flush()
		}
	}))
	defer backend.Close()
	port, _ := strconv.Atoi(mustParseURL(backend.URL).Port())

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "sse-app"},
		Services: map[string]*config.ServiceCfg{
			"stream": {Host: "127.0.0.1", Port: port, Route: "/"},
		},
		Security: config.SecurityConfig{Mode: "public"},
	}
	reg := registry.BuildFromConfig(cfg)
	svc, _ := reg.Get("stream")
	svc.SetStatus(registry.StatusOnline)

	gw := NewGateway(cfg, reg)
	addr, _ := gw.Start()
	defer func() { _ = gw.Stop(context.Background()) }()

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/events", addr))
	if err != nil {
		t.Fatalf("SSE request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Chunk-1") || !strings.Contains(string(body), "Chunk-2") {
		t.Fatalf("expected unbuffered SSE chunks, got: %s", string(body))
	}
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}
