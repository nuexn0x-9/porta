package ui

import (
	"net/url"
	"testing"
)

func TestRedactURLAndPath(t *testing.T) {
	rawURL := "http://localhost:8080/api/v1/data?porta_token=super_secret_token&limit=10"
	redacted := RedactURL(rawURL)
	expected := "http://localhost:8080/api/v1/data?porta_token=[REDACTED]&limit=10"

	if redacted != expected {
		t.Fatalf("expected '%s', got '%s'", expected, redacted)
	}

	u, _ := url.Parse("/api/v1/data?porta_token=super_secret_token&page=2")
	redactedPath := RedactPath(u)
	expectedPath := "/api/v1/data?porta_token=[REDACTED]&page=2"

	if redactedPath != expectedPath {
		t.Fatalf("expected '%s', got '%s'", expectedPath, redactedPath)
	}
}
