package config

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

var (
	projectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Validate performs strict validation on the parsed Config model
func Validate(cfg *Config) error {
	// 1. Project validation
	if strings.TrimSpace(cfg.Project.Name) == "" {
		return fmt.Errorf("validation error: 'project.name' is required")
	}
	if !projectNameRegex.MatchString(cfg.Project.Name) {
		return fmt.Errorf("validation error: 'project.name' must contain only alphanumeric characters, dashes, and underscores (got: %s)", cfg.Project.Name)
	}

	// 2. Cardinality validation (DECISION SERVICE-001)
	if len(cfg.Services) == 0 {
		return fmt.Errorf("validation error (ERR_CFG_NO_SERVICES): at least one service must be defined under 'services'")
	}

	// 3. Service validation & Route collision detection
	routesSeen := make(map[string]string) // route -> serviceName

	for name, svc := range cfg.Services {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("validation error: service name cannot be empty")
		}

		// Port validation
		if svc.Port < 1 || svc.Port > 65535 {
			return fmt.Errorf("validation error: service '%s' has invalid port %d (must be between 1 and 65535)", name, svc.Port)
		}

		// SSRF Guard - Target Host validation
		if err := validateHostIsLoopback(svc.Host); err != nil {
			return fmt.Errorf("security violation in service '%s': %w", name, err)
		}

		// Route validation
		if svc.Route == "" {
			return fmt.Errorf("validation error: service '%s' has an empty route", name)
		}
		if !strings.HasPrefix(svc.Route, "/") {
			return fmt.Errorf("validation error: service '%s' route '%s' must start with '/'", name, svc.Route)
		}

		// Check for duplicate exact route
		normalizedRoute := strings.TrimRight(svc.Route, "/")
		if normalizedRoute == "" {
			normalizedRoute = "/"
		}

		if existingSvc, collision := routesSeen[normalizedRoute]; collision {
			return fmt.Errorf("route collision error (ERR_ROUTE_COLLISION): both service '%s' and service '%s' map to route '%s'", existingSvc, name, svc.Route)
		}
		routesSeen[normalizedRoute] = name
	}

	// 4. Security validation
	switch cfg.Security.Mode {
	case "public", "":
		// Unrestricted
	case "password":
		if strings.TrimSpace(cfg.Security.Password) == "" {
			return fmt.Errorf("validation error: 'security.password' is required when security.mode is 'password'")
		}
	case "token":
		if strings.TrimSpace(cfg.Security.Token) == "" {
			return fmt.Errorf("validation error: 'security.token' is required when security.mode is 'token'")
		}
	default:
		return fmt.Errorf("validation error: invalid security.mode '%s' (must be 'public', 'password', or 'token')", cfg.Security.Mode)
	}

	return nil
}

// validateHostIsLoopback ensures the host is loopback only (SSRF Guard)
func validateHostIsLoopback(host string) error {
	trimmed := strings.TrimSpace(host)
	if trimmed == "" || trimmed == "localhost" || trimmed == "127.0.0.1" || trimmed == "::1" {
		return nil
	}

	ip := net.ParseIP(trimmed)
	if ip != nil {
		if ip.IsLoopback() {
			return nil
		}
		return fmt.Errorf("target host '%s' is not a loopback address (SSRF protection)", host)
	}

	// Resolve hostname to check IPs
	ips, err := net.LookupIP(trimmed)
	if err != nil {
		// If DNS fails during offline dev, check if it's explicitly named localhost
		if strings.EqualFold(trimmed, "localhost") {
			return nil
		}
		return fmt.Errorf("cannot resolve target host '%s': %w", host, err)
	}

	for _, resolvedIP := range ips {
		if !resolvedIP.IsLoopback() {
			return fmt.Errorf("target host '%s' resolves to non-loopback IP '%s' (SSRF protection)", host, resolvedIP.String())
		}
	}

	return nil
}
