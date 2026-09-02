# FINAL PRODUCT REQUIREMENTS DOCUMENT (PRD)

**Product Name:** PORTA  
**Tagline:** Expose your local multi-service application to the world with one command.  
**Release Version:** PORTA v1.0.0 (MVP)  
**Document Status:** OFFICIAL RELEASE BASELINE  
**Maintained By:** Lead Product Architect & Engineering Team  

---

## 1. Product Overview

### 1.1 Core Mission
> **"Expose your local multi-service application to the world with one command."**

PORTA is a standalone, client-side developer CLI tool that transforms multi-service applications running on `localhost` (such as frontend at `:3000`, REST/GraphQL API at `:8000`, and WebSocket server at `:9001`) into a secure, unified public HTTPS environment with zero cloud deployments, zero server provisioning, zero DNS setup, and zero router port forwarding.

### 1.2 Target Personas
1. **Full-Stack Developers:** Needing to demo live feature branches to clients or team members without committing unreviewed code or dealing with CORS errors across multiple tunnels.
2. **Mobile Engineers:** Needing to test iOS/Android apps running on physical devices over 4G/5G networks against local backend APIs.
3. **Webhook Integrators:** Receiving live webhooks from third-party services (Stripe, Midtrans, GitHub) directly on localhost.
4. **QA Engineers:** Reviewing live pull request environments instantly on the engineer's workstation.

### 1.3 Core Value Proposition
- **Unified Ingress:** Replaces disjointed multi-tunnel setups with a single public domain and intelligent path routing (`/` $\rightarrow$ `:3000`, `/api` $\rightarrow$ `:8000`).
- **Zero-Friction Default:** Works out-of-the-box via Cloudflare Quick Tunnel without requiring account registration, API tokens, or credit cards.
- **Unified Cardinality (DECISION SERVICE-001):** Seamlessly supports both single-service (`localhost:3000`) and complex multi-service setups through the identical architecture.
- **Built-in Security:** Protects local workstations with strict Loopback SSRF filtering, Basic Auth/Bearer Token gatekeeping, and automated token redaction in logs.

---

## 2. MVP Product Scope (PORTA v1.0.0)

| Functional Area | Scope Description |
| :--- | :--- |
| **CLI Suite** | `porta init`, `porta start`, `porta status`, `porta logs`, `porta config`, `porta doctor`. |
| **Runtime Model** | Attached foreground CLI runtime; clean teardown via `Ctrl+C` in $<500$ms. |
| **Configuration Engine** | YAML schema (`porta.yaml`) with `${ENV_VAR:-default}` expansion and schema validation. |
| **Service Registry** | In-memory registry supporting 1 service or $N$ services (`services.length >= 1`). |
| **Reverse Proxy** | Embedded Go gateway dynamically binding on ephemeral loopback port (`127.0.0.1:0`). |
| **Routing Engine** | Longest Prefix Match (LPM) dispatcher with path stripping (`strip_path: true`). |
| **Protocol Support** | HTTP/1.1, HTTP/2, WebSocket connection hijacking, and unbuffered SSE/Chunked streaming. |
| **Tunnel Provider** | Cloudflare Quick Tunnel provider driver with on-demand binary download & local caching (`~/.porta/bin/`). |
| **Health Checking** | Active TCP ping and HTTP GET health probing with 502 Bad Gateway isolation. |
| **Security Gatekeeper** | Basic Auth (`401 Challenge`), Bearer Token (`403 Forbidden`), and Strict Loopback SSRF Guard. |
| **Observability** | Interactive ANSI Terminal UI (TUI) and structured JSON logging with token masking. |
| **Cross-Platform** | Native execution on Windows 10/11, macOS (Intel & Apple Silicon), and Linux. |

---

## 3. Explicitly Out of Scope for MVP (Post-MVP Roadmap)

- **Detached Background Daemon Mode (`--detach`, `porta stop`):** Deferred to v1.1.
- **Custom / Fixed Gateway Port Configuration (`proxy.port`):** Deferred to v1.1.
- **Named Persistent Custom Domains (Cloudflare DNS binding):** Deferred to v1.1.
- **Multi-Provider Tunnel Drivers (ngrok, Tailscale Funnel):** Deferred to v1.1 via existing `TunnelProvider` interface.
- **Full Credential Store (`CRED-001` for external provider tokens):** Deferred to v1.1.
- **Process Orchestration (`run: npm run dev`):** Deferred to v1.2.
- **Web-Based Inspection Dashboard:** Deferred to v2.0.
- **Cloud Control Plane & Team Collaboration:** Deferred to v2.0.
- **Raw TCP / UDP Tunneling:** Deferred to v2.0.

---
*End of Final Product Requirements Document.*
