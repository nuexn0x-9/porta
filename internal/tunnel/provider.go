package tunnel

import (
	"context"
	"time"
)

// TunnelSession represents an established public tunnel connection
type TunnelSession struct {
	PublicURL string
	Provider  string
	StartedAt time.Time
}

// TunnelProvider is the generic interface for all tunnel implementations
type TunnelProvider interface {
	// Name returns the provider identifier (e.g. "cloudflare", "ngrok")
	Name() string

	// Start launches the tunnel connection pointing to local gateway address
	Start(ctx context.Context, localGatewayAddr string) (*TunnelSession, error)

	// HealthCheck probes if the tunnel connection is still active and healthy
	HealthCheck(ctx context.Context) error

	// Stop gracefully closes the tunnel session
	Stop() error
}
