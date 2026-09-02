package security

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// SafeDialContext creates a custom net.Dialer that strictly prevents dialing non-loopback IP addresses (SSRF Guard)
func SafeDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address format '%s': %w", addr, err)
		}

		// Strictly permit localhost / 127.0.0.1 / ::1
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return dialer.DialContext(ctx, network, net.JoinHostPort("127.0.0.1", port))
		}

		ip := net.ParseIP(host)
		if ip == nil {
			// Resolve hostname and verify all resolved IPs are loopback
			ips, lookupErr := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if lookupErr != nil {
				return nil, fmt.Errorf("SSRF guard: failed to resolve host '%s': %w", host, lookupErr)
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("SSRF guard: no IP found for host '%s'", host)
			}
			for _, resolvedIP := range ips {
				if !resolvedIP.IsLoopback() {
					return nil, fmt.Errorf("SSRF guard: host '%s' resolves to forbidden non-loopback IP '%s'", host, resolvedIP)
				}
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort("127.0.0.1", port))
		}

		if !ip.IsLoopback() {
			return nil, fmt.Errorf("SSRF guard: connection to non-loopback IP '%s' blocked", ip)
		}

		return dialer.DialContext(ctx, network, addr)
	}
}

// IsLoopbackHost checks whether a host string represents local loopback
func IsLoopbackHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" || h == "localhost" || h == "127.0.0.1" || h == "::1" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}
