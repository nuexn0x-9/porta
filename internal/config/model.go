package config

import (
	"fmt"
	"time"
)

// Config represents the root configuration model for porta.yaml
type Config struct {
	Version  string                 `yaml:"version,omitempty"`
	Project  ProjectConfig          `yaml:"project"`
	Services map[string]*ServiceCfg `yaml:"services"`
	Tunnel   TunnelConfig           `yaml:"tunnel,omitempty"`
	Security SecurityConfig         `yaml:"security,omitempty"`
}

// ProjectConfig represents the project metadata
type ProjectConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment,omitempty"`
}

// ServiceCfg represents an individual service definition
type ServiceCfg struct {
	Host        string            `yaml:"host,omitempty"`
	Port        int               `yaml:"port"`
	Route       string            `yaml:"route,omitempty"`
	StripPath   bool              `yaml:"strip_path,omitempty"`
	WebSocket   *bool             `yaml:"websocket,omitempty"`
	HealthCheck HealthCheckConfig `yaml:"health_check,omitempty"`
}

// HealthCheckConfig represents configuration for target service probing
type HealthCheckConfig struct {
	Path     string        `yaml:"path,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty"`
	Timeout  time.Duration `yaml:"timeout,omitempty"`
}

// TunnelConfig represents the tunneling provider configuration
type TunnelConfig struct {
	Provider  string `yaml:"provider,omitempty"`
	Subdomain string `yaml:"subdomain,omitempty"`
}

// SecurityConfig represents access control & ingress security settings
type SecurityConfig struct {
	Mode       string   `yaml:"mode,omitempty"`
	Password   string   `yaml:"password,omitempty"`
	Token      string   `yaml:"token,omitempty"`
	AllowedIPs []string `yaml:"allowed_ips,omitempty"`
}

// TargetAddress returns the resolved upstream host:port address
func (s *ServiceCfg) TargetAddress() string {
	host := s.Host
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", host, s.Port)
}

// TargetURL returns the HTTP scheme URL for the upstream service
func (s *ServiceCfg) TargetURL() string {
	return fmt.Sprintf("http://%s", s.TargetAddress())
}
