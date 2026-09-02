package proxy

import (
	"fmt"
	"net/http"
	"strings"
)

// Serve502BadGateway sends a friendly, informative 502 page when an upstream service is offline
func Serve502BadGateway(w http.ResponseWriter, r *http.Request, serviceName, targetAddr string) {
	w.WriteHeader(http.StatusBadGateway)

	if acceptsJSON(r) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		msg := fmt.Sprintf(`{"error":"Bad Gateway","service":"%s","target":"%s","message":"Upstream local service is offline or unreachable. PORTA is waiting for the service to start.","porta_version":"1.0.0"}`,
			serviceName, targetAddr)
		_, _ = w.Write([]byte(msg))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>502 Bad Gateway - PORTA</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); max-width: 540px; border: 1px solid #334155; }
        .badge { background: #ef4444; color: white; padding: 4px 10px; border-radius: 9999px; font-size: 0.85rem; font-weight: bold; }
        h1 { margin-top: 1rem; font-size: 1.6rem; color: #f1f5f9; }
        p { color: #94a3b8; line-height: 1.6; }
        code { background: #0f172a; padding: 2px 6px; border-radius: 4px; color: #38bdf8; font-family: monospace; }
        .footer { margin-top: 2rem; border-top: 1px solid #334155; padding-top: 1rem; font-size: 0.8rem; color: #64748b; }
    </style>
</head>
<body>
    <div class="card">
        <span class="badge">502 Bad Gateway</span>
        <h1>Service Offline: <code>%s</code></h1>
        <p>PORTA received the public request, but the local upstream service at <code>%s</code> is currently offline or not responding.</p>
        <p><strong>Next Steps:</strong> Start your local development server on port <code>%s</code>. PORTA will automatically resume routing traffic as soon as the service comes online.</p>
        <div class="footer">PORTA v1.0.0 — Local Exposure Gateway</div>
    </div>
</body>
</html>`, serviceName, targetAddr, targetAddr)

	_, _ = w.Write([]byte(html))
}

// Serve404NotFound sends a 404 response when no matching route exists
func Serve404NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	if acceptsJSON(r) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		msg := fmt.Sprintf(`{"error":"Not Found","path":"%s","message":"No matching service route configured in porta.yaml"}`, r.URL.Path)
		_, _ = w.Write([]byte(msg))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(fmt.Sprintf("404 Not Found: No service configured for path '%s' (PORTA)", r.URL.Path)))
}

func acceptsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}
