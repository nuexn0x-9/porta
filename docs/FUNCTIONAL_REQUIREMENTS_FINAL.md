# FINAL FUNCTIONAL REQUIREMENTS SPECIFICATION

**Product:** PORTA  
**Version:** v1.0.0 (MVP)  
**Document Status:** OFFICIAL RELEASE BASELINE  

---

## 1. Requirement Specification Matrix

```text
Requirement Taxonomy:
PRD-F-xxx: Core Functional Requirements
SERVICE-xxx: Cardinality & Architecture Requirements
```

---

### PRD-F-001: Automatic Workspace Initialization & Safe Port Discovery
- **Name:** Workspace Init & Discovery
- **Description:** Scans active local listening ports on common web development ranges (`3000`, `5173`, `8000`, `8080`, `9000`), explicitly ignores infrastructure/database ports (`5432`, `3306`, `6379`, `27017`), and generates a ready-to-use `porta.yaml`.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/cli/init.go`, `TestInitAndConfigCommands` in `cli_test.go`.

---

### PRD-F-002: Declarative Configuration Parser & Environment Expansion
- **Name:** Declarative Configuration Engine
- **Description:** Parses `porta.yaml`, performs strict schema validation, applies smart defaults, and dynamically expands environment variables (`${VAR:-default}`).
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/config/parser.go`, `TestEnvVarExpansion` in `validator_test.go`.

---

### PRD-F-003: Multi-Service Registry & Route Collision Prevention
- **Name:** Service Registry & Collision Guard
- **Description:** Maintains an in-memory service collection backing the runtime, enforces `services.length >= 1`, validates port ranges (`1..65535`), and prevents duplicate route declarations with `ERR_ROUTE_COLLISION`.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/registry/registry.go`, `TestRouteCollisionFails` in `validator_test.go`.

---

### PRD-F-004: Embedded Reverse Proxy Gateway on Ephemeral Port
- **Name:** Local Reverse Proxy Gateway
- **Description:** Binds an HTTP/1.1 and HTTP/2 reverse proxy on a dynamic OS-assigned ephemeral loopback port (`127.0.0.1:0`), dispatches requests via Longest Prefix Matching (LPM), and supports optional prefix path stripping (`strip_path: true`).
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/proxy/gateway.go`, `TestGatewayEndToEndRouting` in `gateway_test.go`.

---

### PRD-F-005: Transparent WebSocket Hijacking & Real-Time Streaming
- **Name:** Full-Duplex Connection Forwarder
- **Description:** Intercepts `Upgrade: websocket` requests, executes HTTP connection hijacking (`http.Hijacker`), establishes bi-directional TCP piping, and disables buffering (`FlushInterval: -1`) for Server-Sent Events (SSE) and LLM streaming.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/proxy/websocket.go`, `TestWebSocketProxyEcho` in `websocket_test.go`.

---

### PRD-F-006: Header Sanitization & Standard Proxy Forwarding
- **Name:** Proxy Header Forwarding
- **Description:** Injects and standardizes `X-Forwarded-For`, `X-Forwarded-Proto: https`, `X-Forwarded-Host`, and `X-Real-IP` headers while rewriting upstream `Host` headers.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/proxy/gateway.go`, `TestGatewayEndToEndRouting` in `gateway_test.go`.

---

### PRD-F-007: Cloudflare Quick Tunnel Provider & Driver Cache
- **Name:** Zero-Auth Cloudflare Tunnel Driver
- **Description:** Implements generic `TunnelProvider` abstraction, automatically downloads and caches `cloudflared` to `~/.porta/bin/` if missing, manages child process lifecycle, extracts assigned public URL via regex, and auto-reconnects on network drops.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/tunnel/cloudflare/driver.go`, `EnsureCloudflaredBinary` in `installer.go`, `porta doctor`.

---

### PRD-F-008: Target Health Checking & Graceful 502 Degradation
- **Name:** Service Health Monitor & Fault Isolation
- **Description:** Probes registered services via TCP socket ping or HTTP GET every 5 seconds, maintains service status state machine (`UNKNOWN`, `STARTING`, `ONLINE`, `OFFLINE`), and serves responsive 502 Bad Gateway pages for offline services while keeping healthy services operational.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/health/checker.go`, `TestHealthCheckerTCPAndHTTP` in `checker_test.go`, `TestGatewayEndToEndRouting` (Test 3 & 4).

---

### PRD-F-009: Security Gatekeeper with Credential & Token Log Redaction
- **Name:** Ingress Access Control & Safe Logging
- **Description:** Challenges public requests using HTTP Basic Auth (`401 Unauthorized`) or Bearer/Query Tokens (`403 Forbidden`) with constant-time string comparison (`crypto/subtle`). Automatically masks query parameters (`?porta_token=[REDACTED]`) and `Authorization` headers across all logs.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/security/basic_auth.go`, `internal/ui/logger.go`, `TestCheckBasicAuth`, `TestCheckTokenAuth`, `TestRedactURLAndPath`.

---

### PRD-F-010: Strict Loopback SSRF Isolation Guard
- **Name:** Network Security Isolation Guard
- **Description:** Enforces that all upstream targets resolve exclusively to loopback addresses (`127.0.0.1`, `localhost`, `::1`), blocking all private subnets (RFC 1918), cloud metadata endpoints (`169.254.169.254`), and DNS rebinding attacks at dial time.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/security/guard.go`, `TestSSRFAuditMatrix`, `TestSafeDialerBlocksNonLoopbackDial` in `ssrf_audit_test.go`.

---

### PRD-F-011: Foreground Terminal UI (TUI) & Live Access Log Stream
- **Name:** Interactive Foreground Terminal UI
- **Description:** Renders a clean, colorized ANSI status screen displaying project name, public HTTPS URL, registered routes, real-time health indicators, and a live scrolling request log feed in the foreground terminal.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/ui/tui.go`, `internal/ui/logger.go`.

---

### PRD-F-012: System Health Diagnostics (`porta doctor`)
- **Name:** System Doctor Diagnostics Command
- **Description:** Verifies OS architecture, 127.0.0.1 loopback bindability, outbound internet connectivity, Cloudflare driver availability, and `porta.yaml` validity with actionable output.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `internal/cli/doctor.go`, verified via `porta doctor` execution.

---

### SERVICE-001: Unified Service Cardinality (DECISION SERVICE-001)
- **Name:** Unified Single & Multi-Service Architecture
- **Description:** Single-service applications (`localhost:3000`) and multi-service applications (`localhost:3000` + `localhost:8000`) execute through the exact same unified internal pipeline without special-cased separate managers.
- **Priority:** P0 (Must Have - MVP)
- **Implementation Status:** Verified & Implemented
- **Verification Evidence:** `TestSingleServiceSucceeds` & `TestMultiServiceSucceeds` in `validator_test.go`.

---
*End of Final Functional Requirements Specification.*
