# SYSTEM REQUIREMENT SPECIFICATION (SRS)
**Project:** PORTA (Local Multi-Service Application Public Exposure Platform)  
**Document Version:** 1.0.0-PROD-SPEC  
**Author:** Lead Product Architect + System Architect  
**Status:** APPROVED FOR IMPLEMENTATION BASELINE  

---

## 1. SYSTEM OVERVIEW

PORTA adalah sistem runtime client-side independen yang bertindak sebagai **Local Gateway & Secure Egress Orchestrator**. Sistem ini menggabungkan beberapa service lokal yang berjalan di port berbeda (misal: frontend React/Vite di `:3000`, backend Go/Node/Python di `:8000`, WebSocket engine di `:9001`) menjadi satu endpoint gateway lokal yang terpadu, kemudian mengeksposnya ke internet secara aman melalui *encrypted tunneling* (Cloudflare Quick Tunnel / Named Tunnel / ngrok) dengan perlindungan security gatekeeper (Basic Auth / Token / SSRF mitigation).

---

## 2. ARCHITECTURE TOPOLOGY & DATA FLOW

### 2.1 High-Level Architecture Diagram

```text
                                  ┌──────────────────────────┐
                                  │      PUBLIC INTERNET     │
                                  │ (Browsers, Webhooks, QA) │
                                  └─────────────┬────────────┘
                                                │
                                                ▼  HTTPS (TLS 1.3)
                                  ┌──────────────────────────┐
                                  │   Tunnel Edge Provider   │
                                  │ (Cloudflare Edge / Ngrok)│
                                  └─────────────┬────────────┘
                                                │
════════════════════════════════════════════════╪═══════════════════════════════════════════════
 DEVELOPER MACHINE (LOCAL ENVIRONMENT)          │ Encrypted Tunnel Stream (QUIC / gRPC / mTLS)
                                                ▼
                                  ┌──────────────────────────┐
                                  │   PORTA TUNNEL ADAPTER   │  <─── Tunnel Manager (Watchdog)
                                  └─────────────┬────────────┘
                                                │ Local TCP Proxying
                                                ▼
                                  ┌──────────────────────────┐
                                  │  PORTA SECURITY ENGINE   │  <─── Auth / Token / SSRF Filter
                                  └─────────────┬────────────┘
                                                │ (Authorized Requests)
                                                ▼
                                  ┌──────────────────────────┐
                                  │  PORTA REVERSE PROXY     │  <─── Dynamic Routing Table
                                  │ (Embedded Gateway :8080) │       (Longest Prefix Match)
                                  └───────┬──────────┬───────┘
                                          │          │
                 ┌────────────────────────┘          └────────────────────────┐
                 │                                                            │
                 ▼ (Path: `/` / WebSocket)                                   ▼ (Path: `/api/*`)
     ┌───────────────────────┐                                    ┌───────────────────────┐
     │    Frontend Service   │                                    │    Backend Service    │
     │  http://127.0.0.1:3000│                                    │ http://127.0.0.1:8000 │
     └───────────────────────┘                                    └───────────────────────┘
```

### 2.2 End-to-End Request/Response Cycle
1. **Inbound Traffic:** Klien publik memanggil `https://qulineria-dev.trycloudflare.com/api/v1/orders`.
2. **Tunnel Egress:** Edge provider meneruskan payload HTTP terenkripsi ke proses adapter lokal PORTA via QUIC/mTLS tunnel.
3. **Local Ingress:** Tunnel adapter membongkar payload dan mem-forward HTTP request ke PORTA Embedded Reverse Proxy (`127.0.0.1:<allocated_gateway_port>`).
4. **Security Filter:** Middleware Security memeriksa apakah endpoint dilindungi `Basic Auth` atau `Bearer Token`. Jika gagal, return `401/403` seketika sebelum membebani service lokal.
5. **Path Dispatching:** Router mengevaluasi path `/api/v1/orders`. Aturan *Longest Prefix Match* mencocokkannya ke service backend (`127.0.0.1:8000`).
6. **Header Injection:** Gateway menyuntikkan `X-Forwarded-For`, `X-Forwarded-Proto: https`, `X-Forwarded-Host`, dan menyesuaikan `Host` header.
7. **Local Upstream Dispatch:** Gateway melakukan HTTP round-trip ke `http://127.0.0.1:8000/api/v1/orders`.
8. **Response Return:** Response dari backend dialirkan kembali (streaming/chunked) melalui proxy -> tunnel adapter -> edge -> client publik.
9. **Observability Hook:** Logger merekam `[200 OK] GET /api/v1/orders -> :8000 (14ms)` ke live terminal TUI dan `.porta/logs/access.log`.

---

## 3. COMPONENT ARCHITECTURE & SPECIFICATIONS

PORTA dibangun dari 11 sub-komponen modular yang terisolasi dengan kontrak interface yang tegas:

```text
┌────────────────────────────────────────────────────────────────────────┐
│                               PORTA CLI                                │
│                     (Cobra / UI TUI / Args Parser)                     │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼────────────────────────────────────┐
│                               PORTA CORE                               │
│                         (Orchestrator Engine)                          │
├─────────────────┬──────────────────┬─────────────────┬─────────────────┤
│  Config Engine  │ Service Registry │ Health Checker  │ Credential Store│
├─────────────────┼──────────────────┼─────────────────┼─────────────────┤
│ Proxy Manager   │  Tunnel Manager  │ Security Engine │ Runtime Manager │
└─────────────────┴──────────────────┴─────────────────┴─────────────────┘
                                    │
┌───────────────────────────────────▼────────────────────────────────────┐
│                       PORTA OBSERVABILITY & LOGS                       │
└────────────────────────────────────────────────────────────────────────┘
```

### Component Details:

#### 1. PORTA CLI
- **Tanggung Jawab:** Entry point user, flag parsing, formatted console output (ANSI table & spinners), sinyal termination capture.
- **Input:** CLI flags (`porta start -c ./porta.yaml`), terminal resize events, `os.Interrupt`.
- **Output:** Rendered UI view, stderr error diagnostics, OS exit codes (`0`, `1`, `2`).
- **Dependencies:** PORTA Core.

#### 2. PORTA Core (Orchestrator)
- **Tanggung Jawab:** Mengkoordinasikan inisialisasi, memvalidasi dependencies antar komponen, mengontrol siklus startup dan teardown.
- **Komunikasi:** In-process Go channels & event dispatching.

#### 3. Config Engine
- **Tanggung Jawab:** Membaca file `porta.yaml`, mengekspansi environment variables (`${VAR}`), memvalidasi schema YAML, mengaplikasikan smart defaults.
- **Input:** File path `porta.yaml` atau raw YAML string.
- **Output:** Immutable `*config.Config` struct.

#### 4. Service Registry
- **Tanggung Jawab:** Menyimpan tabel pemetaan service upstream (nama, target host, target port, prefix route, rewrite rules, TLS target).
- **Input:** Service definitions dari Config Engine.
- **Output:** Lookup function `MatchRoute(path string) (*Service, error)`.

#### 5. Health Checker
- **Tanggung Jawab:** Memverifikasi liveness service target secara pre-flight dan background periodic polling (TCP ping / HTTP probe).
- **Output:** Real-time state per service (`ONLINE`, `UNHEALTHY`, `OFFLINE`).

#### 6. Proxy Manager (Embedded Gateway)
- **Tanggung Jawab:** Menjalankan HTTP/1.1 & HTTP/2 reverse proxy lokal, routing dispatch, WebSocket connection hijacking, header sanitation, error page generation (502/404).
- **Input:** Inbound HTTP requests dari Tunnel Adapter.
- **Output:** Forwarded requests ke upstream services.

#### 7. Tunnel Manager & Providers
- **Tanggung Jawab:** Mengorkestrasi child process provider (e.g. `cloudflared`) atau direct protocol driver, menangani parsing public URL dari provider output, health watchdog, dan auto-reconnect.
- **Output:** Active public URL (e.g. `https://random-id.trycloudflare.com`).

#### 8. Security Engine
- **Tanggung Jawab:** Interceptor middleware untuk Basic Auth (`401 Challenge`), Bearer Token verification (`403 Forbidden`), CIDR IP Whitelist, dan SSRF Loopback protection.
- **Dependencies:** Config Engine.

#### 9. Runtime Manager
- **Tanggung Jawab:** State machine runtime (`INIT`, `STARTING`, `ONLINE`, `SHUTTING_DOWN`), watchdog process supervisor, synchronization barrier (`sync.WaitGroup`), channel cancel context.

#### 10. Logger & Observability
- **Tanggung Jawab:** Structured JSON logging ke disk (`.porta/logs/access.log`), buffered console stream ke TUI.
- **Output:** Colorized live traffic feed.

#### 11. Credential Store
- **Tanggung Jawab:** Mengambil dan menyimpan auth tokens (Cloudflare API token, ngrok authtoken) dari OS Keychain atau file ber-permission ketat (`0600`).

---

## 4. TECHNOLOGY SELECTION & COMPARATIVE TRADE-OFF ANALYSIS

### 4.1 Core Programming Language

| Parameter | Go (Golang) | Rust | Node.js / TypeScript |
| :--- | :--- | :--- | :--- |
| **Kelebihan** | Single static binary (zero runtime dep), standard library networking kelas dunia (`net/http`, `httputil.ReverseProxy`), concurrency goroutines ringan, cross-compilation instan (`GOOS=windows/linux/darwin`). | Zero-cost abstractions, memory safety tanpa GC, performa ekstrem. | Ekosistem JavaScript/NPM sangat besar, rapid prototyping. |
| **Kekurangan** | Runtime GC kecil (overhead < 10MB RAM, tidak masalah untuk CLI). | Waktu kompilasi lambat, kompleksitas async networking & lifetime borrow checker memperpanjang lead time. | Membutuhkan runtime Node.js terpasang atau bundling binary besar (>60MB via pkg/bun), memory footprint tinggi (>50MB). |
| **Cross-Platform** | **Native & Flawless.** Single binary Windows `.exe`, macOS universal binary, Linux static ELF. | Native cross-compile, namun cross-linking toolchain lebih rumit. | Memerlukan platform-specific binary wrapper. |
| **Suitability** | **9.8 / 10** | 8.5 / 10 | 6.0 / 10 |
| **Rekomendasi** | **GO (Golang 1.22+) DIPILIH.** Ideal untuk developer network tools & CLI (standard industri: Docker, Kubernetes, Caddy, Terraform, Ngrok dibangun dengan Go). |

---

### 4.2 Reverse Proxy Strategy

| Parameter | Embedded Go Gateway (`net/http/httputil`) | Caddy (External Binary / Plugin) | Nginx (External Binary) |
| :--- | :--- | :--- | :--- |
| **Kelebihan** | 100% in-process, zero external binary dependency, kontrol dinamis 100% via Go code, native WebSocket upgrade support, memory footprint ultra-rendah (<15MB). | Fitur automatic HTTPS, konfigurasi JSON/Caddyfile modular. | Performa raw C sangat cepat, battle-tested puluhan tahun. |
| **Kekurangan** | Fitur lanjutan harus ditulis sendiri (namun reverse proxy basic + websocket di Go hanya ~150 baris kode). | Mengharuskan download binary Caddy eksternal (~40MB) atau Caddy Go library yang membuat binary PORTA bengkak (>50MB). | Membutuhkan file konfigurasi Nginx statis di disk, spawn child process, sulit dikontrol programmatically di Windows. |
| **Rekomendasi** | **EMBEDDED GO GATEWAY DIPILIH.** Menghilangkan dependency binary pihak ketiga dan memberikan kontrol mutlak atas lifecycle request. |

---

### 4.3 Tunnel Provider Abstraction

| Provider | Authentication | Setup Friction | Speed / Reliability | Biaya / Limit |
| :--- | :--- | :--- | :--- | :--- |
| **Cloudflare Tunnel (Quick)** | **Zero Auth Needed** (TryCloudflare) | **Instant (Zero Config)** | Global Anycast Cloudflare network, sangat cepat, low latency. | **100% Free**, tidak ada limit bandwidth ketat untuk dev use. |
| **Cloudflare Named Tunnel** | Cloudflare Account + Domain | Sedang (Perlu token) | Sangat stabil, custom subdomain permanen. | **Free Tier**. |
| **ngrok** | Wajib Register Akun + Auth Token | Sedang (Perlu `ngrok config`) | Sangat stabil, terkenal. | Free tier memiliki limit bandwidth & session duration. |
| **Tailscale Funnel** | Tailscale Account + Node | Rumit untuk non-Tailscale user | Sangat aman (WireGuard mesh). | Memerlukan instalasi client Tailscale. |
| **Rekomendasi** | **CLOUDFLARE QUICK TUNNEL sebagai Default MVP Provider**, dengan abstraction layer arsitektur `TunnelProvider` interface agar ngrok dan custom provider dapat ditambahkan dengan mudah tanpa refactor core engine. |

---

### 4.4 Local State & Storage Engine

| Tipe Storage | Complexity | Dependency | Suitability untuk PORTA |
| :--- | :--- | :--- | :--- |
| **File-based (`porta.yaml` + JSON log)** | **Sangat Sederhana** | Zero (Standard File I/O) | **10/10.** Sangat cocok untuk CLI stateless/semi-stateful. Konfigurasi deklaratif berbasis YAML dan file log plain text. |
| **SQLite Embedded** | Sedang (CGO / pure Go driver) | Menambah kompleksitas schema & migration | **3/10 (Over-engineering).** PORTA tidak memerlukan query relational database untuk runtime exposure. |
| **Rekomendasi** | **FILE-BASED DIPILIH.** |

---

## 5. TUNNEL ARCHITECTURE & PROVIDER ABSTRACTION

### 5.1 Provider Interface Contract
Arsitektur tunneling menggunakan Go Interface abstraction pattern:

```go
type TunnelProvider interface {
    // Name returns the identifier of the provider (e.g., "cloudflare", "ngrok")
    Name() string
    
    // Start launches the tunnel connection pointing to local gateway address
    Start(ctx context.Context, localGatewayAddr string, options TunnelOptions) (*TunnelSession, error)
    
    // HealthCheck probes if the tunnel connection is still active and healthy
    HealthCheck(ctx context.Context) error
    
    // Stop gracefully closes the tunnel session
    Stop() error
}

type TunnelSession struct {
    PublicURL string
    Provider  string
    StartedAt time.Time
}
```

### 5.2 Provider Execution Strategy
1. **Cloudflare Quick Tunnel Driver (MVP Default):**
   - PORTA mengelola binary `cloudflared` (menggunakan auto-downloader binary embed yang terisolasi di `~/.porta/bin/` atau system PATH).
   - Menjalankan command `cloudflared tunnel --url http://127.0.0.1:<PORT> --no-autoupdate`.
   - Mengintersepsi output `stderr` via regex parser untuk menangkap URL publik yang dialokasikan:
     `Regex: https://[a-zA-Z0-9-]+\.trycloudflare\.com`
   - Melakukan health monitoring pada child process. Jika child process mati tak terduga, runtime memicu restart otomatis dengan *exponential backoff* (1s, 2s, 4s, max 10s).

---

## 6. REVERSE PROXY & EMBEDDED GATEWAY ARCHITECTURE

```text
Inbound Request: GET /api/v1/users
        │
        ▼
┌─────────────────────────────────────────────────────────────┐
│ 1. Host & SSRF Guard (Validates Loopback binding)           │
├─────────────────────────────────────────────────────────────┤
│ 2. Security Middleware (Basic Auth / Token Verification)     │
├─────────────────────────────────────────────────────────────┤
│ 3. Router: Longest Prefix Match Algorithm                   │
│    - Rule 1: `/api` -> upstream http://127.0.0.1:8000       │
│    - Rule 2: `/`    -> upstream http://127.0.0.1:3000       │
│    => Match Found: Rule 1                                   │
├─────────────────────────────────────────────────────────────┤
│ 4. Path Transformation (if strip_path: true)                │
│    `/api/v1/users` => `/v1/users`                           │
├─────────────────────────────────────────────────────────────┤
│ 5. Header Sanitation & Forwarding Injection                 │
│    - Set X-Forwarded-For, X-Forwarded-Proto: https          │
│    - Set X-Forwarded-Host: project.trycloudflare.com        │
├─────────────────────────────────────────────────────────────┤
│ 6. Transport & Upstream Dispatch (Connection Pooling)       │
│    - Supports HTTP/1.1, HTTP/2 Cleartext, WebSocket upgrade │
├─────────────────────────────────────────────────────────────┤
│ 7. Response Streaming & Error Interception (502 Handler)    │
└─────────────────────────────────────────────────────────────┘
```

### Fitur Spesifik Reverse Proxy:
- **WebSocket Upgrade:** Menggunakan Go `httputil.ReverseProxy` standard yang secara native mendukung HTTP connection hijacking dan full-duplex TCP copy stream.
- **SSE (Server-Sent Events) & Chunked Streaming:** Men-disable response buffering pada reverse proxy transport (`FlushInterval: -1`) sehingga data LLM stream, notification stream, atau file upload/download mengalir seketika tanpa latency delay.
- **Keep-Alive Connection Pooling:** Menggunakan `http.Transport` dengan `MaxIdleConns: 100`, `IdleConnTimeout: 90s` untuk meminimalkan handshake latency ke service backend lokal.

---

## 7. SERVICE DISCOVERY & HEALTH CHECKING ENGINE

### 7.1 Lifecycle State Machine Target Service

```text
  ┌─────────────┐
  │   UNKNOWN   │ ──(Initiate Probing)──┐
  └─────────────┘                       │
         ▲                              ▼
         │                     ┌─────────────────┐
         │                     │    STARTING     │ (Within Grace Period: 5s)
         │                     └────────┬────────┘
         │                              │
         │             ┌────────────────┴────────────────┐
         │             ▼ (Probe OK)                      ▼ (Probe Failed > Grace)
         │    ┌─────────────────┐               ┌─────────────────┐
         │    │     ONLINE      │               │     OFFLINE     │
         │    └────────┬────────┘               └────────┬────────┘
         │             │                                 │
         │             ▼ (Consecutive Failures >= 3)     │ (Probe Success)
         │    ┌─────────────────┐                        │
         └────┤    UNHEALTHY    │ ◄──────────────────────┘
              └─────────────────┘
```

### 7.2 Probing Mechanism:
1. **TCP Probing (Default):** Melakukan `net.DialTimeout("tcp", "127.0.0.1:<port>", 1*time.Second)`. Jika socket listening terbuka, service dinyatakan `ONLINE`.
2. **HTTP Probing:** Jika `health_check.path` didefinisikan (e.g. `/healthz`), mengirimkan request `GET http://127.0.0.1:<port>/healthz` dengan timeout 2 detik. Status `200-399` dianggap sehat.

---

## 8. SECURITY ARCHITECTURE & THREAT MODELING

```text
   THREAT MATRIX & MITIGATION

1. Threat: Server-Side Request Forgery (SSRF) to Internal Corporate Network
   Mitigation: Target Host Whitelist restriction. PORTA reverse proxy secara default
               hanya diizinkan me-route request ke IP loopback (127.0.0.1, ::1).
               Setiap konfigurasi yang mengarah ke IP subnet LAN (192.168.x.x, 10.x.x.x)
               atau cloud metadata IP (169.254.169.254) akan ditolak saat config validation.

2. Threat: Unauthenticated Public Scraping / Brute Force
   Mitigation: Security Gatekeeper (Basic Auth RFC 7617 & Bearer Token).
               Semua request publik dicegat sebelum diteruskan ke local app.

3. Threat: Host Header Poisoning
   Mitigation: Proxy menyuntikkan X-Forwarded-Host asli dari edge tunnel dan membersihkan
               Host header sebelum dialirkan ke backend.

4. Threat: Accidental Local Secret Exposure
   Mitigation: PORTA log file dan credential store di-chmod 0600 (read/write owner only).
```

---

## 9. CREDENTIAL MANAGEMENT

1. **Storage Hierarchy:**
   - **Tingkat 1 (Prioritas Tertinggi):** Environment Variables (`PORTA_CLOUDFLARE_TOKEN`, `PORTA_NGROK_TOKEN`, `PORTA_PASSWORD`).
   - **Tingkat 2:** Local Encrypted File Store (`~/.porta/credentials.json`) dengan permission file `0600` (hanya user aktif OS yang dapat membaca).
2. **Zero Plaintext Rule di YAML:** Konfigurasi `porta.yaml` tidak boleh memaksa hardcode secret. YAML parser mendukung sintaks ekspansi variabel environment: `${SECRET_NAME:-default_value}`.

---

## 10. RUNTIME LIFECYCLE & STATE MACHINE

```text
[State: INIT]
     │
     ▼ (Parse porta.yaml & Validate Rules)
[State: VALIDATING]
     │
     ▼ (Pre-flight TCP Health Probe & Port Reservation)
[State: STARTING]
     │
     ▼ (Start Embedded Reverse Proxy on 127.0.0.1:<gateway_port>)
[State: PROXY_READY]
     │
     ▼ (Spawn & Handshake Tunnel Provider)
[State: TUNNEL_CONNECTING]
     │
     ▼ (Public URL Extracted & Verified)
[State: ONLINE]  ◄── (Normal Operation: Proxying & Streaming Logs)
     │
     ▼ (SIGINT / SIGTERM / 'porta stop')
[State: STOPPING]
     │
     ▼ (Teardown Tunnel Child Process -> Close Proxy Listeners -> Flush Logs)
[State: STOPPED]
```

---

## 11. ERROR TAXONOMY & RECOVERY MATRIX

| Error Code | Root Cause | Detection Point | Action / Recovery Policy | User Message Guidance |
| :--- | :--- | :--- | :--- | :--- |
| `ERR_CFG_SYNTAX` | File YAML corrupt / salah indentasi | Config Parser | Fatal Stop (Exit 1) | `✗ Error parsing porta.yaml: line 12: yaml: line 12: did not find expected key` |
| `ERR_PORT_OFFLINE` | Service lokal belum dinyalakan oleh developer | Pre-flight Check | Warning (Continue with 502 placeholder) | `! Warning: Service 'backend' on 127.0.0.1:8000 is not reachable yet. PORTA will auto-route once it comes online.` |
| `ERR_ROUTE_COLLISION`| Dua service memiliki path route yang identik | Config Engine | Fatal Stop (Exit 1) | `✗ Route conflict: Both 'service-a' and 'service-b' map to route '/api'.` |
| `ERR_GATEWAY_BIND` | Port lokal gateway terpakai process lain | Proxy Init | Automatic Retry with random ephemeral port | `i Local gateway port in use. Auto-allocated random port :54123.` |
| `ERR_TUNNEL_DISCONN`| Internet drop / tunnel provider reset | Tunnel Watchdog | Exponential Backoff Reconnection (1s..10s) | `! Tunnel disconnected. Reconnecting in 3s... (Attempt 2/5)` |
| `ERR_TUNNEL_FATAL` | Binary `cloudflared` crash / network block | Tunnel Manager | Fatal Stop (Exit 2) | `✗ Tunnel failed to start: Please check internet connection or firewall.` |
| `ERR_AUTH_REJECTED` | Visitor gagal memasukkan password valid | Security Engine | Return HTTP 401 Unauthorized | Logged as warning in access log: `[401] GET / - Unauthorized client IP` |

---

## 12. CLI SPECIFICATION & COMMAND REFERENCE

### 12.1 Command Matrix

```bash
porta init       # Interactive / auto-detection setup, menghasilkan porta.yaml
porta start      # Menjalankan reverse proxy gateway dan membuka public tunnel
porta stop       # Menghentikan background gateway session
porta status     # Memeriksa status kesehatan service dan URL tunnel yang aktif
porta logs       # Menampilkan live request stream atau membaca log historis
porta config     # Memvalidasi dan menampilkan parsed configuration
porta doctor     # Memeriksa dependensi sistem, jaringan, dan kesiapan environment
```

### 12.2 Command Details

#### A. `porta init`
- **Syntax:** `porta init [--force]`
- **Flags:**
  - `-f, --force`: Overwrite file `porta.yaml` yang sudah ada.
- **Behavior:** Melakukan port scan cepat di range port development umum (`3000`, `5173`, `8000`, `8080`, `9000`), mendeteksi port yang sedang aktif (LISTEN), dan membuat draft file `porta.yaml` otomatis.

#### B. `porta start`
- **Syntax:** `porta start [-c config_path] [--detach]`
- **Flags:**
  - `-c, --config string`: Path ke custom config file (default: `./porta.yaml`).
  - `-d, --detach`: Berjalan di background mode (daemon).
  - `--provider string`: Override tunnel provider (`cloudflare`, `ngrok`).
- **Standard Output Terminal View (TUI):**
```text
  ┌─────────────────────────────────────────────────────────────┐
  │   PORTA v1.0.0 — Local Exposure Engine                     │
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

#### C. `porta doctor`
- **Syntax:** `porta doctor`
- **Output:**
```text
  [✓] OS & Architecture: windows/amd64
  [✓] Loopback Interface: 127.0.0.1 reachable
  [✓] Tunnel Driver: cloudflared embedded ready
  [✓] Outbound Internet: DNS resolution & HTTPS connection OK
  [✓] Configuration: ./porta.yaml is valid
```

---

## 13. CROSS-PLATFORM SYSTEM SPECIFICATION

| Area | Windows (Win 10/11) | macOS (Intel / Apple Silicon) | Linux (Ubuntu/Debian/Arch) |
| :--- | :--- | :--- | :--- |
| **Binary Output** | `porta.exe` (PE executable) | `porta` (Universal Mach-O) | `porta` (Static ELF binary) |
| **Path Separators** | `filepath.ToSlash` normalization | Standard POSIX `/` | Standard POSIX `/` |
| **Process Management** | Windows Job Objects untuk memastikan child process (`cloudflared`) ter-kill saat main process mati. | POSIX `kill(pid, SIGTERM)` & Process Groups (`Setpgid`). | POSIX `kill(pid, SIGTERM)` & Process Groups. |
| **Signal Handling** | `os.Interrupt` (Ctrl+C), `os.Kill`. | `SIGINT`, `SIGTERM`, `SIGHUP` (config reload). | `SIGINT`, `SIGTERM`, `SIGHUP`. |
| **Credential Store** | Windows Credential Manager | macOS Keychain Services | Secret Service API / File Permissions `0600` |
| **Localhost Resolv**| Force bind to `127.0.0.1` explicitly (menghindari delay dual-stack IPv6 `::1` di Windows). | `127.0.0.1` and `::1`. | `127.0.0.1` and `::1`. |

---

## 14. NETWORKING & PERFORMANCE BASELINE

### 14.1 Networking Protocols
- **HTTP/1.1 & HTTP/2 Cleartext (h2c):** Full support dengan multiplexing.
- **WebSocket (RFC 6455):** Native connection hijacking dengan zero message alteration.
- **Streaming & SSE:** Chunked Transfer-Encoding dengan zero buffer latency (`FlushInterval: -1`).
- **Dual-Stack Network Handling:** Resolusi default eksplisit `127.0.0.1` untuk mencegah latency delay 1000ms pada Windows IPv6 fallback.

### 14.2 Performance Baseline Metrics

| Metric | Target Baseline | Status / Verification |
| :--- | :--- | :--- |
| **Binary Cold Startup Time** | `< 1.2 detik` (hingga tunnel live) | Achievable via Go + Cloudflare Quick Tunnel |
| **Memory Footprint (Idle)** | `< 25 MB RAM` | Go Runtime + Embedded Gateway |
| **Memory Footprint (Under Load)**| `< 50 MB RAM` | Max 200 concurrent connections |
| **CPU Usage (Idle)** | `< 0.5% CPU` | Verified Event Loop |
| **Gateway Routing Latency Overhead**| `< 3 ms` | Local memory lookup table & streaming proxy |
| **Concurrent Active Streams** | `1,000+ concurrent requests` | Verified Go `net/http` capability |

---

## 15. NON-FUNCTIONAL REQUIREMENTS (NFR)

1. **Security:** Zero plaintext secrets in version control; local loopback isolation prevents SSRF attacks; secure TLS 1.3 encryption on public tunnel edge.
2. **Reliability:** Self-healing auto-reconnect on tunnel network drops; 502 graceful error handler prevents gateway crash when backend restarts.
3. **Portability:** Single standalone binary with zero external runtime dependencies (no Node.js, Python, or Docker required on developer machine).
4. **Usability:** 1-command startup (`porta start`); zero configuration needed for single service applications; actionable and colorized error guidance.
5. **Observability:** Real-time TUI log stream; structured JSON access logs for automated inspection.

---

## 16. ARCHITECTURAL DECISION RECORDS (ADR)

### ADR-001: Core Programming Language Selection
- **Status:** APPROVED
- **Decision:** Menggunakan **Go (Golang 1.22+)** sebagai bahasa implementasi utama PORTA.
- **Rasional:** Go menghasilkan single standalone binary tanpa runtime dependency, memiliki standard library `net/http/httputil` yang sangat matang untuk reverse proxying berkecepatan tinggi, konsumsi RAM kecil (<25MB), serta kompilasi cross-platform instan untuk Windows, macOS, dan Linux.
- **Alternatif Dipertimbangkan:** Rust (terlalu lambat waktu kompilasi dan kompleks untuk MVP), Node.js (memerlukan runtime eksternal dan footprint besar).

---

### ADR-002: Embedded Reverse Proxy Gateway Engine
- **Status:** APPROVED
- **Decision:** Membangun reverse proxy menggunakan **Embedded Go `net/http/httputil`** di dalam binary PORTA, bukan membungkus Caddy atau Nginx eksternal.
- **Rasional:** Menghilangkan keharusan mendistribusikan binary proxy terpisah (>40MB), memungkinkan konfigurasi dynamic in-memory routing tanpa file generation / disk I/O, serta kontrol mutlak terhadap lifecycle error handling dan WebSocket hijacking.
- **Alternatif Dipertimbangkan:** Caddy (terlalu besar untuk embedded library), Nginx (sulit diatur secara dinamis di Windows).

---

### ADR-003: Default Tunnel Provider & Driver Abstraction
- **Status:** APPROVED
- **Decision:** Mengadopsi **Cloudflare Quick Tunnel (TryCloudflare)** sebagai provider default MVP dan mengisolasi provider logic di balik generic Go interface `TunnelProvider`.
- **Rasional:** Cloudflare Quick Tunnel memungkinkan developer langsung mengekspos port tanpa registrasi akun, tanpa login, gratis, dan berkecepatan tinggi di jaringan edge Anycast global.
- **Alternatif Dipertimbangkan:** ngrok (mengharuskan pendaftaran akun dan token pada first-run, free tier terbatas).

---

### ADR-004: Configuration Format Specification
- **Status:** APPROVED
- **Decision:** Menggunakan format **YAML (`porta.yaml`)** dengan support environment variable expansion.
- **Rasional:** YAML adalah standar industri paling ramah manusia untuk konfigurasi developer tools (Docker Compose, Kubernetes, GitHub Actions).
- **Alternatif Dipertimbangkan:** JSON (kurang nyaman diedit manual, tidak mendukung comments), TOML (kurang fleksibel untuk nested service maps).

---

### ADR-005: Process Management Scope for MVP
- **Status:** APPROVED
- **Decision:** PORTA **TIDAK** menjalankan/meng-compile proses aplikasi developer (`npm run dev` / `go run`) pada fase MVP, melainkan fokus mendeteksi dan mengekspos service yang sudah aktif.
- **Rasional:** Menjaga kesederhanaan arsitektur dan membiarkan developer tetap melihat logs/debug trace native dari compiler mereka masing-masing. Fitur process orchestration dimasukkan ke roadmap v1.1.

---

## 17. OPEN QUESTIONS & ARCHITECTURAL IMPACT

| Open Question | Pilihan Solusi | Dampak Arsitektural | Status Keputusan |
| :--- | :--- | :--- | :--- |
| **1. Apakah public URL harus statis atau ephemeral di MVP?** | Ephemeral random URL (`*.trycloudflare.com`) untuk default, static subdomain untuk user yang menyertakan Cloudflare/Ngrok token. | Menjaga first-run zero-friction tanpa registrasi akun, namun tetap membuka fleksibilitas untuk advanced user. | **RESOLVED (Hybrid Model)** |
| **2. Apakah PORTA membutuhkan Cloud Control Plane sendiri?** | **TIDAK untuk MVP.** PORTA 100% client-side tool. | Mengurangi biaya operasional server backend ke $0 pada fase awal dan menjamin privasi data developer 100% lokal. | **RESOLVED (Local-first)** |
| **3. Bagaimana menangani SPA (Single Page App) Client-side Routing?** | Gateway meneruskan 404 dari frontend service langsung ke browser agar frontend router (React Router) menangani rendering. | Proxy tidak perlu mengimplementasikan file server fallback HTML jika frontend sudah dijalankan via dev server (Vite/Webpack). | **RESOLVED** |

---

## 18. TRACEABILITY MATRIX (FRM ↔ SRS)

| FRM Requirement ID | SRS Component | Technical Implementation Mechanism |
| :--- | :--- | :--- |
| **PM-001** (Workspace Init) | PORTA Core / CLI | `internal/config/init.go`: Port scanner & template generator |
| **PM-002** (Project Isolation) | Config Engine | `internal/config/validator.go`: Namespace sanitization regex |
| **CFG-001** (YAML Parsing) | Config Engine | `gopkg.in/yaml.v3` + `os.ExpandEnv` schema unmarshaler |
| **CFG-002** (Route Collision) | Service Registry | Trie / Radix path collision verification algorithm |
| **SRV-001** (Service Reg) | Service Registry | `internal/registry/registry.go`: In-memory service map table |
| **SRV-002** (Path Stripping) | Proxy Manager | `httputil.ReverseProxy.Director` URL path rewriting |
| **HC-001** (Health Probing) | Health Checker | Goroutine worker pool with `net.Dialer` & `http.Client` |
| **HC-002** (Graceful Degradation)| Proxy Manager | Custom `ErrorHandler` returning 502 HTML/JSON banner |
| **PROXY-001** (Gateway Listener) | Proxy Manager | `net.Listen("tcp", "127.0.0.1:0")` ephemeral bind |
| **PROXY-002** (Prefix Matching) | Proxy Manager | Longest Prefix Matching HTTP router middleware |
| **PROXY-003** (WebSocket Upgrade)| Proxy Manager | In-process transparent HTTP connection hijacking |
| **PROXY-004** (Header Forwarding)| Proxy Manager | `X-Forwarded-For/Proto/Host` header mutation middleware |
| **TUNNEL-001** (Provider Abstraction)| Tunnel Manager | `internal/tunnel/provider.go`: Go interface contract |
| **TUNNEL-002** (Quick Tunnel) | Tunnel Manager | `internal/tunnel/cloudflare.go`: Process supervisor + regex parser |
| **TUNNEL-003** (Auto Reconnect) | Tunnel Manager | Watchdog exponential backoff reconnect loop |
| **SEC-001** (Basic Auth) | Security Engine | `crypto/subtle.ConstantTimeCompare` HTTP basic auth interceptor |
| **SEC-002** (Bearer Token) | Security Engine | Bearer header & query parameter token interceptor |
| **SEC-003** (SSRF Isolation) | Security Engine | Private IP / Loopback CIDR validator guard |
| **CLI-001** (Command Dispatch) | PORTA CLI | `github.com/spf13/cobra` command routing |
| **CLI-002** (Interactive TUI) | PORTA CLI | ANSI escape sequence live table renderer |
| **MON-001** (Access Logging) | Logger | Structured JSON / stdout pipeline with zap/zerolog |
| **MON-002** (Log Inspection) | Logger | File tailing reader (`porta logs -f`) |

---
*End of System Requirement Specification.*
