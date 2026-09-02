package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
	"github.com/porta-dev/porta/internal/router"
	"github.com/porta-dev/porta/internal/security"
	"github.com/porta-dev/porta/internal/ui"
)

// Gateway is the local reverse proxy server binding on an ephemeral loopback port
type Gateway struct {
	cfg        *config.Config
	reg        *registry.Registry
	router     *router.Router
	server     *http.Server
	listener   net.Listener
	logger     *ui.Logger
	transport  *http.Transport
	listenAddr string
	port       int
}

// NewGateway creates a new reverse proxy gateway instance
func NewGateway(cfg *config.Config, reg *registry.Registry) *Gateway {
	r := router.NewRouter(reg)
	logger := ui.GetLogger()
	transport := NewProxyTransport()

	g := &Gateway{
		cfg:       cfg,
		reg:       reg,
		router:    r,
		logger:    logger,
		transport: transport,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", g.handleRequest)

	g.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return g
}

// Start binds to an ephemeral loopback port (127.0.0.1:0) and begins listening
func (g *Gateway) Start() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to bind ephemeral gateway port: %w", err)
	}

	g.listener = l
	g.listenAddr = l.Addr().String()
	g.port = l.Addr().(*net.TCPAddr).Port

	go func() {
		if err := g.server.Serve(l); err != nil && err != http.ErrServerClosed {
			// server closed
		}
	}()

	return g.listenAddr, nil
}

// GetAddress returns the local listening address (e.g. 127.0.0.1:54123)
func (g *Gateway) GetAddress() string {
	return g.listenAddr
}

// GetPort returns the allocated dynamic port
func (g *Gateway) GetPort() int {
	return g.port
}

// Stop gracefully shuts down the gateway server
func (g *Gateway) Stop(ctx context.Context) error {
	if g.server != nil {
		return g.server.Shutdown(ctx)
	}
	return nil
}

// handleRequest is the primary ingress dispatcher
func (g *Gateway) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 1. Security Gatekeeper Check
	if g.cfg.Security.Mode == "password" {
		if !security.CheckBasicAuth(r, g.cfg.Security.Password) {
			security.RequireBasicAuth(w)
			g.logger.LogRequest(r, http.StatusUnauthorized, time.Since(start), "security", "Unauthorized: Basic credentials invalid")
			return
		}
	} else if g.cfg.Security.Mode == "token" {
		if !security.CheckTokenAuth(r, g.cfg.Security.Token) {
			security.RequireTokenAuth(w)
			g.logger.LogRequest(r, http.StatusForbidden, time.Since(start), "security", "Forbidden: Bearer or query token invalid")
			return
		}
	}

	// 2. Route Matching via Longest Prefix Match
	match, found := g.router.Match(r.URL.Path)
	if !found {
		Serve404NotFound(w, r)
		g.logger.LogRequest(r, http.StatusNotFound, time.Since(start), "none", "No route matched")
		return
	}

	svc := match.Service
	targetAddr := svc.TargetAddr()

	// 3. Check Service Health
	if !svc.IsAvailable() {
		Serve502BadGateway(w, r, svc.Name, targetAddr)
		g.logger.LogRequest(r, http.StatusBadGateway, time.Since(start), targetAddr, "Service offline")
		return
	}

	// 4. WebSocket Upgrade Handling
	if svc.WebSocket && IsWebSocketRequest(r) {
		wsErr := ProxyWebSocket(w, r, targetAddr, match.RewritePath)
		status := http.StatusSwitchingProtocols
		errStr := ""
		if wsErr != nil {
			status = http.StatusBadGateway
			errStr = wsErr.Error()
		}
		g.logger.LogRequest(r, status, time.Since(start), targetAddr, errStr)
		return
	}

	// 5. Standard HTTP Reverse Proxy
	targetURL, err := url.Parse(match.TargetURL)
	if err != nil {
		http.Error(w, "500 Internal Server Error (PORTA: Target URL Parse Error)", http.StatusInternalServerError)
		g.logger.LogRequest(r, http.StatusInternalServerError, time.Since(start), targetAddr, err.Error())
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = g.transport
	proxy.FlushInterval = -1 // Zero buffer delay for SSE & LLM streaming

	// Custom Director for path rewrite and headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = match.RewritePath
		req.Host = targetAddr

		// Header Injection & Sanitation
		if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			if prior, ok := req.Header["X-Forwarded-For"]; ok {
				clientIP = strings.Join(prior, ", ") + ", " + clientIP
			}
			req.Header.Set("X-Forwarded-For", clientIP)
		}
		req.Header.Set("X-Forwarded-Proto", "https")
		if r.Host != "" {
			req.Header.Set("X-Forwarded-Host", r.Host)
		}
	}

	// Response Interceptor for status recording
	rw := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		svc.SetStatus(registry.StatusOffline)
		Serve502BadGateway(w, r, svc.Name, targetAddr)
		rw.statusCode = http.StatusBadGateway
	}

	proxy.ServeHTTP(rw, r)
	g.logger.LogRequest(r, rw.statusCode, time.Since(start), targetAddr, "")
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
