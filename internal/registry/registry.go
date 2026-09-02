package registry

import (
	"fmt"
	"sync"

	"github.com/porta-dev/porta/internal/config"
)

// Registry manages the collection of active target services
type Registry struct {
	mu       sync.RWMutex
	services map[string]*Service
	order    []string
}

// NewRegistry creates a new Service Registry
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]*Service),
		order:    make([]string, 0),
	}
}

// BuildFromConfig populates the Registry from a parsed Config
func BuildFromConfig(cfg *config.Config) *Registry {
	reg := NewRegistry()
	for name, svcCfg := range cfg.Services {
		reg.Register(NewService(name, svcCfg))
	}
	return reg
}

// Register adds or updates a service in the registry
func (r *Registry) Register(svc *Service) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[svc.Name]; !exists {
		r.order = append(r.order, svc.Name)
	}
	r.services[svc.Name] = svc
}

// Get retrieves a service by name
func (r *Registry) Get(name string) (*Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", name)
	}
	return svc, nil
}

// List returns all registered services in deterministic order
func (r *Registry) List() []*Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Service, 0, len(r.order))
	for _, name := range r.order {
		if svc, exists := r.services[name]; exists {
			list = append(list, svc)
		}
	}
	return list
}

// Count returns the number of registered services
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.services)
}
