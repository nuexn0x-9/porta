package router

import (
	"testing"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
)

func TestRouterSingleService(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(registry.NewService("app", &config.ServiceCfg{
		Port:  3000,
		Route: "/",
	}))

	r := NewRouter(reg)

	// Test exact root
	match, ok := r.Match("/")
	if !ok || match.Service.Name != "app" || match.RewritePath != "/" {
		t.Fatalf("expected match for '/', got match: %+v, ok: %v", match, ok)
	}

	// Test subpath
	match, ok = r.Match("/products/123")
	if !ok || match.Service.Name != "app" || match.RewritePath != "/products/123" {
		t.Fatalf("expected match for '/products/123', got match: %+v, ok: %v", match, ok)
	}
}

func TestRouterLongestPrefixMatch(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(registry.NewService("frontend", &config.ServiceCfg{
		Port:  3000,
		Route: "/",
	}))
	reg.Register(registry.NewService("backend", &config.ServiceCfg{
		Port:      8000,
		Route:     "/api",
		StripPath: false,
	}))
	reg.Register(registry.NewService("backend_v2", &config.ServiceCfg{
		Port:      8001,
		Route:     "/api/v2",
		StripPath: true,
	}))

	r := NewRouter(reg)

	// 1. Should match frontend
	match, ok := r.Match("/about")
	if !ok || match.Service.Name != "frontend" {
		t.Fatalf("expected frontend match, got: %+v", match)
	}

	// 2. Should match backend /api
	match, ok = r.Match("/api/users")
	if !ok || match.Service.Name != "backend" || match.RewritePath != "/api/users" {
		t.Fatalf("expected backend match without strip, got: %+v", match)
	}

	// 3. Should match backend_v2 /api/v2 and strip path
	match, ok = r.Match("/api/v2/orders")
	if !ok || match.Service.Name != "backend_v2" || match.RewritePath != "/orders" {
		t.Fatalf("expected backend_v2 match with strip path '/orders', got: %+v", match)
	}
}
