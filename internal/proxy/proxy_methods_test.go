package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
)

func TestHTTPMethodsAndPayloadAudit(t *testing.T) {
	// 1. Mock Backend handling various HTTP verbs and streaming
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/echo-body":
			body, _ := io.ReadAll(r.Body)
			w.Header().Set("X-Received-Method", r.Method)
			w.Header().Set("X-Custom-Echo", r.Header.Get("X-Custom-Header"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		case "/query-check":
			q := r.URL.Query().Get("test_param")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("query=" + q))
		case "/events":
			// SSE Stream
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
			flusher, ok := w.(http.Flusher)
			if ok {
				_, _ = fmt.Fprintf(w, "data: event1\n\n")
				flusher.Flush()
				_, _ = fmt.Fprintf(w, "data: event2\n\n")
				flusher.Flush()
			}
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	port, _ := strconv.Atoi(u.Port())

	// 2. Start Gateway
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "audit-gateway"},
		Services: map[string]*config.ServiceCfg{
			"api": {
				Host:  "127.0.0.1",
				Port:  port,
				Route: "/",
			},
		},
		Security: config.SecurityConfig{Mode: "public"},
	}

	reg := registry.BuildFromConfig(cfg)
	svc, _ := reg.Get("api")
	svc.SetStatus(registry.StatusOnline)

	gw := NewGateway(cfg, reg)
	addr, err := gw.Start()
	if err != nil {
		t.Fatalf("failed to start gateway: %v", err)
	}
	defer func() {
		_ = gw.Stop(context.Background())
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	base := fmt.Sprintf("http://%s", addr)

	// A. Test POST with body & custom header
	postBody := []byte(`{"key":"hello-world","number":42}`)
	req, _ := http.NewRequest(http.MethodPost, base+"/echo-body", bytes.NewReader(postBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom-Header", "PortaAudit123")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for POST, got %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Received-Method") != "POST" {
		t.Fatalf("expected backend to receive POST, got %s", resp.Header.Get("X-Received-Method"))
	}
	if resp.Header.Get("X-Custom-Echo") != "PortaAudit123" {
		t.Fatalf("expected custom header to be preserved, got %s", resp.Header.Get("X-Custom-Echo"))
	}
	if string(respBody) != string(postBody) {
		t.Fatalf("expected body '%s', got '%s'", string(postBody), string(respBody))
	}

	// B. Test PUT, DELETE, PATCH
	for _, method := range []string{http.MethodPut, http.MethodDelete, http.MethodPatch} {
		mReq, _ := http.NewRequest(method, base+"/echo-body", strings.NewReader("method-data"))
		mResp, mErr := client.Do(mReq)
		if mErr != nil {
			t.Fatalf("%s request failed: %v", method, mErr)
		}
		_ = mResp.Body.Close()
		if mResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", method, mResp.StatusCode)
		}
		if mResp.Header.Get("X-Received-Method") != method {
			t.Fatalf("expected method %s, got %s", method, mResp.Header.Get("X-Received-Method"))
		}
	}

	// C. Test Query Parameters Preservation
	qResp, err := client.Get(base + "/query-check?test_param=special_value_123")
	if err != nil {
		t.Fatalf("query test failed: %v", err)
	}
	qBody, _ := io.ReadAll(qResp.Body)
	_ = qResp.Body.Close()
	if string(qBody) != "query=special_value_123" {
		t.Fatalf("expected 'query=special_value_123', got '%s'", string(qBody))
	}

	// D. Test SSE Streaming (Unbuffered response)
	sseResp, err := client.Get(base + "/events")
	if err != nil {
		t.Fatalf("SSE request failed: %v", err)
	}
	sseBody, _ := io.ReadAll(sseResp.Body)
	_ = sseResp.Body.Close()
	if !strings.Contains(string(sseBody), "data: event1") || !strings.Contains(string(sseBody), "data: event2") {
		t.Fatalf("expected SSE events, got: %s", string(sseBody))
	}

	// E. Test Large File Upload (2MB)
	largeData := bytes.Repeat([]byte("A"), 2*1024*1024)
	largeReq, _ := http.NewRequest(http.MethodPost, base+"/echo-body", bytes.NewReader(largeData))
	largeResp, err := client.Do(largeReq)
	if err != nil {
		t.Fatalf("large upload failed: %v", err)
	}
	largeRespBody, _ := io.ReadAll(largeResp.Body)
	_ = largeResp.Body.Close()
	if len(largeRespBody) != len(largeData) {
		t.Fatalf("expected 2MB body echoed, got %d bytes", len(largeRespBody))
	}
}
