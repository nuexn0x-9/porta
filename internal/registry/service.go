package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/porta-dev/porta/internal/config"
)

// HealthStatus represents the health state enum of a service
type HealthStatus string

const (
	StatusUnknown   HealthStatus = "UNKNOWN"
	StatusStarting  HealthStatus = "STARTING"
	StatusOnline    HealthStatus = "ONLINE"
	StatusUnhealthy HealthStatus = "UNHEALTHY"
	StatusOffline   HealthStatus = "OFFLINE"
)

// Service represents an active in-memory target service
type Service struct {
	Name        string
	Host        string
	Port        int
	Route       string
	StripPath   bool
	WebSocket   bool
	HealthCheck config.HealthCheckConfig

	mu           sync.RWMutex
	Status       HealthStatus
	LastChecked  time.Time
	FailureCount int
}

// NewService creates a new Service instance from configuration
func NewService(name string, cfg *config.ServiceCfg) *Service {
	ws := true
	if cfg.WebSocket != nil {
		ws = *cfg.WebSocket
	}

	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}

	return &Service{
		Name:        name,
		Host:        host,
		Port:        cfg.Port,
		Route:       cfg.Route,
		StripPath:   cfg.StripPath,
		WebSocket:   ws,
		HealthCheck: cfg.HealthCheck,
		Status:      StatusUnknown,
	}
}

// TargetAddr returns host:port
func (s *Service) TargetAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// TargetURL returns http://host:port
func (s *Service) TargetURL() string {
	return fmt.Sprintf("http://%s", s.TargetAddr())
}

// SetStatus safely updates the service health status
func (s *Service) SetStatus(status HealthStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
	s.LastChecked = time.Now()
}

// GetStatus safely retrieves current health status
func (s *Service) GetStatus() HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// IsAvailable returns true if the service is ready to receive traffic
func (s *Service) IsAvailable() bool {
	st := s.GetStatus()
	return st == StatusOnline || st == StatusStarting || st == StatusUnknown
}
