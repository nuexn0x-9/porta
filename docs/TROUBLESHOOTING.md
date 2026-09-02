# PORTA TROUBLESHOOTING GUIDE

This guide provides actionable solutions for common issues encountered when using **PORTA**.

---

## 1. Quick Diagnostic Check

Before troubleshooting specific issues, run the built-in system doctor:
```bash
porta doctor
```
`porta doctor` checks loopback bindability, outbound DNS/HTTPS connectivity, Cloudflare driver availability, and configuration validity.

---

## 2. Common Issues & Solutions

### 2.1 `502 Bad Gateway (Service Offline)`
- **Symptom:** Visiting the public URL returns a PORTA 502 Bad Gateway page.
- **Cause:** PORTA is running, but the local development server on the configured port (e.g., `:3000` or `:8000`) is stopped, crashed, or still booting.
- **Solution:** Start your local application dev server (`npm run dev`, `go run main.go`, `uvicorn main:app`). PORTA will automatically detect active listening sockets and resume forwarding traffic within 5 seconds without restarting PORTA.

---

### 2.2 `Error: failed to locate or download cloudflared binary`
- **Symptom:** `porta start` halts during tunnel initialization with a driver download error.
- **Cause:** Workstation is offline, behind a corporate proxy blocking GitHub releases, or lacks write permissions in `~/.porta/bin/`.
- **Solution:**
  1. Ensure your internet connection is active.
  2. Alternatively, manually install `cloudflared` into your system `PATH` (e.g. via `brew install cloudflared`, `winget install Cloudflare.cloudflared`, or `apt install cloudflared`). PORTA will automatically detect and use the system binary.

---

### 2.3 `Error: route collision error (ERR_ROUTE_COLLISION)`
- **Symptom:** `porta start` or `porta config` fails with `ERR_ROUTE_COLLISION`.
- **Cause:** Two or more services in `porta.yaml` define identical `route` paths (e.g., both Service A and Service B bind `/api`).
- **Solution:** Update `porta.yaml` so that each service has a unique route prefix (e.g., `/` for frontend and `/api` for backend API).

---

### 2.4 `Security Violation: target host is not a loopback address (SSRF protection)`
- **Symptom:** `porta start` rejects configuration during validation.
- **Cause:** A service is configured with a non-loopback IP address (such as `192.168.1.50` or `10.0.0.1`).
- **Solution:** Change `host` to `127.0.0.1` or `localhost`. PORTA strictly forbids tunneling traffic to remote LAN addresses to prevent SSRF vulnerabilities.

---

### 2.5 `401 Unauthorized` or `403 Forbidden`
- **Symptom:** Browser or webhook client receives a 401 or 403 error.
- **Cause:** `security.mode` is set to `password` or `token`, but valid credentials were not provided in the request.
- **Solution:**
  - For Basic Auth (`password`): supply the correct username and password configured in `porta.yaml` or `${PORTA_PASSWORD}`.
  - For Token Auth (`token`): pass the header `Authorization: Bearer <token>` or URL query parameter `?porta_token=<token>`.

---

### 2.6 `Error: at least one service must be defined under 'services'`
- **Symptom:** `porta start` exits with `ERR_CFG_NO_SERVICES`.
- **Cause:** The `services:` block in `porta.yaml` is empty.
- **Solution:** Run `porta init` to auto-detect active listening ports or define at least one service with a `port` in `porta.yaml`.

---
*End of Troubleshooting Guide.*
