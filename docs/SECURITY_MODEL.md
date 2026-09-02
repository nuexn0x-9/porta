# PORTA SECURITY MODEL & ARCHITECTURE

**Document Version:** 1.0.0  
**Target:** PORTA v1.0.0 (MVP)  

---

## 1. Security Philosophy & Ingress Boundary

PORTA operates under a **Strict Least-Privilege Exposure Model**:
> **"PORTA is an intentional local application exposure engine, NOT an open network proxy."**

PORTA's primary security responsibility is to bridge public traffic exclusively to designated, approved loopback services on the developer's local machine, while preventing malicious external actors from exploiting the gateway to attack internal networks or intercept credentials.

```text
[ PUBLIC INTERNET ]
        │
        ▼ (TLS 1.3 Anycast Edge)
[ Cloudflare Quick Tunnel ]
        │
        ▼ (Encrypted QUIC / mTLS Egress)
[ PORTA Ingress Boundary ]
        ├─ 1. SSRF & Host Validation Guard (Rejects non-loopback)
        ├─ 2. Security Gatekeeper (Basic Auth & Token Challenges)
        ├─ 3. Longest Prefix Match Routing Table
        ├─ 4. Safe Logging Interceptor (Token & Header Redaction)
        ▼
[ Local Loopback Services Only ] (127.0.0.1:3000, 127.0.0.1:8000)
```

---

## 2. Threat Model & Mitigations

### 2.1 Server-Side Request Forgery (SSRF) & Internal LAN Pivoting
- **Threat:** An attacker crafts requests or configures upstream target hosts pointing to internal corporate subnets (`10.0.0.0/8`, `192.168.0.0/16`) or cloud metadata endpoints (`169.254.169.254`) to steal IAM credentials or exploit internal microservices.
- **PORTA Defense:** **Two-layer SSRF Protection:**
  1. *Static Schema Validation:* `config.Validate()` verifies all configured `host` properties belong strictly to `127.0.0.1`, `localhost`, or `::1`.
  2. *Dial-Time Enforcement:* `security.SafeDialContext()` hooks directly into Go's `net.Dialer`. It resolves destination hostnames and rejects the socket connection if any resolved IP address is non-loopback.

### 2.2 DNS Rebinding Attacks
- **Threat:** An attacker configures a domain name that initially resolves to `127.0.0.1` during config validation, but switches DNS resolution to an internal IP (`192.168.1.1`) during runtime HTTP requests.
- **PORTA Defense:** `SafeDialContext()` evaluates IP resolution on *every outbound connection dial*, completely mitigating DNS rebinding attacks.

### 2.3 Sensitive Credential & Token Log Leakage
- **Threat:** Access tokens passed in webhook URLs (`?porta_token=secret123`) or `Authorization` headers are written in plaintext to disk logs (`.porta/logs/access.log`) or displayed in shared screen sessions.
- **PORTA Defense:** `internal/ui/logger.go` automatically executes regex token masking on all URL query parameters and headers before writing to any log sink (`?porta_token=[REDACTED]`).

### 2.4 Unauthenticated Internet Scanners & Brute Force
- **Threat:** Automated bots scanning Cloudflare Anycast domains hit unreleased staging applications.
- **PORTA Defense:** Setting `security.mode: password` or `security.mode: token` challenges incoming traffic at the gateway layer (`401 Unauthorized` / `403 Forbidden`) with constant-time string comparisons (`crypto/subtle.ConstantTimeCompare`), preventing requests from reaching the developer's application without valid authentication.

---

## 3. Storage Security & File Permissions

- **Local Runtime State (`.porta/`):** All directories and files created by PORTA enforce POSIX permissions `0700` (directory) and `0600` (files).
- **Environment Variables:** Credentials and secrets are never written to disk in plaintext if supplied via environment variables (`${PORTA_SECRET}`).

---
*End of Security Model.*
