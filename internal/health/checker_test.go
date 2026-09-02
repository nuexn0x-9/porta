package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
)

func TestHealthCheckerTCPAndHTTP(t *testing.T) {
	// 1. Mock HTTP healthy server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	port, _ := strconv.Atoi(u.Port())

	// 2. Build registry with 1 healthy service and 1 offline service
	reg := registry.NewRegistry()
	reg.Register(registry.NewService("healthy_svc", &config.ServiceCfg{
		Host: "127.0.0.1",
		Port: port,
		HealthCheck: config.HealthCheckConfig{
			Path: "/healthz",
		},
	}))
	reg.Register(registry.NewService("offline_svc", &config.ServiceCfg{
		Host: "127.0.0.1",
		Port: 64999,
	}))

	checker := NewChecker(reg)
	checker.Start(context.Background())
	defer checker.Stop()

	// Wait for probe
	time.Sleep(200 * time.Millisecond)

	hSvc, _ := reg.Get("healthy_svc")
	if hSvc.GetStatus() != registry.StatusOnline {
		t.Fatalf("expected healthy_svc to be ONLINE, got %s", hSvc.GetStatus())
	}

	offSvc, _ := reg.Get("offline_svc")
	if offSvc.GetStatus() != registry.StatusOffline {
		t.Fatalf("expected offline_svc to be OFFLINE, got %s", offSvc.GetStatus())
	}
}
