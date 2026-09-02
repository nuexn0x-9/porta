package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckBasicAuth(t *testing.T) {
	// Test standard user:pass
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "secret123")

	if !CheckBasicAuth(req, "admin:secret123") {
		t.Fatalf("expected valid basic auth for admin:secret123")
	}

	if CheckBasicAuth(req, "admin:wrong") {
		t.Fatalf("expected invalid basic auth for wrong password")
	}

	// Test default user "porta" with single password
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.SetBasicAuth("porta", "mypassword")
	if !CheckBasicAuth(req2, "mypassword") {
		t.Fatalf("expected valid basic auth with default porta user")
	}
}

func TestCheckTokenAuth(t *testing.T) {
	configuredToken := "valid-token-xyz"

	// 1. Authorization: Bearer
	reqBearer := httptest.NewRequest(http.MethodGet, "/", nil)
	reqBearer.Header.Set("Authorization", "Bearer valid-token-xyz")
	if !CheckTokenAuth(reqBearer, configuredToken) {
		t.Fatalf("expected valid token from Authorization header")
	}

	// 2. Query param ?porta_token=
	reqQuery := httptest.NewRequest(http.MethodGet, "/?porta_token=valid-token-xyz", nil)
	if !CheckTokenAuth(reqQuery, configuredToken) {
		t.Fatalf("expected valid token from query param")
	}

	// 3. Invalid token
	reqInvalid := httptest.NewRequest(http.MethodGet, "/?porta_token=invalid", nil)
	if CheckTokenAuth(reqInvalid, configuredToken) {
		t.Fatalf("expected invalid token to fail")
	}
}
