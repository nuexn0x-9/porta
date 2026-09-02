# FINAL SYSTEM REQUIREMENTS SPECIFICATION (SRS)

**Product:** PORTA  
**Version:** v1.0.0 (MVP)  
**Document Status:** OFFICIAL RELEASE BASELINE  

---

## 1. System Architecture Topology

```text
                                  ┌──────────────────────────┐
                                  │      PUBLIC INTERNET     │
                                  │ (Browsers, Webhooks, QA) │
                                  └─────────────┬────────────┘
                                                │
                                                ▼  HTTPS (TLS 1.3)
                                  ┌──────────────────────────┐
                                  │   Tunnel Edge Provider   │
                                  │ (Cloudflare Anycast Edge)│
                                  └─────────────┬────────────┘
                                                │
════════════════════════════════════════════════╪═══════════════════════════════════════════════
 DEVELOPER WORKSTATION (PORTA FOREGROUND)       │ Encrypted QUIC / mTLS Tunnel
                                                ▼
                                  ┌──────────────────────────┐
                                  │   PORTA TUNNEL ADAPTER   │  <─── Tunnel Watchdog
                                  │  (cloudflared supervisor)│
                                  └─────────────┬────────────┘
                                                │ 127.0.0.1:<ephemeral_port>
                                                ▼
                                  ┌──────────────────────────┐
                                  │  SECURITY GATEKEEPER     │  <─── Basic Auth / Token / SSRF Guard
                                  └─────────────┬────────────┘
                                                │ Authorized Requests
                                                ▼
                                  ┌──────────────────────────┐
                                  │  EMBEDDED REVERSE PROXY  │  <─── Longest Prefix Match (LPM)
                                  │  (net/http/httputil)     │
                                  └───────┬──────────┬───────┘
                                          │          │
                 ┌────────────────────────┘          └────────────────────────┐
                 │                                                            │
                 ▼ Path: `/` (WebSocket Support)                             ▼ Path: `/api/*`
     ┌───────────────────────┐                                    ┌───────────────────────┐
     │    Frontend Service   │                                    │    Backend Service    │
     │  http://127.0.0.1:3000│                                    │ http://127.0.0.1:8000 │
     └───────────────────────┘                                    └───────────────────────┘
```

---

## 2. Component Specifications

### 2.1 PORTA CLI (`internal/cli`, `cmd/porta`)
- **Purpose:** Primary command-line interface entry point.
- **Responsibilities:** Command routing via Cobra, flag parsing, formatted help text, signal trapping (`SIGINT`, `SIGTERM`), OS exit codes.
- **Dependencies:** `internal/config`, `internal/registry`, `internal/proxy`, `internal/tunnel`, `internal/ui`.

### 2.2 Config Engine (`internal/config`)
- **Purpose:** Parse, validate, and inject defaults into `porta.yaml`.
- **Responsibilities:** YAML unmarshaling, `${ENV:-default}` expansion, port validation (`1..65535`), cardinality check (`len(services) >= 1`), route collision prevention.
- **Dependencies:** `gopkg.in/yaml.v3`.

### 2.3 Service Registry (`internal/registry`)
- **Purpose:** In-memory collection of target upstream applications.
- **Responsibilities:** Thread-safe service storage, health state management (`UNKNOWN`, `STARTING`, `ONLINE`, `OFFLINE`), deterministic listing.
- **Dependencies:** `internal/config`.

### 2.4 Router (`internal/router`)
- **Purpose:** Ingress path resolution engine.
- **Responsibilities:** Longest Prefix Match (LPM) route matching, prefix path stripping (`strip_path: true`), route normalization.
- **Dependencies:** `internal/registry`.

### 2.5 Reverse Proxy Gateway (`internal/proxy`)
- **Purpose:** Local HTTP and WebSocket traffic forwarding gateway.
- **Responsibilities:** Ephemeral loopback port binding (`127.0.0.1:0`), standard header injection (`X-Forwarded-*`), WebSocket connection hijacking (`http.Hijacker`), unbuffered SSE streaming (`FlushInterval: -1`), responsive 502/404 HTML and JSON error templates.
- **Dependencies:** `internal/config`, `internal/registry`, `internal/router`, `internal/security`, `internal/ui`.

### 2.6 Security Gatekeeper & SSRF Guard (`internal/security`)
- **Purpose:** Ingress access control and network isolation firewall.
- **Responsibilities:** HTTP Basic Auth (RFC 7617), Bearer/Query Token verification with constant-time comparison (`crypto/subtle`), dial-time loopback IP enforcement (`SafeDialContext`).
- **Dependencies:** Go Standard Library (`crypto/subtle`, `net`, `net/http`).

### 2.7 Health Checker Engine (`internal/health`)
- **Purpose:** Active background target probing and fault isolation.
- **Responsibilities:** TCP socket ping, HTTP GET path probing every 5 seconds, state transition updates.
- **Dependencies:** `internal/registry`, `internal/security`.

### 2.8 Tunnel Manager (`internal/tunnel`, `internal/tunnel/cloudflare`)
- **Purpose:** Public egress tunnel provider orchestration.
- **Responsibilities:** Generic `TunnelProvider` interface implementation, on-demand driver downloader (`EnsureCloudflaredBinary`), process supervision (`CREATE_NO_WINDOW` / `Setpgid`), stderr regex URL extraction, exponential backoff reconnection.
- **Dependencies:** Go Standard Library (`os/exec`, `bufio`, `regexp`).

### 2.9 UI Subsystem & Logger (`internal/ui`)
- **Purpose:** Live terminal status rendering and safe access logging.
- **Responsibilities:** ANSI colorized TUI table, structured JSON logging to `.porta/logs/access.log`, mandatory token masking (`?porta_token=[REDACTED]`).
- **Dependencies:** Go Standard Library (`fmt`, `sync`, `regexp`, `encoding/json`).

---
*End of Final System Requirements Specification.*
