package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var envVarRegex = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)(?::-([^}]*))?\}`)

// ExpandEnv expands ${VAR} and ${VAR:-default} patterns in the raw configuration string
func ExpandEnv(input string) string {
	return envVarRegex.ReplaceAllStringFunc(input, func(match string) string {
		submatches := envVarRegex.FindStringSubmatch(match)
		if len(submatches) >= 2 {
			varName := submatches[1]
			val, exists := os.LookupEnv(varName)
			if exists && val != "" {
				return val
			}
			if len(submatches) >= 3 && submatches[2] != "" {
				return submatches[2]
			}
		}
		return ""
	})
}

// Load reads and parses a porta.yaml file from the given path
func Load(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", filePath, err)
	}

	return Parse(data)
}

// Parse parses raw YAML bytes with environment variable expansion and defaults
func Parse(data []byte) (*Config, error) {
	expandedStr := ExpandEnv(string(data))

	var cfg Config
	decoder := yaml.NewDecoder(strings.NewReader(expandedStr))
	decoder.KnownFields(false)

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("YAML parsing error: %w", err)
	}

	ApplyDefaults(&cfg)

	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ApplyDefaults applies smart defaults to optional configuration fields
func ApplyDefaults(cfg *Config) {
	if cfg.Project.Environment == "" {
		cfg.Project.Environment = "development"
	}

	if cfg.Tunnel.Provider == "" {
		cfg.Tunnel.Provider = "cloudflare"
	}

	if cfg.Security.Mode == "" {
		cfg.Security.Mode = "public"
	}

	// Single service auto-routing default
	if len(cfg.Services) == 1 {
		for _, svc := range cfg.Services {
			if svc.Route == "" {
				svc.Route = "/"
			}
		}
	}

	// Default host and health check intervals
	for name, svc := range cfg.Services {
		if svc.Host == "" {
			svc.Host = "127.0.0.1"
		}
		if svc.Route == "" {
			svc.Route = "/" + strings.ToLower(name)
		}
		if svc.HealthCheck.Interval == 0 {
			svc.HealthCheck.Interval = 5 * 1000 * 1000 * 1000 // 5s
		}
		if svc.HealthCheck.Timeout == 0 {
			svc.HealthCheck.Timeout = 2 * 1000 * 1000 * 1000 // 2s
		}
		if svc.WebSocket == nil {
			t := true
			svc.WebSocket = &t
		}
	}
}
