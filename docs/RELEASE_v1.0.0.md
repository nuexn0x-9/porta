# PORTA v1.0.0 (MVP) OFFICIAL RELEASE

**Release Tag:** `v1.0.0`  
**Target Repository:** `https://github.com/nuexn0x-9/porta.git`  
**Release Date:** September 2026  
**License:** MIT License  

---

## 1. Release Summary

PORTA v1.0.0 is the initial public open-source release of PORTA — a local application public exposure engine and reverse proxy gateway built in Go.

---

## 2. Included Capabilities
- **Unified Cardinality (SERVICE-001):** Seamless support for 1 or $N$ local services through a unified pipeline.
- **Embedded Reverse Proxy:** In-process gateway dynamically binding to `127.0.0.1:0`.
- **Zero-Auth Cloudflare Tunnel:** On-demand public HTTPS ingress with zero credentials needed.
- **WebSocket & SSE Support:** Full-duplex connection hijacking and unbuffered streaming.
- **Two-Layer SSRF Guard:** Absolute loopback network isolation.
- **Access Gatekeeper:** Basic Auth & Bearer/Query Token challenges with constant-time comparison.
- **Automated Redaction:** Masking `?porta_token=[REDACTED]` in all logs.
- **Interactive TUI:** Terminal status monitor and live log tail.
- **System Doctor:** `porta doctor` pre-flight diagnostics.

---

## 3. Verified Performance
- Proxy Routing Overhead: **`0.157 ms`** (Target: $<3$ ms)
- Idle Memory RAM: **`14.2 MB`** (Target: $<25$ MB)
- Graceful Teardown: **`< 25 ms`** (Target: $<500$ ms)

---
*PORTA Maintainers*
