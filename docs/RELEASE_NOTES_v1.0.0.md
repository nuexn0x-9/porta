# RELEASE NOTES — PORTA v1.0.0 (MVP)

**Release Version:** v1.0.0 (MVP)  
**Release Date:** September 2026  
**License:** Apache 2.0 / MIT (Dual Licensed)  
**Git Tag:** `v1.0.0`  

---

## 1. Overview

We are excited to announce the official release of **PORTA v1.0.0 (MVP)** — a developer tool and runtime engine designed to expose multi-service local development environments to the public internet with a single command.

PORTA eliminates the friction of configuring cloud staging servers, manual ngrok tunneling, complex CORS setups, and router port forwarding.

---

## 2. Major Features in v1.0.0

- **Unified Single & Multi-Service Exposure (DECISION SERVICE-001):** Expose a single frontend (`localhost:3000`) or an entire microservice stack (`:3000` + `:8000` + `:9001`) through a single public HTTPS URL.
- **Zero-Auth Cloudflare Quick Tunnel:** Establish secure, encrypted public HTTPS tunnels immediately without creating accounts, registering API keys, or paying fees.
- **Embedded Longest Prefix Match (LPM) Reverse Proxy:** High-performance in-process gateway with optional path prefix stripping (`strip_path: true`).
- **Full-Duplex WebSocket Hijacking:** Transparent bi-directional streaming for Vite/Webpack HMR, Socket.io, Chat servers, and multiplayer applications.
- **Unbuffered SSE & LLM Streaming:** Zero-delay response flushing for Server-Sent Events and AI streaming completions.
- **Two-Layer Loopback SSRF Guard:** Absolute isolation preventing access to private subnets (RFC 1918), cloud metadata endpoints (`169.254.169.254`), and DNS rebinding vectors.
- **Ingress Access Gatekeeper:** Protect public endpoints with HTTP Basic Auth or Bearer/Query Tokens.
- **Safe Structured Access Logging:** Disk persistence (`.porta/logs/access.log`) with automated token redaction (`?porta_token=[REDACTED]`).
- **Interactive Foreground TUI:** Real-time ANSI terminal status table with live request streaming and sub-500ms graceful `Ctrl+C` teardown.
- **CLI Suite:** `porta init` (with safe port scanning ignoring databases), `porta start`, `porta status`, `porta logs`, `porta config`, and `porta doctor`.

---

## 3. Performance Benchmark Summary

- **Reverse Proxy Latency Overhead:** `0.157 ms` (Target: $<3.0$ ms)
- **Idle Memory Footprint:** `14.2 MB RAM` (Target: $<25.0$ MB)
- **Cold Startup Time:** `0.030 s` (Target: $<1.5$ s)
- **Teardown Duration:** `< 25 ms` (Target: $<500$ ms)

---

## 4. Known Limitations & Roadmap

### Known Limitations in MVP:
- Foreground runtime only (`Ctrl+C` halts tunnel). Detached background daemon deferred to v1.1.
- Random generated subdomains via Cloudflare Quick Tunnel. Persistent custom domains deferred to v1.1.
- HTTP/WebSocket tunneling only. Raw TCP/UDP protocols deferred to v2.0.

### Upcoming in v1.1:
- Detached Daemon Mode (`porta start --detach` / `porta stop`)
- Persistent Custom Domains (`tunnel.subdomain`)
- Multiple Tunnel Provider Drivers (ngrok, Tailscale Funnel)

---
*PORTA Team — September 2026*
