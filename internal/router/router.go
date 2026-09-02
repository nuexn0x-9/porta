package router

import (
	"sort"
	"strings"

	"github.com/porta-dev/porta/internal/registry"
)

// RouteEntry represents a registered routing rule
type RouteEntry struct {
	Prefix    string
	Service   *registry.Service
	StripPath bool
}

// Router provides Longest Prefix Match (LPM) path routing to services
type Router struct {
	entries []RouteEntry
}

// NewRouter creates a new Router from a Service Registry
func NewRouter(reg *registry.Registry) *Router {
	r := &Router{
		entries: make([]RouteEntry, 0),
	}

	for _, svc := range reg.List() {
		prefix := strings.TrimRight(svc.Route, "/")
		if prefix == "" {
			prefix = "/"
		}
		r.entries = append(r.entries, RouteEntry{
			Prefix:    prefix,
			Service:   svc,
			StripPath: svc.StripPath,
		})
	}

	// Sort routes descending by prefix length so Longest Prefix Match is prioritized
	sort.Slice(r.entries, func(i, j int) bool {
		// Root "/" is lowest priority
		if r.entries[i].Prefix == "/" {
			return false
		}
		if r.entries[j].Prefix == "/" {
			return true
		}
		return len(r.entries[i].Prefix) > len(r.entries[j].Prefix)
	})

	return r
}

// MatchResult contains the matched service and the rewritten path
type MatchResult struct {
	Service      *registry.Service
	MatchedRoute string
	TargetURL    string
	RewritePath  string
}

// Match resolves an incoming request path against registered route entries
func (r *Router) Match(reqPath string) (*MatchResult, bool) {
	cleanPath := reqPath
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	for _, entry := range r.entries {
		if entry.Prefix == "/" {
			// Root route matches everything
			rewritePath := cleanPath
			return &MatchResult{
				Service:      entry.Service,
				MatchedRoute: "/",
				TargetURL:    entry.Service.TargetURL(),
				RewritePath:  rewritePath,
			}, true
		}

		// Exact match or prefix match with path boundary
		if cleanPath == entry.Prefix || strings.HasPrefix(cleanPath, entry.Prefix+"/") {
			rewritePath := cleanPath
			if entry.StripPath {
				rewritePath = strings.TrimPrefix(cleanPath, entry.Prefix)
				if !strings.HasPrefix(rewritePath, "/") {
					rewritePath = "/" + rewritePath
				}
			}
			return &MatchResult{
				Service:      entry.Service,
				MatchedRoute: entry.Prefix,
				TargetURL:    entry.Service.TargetURL(),
				RewritePath:  rewritePath,
			}, true
		}
	}

	return nil, false
}
