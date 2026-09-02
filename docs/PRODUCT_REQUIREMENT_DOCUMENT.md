# PRODUCT REQUIREMENT DOCUMENT (PRD)

**Product Name:** PORTA  
**Tagline:** Expose your local multi-service application to the world with one command.  
**Document Version:** 1.0.1  
**Document Status:** REVISED PRODUCT REQUIREMENT BASELINE  
**Product Lead / Author:** Senior Product Manager & Technical Product Architect  
**Target Release:** PORTA v1.0.0 (MVP)  
**Last Updated:** September 2026  

---

## 1. DOCUMENT CONTROL

| Attribute | Details |
| :--- | :--- |
| **Document Owner** | Senior Product Manager & Technical Product Architect |
| **Target Audience** | Core Engineering Team, QA Engineers, Implementation Planning Agents |
| **Specification Source of Truth** | `docs/FUNCTIONAL_REQUIREMENT_MAP.md` (FRM v1.0.0)<br>`docs/SYSTEM_REQUIREMENT_SPECIFICATION.md` (SRS v1.0.0) |
| **Approved Decision Baseline** | Approved Product Decision Baseline (Zero cloud hosting, local-first client, CLI+lightweight TUI, foreground execution, Cloudflare Quick Tunnel default with on-demand download & caching, ephemeral gateway port, embedded Go reverse proxy) |
| **Revision History** | • **v1.0.0:** Initial comprehensive baseline from approved FRM & SRS.<br>• **v1.0.1:** Clarified MVP foreground runtime model; resolved Cloudflare binary distribution decision (on-demand download + local cache); locked local gateway to ephemeral loopback port; updated performance metrics to engineering benchmark targets (To Be Validated); strengthened security requirements (mandatory token/credential redaction in all logs, strict SSRF/loopback isolation); clarified Credential Store scope (deferred to post-MVP provider support); tightened MVP boundaries. |

---

## 2. EXECUTIVE SUMMARY

**PORTA** is a lightweight, single-binary developer tool that transforms a local multi-service application running on `localhost` (such as frontend at `:3000`, backend API at `:8000`, and realtime WebSocket server at `:9001`) into a secure, publicly accessible HTTPS environment via a single CLI command.

PORTA removes all traditional developer friction associated with external exposure—eliminating the need for manual cloud deployments, server provisioning, static public IP acquisition, router port forwarding, or juggling multiple tunnel URLs with CORS misconfigurations. Operating entirely on the client side with zero cloud hosting overhead, PORTA executes as a foreground terminal runtime that orchestrates an embedded Go reverse proxy gateway on an ephemeral loopback port, runs active service health checks, enforces security gatekeeping with credential-safe logging, and establishes an encrypted public egress tunnel using Cloudflare Quick Tunnel as the default zero-configuration provider via an on-demand cached binary.

---

## 3. PRODUCT OVERVIEW

Modern web applications are inherently multi-service, comprising web frontends, REST/GraphQL APIs, background workers, and WebSocket streams running across multiple local ports. When developers need to share their work with external stakeholders (clients, remote QA testers, mobile devices on cellular networks, or third-party webhook dispatchers like Stripe or Midtrans), existing tools force them into two painful extremes:
1. Running complex, multi-instance free tunnels that produce separate disjointed URLs causing severe CORS breakage.
2. Deploying unfinished branch code to staging cloud infrastructure (Vercel, AWS, Render) which is slow and disrupts local iteration speed.

PORTA unifies all locally running services behind a single declarative file (`porta.yaml`), binds them into a consolidated local reverse proxy gateway, and exposes them through a single public HTTPS domain with intelligent path-based routing.

---

## 4. PRODUCT VISION

To become the standard developer utility for instant local application exposure—enabling any software engineer anywhere in the world to securely share their complete local working environment with team members, clients, and automated testing services in under two seconds.

---

## 5. PRODUCT MISSION

Deliver a blazing-fast, robust, zero-dependency CLI tool that executes the promise:
> **"Expose your local multi-service application to the world with one command."**

---

## 6. PROBLEM STATEMENT

1. **Multi-Service Exposure Friction:** Traditional tunneling utilities (e.g., standard free ngrok) only expose a single port per tunnel session. Exposing a full-stack app requires multiple tunnels, different hostnames, and painful manual CORS configuration.
2. **Slow Collaboration Feedback Loop:** Sharing a feature branch with a client or QA tester requires building, pushing, CI/CD pipeline execution, and cloud deployment, wasting 10–30 minutes per iteration.
3. **Webhook Testing Bottlenecks:** External webhook providers (payment gateways, OAuth, GitHub apps) require valid public HTTPS endpoints that cannot hit `http://localhost:8000` directly.
4. **Setup Complexity:** Enterprise tunneling or VPN tools require account registrations, credit cards, DNS configuration, and router NAT adjustments that break zero-friction workflows.

---

## 7. VALUE PROPOSITION

| Stakeholder / Persona | Without PORTA | With PORTA |
| :--- | :--- | :--- |
| **Full-Stack Developer** | Runs 2 separate tunnels, modifies frontend API baseUrl, wrestles with browser CORS errors. | Runs `porta start`. Frontend and Backend are available under 1 unified HTTPS URL (`/` and `/api`). |
| **Frontend / Mobile Engineer** | Deploys backend to staging just to test mobile app on a real smartphone. | Points smartphone browser or app to PORTA's public HTTPS URL in real-time. |
| **QA / Reviewer** | Waits for CI/CD builds to finish staging deployment before testing bugs. | Immediately tests the developer's live local code branch with zero wait time. |
| **Webhook Integrator** | Manually mocks payloads or uses third-party relay proxies. | Provides live PORTA HTTPS URL directly to payment gateway dashboard. |

---

## 8. PRODUCT POSITIONING

```text
               High Setup Complexity / Enterprise
                               ▲
                               │   Tailscale / Cloudflare Zero Trust
                               │   (Requires Account, Domain, VPN Client)
                               │
Single-Port Focus ─────────────┼───────────── Multi-Service Consolidated
(ngrok, Localtunnel,           │
 TryCloudflare CLI raw)        │   ★ PORTA (Single command, Multi-Service,
                               │            Embedded Proxy, Zero-Auth Default,
                               │            Foreground CLI Runtime)
                               │
                               ▼
               Zero Setup Friction / Instant CLI
```

PORTA is positioned as a **Developer Productivity Tool**—not a hosting platform, not a cloud infrastructure provider, but a local developer companion that makes local environments temporarily public.

---

## 9. TARGET USERS / PERSONAS

### Persona 1: Full-Stack Web Developer (Primary Operator)
- **Profile:** Alex, Senior Full-Stack Engineer working with Next.js (`:3000`) and Go backend (`:8000`).
- **Needs:** Wants to demo live checkout flow to a client during a video call without deploying uncommitted code.
- **Pain Point:** Client cannot access `localhost`. Deploying to Vercel/Fly.io requires merging unfinished PR.

### Persona 2: Mobile App Developer & Integrator (Primary Operator)
- **Profile:** Rian, Flutter/React Native Developer.
- **Needs:** Testing the mobile app on a physical Android/iOS phone over 4G LTE against local backend APIs.
- **Pain Point:** Phone cannot connect to `127.0.0.1` on developer's laptop over cellular data.

### Persona 3: QA Engineer / External Reviewer (Consumer Persona)
- **Profile:** Sarah, QA Lead validating user stories.
- **Needs:** Accessing the developer's exact branch environment to verify bug fixes before signing off.
- **Pain Point:** Staging environment is often broken or blocked by other PR deployments.

---

## 10. USER PROBLEMS & PRODUCT REMEDIES

| ID | User Problem | PORTA Product Remedy |
| :--- | :--- | :--- |
| **UP-01** | Frontend can't talk to backend over separate public tunnels due to CORS. | Path-based reverse proxy routing (`/` $\rightarrow$ frontend, `/api` $\rightarrow$ backend) under single origin. |
| **UP-02** | Exposing localhost opens security vulnerabilities to unintended internet scanners. | Integrated Security Gatekeeper offering Basic Auth (`user:pass`) and Bearer Token protection. |
| **UP-03** | Upstream backend crashes or restarts, breaking the entire tunnel session. | Self-healing gateway: returns 502 placeholder for offline services while keeping tunnel alive. |
| **UP-04** | Installing runtime requires installing Node.js/Python or heavy Docker daemon. | Single static Go binary (`porta` / `porta.exe`) with zero external runtime dependencies. |
| **UP-05** | Real-time features (WebSocket, Chat, SSE) fail through standard proxies. | Native HTTP connection hijacking and unbuffered chunked stream forwarding. |
| **UP-06** | Sensitive auth tokens leaked into terminal or disk access logs. | Mandatory credential/token redaction across all logging sinks (TUI, stdout, stderr, `access.log`). |

---

## 11. PRODUCT GOALS

1. **G-01 (One-Command Experience):** Once configured, allow developers to start a multi-service public environment with just `porta start`.
2. **G-02 (Zero-Friction Default):** Provide an instant public HTTPS URL on the first run without requiring account registration, login tokens, or credit cards.
3. **G-03 (Unified Ingress):** Route multiple local ports through a single public HTTPS endpoint with deterministic path-matching.
4. **G-04 (Terminal Native UX):** Deliver a clean, lightweight TUI and structured CLI output that provides total runtime visibility in the foreground without needing a browser dashboard.
5. **G-05 (Robust Cross-Platform Support):** Guarantee identical behavior and stability across Windows, macOS, and Linux.

---

## 12. NON-GOALS

The following capabilities are explicitly declared **OUT OF SCOPE** for PORTA MVP:

- **NG-01 (No Cloud Hosting):** PORTA will not store application code, host databases, or run servers on cloud infrastructure.
- **NG-02 (No Process Compilation / Management):** PORTA MVP will not replace `npm start`, `go run`, or `docker compose`. Developers manage their own local application lifecycles.
- **NG-03 (No Enterprise Identity / SSO):** PORTA will not act as a SAML/OAuth2 Identity Provider (IdP).
- **NG-04 (No Web Management Dashboard):** PORTA MVP will not serve a browser-based admin UI.
- **NG-05 (No Raw TCP/UDP Tunneling):** PORTA MVP focuses strictly on HTTP, HTTPS, WebSocket, and SSE traffic.
- **NG-06 (No Permanent Production Hosting):** PORTA environments are ephemeral and intended for testing, review, and development.
- **NG-07 (No Detached Daemon Architecture in MVP):** PORTA MVP executes exclusively as an attached foreground CLI process. Detached background daemon management is post-MVP.
- **NG-08 (No Custom Gateway Port Configuration in MVP):** Local reverse proxy binds strictly to an ephemeral loopback port (`127.0.0.1:0`).

---

## 13. PRODUCT PRINCIPLES

1. **Zero-Configuration by Default:** If standard conventions exist (e.g., `127.0.0.1`, port 3000), PORTA infers them automatically.
2. **Local-First & Client-Side:** Everything runs on the developer's machine. Zero external control plane servers required for MVP.
3. **Foreground Simplicity:** The MVP runs attached to the terminal; `Ctrl+C` initiates instant, clean teardown.
4. **Fail Gracefully, Never Panic:** If an upstream service is down, the tunnel remains active, and a clear diagnostic message is shown to both the terminal and public client.
5. **Security by Awareness & Isolation:** Strictly reject non-loopback targets to prevent SSRF; redact all tokens in logs; challenge public traffic when configured.
6. **High Performance, Minimal Footprint:** Lightweight memory and CPU profile with minimal routing overhead.

---

## 14. CORE PRODUCT CONCEPT & ARCHITECTURAL TOPOLOGY

```text
                            ┌───────────────────────────────────┐
                            │          PUBLIC INTERNET          │
                            │ (Client, QA, Smartphone, Webhook) │
                            └─────────────────┬─────────────────┘
                                              │
                                              ▼ HTTPS (TLS 1.3 Edge)
                            ┌───────────────────────────────────┐
                            │       Tunnel Edge Provider        │
                            │     (Cloudflare Quick Tunnel)     │
                            └─────────────────┬─────────────────┘
                                              │
══════════════════════════════════════════════╪═══════════════════════════════════════════════
 DEVELOPER WORKSTATION (PORTA FOREGROUND RUNTIME) Encrypted Stream (QUIC/mTLS)
                                              ▼
                            ┌───────────────────────────────────┐
                            │       Tunnel Engine Adapter       │
                            │ (On-Demand Download & Local Cache)│
                            └─────────────────┬─────────────────┘
                                              │ 127.0.0.1:<ephemeral_port>
                                              ▼
                            ┌───────────────────────────────────┐
                            │      Security Gatekeeper          │
                            │ (Basic Auth / Bearer / SSRF Guard)│
                            └─────────────────┬─────────────────┘
                                              │ Authorized Request
                                              ▼
                            ┌───────────────────────────────────┐
                            │    Embedded Go Reverse Proxy      │
                            │    (Longest Prefix Match Router)  │
                            └─────────┬───────────────┬─────────┘
                                      │               │
                 ┌────────────────────┘               └────────────────────┐
                 │                                                         │
                 ▼ Path: `/` (WebSocket Support)                           ▼ Path: `/api/*`
     ┌───────────────────────┐                                 ┌───────────────────────┐
     │    Frontend App       │                                 │      Backend API      │
     │ http://127.0.0.1:3000 │                                 │ http://127.0.0.1:8000 │
     └───────────────────────┘                                 └───────────────────────┘
```

---

## 15. CORE USER JOURNEY

```text
Step 1: Install Binary
        └─ Developer installs single binary (via curl, brew, or direct download).

Step 2: Project Initialization (`porta init`)
        └─ Developer runs `porta init` in project root when setting up a workspace.
        └─ PORTA scans active listening ports (e.g., :3000, :8000) and drafts `porta.yaml`.

Step 3: Configuration & Customization
        └─ Developer adjusts routes (`/` -> 3000, `/api` -> 8000) or adds security mode.

Step 4: Launching Environment (`porta start`) [Primary One-Command Runtime Experience]
        └─ PORTA performs pre-flight validation on config and targets.
        └─ PORTA allocates ephemeral loopback port (127.0.0.1:0) for embedded reverse proxy.
        └─ PORTA verifies/downloads cached `cloudflared` binary and connects Quick Tunnel.
        └─ PORTA captures allocated public HTTPS URL.
        └─ PORTA displays real-time interactive TUI in foreground with active routes and status.

Step 5: Testing & Observation
        └─ Developer / QA / Client visits public HTTPS URL.
        └─ Terminal streams live incoming request logs (method, path, status, latency) with credentials redacted.

Step 6: Graceful Teardown (`Ctrl+C`)
        └─ Developer presses Ctrl+C.
        └─ PORTA cleanly terminates tunnel child process, closes proxy socket, flushes logs.
```

---

## 16. CORE USE CASES

- **UC-01: Consolidated Full-Stack Web App Preview:** Exposing a Vite/Next.js frontend and Node/Go backend under `https://<slug>.trycloudflare.com` with unified cookie and CORS domain.
- **UC-02: Authenticated Client Demo:** Exposing local build with `security.mode: password` so only the client with credentials can view unreleased work.
- **UC-03: Real-Time Webhook Development:** Receiving live asynchronous Stripe / Midtrans webhook callbacks on `https://<slug>.trycloudflare.com/api/webhooks/stripe`.
- **UC-04: Cross-Device Mobile Testing:** Navigating to the public URL on iOS Safari / Android Chrome over mobile cellular network to test touch UX and responsiveness.
- **UC-05: Real-Time WebSocket Streaming:** Connecting a public client to local socket.io / SSE server through transparent proxy connection hijacking.

---

## 17. MVP DEFINITION

The **PORTA MVP (v1.0.0)** is a standalone, cross-platform, foreground CLI tool written in Go that reads a declarative `porta.yaml`, binds multiple local HTTP/WebSocket services into an embedded reverse proxy gateway on an ephemeral loopback port, establishes an ephemeral public HTTPS URL via Cloudflare Quick Tunnel using on-demand binary downloading and local caching without requiring account credentials, and displays a live interactive TUI in the terminal.

---

## 18. MVP SCOPE

- **CLI Suite:** `porta init`, `porta start` (foreground execution), `porta status`, `porta logs`, `porta config`, `porta doctor`.
- **Configuration Engine:** YAML parser supporting environment variable expansion (`${VAR}`) and smart default inference.
- **Routing Engine:** Path-based *Longest Prefix Match* reverse proxy with path stripping (`strip_path: true`).
- **Gateway Binding:** Ephemeral loopback port allocation (`127.0.0.1:0`) dynamically selected by the OS.
- **Protocol Support:** HTTP/1.1, HTTP/2, WebSocket connection hijacking, and unbuffered SSE/Chunked streaming.
- **Tunneling:** Cloudflare Quick Tunnel provider driver with on-demand binary download, local caching (`~/.porta/bin/`), process supervision, and URL extraction.
- **Health Probing:** Pre-flight and periodic background TCP/HTTP health checks with graceful 502 degradation.
- **Security Gatekeeper:** HTTP Basic Authentication (`401 Challenge`), Bearer Token verification (`403 Forbidden`), and Strict Localhost Loopback SSRF protection.
- **Log Redaction:** Mandatory automatic redaction of sensitive query parameters (`?porta_token=`) and `Authorization` headers across all logs.
- **Terminal UI (TUI):** Lightweight ANSI live status display with real-time request access log streaming.
- **Cross-Platform:** Native compilation and verification for Windows (x64/ARM64), macOS (Intel/Apple Silicon), and Linux (x64/ARM64).

---

## 19. MVP NON-SCOPE (DEFERRED TO POST-MVP)

- **Detached Background Daemon Mode (`--detach`, `porta stop`, background daemon supervisor):** Deferred to v1.1. MVP is strictly a foreground CLI process.
- **Custom / Fixed Gateway Port Configuration (`proxy.port`):** Deferred to v1.1. MVP strictly uses ephemeral port `127.0.0.1:0`.
- **Full Credential Store (`CRED-001` for external provider tokens):** Deferred to v1.1 since Cloudflare Quick Tunnel requires zero user credentials.
- **Named Persistent Custom Domains (Cloudflare DNS binding):** Deferred to v1.1.
- **Multi-Provider Drivers (ngrok, Tailscale Funnel):** Deferred to v1.1 via existing `TunnelProvider` abstraction.
- **Process Orchestration (`run: npm run dev`):** Deferred to v1.2.
- **Web-Based Inspection Dashboard:** Deferred to v2.0.
- **Team Collaboration Workspaces & Cloud Control Plane:** Deferred to v2.0.
- **Raw TCP / UDP Tunneling:** Deferred to v2.0.

---

## 20. FEATURE MAP

```text
PORTA v1.0.0 Feature Map (MVP)
│
├── [CLI & UX Layer]
│   ├── Interactive Init & Port Discovery (CLI-001) [SOURCE-BACKED]
│   ├── Foreground Execution & Signal Trapping (CLI-001) [APPROVED PRODUCT DECISION]
│   ├── Terminal User Interface (TUI) (CLI-002) [SOURCE-BACKED]
│   ├── Diagnostic Subsystem (`porta doctor`) (CLI-001) [SOURCE-BACKED]
│   └── Log Streaming & Inspection (`porta logs`) (MON-002) [SOURCE-BACKED]
│
├── [Core Gateway & Routing Layer]
│   ├── Declarative YAML Parser & Validator (CFG-001, CFG-002) [SOURCE-BACKED]
│   ├── Multi-Service Registry (SRV-001) [SOURCE-BACKED]
│   ├── Ephemeral Loopback Gateway Listener (PROXY-001) [APPROVED PRODUCT DECISION]
│   ├── Longest Prefix Match Proxy (PROXY-002) [SOURCE-BACKED]
│   ├── WebSocket Hijacking & SSE Streaming (PROXY-003) [SOURCE-BACKED]
│   ├── Header Sanitation & Forwarding (PROXY-004) [SOURCE-BACKED]
│   └── Path Rewrite / Stripping (SRV-002) [SOURCE-BACKED]
│
├── [Tunnel & Egress Layer]
│   ├── Unified `TunnelProvider` Abstraction (TUNNEL-001) [SOURCE-BACKED]
│   ├── Cloudflare Quick Tunnel Driver (TUNNEL-002) [SOURCE-BACKED]
│   ├── On-Demand Binary Downloader & Local Cache (TUNNEL-002) [APPROVED PRODUCT DECISION]
│   ├── Watchdog & Reconnection Supervisor (TUNNEL-003) [SOURCE-BACKED]
│   └── Public URL Extractor (TUNNEL-002) [SOURCE-BACKED]
│
├── [Health & Security Layer]
│   ├── TCP & HTTP Target Prober (HC-001) [SOURCE-BACKED]
│   ├── Graceful 502 Degradation Handler (HC-002) [SOURCE-BACKED]
│   ├── Basic Auth Gatekeeper (SEC-001) [SOURCE-BACKED]
│   ├── Bearer Token Gatekeeper (SEC-002) [SOURCE-BACKED]
│   ├── Credential & Token Log Redaction (SEC-002) [APPROVED PRODUCT DECISION]
│   └── Strict Loopback SSRF Isolation (SEC-003) [APPROVED PRODUCT DECISION]
│
└── [Storage & State]
    └── File-Based Local State & Log Rotation (`.porta/`) (MON-001) [SOURCE-BACKED]
```

---

## 21. DETAILED PRODUCT REQUIREMENTS

```text
Requirement Taxonomy: PRD-F-xxx
Classification Labels: [SOURCE-BACKED], [APPROVED PRODUCT DECISION], [INFERENCE], [PROPOSED], [DECISION REQUIRED]
```

### PRD-F-001: Automatic Workspace Initialization & Port Scanning
- **Classification:** `[SOURCE-BACKED]`
- **Feature:** `porta init`
- **Purpose:** Eliminate manual configuration boilerplate by inspecting the developer's local environment.
- **User Value:** A developer can configure a multi-service app in seconds without consulting documentation.
- **Priority:** P0 (MVP)
- **Actor:** Developer
- **Preconditions:** Developer runs the command inside their project repository directory.
- **Inputs:** Current working directory, local network listening sockets.
- **Expected Behavior:** Scans standard dev ports (`3000`, `5173`, `8000`, `8080`, `9000`), detects active listening processes, infers frontend vs backend, and writes a valid `porta.yaml`. If `porta.yaml` already exists, prompts before overwriting unless `--force` is provided.
- **Outputs:** Starter `./porta.yaml` file.
- **Success State:** `porta.yaml` created with valid syntax matching active local ports.
- **Failure State:** Write permission error or corrupted existing file.
- **Edge Cases:** No ports are currently listening $\rightarrow$ generates a clean starter template with helpful comments.
- **Dependencies:** None.
- **Acceptance Criteria:**  
  `Given` no `porta.yaml` and port 3000 active,  
  `When` developer runs `porta init`,  
  `Then` `porta.yaml` is generated with port 3000 mapped to `/`.
- **FRM Traceability:** PM-001  
- **SRS Traceability:** Section 12.2.A (`init.go`)

---

### PRD-F-002: Declarative Configuration Engine with Variable Expansion
- **Classification:** `[SOURCE-BACKED]`
- **Feature:** `porta.yaml` Configuration Parser
- **Purpose:** Parse, validate, and populate default values for runtime execution.
- **User Value:** Human-readable configuration that supports storing secrets via environment variables.
- **Priority:** P0 (MVP)
- **Actor:** PORTA Runtime
- **Preconditions:** `porta.yaml` exists in working directory or is passed via `-c <path>`.
- **Inputs:** File content of `porta.yaml`, OS environment variables.
- **Expected Behavior:** Parses YAML, expands `${ENV_VAR:-default}`, applies smart defaults (`host=127.0.0.1`, `route=/`, `security.mode=public`), and performs strict schema validation.
- **Outputs:** In-memory configuration model.
- **Success State:** Valid config object ready for gateway bootstrapping.
- **Failure State:** Returns specific YAML syntax error with line number and actionable hint.
- **Edge Cases:** Nested environment variable syntax, missing port definitions.
- **Dependencies:** None.
- **Acceptance Criteria:**  
  `Given` a config with `${PORTA_SECRET:-mysecret}`,  
  `When` config is parsed without `PORTA_SECRET` set,  
  `Then` the secret evaluates to `mysecret`.
- **FRM Traceability:** CFG-001  
- **SRS Traceability:** Section 3.3 (`config/validator.go`)

---

### PRD-F-003: Multi-Service Registry & Route Collision Prevention
- **Classification:** `[SOURCE-BACKED]`
- **Feature:** Service Registry Engine
- **Purpose:** Maintain an in-memory routing table mapping URL paths to local service targets.
- **User Value:** Guarantees deterministic routing and prevents ambiguous route conflicts.
- **Priority:** P0 (MVP)
- **Actor:** PORTA Core
- **Preconditions:** Parsed configuration from PRD-F-002.
- **Inputs:** Map of declared services.
- **Expected Behavior:** Validates that all ports are within `1–65535`, all routes start with `/`, and no two services claim the exact same prefix route.
- **Outputs:** Active Service Registry lookup table.
- **Success State:** Registry ready for reverse proxy dispatching.
- **Failure State:** Exits with `ERR_ROUTE_COLLISION` indicating conflicting services.
- **Edge Cases:** Multiple nested routes (e.g., `/api` and `/api/v1`) $\rightarrow$ allowed and resolved via longest prefix matching.
- **Dependencies:** PRD-F-002.
- **Acceptance Criteria:**  
  `Given` Service A on `/api` and Service B on `/api`,  
  `When` PORTA validates config,  
  `Then` it halts startup with exit code 1 and logs the duplicate route conflict.
- **FRM Traceability:** CFG-002, SRV-001  
- **SRS Traceability:** Section 3.4 (`registry/registry.go`)

---

### PRD-F-004: Embedded Reverse Proxy Gateway on Ephemeral Port
- **Classification:** `[SOURCE-BACKED]` + `[APPROVED PRODUCT DECISION]`
- **Feature:** Local Ingress Gateway
- **Purpose:** Route incoming HTTP requests to the appropriate local upstream service via an ephemeral loopback port.
- **User Value:** Multiple services appear under a single origin without CORS issues or local port conflicts.
- **Priority:** P0 (MVP)
- **Actor:** Inbound HTTP Client
- **Preconditions:** Service Registry is populated; gateway socket bound to ephemeral loopback (`127.0.0.1:0`).
- **Inputs:** Incoming HTTP request from tunnel adapter.
- **Expected Behavior:** Binds to dynamic OS-assigned port `127.0.0.1:0`. Matches request path against routing table using Longest Prefix Match. Forwards request to target `127.0.0.1:<target_port>`. If `strip_path: true`, strips the matching prefix before forwarding.
- **Outputs:** Response forwarded from upstream back to client.
- **Success State:** 2xx/3xx/4xx/5xx response accurately relayed.
- **Failure State:** No route matched $\rightarrow$ returns 404 with custom PORTA error payload.
- **Edge Cases:** Trailing slashes in request URIs (`/api` vs `/api/`).
- **Dependencies:** PRD-F-003.
- **Acceptance Criteria:**  
  `Given` routes `/` (:3000) and `/api` (:8000),  
  `When` request is `GET /api/users`,  
  `Then` request is proxied to `127.0.0.1:8000/api/users` via an ephemeral gateway port.
- **FRM Traceability:** PROXY-001, PROXY-002, SRV-002  
- **SRS Traceability:** Section 6 (`proxy/gateway.go`)

---

### PRD-F-005: Transparent WebSocket Hijacking & Real-Time Streaming
- **Classification:** `[SOURCE-BACKED]`
- **Feature:** Full-Duplex Connection Forwarder
- **Purpose:** Support real-time features like Socket.io, GraphQL subscriptions, and LLM SSE streams.
- **User Value:** Developers building modern AI streaming or real-time apps experience zero protocol breakage.
- **Priority:** P0 (MVP)
- **Actor:** WebSocket Client / Streaming Consumer
- **Preconditions:** Reverse Proxy Gateway is active.
- **Inputs:** HTTP Request with `Upgrade: websocket` or `Accept: text/event-stream`.
- **Expected Behavior:** Performs HTTP connection hijacking for WebSockets, establishing a bidirectional raw TCP stream. Disables response buffering (`FlushInterval: -1`) for SSE streams.
- **Outputs:** Continuous real-time data flow.
- **Success State:** WebSocket connection stays open with bidirectional messaging.
- **Failure State:** Upstream service rejects upgrade handshake.
- **Edge Cases:** Network interruption during long-running stream.
- **Dependencies:** PRD-F-004.
- **Acceptance Criteria:**  
  `Given` a WebSocket server at `:9001`,  
  `When` client connects with `Upgrade: websocket`,  
  `Then` gateway upgrades connection and forwards messages without frame corruption.
- **FRM Traceability:** PROXY-003  
- **SRS Traceability:** Section 6 (`proxy/websocket.go`)

---

### PRD-F-006: Header Sanitization & Standard Proxy Forwarding
- **Classification:** `[SOURCE-BACKED]`
- **Feature:** Reverse Proxy Header Injector
- **Purpose:** Inform upstream local applications of the original public client's IP, scheme, and host.
- **User Value:** Web frameworks (Express, Django, Rails) generate correct absolute URLs and redirect targets.
- **Priority:** P0 (MVP)
- **Actor:** Reverse Proxy Engine
- **Preconditions:** Inbound request arrives at gateway.
- **Inputs:** Inbound HTTP Headers.
- **Expected Behavior:** Injects `X-Forwarded-For`, `X-Forwarded-Proto: https`, `X-Forwarded-Host: <public_domain>`, and `X-Real-IP`. Rewrites internal `Host` header to upstream target host.
- **Outputs:** Modified HTTP request dispatched to local service.
- **Success State:** Upstream framework correctly identifies `req.secure == true`.
- **Failure State:** Header buffer overflow.
- **Edge Cases:** Client supplies spoofed `X-Forwarded-For` $\rightarrow$ gateway appends client IP rather than overwriting blindly.
- **Dependencies:** PRD-F-004.
- **Acceptance Criteria:**  
  `Given` request from public HTTPS domain,  
  `When` forwarded to localhost backend,  
  `Then` header `X-Forwarded-Proto` contains `https`.
- **FRM Traceability:** PROXY-004  
- **SRS Traceability:** Section 6

---

### PRD-F-007: Cloudflare Quick Tunnel Provider with On-Demand Download & Local Cache
- **Classification:** `[SOURCE-BACKED]` + `[APPROVED PRODUCT DECISION]`
- **Feature:** Zero-Auth Tunnel Provider Integration & Driver Manager
- **Purpose:** Establish encrypted public ingress without requiring account registration, using on-demand binary download and local caching.
- **User Value:** Instant public exposure on first run with small base binary footprint.
- **Priority:** P0 (MVP)
- **Actor:** Tunnel Manager
- **Preconditions:** Local gateway reverse proxy listening on ephemeral loopback port.
- **Inputs:** Gateway local address (`127.0.0.1:<port>`).
- **Expected Behavior:** Checks for compatible `cloudflared` binary in local cache (`~/.porta/bin/`). If missing, downloads compatible binary, verifies integrity, and caches locally. Launches `cloudflared`, extracts public HTTPS URL from stderr, monitors process liveness, and auto-reconnects with exponential backoff on disconnects.
- **Outputs:** Public HTTPS URL (`https://<hash>.trycloudflare.com`).
- **Success State:** Active, stable encrypted tunnel stream.
- **Failure State:** Network failure during download or execution $\rightarrow$ exits with clear remediation error.
- **Edge Cases:** Transient network drops $\rightarrow$ auto-reconnects up to 5 attempts without terminating reverse proxy.
- **Dependencies:** PRD-F-004.
- **Acceptance Criteria:**  
  `Given` a healthy local gateway,  
  `When` `porta start` is executed,  
  `Then` `cloudflared` is verified/downloaded to local cache, started, and a valid `https://*.trycloudflare.com` URL is presented in the terminal.
- **FRM Traceability:** TUNNEL-001, TUNNEL-002, TUNNEL-003  
- **SRS Traceability:** Section 5 (`tunnel/cloudflare.go`)

---

### PRD-F-008: Target Health Checking & Graceful 502 Degradation
- **Classification:** `[SOURCE-BACKED]`
- **Feature:** Service Health Monitor
- **Purpose:** Continuously verify availability of upstream local services.
- **User Value:** Gateway stays alive even if one microservice crashes or is restarted by the developer.
- **Priority:** P0 (MVP)
- **Actor:** Health Checker Engine
- **Preconditions:** Services registered in Service Registry.
- **Inputs:** Target endpoints, optional `health_check.path`.
- **Expected Behavior:** Probes target via TCP socket ping or HTTP GET probe every 5 seconds. Maintains state (`ONLINE`, `UNHEALTHY`, `OFFLINE`). If an offline service receives traffic, returns a friendly `502 Bad Gateway (PORTA: Upstream Service Offline)` response.
- **Outputs:** Live service status flags.
- **Success State:** Service marked `ONLINE` when responding, `OFFLINE` when closed.
- **Failure State:** Probe timeouts properly recorded without hanging runtime.
- **Edge Cases:** Service starts up slowly during initial launch $\rightarrow$ 5s grace period before marking `OFFLINE`.
- **Dependencies:** PRD-F-003.
- **Acceptance Criteria:**  
  `Given` backend service at :8000 is stopped,  
  `When` public client visits `/api`,  
  `Then` gateway returns 502 indicating backend is offline while keeping frontend at `/` available.
- **FRM Traceability:** HC-001, HC-002  
- **SRS Traceability:** Section 7 (`health/checker.go`)

---

### PRD-F-009: Security Gatekeeper with Credential & Token Log Redaction
- **Classification:** `[SOURCE-BACKED]` + `[APPROVED PRODUCT DECISION]`
- **Feature:** Ingress Access Control & Safe Logging
- **Purpose:** Protect exposed local applications from unauthorized access while guaranteeing zero credential leakage into log files or console streams.
- **User Value:** Confidently share preview URLs knowing unauthorized traffic is challenged and tokens are never logged in plaintext.
- **Priority:** P0 (MVP)
- **Actor:** Public Visitor / API Client / Logging Subsystem
- **Preconditions:** `security.mode` set to `password` or `token` in `porta.yaml`.
- **Inputs:** HTTP `Authorization` header, query parameter `?porta_token=`.
- **Expected Behavior:** Challenges unauthenticated requests (`401` for password, `403` for token). When logging incoming requests, the logging engine automatically masks/redacts token query parameters (`?porta_token=[REDACTED]`) and `Authorization` headers across all sinks (TUI, stdout, stderr, `access.log`).
- **Outputs:** Allowed request forwarding or blocked HTTP challenge; sanitized access logs.
- **Success State:** Valid credentials allow access; log files contain zero plaintext secrets.
- **Failure State:** Invalid credentials rejected before hitting local app.
- **Edge Cases:** Timing attack protection via constant-time string comparison (`crypto/subtle`).
- **Dependencies:** PRD-F-004.
- **Acceptance Criteria:**  
  `Given` `security.mode: token` and a request with `?porta_token=secret123`,  
  `When` the request is processed and logged,  
  `Then` the gateway permits access and the access log records `?porta_token=[REDACTED]`.
- **FRM Traceability:** SEC-001, SEC-002  
- **SRS Traceability:** Section 8 (`security/gatekeeper.go`)

---

### PRD-F-010: Strict Loopback SSRF Guard & Private Network Rejection
- **Classification:** `[SOURCE-BACKED]` + `[APPROVED PRODUCT DECISION]`
- **Feature:** Network Security Isolation Guard
- **Purpose:** Prevent attackers or misconfigurations from using PORTA to reach arbitrary private, internal corporate, or cloud metadata network resources.
- **User Value:** Eliminates the risk of accidental internal data breaches or Server-Side Request Forgery via the public tunnel.
- **Priority:** P0 (MVP)
- **Actor:** Security Engine / Config Validator
- **Preconditions:** Service target parsing and upstream connection establishment.
- **Inputs:** Configured service host targets, resolved IP addresses.
- **Expected Behavior:** Strictly enforces that all target destinations resolve exclusively to loopback addresses (`127.0.0.1`, `localhost`, `::1`). Rejects any configuration or DNS resolution resolving to private RFC 1918 subnets (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), link-local/cloud metadata (`169.254.169.254`), or non-loopback IPv6.
- **Outputs:** Validation approval or fatal security rejection.
- **Success State:** Only local machine loopback services are reachable.
- **Failure State:** Startup halted with `ERR_SECURITY_POLICY_VIOLATION`.
- **Edge Cases:** Hostnames that dynamically resolve or rebind to internal LAN IP addresses are blocked.
- **Dependencies:** PRD-F-002.
- **Acceptance Criteria:**  
  `Given` a service configured to target `192.168.1.50` or resolving to a non-loopback address,  
  `When` PORTA validates or establishes the upstream connection,  
  `Then` PORTA rejects the target and does not expose that internal resource publicly.
- **FRM Traceability:** SEC-003  
- **SRS Traceability:** Section 8

---

### PRD-F-011: Foreground Terminal UI (TUI) & Live Request Logging
- **Classification:** `[SOURCE-BACKED]` + `[APPROVED PRODUCT DECISION]`
- **Feature:** Interactive Foreground Terminal Status & Live Log Feed
- **Purpose:** Provide immediate operational visibility into tunnel state, service health, and live traffic during foreground execution.
- **User Value:** Developer sees exactly what is happening in real time without needing a separate browser dashboard.
- **Priority:** P0 (MVP)
- **Actor:** Developer
- **Preconditions:** `porta start` running interactively in a TTY terminal.
- **Inputs:** Runtime state events, sanitized request access log events.
- **Expected Behavior:** Displays clean ANSI table summarizing project name, public HTTPS URL, active routes, service statuses, and a scrolling live log feed of incoming requests (status code, method, path, target, duration).
- **Outputs:** Formatted terminal screen.
- **Success State:** Dynamic, non-flickering terminal status display.
- **Failure State:** Terminal does not support ANSI $\rightarrow$ falls back to linear standard text logging.
- **Edge Cases:** Terminal window resize $\rightarrow$ re-renders table cleanly.
- **Dependencies:** PRD-F-004, PRD-F-007.
- **Acceptance Criteria:**  
  `Given` an active foreground PORTA session,  
  `When` an external request hits the public URL,  
  `Then` the sanitized request log line appears in the TUI within 100ms.
- **FRM Traceability:** CLI-002, MON-001  
- **SRS Traceability:** Section 12.2.B (`ui/tui.go`)

---

### PRD-F-012: System Health Diagnostics (`porta doctor`)
- **Classification:** `[SOURCE-BACKED]`
- **Feature:** Diagnostic Command
- **Purpose:** Verify system dependencies, networking, loopback adapters, and config validity before running.
- **User Value:** Immediate troubleshooting assistance when developer machine has network or firewall issues.
- **Priority:** P0 (MVP)
- **Actor:** Developer
- **Preconditions:** Binary installed.
- **Inputs:** OS environment, network adapters, DNS resolver, local config file.
- **Expected Behavior:** Runs diagnostic checks: OS architecture compatibility, 127.0.0.1 loopback bindability, tunnel driver/cache availability, outbound internet connectivity, and `porta.yaml` schema validity.
- **Outputs:** Formatted checklist with green checkmarks or actionable red error guidance.
- **Success State:** All checks pass with exit code 0.
- **Failure State:** Outputs specific diagnostic failure with suggested fix.
- **Edge Cases:** Offline development machine $\rightarrow$ flags outbound internet failure.
- **Dependencies:** None.
- **Acceptance Criteria:**  
  `Given` an invalid YAML config,  
  `When` developer runs `porta doctor`,  
  `Then` it flags the configuration syntax error specifically.
- **FRM Traceability:** CLI-001  
- **SRS Traceability:** Section 12.2.C (`cmd/doctor.go`)

---

## 22. CLI UX SPECIFICATION

### 22.1 Command Grammar & Hierarchy
```bash
porta <command> [flags] [arguments]
```

### 22.2 Command Reference Table (MVP)

| Command | Purpose | Key Flags | Exit Codes | MVP Status |
| :--- | :--- | :--- | :--- | :--- |
| `porta init` | Scans workspace and generates starter `porta.yaml` | `-f, --force` | 0: Success, 1: File write error | **P0 (MVP)** |
| `porta start` | Starts reverse proxy, health checker, and public tunnel (Foreground) | `-c, --config <file>` | 0: Clean exit, 1: Config/Validation error, 2: Tunnel/Runtime fatal | **P0 (MVP)** |
| `porta status` | Shows status of current active configuration and targets | `--json` | 0: Active, 1: Inactive | **P0 (MVP)** |
| `porta logs` | Tails structured access logs from `.porta/logs/` | `-f, --follow`<br>`--level <info\|warn\|error>`<br>`--service <name>` | 0: Success, 1: Log file unreadable | **P0 (MVP)** |
| `porta config` | Validates and dumps parsed configuration model | `-c, --config <file>` | 0: Valid, 1: Syntax/Validation failure | **P0 (MVP)** |
| `porta doctor` | Runs system and network diagnostic checks | None | 0: All pass, 1: Diagnostic failure | **P0 (MVP)** |

> [!NOTE]
> `porta stop` and `porta start --detach` (background daemon mode) are reclassified as Post-MVP (v1.1) features. In MVP, `porta start` runs in the foreground and terminates cleanly via `Ctrl+C`.

### 22.3 Signal Handling & Shutdown UX
When the developer presses `Ctrl+C` (`SIGINT`) or sends `SIGTERM`:
```text
^C
[i] Gracefully shutting down PORTA...
[✓] Public tunnel disconnected.
[✓] Reverse proxy gateway stopped.
[✓] Port bindings released.
[✓] Goodbye!
```
Shutdown duration target is under **500 milliseconds**.

---

## 23. TUI UX SPECIFICATION

### 23.1 Information Architecture & Visual Layout
The interactive TUI answers five fundamental questions in one screen:
1. Is PORTA running?
2. Which services are ONLINE vs OFFLINE?
3. Is the reverse proxy active?
4. Is the tunnel connected?
5. What is the public URL?

```text
  ┌─────────────────────────────────────────────────────────────┐
  │   PORTA v1.0.1 — Local Exposure Engine                     │
  ├─────────────────────────────────────────────────────────────┤
  │   Project     : qulineria (development)                    │
  │   Public URL  : https://qulineria-dev.trycloudflare.com     │
  │   Security    : Basic Auth Protected (user: porta)          │
  ├─────────────────────────────────────────────────────────────┤
  │   SERVICES & ROUTES:                                        │
  │   • frontend  :3000   ->  /           [ONLINE]              │
  │   • backend   :8000   ->  /api        [ONLINE]              │
  │   • socket    :9001   ->  /socket.io  [ONLINE]              │
  ├─────────────────────────────────────────────────────────────┤
  │   LIVE LOGS (Ctrl+C to stop):                               │
  │   11:15:02 [200] GET  /                  -> :3000 (12ms)    │
  │   11:15:05 [200] GET  /api/v1/products   -> :8000 (34ms)    │
  │   11:15:10 [101] GET  /socket.io/?EIO=4  -> :9001 (Upgrade) │
  └─────────────────────────────────────────────────────────────┘
```

### 23.2 Terminal Compatibility & Fallback
- **Standard TTY:** Full ANSI box-drawing and colorized status indicators.
- **Non-TTY / CI / Redirected Output:** Automatically switches to standard linear stdout logging without cursor jumping codes.

---

## 24. CONFIGURATION UX (`porta.yaml`)

### 24.1 Official MVP Configuration Schema
```yaml
# porta.yaml - Configuration Specification
version: "1"

project:
  name: store-platform # Required: Project identifier

services:
  web:
    port: 3000         # Required: Target localhost port
    route: /           # Optional: Ingress route prefix (default: /)
    host: 127.0.0.1    # Optional: Target host (default: 127.0.0.1)

  api:
    port: 8000
    route: /api
    strip_path: true   # Optional: Strip '/api' prefix when forwarding (default: false)
    health_check:
      path: /healthz   # Optional: HTTP GET health check path (default: TCP ping)
      interval: 5s     # Optional: Check interval (default: 5s)

tunnel:
  provider: cloudflare # Optional: Tunnel provider (default: cloudflare)

security:
  mode: public         # Optional: public | password | token (default: public)
  password: ${PORTA_PASSWORD:-admin:SecretPass123!} # Supported if mode=password
```

### 24.2 Configuration Validation UX
If validation fails, PORTA outputs pinpoint error diagnostics:
```text
✗ Configuration Error in ./porta.yaml:
  Line 8: services.api.port: value '99999' is invalid. Port must be between 1 and 65535.
```

---

## 25. SERVICE MODEL

- **Service Identity:** Each service has a unique alphanumeric ID (e.g., `web`, `api`, `auth`).
- **Target Resolution:** Defaults to `127.0.0.1:<port>`. Host targets resolving outside loopback are rejected by the SSRF Guard.
- **Service Independence:** Services operate independently. If `api` is offline, `web` continues to serve requests.
- **No Process Management in MVP:** PORTA does not spawn or manage developer application processes. Services must be started by the developer.

---

## 26. ROUTING MODEL

- **Matching Algorithm:** *Longest Prefix Match*.
  - Request `GET /api/v1/users` matches `/api` over `/`.
  - Request `GET /assets/main.js` matches `/`.
- **Prefix Stripping (`strip_path`):**
  - If `strip_path: false`: `GET /api/v1/users` $\rightarrow$ upstream receives `GET /api/v1/users`.
  - If `strip_path: true`: `GET /api/v1/users` $\rightarrow$ upstream receives `GET /v1/users`.
- **Unmatched Routes:** If no route matches, gateway returns `404 Not Found (PORTA: No matching route)`.

---

## 27. REVERSE PROXY EXPERIENCE

- **Local Ephemeral Port Allocation:** Gateway binds to a dynamic loopback port (`127.0.0.1:0`), completely preventing local port conflicts with developer apps.
- **Connection Pooling:** High-throughput `http.Transport` connection reuse minimizing local latency.
- **Streaming & SSE:** Unbuffered chunked transfer encoding ensuring real-time token streaming for LLM apps.
- **Error Pages:** Custom responsive HTML/JSON 502 error templates explaining when a local service is offline.

---

## 28. TUNNEL EXPERIENCE

- **Zero-Friction Ingress:** Cloudflare Quick Tunnel provisions an Anycast public domain without credentials.
- **On-Demand Driver Caching:** If `cloudflared` is not found in `~/.porta/bin/`, PORTA downloads the platform-specific binary, verifies integrity, and caches it locally.
- **Watchdog & Self-Healing:** If network connection drops, tunnel manager attempts exponential backoff reconnection up to 5 times (1s, 2s, 4s, 8s, 10s) before alerting the user.
- **Provider Abstraction:** The core gateway communicates with tunnels via the `TunnelProvider` interface, allowing future providers (ngrok, Tailscale) without architectural changes.

---

## 29. HEALTH & MONITORING EXPERIENCE

### 29.1 Status Transitions
```text
UNKNOWN ──► STARTING (Grace Period 5s) ──► ONLINE ──► UNHEALTHY ──► OFFLINE
```

### 29.2 Status Display in TUI
- `[ONLINE]` (Green): Responding successfully to TCP/HTTP probes.
- `[STARTING]` (Yellow): In grace period, awaiting initial socket response.
- `[OFFLINE]` (Red): Port unreachable; gateway serves 502 for this route.

---

## 30. SECURITY REQUIREMENTS

1. **Strict Loopback SSRF Guard:** Rejects any upstream target resolving outside `127.0.0.1`, `localhost`, or `::1`, protecting internal LAN resources and cloud metadata endpoints.
2. **Access Protection Modes:**
   - `public`: Unrestricted internet access (ideal for webhooks and public demos).
   - `password`: HTTP Basic Auth challenge on public gateway.
   - `token`: Bearer token in header or `?porta_token=` query param.
3. **Mandatory Log Redaction:** All authentication tokens (`?porta_token=[REDACTED]`) and `Authorization` headers MUST be masked before emitting to TUI, stdout, stderr, or `access.log`.
4. **Secret Protection:** Secrets in `porta.yaml` can reference environment variables (`${VAR}`). Local runtime state files are saved with `0600` file permissions.
5. **Header Defense:** Cleans and normalizes `Host` and `X-Forwarded-*` headers to prevent host header injection attacks.

---

## 31. RUNTIME LIFECYCLE & STATE MODEL

| State | Entry Condition | Behavior | Exit Condition | Failure State |
| :--- | :--- | :--- | :--- | :--- |
| **INIT** | `porta start` executed | Initializes workspace context | Config loaded | `FAILED (Exit 1)` |
| **VALIDATING** | Config loaded | Validates YAML, schema, SSRF rules | Validation passed | `FAILED (Exit 1)` |
| **STARTING** | Validation passed | Probes target ports; allocates ephemeral gateway port | Proxy socket listening | `FAILED (Exit 1)` |
| **PROXY_READY**| Gateway listening | Verifies/downloads cached `cloudflared` | Process running | `FAILED (Exit 2)` |
| **TUNNEL_CONNECTING**| Tunnel spawned | Parses stderr for public URL | URL extracted | `FAILED (Exit 2)` |
| **ONLINE** | URL active | Foreground proxying, live TUI & log streaming | `SIGINT` (`Ctrl+C`) | `RECOVERING` |
| **RECOVERING** | Network drop | Exponential backoff tunnel reconnection | Reconnected $\rightarrow$ `ONLINE` | `FAILED (Max retries)` |
| **STOPPING** | Termination signal | Kills tunnel child process, closes gateway socket | Resources freed | `STOPPED` |
| **STOPPED** | Cleanup complete | Flushes logs, exits process (0) | Process exited | None |

---

## 32. ERROR & FAILURE EXPERIENCE (FAILURE MATRIX)

| Failure Scenario | Detection Mechanism | User Experience (TUI / Terminal) | System Behavior | Recovery Action |
| :--- | :--- | :--- | :--- | :--- |
| **Malformed `porta.yaml`** | YAML Unmarshaler | `✗ Error: Line 5: invalid syntax` | Halts execution immediately | Developer fixes syntax in YAML. |
| **Route Collision** | Service Registry | `✗ Route conflict: Both 'a' and 'b' map to '/api'` | Halts execution immediately | Developer changes route prefix. |
| **Target Service Offline** | Health Check Prober | TUI shows `• backend :8000 [OFFLINE]` in red | Gateway serves 502 for `/api`; keeps tunnel alive | Developer starts backend; status turns `ONLINE` automatically. |
| **Tunnel Binary Missing** | Driver Initializer | `i Downloading tunnel driver to ~/.porta/bin/...` | Auto-downloads compatible binary to cache | Automatic recovery. |
| **Tunnel Connection Drop** | Watchdog Monitor | `! Tunnel disconnected. Reconnecting (Attempt 2/5)...` | Exponential backoff reconnection | Reconnects automatically when internet restores. |
| **SSRF Target Attempt** | Security Engine | `✗ Security Violation: Target IP '192.168.1.1' not allowed` | Halts execution immediately | Developer changes host to `127.0.0.1`. |
| **Unauthorized Access** | Security Gatekeeper | Browser displays HTTP Basic Auth login prompt | Returns 401 Unauthorized | Visitor enters valid credentials. |

---

## 33. RECOVERY BEHAVIOR

- **Ephemeral Auto-Reconnection:** Tunnel disconnections trigger retries at intervals of 1s, 2s, 4s, 8s, 10s.
- **Non-Destructive Target Reconnect:** When a crashed local backend restarts, the Health Checker detects active TCP listening within 5s and seamlessly resumes traffic forwarding without tunnel restarts.

---

## 34. LOCAL STATE & STORAGE

- **State Directory:** Stored at `.porta/` inside project directory and `~/.porta/bin/` for cached tunnel driver binaries.
- **Log Storage:** `.porta/logs/access.log` (structured JSON, rotating max 10MB, max 3 backups).
- **Permissions:** All created directories and files enforce POSIX `0700` / `0600` security permissions.

---

## 35. LOGGING & OBSERVABILITY

- **Console Log Stream:** Colorized live request stream in foreground TUI.
- **Mandatory Redaction:** All logging sinks sanitize sensitive query parameters and `Authorization` headers.
- **File Access Log:** Structured JSON format:
  ```json
  {"timestamp":"2026-09-02T11:30:00Z","level":"info","method":"GET","path":"/api/v1/users?porta_token=[REDACTED]","status":200,"duration_ms":14,"upstream":"127.0.0.1:8000","client_ip":"198.51.100.42"}
  ```
- **CLI Log Inspection:** `porta logs -f --service=backend --level=error`.

---

## 36. CROSS-PLATFORM EXPERIENCE

| Platform | Binary | Process Supervision | Path Separators | Loopback Behavior |
| :--- | :--- | :--- | :--- | :--- |
| **Windows 10/11** | `porta.exe` | Windows Job Objects (guarantees child process termination on exit) | Backslash normalized via `filepath.ToSlash` | Explicit `127.0.0.1` binding to avoid IPv6 `::1` delay |
| **macOS (Intel/M1/M2/M3)** | `porta` | POSIX Process Groups (`Setpgid`) & `SIGINT` trapping | POSIX `/` | Standard `127.0.0.1` / `::1` |
| **Linux (Ubuntu/Arch/etc)** | `porta` | POSIX Process Groups & `SIGTERM` trapping | POSIX `/` | Standard `127.0.0.1` / `::1` |

---

## 37. PERFORMANCE EXPECTATIONS

| Dimension | Target Baseline | Classification | Status |
| :--- | :--- | :--- | :--- |
| **Cold Startup Time (to Public URL live)** | `< 1.5 seconds` (on broadband) | Engineering Benchmark Target | To Be Validated in Implementation Phase |
| **Memory Footprint (Idle)** | `< 25 MB RAM` | Engineering Benchmark Target | To Be Validated in Implementation Phase |
| **Memory Footprint (Under Load - 200 req/s)** | `< 50 MB RAM` | Engineering Benchmark Target | To Be Validated in Implementation Phase |
| **CPU Usage (Idle)** | `< 0.5% CPU` | Engineering Benchmark Target | To Be Validated in Implementation Phase |
| **Gateway Routing Overhead Latency** | `< 3 ms` | Engineering Benchmark Target | To Be Validated in Implementation Phase |
| **Concurrent Active Streams** | `1,000+ concurrent requests` | Engineering Benchmark Target | To Be Validated in Implementation Phase |

---

## 38. RELIABILITY EXPECTATIONS

- **Zero-Orphan Guarantee:** Exiting PORTA (`Ctrl+C` or crash) will never leave orphaned background `cloudflared` processes running on developer machines.
- **Fault-Tolerant Upstream:** Single-service failures will never crash the gateway or terminate the public tunnel.
- **Self-Healing Connectivity:** Tunnel handles transient WiFi drops automatically.

---

## 39. DEPENDENCIES

- **Build-Time / Core Runtime:** Go 1.22+ Standard Library (`net/http`, `net/http/httputil`, `context`, `sync`, `os/signal`).
- **CLI Framework:** `github.com/spf13/cobra` (standard CLI commander).
- **YAML Parser:** `gopkg.in/yaml.v3` (strict YAML unmarshaling).
- **Tunnel Driver:** Cloudflare Tunnel (`cloudflared` standalone binary cached in `~/.porta/bin/`).

---

## 40. PRODUCT CONSTRAINTS

- **No Root / Admin Requirement:** PORTA must run completely in user-space without requiring `sudo` or Administrator privileges.
- **No Inbound Open Ports:** PORTA must not require opening firewall ports or configuring router NAT port-forwarding.
- **No Cloud Backend Requirement:** PORTA must operate without any proprietary cloud control plane servers.

---

## 41. USER STORIES & ACCEPTANCE CRITERIA

### US-001: First-Time Setup & Zero-Config Init
**As a** Full-Stack Developer,  
**I want to** run `porta init` in my repository,  
**so that** PORTA automatically creates a working `porta.yaml` based on my active ports.

```gherkin
Scenario: Successful automatic port discovery
  Given my local Vite app is running on port 3000 and FastAPI is running on port 8000
  When I execute "porta init"
  Then PORTA detects ports 3000 and 8000
  And generates a valid "porta.yaml" with routes "/" and "/api"
  And prints a success confirmation with next steps.
```

---

### US-002: One-Command Multi-Service Exposure (Foreground)
**As a** Full-Stack Developer,  
**I want to** run `porta start`,  
**so that** my frontend and backend are exposed under a single public HTTPS URL without CORS errors.

```gherkin
Scenario: Successful unified multi-service startup in foreground
  Given a valid "porta.yaml" defining frontend (:3000 -> /) and backend (:8000 -> /api)
  When I execute "porta start"
  Then PORTA starts the embedded reverse proxy on an ephemeral loopback port
  And connects a Cloudflare Quick Tunnel using cached or downloaded driver
  And renders the interactive TUI showing the public URL "https://*.trycloudflare.com"
  And requests to "https://*.trycloudflare.com/api/users" route directly to localhost:8000.
```

---

### US-003: Password-Protected Client Demo with Safe Logging
**As a** Freelance Developer,  
**I want to** protect my public URL with credentials and ensure tokens are not logged in plaintext,  
**so that** unauthorized users cannot access work and credentials are kept safe.

```gherkin
Scenario: Unauthorized visitor receives Basic Auth challenge
  Given "porta.yaml" configured with "security.mode: password" and "password: admin:Secret123"
  When an external user navigates to the public URL in a web browser without credentials
  Then PORTA intercepts the request and responds with HTTP 401 Unauthorized
  And displays a browser Basic Auth login prompt.

Scenario: Authorized visitor access with token redaction in logs
  Given "porta.yaml" configured with "security.mode: token"
  When a client accesses the URL with "?porta_token=secret_token"
  Then PORTA permits access to the upstream service
  And writes the log entry with "?porta_token=[REDACTED]".
```

---

### US-004: Resilient Upstream Restarting
**As a** Backend Developer,  
**I want** PORTA to keep the tunnel open while I restart my backend server,  
**so that** I don't have to restart PORTA and generate a new public URL every time my code changes.

```gherkin
Scenario: Upstream server temporarily restarts
  Given PORTA is running with backend service at localhost:8000
  When I stop and restart my backend server
  Then PORTA marks the backend as [OFFLINE] in the TUI
  And serves a friendly 502 page for "/api" requests during downtime
  And automatically marks backend [ONLINE] and resumes routing when port 8000 opens again.
```

---

## 42. ACCEPTANCE CRITERIA SUMMARY

1. `porta init` generates a valid `porta.yaml` within 500ms.
2. `porta start` establishes a live public HTTPS URL in foreground mode.
3. Path routing correctly dispatches `/` and `/api` to different ports under the same domain.
4. WebSocket connection hijacking preserves real-time bi-directional streaming.
5. `Ctrl+C` terminates all child processes and releases ports in $< 500$ms.
6. Target services outside loopback IP addresses are blocked by the SSRF Guard.
7. Token parameters and Authorization headers are masked in all log outputs.

---

## 43. FEATURE PRIORITIES

| Priority Level | Meaning | Features Included |
| :--- | :--- | :--- |
| **P0 (Must Have - MVP)** | Core product promise; required for v1.0.0 release. | `init`, `start` (foreground), `status`, `logs`, `doctor`, Config Parser, Ephemeral Reverse Proxy, Path Matching, WebSocket, Cloudflare Quick Tunnel (On-demand download & cache), Health Checker, Basic Auth, Token Auth with Log Redaction, SSRF Guard, TUI. |
| **P1 (Important - v1.1)** | High-value enhancements immediately post-MVP. | Named Persistent Cloudflare Tunnels (custom domains), ngrok provider driver, detached background daemon mode (`--detach`, `porta stop`), custom gateway port config (`proxy.port`), Credential Store (`CRED-001`). |
| **P2 (Nice to Have - v1.2)** | Process convenience features. | Optional process runner (`run: npm run dev`) inside `porta.yaml`. |
| **P3 (Future - v2.0)** | Long-term roadmap items. | Web Inspector UI, Team Workspaces, Cloud Control Plane. |

---

## 44. SUCCESS METRICS

| Metric Category | Metric Definition | Measurement Method | Target Baseline | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Time to Value** | Time elapsed from `porta start` to active public URL | Benchmark timer | $< 1.5\text{ seconds}$ | Engineering Benchmark Target |
| **Startup Success Rate** | % of `porta start` executions resulting in live URL | Automated test suite | $> 99.0\%$ | Engineering Benchmark Target |
| **Routing Reliability** | % of requests correctly dispatched without gateway errors | HTTP integration tests | $100.0\%$ | Engineering Benchmark Target |
| **Resource Efficiency** | RAM consumption during idle runtime | OS process monitor | $< 25\text{ MB RAM}$ | Engineering Benchmark Target |
| **Teardown Cleanliness** | % of shutdowns leaving zero orphan processes | Cross-platform test suite | $100.0\%$ | Engineering Benchmark Target |

---

## 45. RELEASE CRITERIA

1. **Functional Completeness:** 100% of P0 requirements implemented and passing automated tests.
2. **Cross-Platform Verification:** Automated green CI/CD test builds on Windows 11, macOS (ARM64/x64), and Ubuntu Linux.
3. **Security Audit:** Pass static analysis with zero high/critical vulnerabilities; SSRF filter verified against internal IP spoofing; log redaction verified against credential leakage.
4. **Zero-Dependency Check:** Standalone binary runs on clean OS VMs without pre-installed Go, Node.js, Python, or Docker.
5. **Documentation Baseline:** Accurate README, CLI help texts, and configuration documentation.

---

## 46. MVP ACCEPTANCE CHECKLIST

```text
[✓] Single static executable binary created for Windows, macOS, and Linux
[✓] `porta init` auto-detects active ports and creates valid `porta.yaml`
[✓] `porta start` launches foreground reverse proxy on dynamic ephemeral port (127.0.0.1:0)
[✓] Path-based routing successfully dispatches multiple services (e.g. `/` and `/api`)
[✓] WebSocket connections successfully upgraded and sustained without drops
[✓] Cloudflare Quick Tunnel auto-downloads/caches driver and establishes HTTPS URL
[✓] Health Checker detects offline services and serves friendly 502 pages
[✓] Basic Auth and Token Auth gatekeepers block unauthorized public access
[✓] Token query parameters and Authorization headers are redacted in all logs
[✓] Localhost SSRF Guard strictly blocks non-loopback private network targets
[✓] TUI renders live route table, service health, and streaming request logs
[✓] `Ctrl+C` cleans up child processes and releases port bindings in < 500ms
```

---

## 47. POST-MVP ROADMAP

```text
┌────────────────────────────────────────────────────────────────────────┐
│                              PORTA ROADMAP                             │
├───────────────────┬───────────────────┬────────────────────────────────┤
│    MVP (v1.0.0)   │   Phase 2 (v1.1)  │         Future (v2.0)          │
├───────────────────┼───────────────────┼────────────────────────────────┤
│ • Foreground CLI  │ • Detached Daemon │ • Web-based Traffic Inspector  │
│ • Ephemeral Proxy │ • Fixed Port Opt  │ • Process Orchestrator (`run:`)│
│ • Quick Tunnel    │ • Named Tunnels   │ • Team Collaboration Spaces    │
│ • Driver Caching  │ • Ngrok / Tailscale│ • Cloud Control Plane         │
│ • Token Redaction │ • Credential Store│ • Raw TCP/UDP Tunneling        │
│ • SSRF Guard      │ • Config HotReload│                                │
└───────────────────┴───────────────────┴────────────────────────────────┘
```

---

## 48. RISKS & MITIGATIONS

| Risk ID | Risk Description | Severity | Likelihood | Mitigation Strategy |
| :--- | :--- | :--- | :--- | :--- |
| **RSK-01** | Cloudflare Quick Tunnel service outage or rate-limiting. | High | Low | Implement clean user diagnostic message; architect `TunnelProvider` so developers can switch to ngrok driver easily in v1.1. |
| **RSK-02** | `cloudflared` binary download failure or corrupted local cache. | High | Medium | Implement integrity check on cached binary; auto-retry download with clear network failure message. |
| **RSK-03** | Anti-virus software on Windows false-flagging child process spawning. | Medium | Medium | Sign Windows binaries with code signing certificate; keep child process execution transparent via standard Windows Job Objects. |
| **RSK-04** | Port collision on local gateway listener. | Low | Low | Always bind to dynamic ephemeral port `127.0.0.1:0` assigned by OS kernel. |
| **RSK-05** | Accidental exposure of internal corporate LAN via target host / DNS rebinding. | Critical | Low | Enforce strict Loopback SSRF Guard in Security Engine rejecting all non-loopback IPs upon connection dial. |
| **RSK-06** | Credential / Token leakage in terminal output or shared log files. | High | Medium | Implement mandatory automatic token masking/redaction in logging pipeline before writing to any sink. |

---

## 49. TRACEABILITY MATRIX

| FRM ID | SRS Component ID | PRD ID | Feature Description | Classification | Primary Acceptance Criteria | MVP Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **PM-001** | CLI / Init Engine | **PRD-F-001** | Automatic Workspace Init | `[SOURCE-BACKED]` | Generates `porta.yaml` from active ports | **P0 (MVP)** |
| **PM-002** | Config Validator | **PRD-F-002** | Project Naming & Isolation | `[SOURCE-BACKED]` | Sanitizes project namespace | **P0 (MVP)** |
| **CFG-001** | Config Engine | **PRD-F-002** | YAML Parser & Env Expansion | `[SOURCE-BACKED]` | Parses `${VAR}` and applies defaults | **P0 (MVP)** |
| **CFG-002** | Service Registry | **PRD-F-003** | Route Collision Prevention | `[SOURCE-BACKED]` | Blocks duplicate route declarations | **P0 (MVP)** |
| **SRV-001** | Service Registry | **PRD-F-003** | Multi-Service Registration | `[SOURCE-BACKED]` | Registers port/route mappings in memory | **P0 (MVP)** |
| **SRV-002** | Proxy Director | **PRD-F-004** | Path Stripping & Rewriting | `[SOURCE-BACKED]` | Strips path prefix if `strip_path: true` | **P0 (MVP)** |
| **PROXY-001**| Reverse Proxy | **PRD-F-004** | Ephemeral Gateway Listener | `[APPROVED PRODUCT DECISION]` | Binds dynamic loopback reverse proxy (127.0.0.1:0) | **P0 (MVP)** |
| **PROXY-002**| Reverse Proxy | **PRD-F-004** | Longest Prefix Match Routing | `[SOURCE-BACKED]` | Routes `/api` to backend, `/` to frontend | **P0 (MVP)** |
| **PROXY-003**| Reverse Proxy | **PRD-F-005** | WebSocket Hijacking & SSE | `[SOURCE-BACKED]` | Transparent bi-directional stream | **P0 (MVP)** |
| **PROXY-004**| Reverse Proxy | **PRD-F-006** | Header Forwarding | `[SOURCE-BACKED]` | Sets `X-Forwarded-For/Proto/Host` | **P0 (MVP)** |
| **TUNNEL-001**| Tunnel Manager | **PRD-F-007** | Provider Abstraction | `[SOURCE-BACKED]` | Implements generic `TunnelProvider` | **P0 (MVP)** |
| **TUNNEL-002**| Cloudflare Driver| **PRD-F-007** | Quick Tunnel with Driver Cache| `[APPROVED PRODUCT DECISION]` | Downloads/caches driver & provisions HTTPS URL | **P0 (MVP)** |
| **TUNNEL-003**| Tunnel Watchdog | **PRD-F-007** | Auto-Reconnect Supervisor | `[SOURCE-BACKED]` | Exponential backoff on disconnect | **P0 (MVP)** |
| **HC-001** | Health Checker | **PRD-F-008** | Pre-flight & Periodic Probing| `[SOURCE-BACKED]` | Probes TCP/HTTP target every 5s | **P0 (MVP)** |
| **HC-002** | Health Checker | **PRD-F-008** | Graceful 502 Degradation | `[SOURCE-BACKED]` | Serves 502 for offline service | **P0 (MVP)** |
| **SEC-001** | Security Engine | **PRD-F-009** | Basic Auth Gatekeeper | `[SOURCE-BACKED]` | Challenges unauthorized visitors (401) | **P0 (MVP)** |
| **SEC-002** | Security Engine | **PRD-F-009** | Bearer Token & Log Redaction | `[APPROVED PRODUCT DECISION]` | Blocks invalid tokens & redacts query/header tokens | **P0 (MVP)** |
| **SEC-003** | Security Engine | **PRD-F-010** | Strict Loopback SSRF Guard | `[APPROVED PRODUCT DECISION]` | Rejects all non-loopback targets | **P0 (MVP)** |
| **CLI-001** | CLI Dispatcher | **PRD-F-012** | Command Suite & Doctor | `[SOURCE-BACKED]` | Executes CLI commands & diagnostics | **P0 (MVP)** |
| **CLI-002** | UI Subsystem | **PRD-F-011** | Foreground TUI & Live Logs | `[APPROVED PRODUCT DECISION]` | Renders live route table & traffic feed in foreground | **P0 (MVP)** |
| **MON-001** | Logger Subsystem | **PRD-F-011** | Redacted Access Logging | `[APPROVED PRODUCT DECISION]` | Writes sanitized JSON logs to `.porta/logs/` | **P0 (MVP)** |
| **MON-002** | Logger Subsystem | **PRD-F-011** | Log Inspection (`porta logs`)| `[SOURCE-BACKED]` | Tails and filters access logs in CLI | **P0 (MVP)** |

---

## 50. OPEN QUESTIONS & DECISIONS REQUIRED

All prior open questions have been formally resolved through the Approved Product Decision Baseline:
- **DQ-001 (Binary Distribution):** RESOLVED $\rightarrow$ On-demand download + local cache in `~/.porta/bin/`.
- **DQ-002 (Local Gateway Port):** RESOLVED $\rightarrow$ Ephemeral loopback port `127.0.0.1:0` for MVP; custom/fixed port deferred to post-MVP.
- **DQ-003 (Runtime Model):** RESOLVED $\rightarrow$ Foreground CLI execution model for MVP; detached daemon mode deferred to v1.1.

*Zero implementation-blocking product questions remain open.*

---

## 51. FINAL PRODUCT DEFINITION

### What PORTA IS
PORTA is a **lightweight, standalone developer CLI tool** that binds multiple local application services into an integrated reverse proxy gateway on an ephemeral loopback port and exposes them to the public internet via a secure, encrypted HTTPS tunnel with a single command in the foreground.

### What PORTA IS NOT
PORTA is **NOT a cloud hosting platform, NOT a deployment server, NOT a process orchestrator, NOT a background daemon manager, NOT a VPN, and NOT an enterprise identity provider**. It does not host code in the cloud or manage background production workloads.

### Who PORTA IS FOR
PORTA is built for **Full-Stack Developers, Mobile App Engineers, and Software Teams** who need to instantly share, demo, review, and test their live localhost applications with team members, clients, mobile devices, and external webhook services.

### What MVP Delivers
The MVP delivers a **rock-solid, cross-platform single binary (`porta`)** that provides zero-configuration port discovery (`porta init`), one-command foreground startup (`porta start`), path-based multi-service reverse proxying on an ephemeral port, zero-auth Cloudflare Quick Tunnel exposure with automatic driver caching, live terminal TUI status and sanitized log streaming, Basic/Token auth gatekeeping with token redaction, strict SSRF isolation, and graceful error handling.

### What Success Looks Like
A developer has their React frontend running on `:3000` and Node API on `:8000`. They type `porta start`. Within 1.5 seconds, their terminal displays an interactive status screen with a public URL: `https://app-slug.trycloudflare.com`. They share this link with their client; the client navigates to the app, logs in, and tests live features seamlessly while the developer watches sanitized real-time request logs scroll by in their terminal. When the demo is over, the developer hits `Ctrl+C`, and PORTA shuts down cleanly in under 500 milliseconds.

---

## 52. PRD REVISION VALIDATION REPORT

```text
PRD Status:                         REVISED PRODUCT REQUIREMENT BASELINE
PRD Version:                        1.0.1
PRD File:                           docs/PRODUCT_REQUIREMENT_DOCUMENT.md
FRM Reviewed:                       YES (docs/FUNCTIONAL_REQUIREMENT_MAP.md v1.0.0)
SRS Reviewed:                       YES (docs/SYSTEM_REQUIREMENT_SPECIFICATION.md v1.0.0)
Previous PRD Reviewed:              YES (PRD v1.0.0 baseline)
Approved Product Decisions Applied: YES (All 7 Revision Decisions strictly incorporated)

Changes Made:
1. Clarified MVP runtime model: locked to foreground execution (`porta start` + `Ctrl+C` teardown).
2. Resolved Decision 01: Adopted on-demand download + local cache (`~/.porta/bin/`) for Cloudflare Quick Tunnel driver.
3. Resolved Decision 02: Locked local gateway to ephemeral loopback port (`127.0.0.1:0`); removed custom gateway port config from MVP.
4. Resolved Decision 03: Reclassified detached daemon mode (`--detach`, `porta stop`, daemon supervisor) to Post-MVP (v1.1).
5. Resolved Decision 04: Performance metrics reclassified from "Validated" to "Engineering Benchmark Target (To Be Validated)".
6. Resolved Decision 05: Added mandatory credential & token query parameter redaction in all logging sinks (TUI, stdout, stderr, access.log).
7. Resolved Decision 06: Strengthened Loopback SSRF Guard to strictly prohibit non-loopback, private LAN, cloud metadata, and DNS rebinding targets.
8. Resolved Decision 07: Reclassified full Credential Store (`CRED-001`) to Post-MVP provider support (since Quick Tunnel is zero-auth).
9. Refined One-Command concept: `porta start` is the primary one-command runtime; `porta init` is workspace initialization.
10. Explicitly classified all requirements as [SOURCE-BACKED] or [APPROVED PRODUCT DECISION].

Requirements Covered:               100% of MVP Functional Requirements (PRD-F-001 through PRD-F-012)
Requirements Not Yet Traceable:     NONE
Conflicts Found:                    NONE (All specifications are harmonized)
Open Questions:                     0 Implementation-blocking questions remaining (All resolved in Section 50)
Assumptions:                        Classified and documented (Standard Go networking, Cloudflare Quick Tunnel availability)
Proposed Items:                     Quarantined to Post-MVP Roadmap (v1.1 / v1.2 / v2.0)
Decision Required Items:            NONE

Scope Creep Found:                  Detached daemon management, fixed gateway port config, full credential store in MVP.
Scope Creep Removed:                Moved to Post-MVP (v1.1) roadmap.

MVP Features:                       • CLI Suite (`init`, `start` foreground, `status`, `logs`, `config`, `doctor`)
                                    • Declarative YAML Parser (`porta.yaml`) with ${ENV} expansion
                                    • Multi-Service Registry & Longest Prefix Match Reverse Proxy
                                    • Ephemeral Gateway Binding (127.0.0.1:0)
                                    • WebSocket & SSE Connection Hijacking
                                    • Cloudflare Quick Tunnel Driver with On-Demand Download & Local Cache
                                    • TCP & HTTP Health Checking with Graceful 502 Degradation
                                    • Basic Auth & Bearer Token Gatekeeper
                                    • Mandatory Credential & Token Log Redaction
                                    • Strict Loopback SSRF Isolation Guard
                                    • Foreground Interactive Terminal UI (TUI) & Live Request Stream
                                    • Native Cross-Platform Support (Windows, macOS, Linux)

Post-MVP Features:                  • Detached Daemon Mode (`--detach`, `porta stop`) (v1.1)
                                    • Custom/Fixed Gateway Port Config (`proxy.port`) (v1.1)
                                    • Named Cloudflare Tunnels with Custom Domains (v1.1)
                                    • Ngrok & Tailscale Funnel Provider Drivers (v1.1)
                                    • Credential Store for Provider API Keys (v1.1)
                                    • Process Orchestration (`run: npm run dev`) (v1.2)
                                    • Web-based Request Inspector & Cloud Control Plane (v2.0)

Security Concerns Reviewed:         SSRF protection, Loopback-only enforcement, Token log redaction, Constant-time auth checks.
Performance Targets Reviewed:       Startup < 1.5s, RAM < 25MB, Routing overhead < 3ms (All marked as Engineering Benchmark Targets).
Traceability Reviewed:              Bi-directional mapping across FRM, SRS, PRD, and Acceptance Criteria fully verified.

Implementation-Blocking Decisions Remaining: NONE

Final Recommendation:               READY FOR IMPLEMENTATION PLAN
```

---
*End of Product Requirement Document v1.0.1.*
