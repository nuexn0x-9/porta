package security

import (
	"context"
	"net"
	"testing"
)

func TestSSRFAuditMatrix(t *testing.T) {
	tests := []struct {
		name        string
		targetHost  string
		shouldAllow bool
	}{
		// IPv4 Loopback
		{"IPv4 Loopback 127.0.0.1", "127.0.0.1", true},
		{"IPv4 Loopback 127.0.0.2", "127.0.0.2", true},
		{"IPv4 Localhost String", "localhost", true},

		// IPv4 Private Subnets (RFC 1918)
		{"IPv4 Class A Private 10.0.0.1", "10.0.0.1", false},
		{"IPv4 Class B Private 172.16.0.1", "172.16.0.1", false},
		{"IPv4 Class B Private 172.31.255.254", "172.31.255.254", false},
		{"IPv4 Class C Private 192.168.1.1", "192.168.1.1", false},

		// Cloud Metadata & Link-Local
		{"AWS/GCP Cloud Metadata 169.254.169.254", "169.254.169.254", false},
		{"Link-Local Subnet 169.254.1.1", "169.254.1.1", false},

		// Public Internet IPs
		{"Public DNS 8.8.8.8", "8.8.8.8", false},
		{"Public Cloudflare 1.1.1.1", "1.1.1.1", false},

		// IPv6 Loopback & Non-Loopback
		{"IPv6 Loopback ::1", "::1", true},
		{"IPv6 Unique Local fc00::1", "fc00::1", false},
		{"IPv6 Link-Local fe80::1", "fe80::1", false},
		{"IPv6 Global 2001:4860:4860::8888", "2001:4860:4860::8888", false},

		// IPv4-mapped IPv6
		{"IPv4-mapped IPv6 Loopback ::ffff:127.0.0.1", "::ffff:127.0.0.1", true},
		{"IPv4-mapped IPv6 Private ::ffff:192.168.1.1", "::ffff:192.168.1.1", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allowed := IsLoopbackHost(tc.targetHost)
			if allowed != tc.shouldAllow {
				t.Errorf("IsLoopbackHost(%s) = %v; want %v", tc.targetHost, allowed, tc.shouldAllow)
			}
		})
	}
}

func TestSafeDialerBlocksNonLoopbackDial(t *testing.T) {
	dialer := SafeDialContext()
	ctx := context.Background()

	// Attempt dialing forbidden private IP
	_, err := dialer(ctx, "tcp", "192.168.1.1:80")
	if err == nil {
		t.Fatalf("expected SafeDialContext to block dialing 192.168.1.1:80, but it succeeded")
	}

	// Attempt dialing cloud metadata IP
	_, err = dialer(ctx, "tcp", "169.254.169.254:80")
	if err == nil {
		t.Fatalf("expected SafeDialContext to block dialing 169.254.169.254:80, but it succeeded")
	}

	// Dialing valid local loopback listener
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind test listener: %v", err)
	}
	defer l.Close()

	conn, err := dialer(ctx, "tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("expected SafeDialContext to allow dialing %s, got error: %v", l.Addr().String(), err)
	}
	_ = conn.Close()
}
