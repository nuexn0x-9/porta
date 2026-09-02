package proxy

import (
	"net/http"
	"time"

	"github.com/porta-dev/porta/internal/security"
)

// NewProxyTransport creates an optimized http.Transport for the local reverse proxy
func NewProxyTransport() *http.Transport {
	return &http.Transport{
		DialContext:           security.SafeDialContext(),
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		DisableCompression:    false,
		ForceAttemptHTTP2:     false,
	}
}
