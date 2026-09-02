# PORTA USER GUIDE

Welcome to **PORTA** — the developer tool that exposes your local multi-service applications to the world with a single command.

---

## 1. Installation

PORTA is distributed as a single static binary with zero external runtime dependencies.

### Windows
1. Download `porta.exe` from the latest GitHub Release.
2. Place `porta.exe` in a folder included in your system `PATH` (e.g., `C:\Program Files\PORTA\` or `%USERPROFILE%\bin\`).
3. Open PowerShell or Command Prompt and verify:
   ```powershell
   porta doctor
   ```

### macOS
1. Download the macOS binary (`porta`) for your architecture (Apple Silicon / Intel).
2. Make the binary executable and move it to `/usr/local/bin`:
   ```bash
   chmod +x porta
   sudo mv porta /usr/local/bin/porta
   porta doctor
   ```

### Linux
1. Download the Linux binary (`porta`) for your architecture (x86_64 / ARM64).
2. Make it executable and place it in `/usr/local/bin`:
   ```bash
   chmod +x porta
   sudo mv porta /usr/local/bin/porta
   porta doctor
   ```

---

## 2. Quickstart Tutorial: Single Service (e.g., Vite/React on `:3000`)

### Step 1: Initialize Configuration
In your project root directory:
```bash
porta init
```
PORTA automatically detects port `3000` and creates `porta.yaml`:
```yaml
# porta.yaml
version: "1"

project:
  name: my-web-app

services:
  app:
    port: 3000
    route: /
```

### Step 2: Start PORTA
```bash
porta start
```

### Step 3: Access Public URL
Within 2 seconds, PORTA renders the interactive TUI:
```text
┌─────────────────────────────────────────────────────────────┐
│  PORTA v1.0.0 — Local Multi-Service Public Exposure Engine  │
├─────────────────────────────────────────────────────────────┤
│   Project     : my-web-app                                  │
│   Public URL  : https://random-slug.trycloudflare.com       │
│   Security    : Public / Unrestricted                       │
├─────────────────────────────────────────────────────────────┤
│   SERVICES & ROUTES:                                        │
│   • app (127.0.0.1:3000) -> /                 [ONLINE]      │
├─────────────────────────────────────────────────────────────┤
│   LIVE LOGS (Press Ctrl+C to stop):                         │
│   14:30:12 [200] GET  /index.html   -> 127.0.0.1:3000 (8ms) │
└─────────────────────────────────────────────────────────────┘
```
Share the `https://random-slug.trycloudflare.com` URL with your client or teammate.

### Step 4: Stop Session
Press `Ctrl+C` in your terminal. PORTA terminates the tunnel and releases port bindings in under 500ms.

---

## 3. Multi-Service Tutorial: Frontend + Backend API

### Scenario:
- **Frontend App:** Running at `http://localhost:3000`
- **Backend API Server:** Running at `http://localhost:8000`

### Step 1: Create `porta.yaml`
```yaml
# porta.yaml
version: "1"

project:
  name: store-platform

services:
  frontend:
    port: 3000
    route: /

  backend:
    port: 8000
    route: /api
    strip_path: false
    health_check:
      path: /healthz
      interval: 5s

security:
  mode: password
  password: ${PORTA_PASSWORD:-admin:SecretDemo123!}
```

### Step 2: Start Exposure
```bash
porta start
```

### How Ingress Routing Works:
- `https://<slug>.trycloudflare.com/` $\rightarrow$ proxied to `http://localhost:3000/`
- `https://<slug>.trycloudflare.com/api/v1/products` $\rightarrow$ proxied to `http://localhost:8000/api/v1/products`
- Visitors are prompted for username/password (`admin` / `SecretDemo123!`).

---

## 4. Complete CLI Reference

### 4.1 `porta init`
- **Purpose:** Scans active local ports and generates starter `porta.yaml`.
- **Usage:** `porta init [--force]`
- **Flags:**
  - `-f, --force`: Overwrite existing `porta.yaml` without confirmation.
- **Example Output:**
  ```text
  [i] Scanning local listening ports for active web applications...
  [✓] Detected 2 active services on ports: [3000 8000]
  [✓] Successfully initialized porta.yaml
  [i] Run 'porta start' to launch your public environment!
  ```

---

### 4.2 `porta start`
- **Purpose:** Starts the reverse proxy gateway, connects the public tunnel, and displays the foreground TUI.
- **Usage:** `porta start [-c porta.yaml]`
- **Flags:**
  - `-c, --config string`: Path to configuration file (default: `porta.yaml`).
- **Shutdown:** Press `Ctrl+C` (`SIGINT`) for graceful teardown.

---

### 4.3 `porta status`
- **Purpose:** Inspects configured services and routing mappings.
- **Usage:** `porta status [--json]`
- **Flags:**
  - `--json`: Output status in structured JSON format.
- **Example Output:**
  ```text
  Project: store-platform (Environment: development)
  Registered Services (2):
    • frontend     127.0.0.1:3000 -> /
    • backend      127.0.0.1:8000 -> /api
  ```

---

### 4.4 `porta logs`
- **Purpose:** Tails and filters structured access logs from `.porta/logs/access.log`.
- **Usage:** `porta logs [-f] [--service <name>] [--level <info|warn|error>]`
- **Flags:**
  - `-f, --follow`: Stream new log entries in real time.
  - `--service string`: Filter logs for a specific service name or address.
  - `--level string`: Filter logs by minimum severity level.

---

### 4.5 `porta setup`
- **Purpose:** Initializes user directory structure (`~/.porta/`), generates default `~/.porta/config/global.yaml`, verifies loopback bindability, and pre-caches the Cloudflare tunnel driver.
- **Usage:** `porta setup`
- **Example Output:**
  ```text
  [i] Initializing PORTA environment...
    [✓] System detected       : windows/amd64
    [✓] Directory structure   : ~/.porta/ initialized
    [✓] Loopback interface    : 127.0.0.1 bindable
    [✓] Internet connectivity : DNS & HTTPS connection OK
    [✓] Tunnel driver         : Ready (~/.porta/bin/cloudflared.exe)
  ```

---

### 4.6 `porta upgrade`
- **Purpose:** Checks GitHub Releases for updates and performs atomic in-place binary upgrades with SHA-256 checksum verification.
- **Usage:** `porta upgrade`
- **Example Output:**
  ```text
  [i] Current PORTA version: v1.1.0
  [i] Checking GitHub Releases for updates...
  [✓] You are already on the latest version (v1.1.0)!
  ```

---

### 4.7 `porta version`
- **Purpose:** Prints version, commit hash, build date, and Go runtime environment.
- **Usage:** `porta version [--json]`
- **Flags:**
  - `--json`: Output metadata in structured JSON format.
- **Example Output:**
  ```text
  PORTA version v1.1.0 (release) windows/amd64
  Build date : 2026-09-02
  Go runtime : go1.27.0
  ```

---

### 4.8 `porta config`
- **Purpose:** Validates syntax and displays parsed configuration model.
- **Usage:** `porta config [-c porta.yaml]`
- **Example Output:**
  ```text
  [✓] Configuration syntax and rules are VALID.
  ---
  version: "1"
  project:
      name: store-platform
  ...
  ```

---

### 4.9 `porta doctor`
- **Purpose:** Runs diagnostic health checks on system, networking, and drivers.
- **Usage:** `porta doctor`
- **Example Output:**
  ```text
  PORTA System Doctor Diagnostics
  ===============================
  [✓] OS & Architecture: windows/amd64
  [✓] Loopback Interface: 127.0.0.1 bindable
  [✓] Outbound Internet: DNS resolution & HTTPS connection OK
  [✓] Tunnel Driver: Ready (C:\Program Files (x86)\cloudflared\cloudflared.exe)
  [✓] Configuration (porta.yaml): Valid (2 service(s) configured)
  ```

---
*End of User Guide.*
