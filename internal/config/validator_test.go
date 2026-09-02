package config

import (
	"testing"
)

func TestCardinalityZeroServicesFails(t *testing.T) {
	yamlData := `
project:
  name: test-app
services: {}
`
	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Fatalf("expected error for 0 services, got nil")
	}
}

func TestSingleServiceSucceeds(t *testing.T) {
	yamlData := `
project:
  name: single-app
services:
  app:
    port: 3000
`
	cfg, err := Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error for single service: %v", err)
	}

	if len(cfg.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(cfg.Services))
	}

	svc := cfg.Services["app"]
	if svc.Route != "/" {
		t.Fatalf("expected single service route to default to '/', got '%s'", svc.Route)
	}
	if svc.Host != "127.0.0.1" {
		t.Fatalf("expected host '127.0.0.1', got '%s'", svc.Host)
	}
}

func TestMultiServiceSucceeds(t *testing.T) {
	yamlData := `
project:
  name: multi-app
services:
  frontend:
    port: 3000
    route: /
  backend:
    port: 8000
    route: /api
`
	cfg, err := Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error for multi service: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}
}

func TestRouteCollisionFails(t *testing.T) {
	yamlData := `
project:
  name: collision-app
services:
  svc1:
    port: 3000
    route: /api
  svc2:
    port: 8000
    route: /api
`
	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Fatalf("expected route collision error, got nil")
	}
}

func TestSSRFNonLoopbackHostFails(t *testing.T) {
	yamlData := `
project:
  name: ssrf-app
services:
  bad:
    host: 192.168.1.100
    port: 8080
    route: /
`
	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Fatalf("expected SSRF error for non-loopback host, got nil")
	}
}

func TestEnvVarExpansion(t *testing.T) {
	t.Setenv("PORTA_TEST_PASS", "SuperSecret123")
	yamlData := `
project:
  name: env-app
services:
  app:
    port: 3000
    route: /
security:
  mode: password
  password: ${PORTA_TEST_PASS:-default}
`
	cfg, err := Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Security.Password != "SuperSecret123" {
		t.Fatalf("expected password 'SuperSecret123', got '%s'", cfg.Security.Password)
	}
}
