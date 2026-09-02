# FUNCTIONAL REQUIREMENT MAP (FRM)
**Project:** PORTA (Local Multi-Service Application Public Exposure Platform)  
**Document Version:** 1.0.0-PROD-SPEC  
**Author:** Lead Product Architect + System Architect  
**Status:** APPROVED FOR SPECIFICATION BASELINE  

---

## 1. PRODUCT SCOPE

### 1.1 Core Mission Statement
> **"Expose your local multi-service application to the world with one command."**

PORTA adalah developer utility CLI and runtime engine yang mengorkestrasi *local reverse proxying*, *traffic routing*, dan *secure tunneling* untuk mengekspos aplikasi multi-service (frontend, backend API, WebSocket, microservices) yang berjalan pada `localhost` ke sebuah Public HTTPS URL yang aman dan dapat diakses publik secara instan tanpa memerlukan IP publik statis, konfigurasi router/NAT forwarding, maupun deployment cloud.

---

### 1.2 In-Scope vs Out-of-Scope Matrix

| Domain | In-Scope (Apa yang Dilakukan PORTA) | Out-of-Scope (Apa yang TIDAK Dilakukan PORTA) |
| :--- | :--- | :--- |
| **Exposure** | Menghubungkan service `localhost:<port>` ke public tunnel URL. | Hosting kode aplikasi atau database di cloud server PORTA. |
| **Routing** | Path-based routing (`/` -> :3000, `/api` -> :8000), Host/Subdomain routing, and header rewrites. | Web Application Firewall (WAF) level enterprise (DDoS L7 mitigasi kompleks di level lokal). |
| **Process Control** | Memeriksa ketersediaan (health check) service lokal yang sudah dijalankan developer. | Berfungsi sebagai full Process Manager seperti PM2 / Docker Compose (tidak meng-compile, start npm/python script secara default pada MVP). |
| **Security** | Basic Auth, Bearer Token gatekeeper, IP Whitelisting pada public gateway, TLS termination di tunnel edge. | Identity Provider (IdP), SAML/SSO Enterprise auth server. |
| **Tunneling** | Menjalankan dan mengelola lifecycle tunnel provider (e.g. Cloudflare Quick Tunnel / Named Tunnel, ngrok) via unified abstraction. | Menggantikan ISP atau membuat protokol tunneling proprietary VPN dari nol. |
| **Networking** | WebSocket forwarding, SSE, chunked streaming, HTTP/1.1 & HTTP/2 reverse proxying. | Raw TCP/UDP tunnel untuk non-HTTP protocol (e.g. raw MySQL 3306, SSH) pada fase MVP. |

---

### 1.3 Target Persona & User Profiles

1. **Full-stack Developer:**
   - *Karakteristik:* Menjalankan frontend Next.js/Vite di `:3000` dan backend Go/Node/Django di `:8000`.
   - *Pain Point:* Kesulitan membagikan progress staging lokal ke tim/klien tanpa setup CORS yang rumit, deployment manual ke Vercel/Fly.io, atau tunneling 2 URL terpisah via ngrok gratis.
2. **Frontend / Mobile Engineer:**
   - *Karakteristik:* Membutuhkan live webhook testing dari pihak ketiga (Stripe, Midtrans, GitHub) ke localhost backend, atau testing aplikasi mobile fisik yang menembak API di laptop.
   - *Pain Point:* Webhook provider membutuhkan public HTTPS URL yang stabil dan valid SSL.
3. **QA Engineer / Product Manager:**
   - *Karakteristik:* Menguji fitur baru langsung pada branch lokal developer sebelum PR dimerge.
   - *Pain Point:* Menunggu CI/CD pipeline staging lambat; ingin instant preview dari mesin engineer.

---

### 1.4 Use Cases

#### Primary Use Cases
- **UC-01: Multi-Service Unified Exposure:** Mengekspos frontend (`localhost:3000`) dan backend (`localhost:8000`) melalui 1 domain publik HTTPS (`https://project.porta.dev`) dengan routing path `/` dan `/api` tanpa CORS error.
- **UC-02: Instant Client / Peer Review:** Membuka akses publik ber-password ke aplikasi lokal dalam hitungan detik untuk demo ke stakeholder.
- **UC-03: Webhook & Third-Party Integration Testing:** Menerima payload webhook HTTPS dari payment gateway ke API backend lokal secara real-time.

#### Secondary Use Cases
- **UC-04: Mobile Device Testing:** Membuka akses URL publik HTTPS dari smartphone fisik di jaringan seluler/luar WiFi kantor untuk menguji responsivitas aplikasi.
- **UC-05: WebSocket / Realtime Streaming Inspection:** Menguji koneksi socket.io / GraphQL subscriptions / SSE melewati reverse proxy gateway.

---

## 2. USER JOURNEY & LIFECYCLE FLOW

```text
[ Developer Machine ]
        │
        ▼
   ┌─────────┐
   │ Install │  (Single binary via curl/brew/scoop)
   └────┬────┘
        ▼
   ┌────────────┐
   │ porta init │  (Scans directory, detects running ports, creates porta.yaml)
   └────┬───────┘
        ▼
   ┌───────────────────────┐
   │ Project/Service Config│  (Declarative YAML: frontend, backend, routes, auth)
   └────┬──────────────────┘
        ▼
   ┌────────────┐
   │ Validation │  (Pre-flight check: syntax, port conflict, target reachability)
   └────┬───────┘
        ▼
   ┌─────────────┐
   │ porta start │  (Spawns PORTA Gateway Engine)
   └────┬────────┘
        ├──────────────────────────────────────────────────────┐
        ▼                                                      ▼
   ┌──────────────────────┐                         ┌──────────────────────┐
   │ Service Health Check │ (TCP/HTTP probes)       │ Reverse Proxy Boot   │ (Binds local gateway)
   └────┬─────────────────┘                         └────┬─────────────────┘
        │                                                      │
        └──────────────────────────┬───────────────────────────┘
                                   ▼
                        ┌──────────────────────┐
                        │ Tunnel Establishment │ (Cloudflare / Provider adapter)
                        └──────────┬───────────┘
                                   ▼
                        ┌──────────────────────┐
                        │ Security Gatekeeper  │ (Enforces Auth / Token / Headers)
                        └──────────┬───────────┘
                                   ▼
                        ┌──────────────────────┐
                        │ Ready: Public HTTPS  │ (Outputs https://<slug>.porta.dev)
                        └──────────┬───────────┘
                                   ▼
                        ┌──────────────────────┐
                        │ Active Observability │ (Live traffic log & status monitor)
                        └──────────┬───────────┘
                                   ▼
                        ┌──────────────────────┐
                        │ porta stop / SIGINT  │ (Graceful teardown proxy & tunnel)
                        └──────────────────────┘
```

### Tahapan Detail User Journey:
1. **Install:** Developer memasang binary PORTA (zero external dependency).
2. **`porta init`:** PORTA mendeteksi project structure dan port aktif lokal, menghasilkan template file `porta.yaml` yang minimalis dan teruji.
3. **Configuration Review:** Developer menyesuaikan routing path (`/` -> 3000, `/api` -> 8000) dan security mode (`public` / `protected`).
4. **Validation (Pre-flight):** Sebelum membuka jaringan publik, PORTA memverifikasi integritas konfigurasi, tidak adanya route collision, dan ketersediaan target service.
5. **`porta start`:**
   - Inisialisasi **Embedded Reverse Proxy** pada local loopback.
   - Polling **Health Checker** untuk menandai service yang `ONLINE`.
   - Inisialisasi **Tunnel Engine** (e.g. Cloudflare Quick/Named Tunnel).
   - Inisialisasi **Security Gatekeeper** (memvalidasi auth header / basic credentials).
6. **Public Exposure & Testing:** Terminal menampilkan status table, active routes, dan Public HTTPS URL. Traffic masuk diteruskan ke target service dengan streaming log.
7. **`porta stop` / Exit:** Sinyal `SIGINT` (Ctrl+C) atau command `porta stop` memutus tunnel session secara bersih, menghentikan proxy, dan melepaskan port lokal tanpa meninggalkan orphan processes.

---

## 3. FUNCTIONAL MODULES ANALYSIS & EVALUATION

| Modul | Status | Evaluasi & Rasional Arsitektural |
| :--- | :--- | :--- |
| **Project Manager** | **Wajib MVP** | Mengidentifikasi context project aktif berdasarkan working directory, membaca metadata, dan mengisolasi environment execution. |
| **Configuration Manager** | **Wajib MVP** | Membaca, memvalidasi schema, mengaplikasikan default values, dan mem-parse `porta.yaml` dengan strict error reporting. |
| **Service Registry** | **Wajib MVP** | In-memory registry yang memetakan service ID, endpoint target (`host:port`), protocol, dan path routing rules. |
| **Health Checker** | **Wajib MVP** | Memonitor status liveness service target (`ONLINE`, `DEGRADED`, `OFFLINE`) via TCP socket / HTTP probe berkala agar gateway tidak meneruskan request ke service mati. |
| **Reverse Proxy Manager** | **Wajib MVP** | Inti traffic routing lokal. Menggabungkan multiple upstream services ke dalam single local gateway listener dengan support WebSocket, header forwarding (`X-Forwarded-*`), dan path stripping/rewriting. |
| **Tunnel Manager** | **Wajib MVP** | Mengelola adapter penyedia tunnel (Cloudflare Tunnel, ngrok), menangani auth, lifecycle tunnel process/connection, dan auto-reconnect. |
| **Public URL Manager** | **Wajib MVP** | Menangkap assigned public domain dari tunnel provider, memvalidasi SSL handshake, dan mempresentasikan URL final ke developer. |
| **Security / Access Control** | **Wajib MVP** | Mencegah akses publik yang tidak diinginkan dengan fitur Basic Auth (`user:pass`), Bearer Token, dan Localhost isolation filter. |
| **Runtime Manager** | **Wajib MVP** | State machine utama yang mengorkestrasi urutan startup, health polling, signal handling (`SIGINT`, `SIGTERM`), dan graceful shutdown. |
| **Logging & Observability** | **Wajib MVP** | Menampilkan live streaming request logs (Method, Path, Status Code, Latency, Upstream Target) dan command `porta logs`. |
| **CLI Engine** | **Wajib MVP** | Antarmuka pengguna berbasis terminal yang interaktif, informatif, dan mendukung formatted table outputs serta non-zero exit codes. |
| **Process Manager (Orchestrator)** | **Opsional MVP (Target v1.1)** | Menjalankan script npm/python developer (`npm run dev`) secara otomatis. **Rasional:** Pada MVP, developer lebih suka mengontrol proses server mereka sendiri (melihat stack trace langsung). PORTA fokus mengekspos service yang sudah berjalan. |
| **Credential Manager** | **Wajib MVP (Basic)** | Menyimpan auth token provider tunnel secara aman (OS Keychain / restricted file `~/.porta/credentials`). |
| **Environment Manager** | **Tidak Diperlukan MVP** | Menyuntikkan `.env` ke dalam app developer. **Rasional:** Merupakan tanggung jawab process runner lokal / dotenv framework aplikasi developer. |
| **Service Discovery (Auto)** | **Opsional MVP (Basic)** | Heuristic port scanning saat `porta init` untuk auto-populate `porta.yaml`. Full dynamic runtime discovery ditunda ke v2.0. |
| **Dashboard / Web UI** | **Future (Post-MVP)** | Web-based inspector UI (seperti ngrok web inspector). CLI console output sudah sangat mencukupi untuk MVP. |

---

## 4. FUNCTIONAL REQUIREMENTS SPECIFICATION

```text
Requirement ID Taxonomy:
PM-xxx     : Project Management
CFG-xxx    : Configuration Engine
SRV-xxx    : Service Registry
HC-xxx     : Health Checking
PROXY-xxx  : Reverse Proxy Gateway
TUNNEL-xxx : Tunnel Provider Orchestration
SEC-xxx    : Security & Access Control
CLI-xxx    : Command Line Interface
MON-xxx    : Logging & Runtime Observability
```

### 4.1 Project Management (PM)
* **PM-001: Project Workspace Initialization**
  - **Actor:** Developer
  - **Input:** Directory context saat ini.
  - **Processing:** Memeriksa keberadaan file `porta.yaml`. Jika belum ada, lakukan auto-detection port dan buat file konfigurasi baru.
  - **Output:** File `porta.yaml` terinisialisasi.
  - **Error:** Write permission denied pada direktori.
  - **Priority:** High | **Scope:** MVP
* **PM-002: Project Isolation & Naming**
  - **Actor:** Developer
  - **Input:** Parameter `project.name` dari config.
  - **Processing:** Memastikan nama project valid (alphanumeric + hyphen) untuk penamaan session tunnel dan log context.
  - **Output:** Project ID tervalidasi.
  - **Error:** Invalid character format / empty project name.
  - **Priority:** Critical | **Scope:** MVP

---

### 4.2 Configuration Engine (CFG)
* **CFG-001: Declarative YAML Parsing & Validation**
  - **Actor:** PORTA Runtime
  - **Input:** File `porta.yaml`.
  - **Processing:** Parse file YAML, validasi tipe data dan schema constraints, inject default values untuk parameter opsional.
  - **Output:** Parsed in-memory configuration object.
  - **Error:** Malformed YAML syntax, missing mandatory fields (e.g. `services.<name>.port`).
  - **Priority:** Critical | **Scope:** MVP
* **CFG-002: Route Collision Detection**
  - **Actor:** Configuration Engine
  - **Input:** List of routes dari semua registered services.
  - **Processing:** Memeriksa apakah ada duplikasi exact path route yang sama tanpa pembeda prefix.
  - **Output:** Validation status OK.
  - **Error:** Ambiguous route definition (e.g. Service A & Service B both binding `/api`).
  - **Priority:** Critical | **Scope:** MVP

---

### 4.3 Service Registry (SRV)
* **SRV-001: Multi-Service Registration**
  - **Actor:** Developer
  - **Input:** Map services (name, host, port, route, websocket flag).
  - **Processing:** Mendaftarkan setiap target service ke routing table internal. Resolusi default `host=127.0.0.1` jika tidak didefinisikan.
  - **Output:** In-memory upstream routing map.
  - **Error:** Port number out of range (`1-65535`), invalid route prefix syntax.
  - **Priority:** Critical | **Scope:** MVP
* **SRV-002: Path Stripping / Rewriting Definition**
  - **Actor:** Developer
  - **Input:** Parameter `strip_path: true|false` pada service config.
  - **Processing:** Jika `true`, gateway memotong prefix route sebelum request diteruskan ke upstream (e.g. `/api/v1/users` -> diteruskan ke upstream sebagai `/v1/users`).
  - **Output:** Adjusted request URI pada proxy forwarder.
  - **Error:** Invalid regex / rewrite definition.
  - **Priority:** Medium | **Scope:** MVP

---

### 4.4 Health Checking (HC)
* **HC-001: Pre-flight & Periodic Target Probing**
  - **Actor:** Health Checker Engine
  - **Input:** Service upstream `host:port` dan path probe.
  - **Processing:** Melakukan TCP connection check atau HTTP GET probe pada interval reguler (default: 5 detik).
  - **Output:** Status enum (`ONLINE`, `UNHEALTHY`, `OFFLINE`).
  - **Error:** Target service tidak merespon / connection refused.
  - **Priority:** High | **Scope:** MVP
* **HC-002: Graceful Degradation Handling**
  - **Actor:** Health Checker Engine
  - **Input:** State `OFFLINE` pada salah satu service.
  - **Processing:** Menandai service sebagai unavailable. Jika request masuk ke route service tersebut, gateway mengembalikan custom error page `502 Bad Gateway (PORTA: Upstream Service Offline)` tanpa mematikan seluruh tunnel.
  - **Output:** 502 HTTP Response bersahabat.
  - **Error:** None.
  - **Priority:** High | **Scope:** MVP

---

### 4.5 Reverse Proxy Gateway (PROXY)
* **PROXY-001: Unified Local Gateway Listener**
  - **Actor:** Reverse Proxy Engine
  - **Input:** In-memory routing table dan random available loopback port.
  - **Processing:** Membuka HTTP/1.1 & HTTP/2 reverse proxy server pada loopback lokal (`127.0.0.1:<random-port>`).
  - **Output:** Active local gateway endpoint.
  - **Error:** Gagal mengalokasikan port loopback.
  - **Priority:** Critical | **Scope:** MVP
* **PROXY-002: Path-Based Dispatching & Longest Prefix Matching**
  - **Actor:** Inbound HTTP Request
  - **Input:** HTTP Request URI (e.g. `GET /api/v1/orders`).
  - **Processing:** Evaluasi path menggunakan algoritma *Longest Prefix Match* (e.g. `/api` cocok ke backend `:8000`, `/` cocok ke frontend `:3000`).
  - **Output:** Request di-proxy ke upstream yang sesuai.
  - **Error:** No matching route found (returns 404 PORTA Not Found).
  - **Priority:** Critical | **Scope:** MVP
* **PROXY-003: WebSocket & Full Duplex Upgrade**
  - **Actor:** Client Request
  - **Input:** Request dengan header `Upgrade: websocket` dan `Connection: Upgrade`.
  - **Processing:** Gateway melakukan HTTP connection hijacking dan membidireksionalkan raw TCP stream antara client dan upstream service.
  - **Output:** Transparent WebSocket streaming.
  - **Error:** Upstream service menolak upgrade request.
  - **Priority:** Critical | **Scope:** MVP
* **PROXY-004: Standard Header Rewriting & Forwarding**
  - **Actor:** Reverse Proxy Engine
  - **Input:** Inbound HTTP Headers.
  - **Processing:** Menambahkan/memodifikasi header standar: `X-Forwarded-For`, `X-Forwarded-Proto` (`https`), `X-Forwarded-Host`, `X-Real-IP`, serta rewrite `Host` header sesuai target upstream jika dikonfigurasi.
  - **Output:** Sanitized outbound HTTP request ke upstream.
  - **Error:** Header overflow / parsing failure.
  - **Priority:** Critical | **Scope:** MVP

---

### 4.6 Tunnel Provider Orchestration (TUNNEL)
* **TUNNEL-001: Tunnel Provider Abstraction Layer**
  - **Actor:** Tunnel Manager
  - **Input:** Provider type (`cloudflare`, `ngrok`).
  - **Processing:** Menginisialisasi driver provider yang sesuai via generic interface (`Start()`, `Stop()`, `GetPublicURL()`, `Status()`).
  - **Output:** Active secure tunnel connection.
  - **Error:** Unsupported provider type / missing driver binary.
  - **Priority:** Critical | **Scope:** MVP
* **TUNNEL-002: Ephemeral Quick Tunnel Provisioning**
  - **Actor:** Tunnel Manager (Default Mode)
  - **Input:** Local proxy gateway port.
  - **Processing:** Membuka secure TLS tunnel tanpa mengharuskan developer login / mendaftar akun (e.g. Cloudflare Quick Tunnel / TryCloudflare).
  - **Output:** Generated random public URL (`https://<random-hash>.trycloudflare.com` / `https://<slug>.porta.dev`).
  - **Error:** Rate limit provider / network blocked.
  - **Priority:** Critical | **Scope:** MVP
* **TUNNEL-003: Auto-Reconnect & Connection Health Watchdog**
  - **Actor:** Tunnel Manager
  - **Input:** Tunnel connection status event.
  - **Processing:** Jika koneksi tunnel terputus (jaringan drop), lakukan exponential backoff reconnection tanpa mematikan local reverse proxy.
  - **Output:** Re-established tunnel session.
  - **Error:** Max retries exceeded (exit with fatal error).
  - **Priority:** High | **Scope:** MVP

---

### 4.7 Security & Access Control (SEC)
* **SEC-001: HTTP Basic Authentication Gatekeeper**
  - **Actor:** Unauthenticated Public Visitor
  - **Input:** Request HTTP tanpa header `Authorization` yang valid saat security mode aktif (`protected`).
  - **Processing:** Gateway mengintersepsi request dan mengembalikan response `401 Unauthorized` dengan header `WWW-Authenticate: Basic realm="PORTA Protected Environment"`.
  - **Output:** Challenge prompt di browser visitor.
  - **Error:** Invalid username/password combination.
  - **Priority:** High | **Scope:** MVP
* **SEC-002: Bearer Token Authorization Gatekeeper**
  - **Actor:** API Client / Webhook Caller
  - **Input:** Header `Authorization: Bearer <token>` atau query parameter `?porta_token=<token>`.
  - **Processing:** Gateway mencocokkan token dengan konfigurasi. Jika tidak cocok, request ditolak dengan status `403 Forbidden`.
  - **Output:** Forward request jika token valid.
  - **Error:** Invalid or expired token.
  - **Priority:** High | **Scope:** MVP
* **SEC-003: Localhost Loopback Isolation Enforcement**
  - **Actor:** System Security Monitor
  - **Input:** Target service definition.
  - **Processing:** Memastikan gateway hanya me-route traffic ke IP loopback (`127.0.0.1`, `localhost`, `::1`) atau interface private LAN yang diizinkan secara eksplisit, mencegah eksploitasi Server-Side Request Forgery (SSRF) ke jaringan internal sensitif.
  - **Output:** Blocked request jika target mengarah ke IP terlarang (e.g. cloud metadata `169.254.169.254`).
  - **Error:** Security policy violation.
  - **Priority:** Critical | **Scope:** MVP

---

### 4.8 Command Line Interface (CLI)
* **CLI-001: Command Dispatching & TTY Detection**
  - **Actor:** Developer
  - **Input:** Command arguments (`porta init`, `porta start`, `porta stop`, `porta status`, `porta logs`, `porta config`).
  - **Processing:** Parse flags, jalankan handler, dan sesuaikan formatting output (Rich TTY styling vs Plain text untuk non-interactive scripts).
  - **Output:** Rendered terminal output.
  - **Error:** Unknown command / invalid flag.
  - **Priority:** Critical | **Scope:** MVP
* **CLI-002: Real-time Interactive Status TUI**
  - **Actor:** Developer
  - **Input:** `porta start` execution state.
  - **Processing:** Menampilkan formatted live table status di terminal: Project Name, Public URL, Route mapping list, Latency, dan Status Service Target.
  - **Output:** Live interactive summary.
  - **Error:** Terminal resize / unsupported ANSI escape codes.
  - **Priority:** High | **Scope:** MVP

---

### 4.9 Logging & Observability (MON)
* **MON-001: Structured Request Access Logging**
  - **Actor:** Inbound Public Request
  - **Input:** HTTP Request & Response metadata.
  - **Processing:** Format log line: `[TIMESTAMP] [STATUS] METHOD /path -> upstream_target (duration_ms) [CLIENT_IP]`.
  - **Output:** Output stream ke stdout dan file rotating log di `.porta/logs/access.log`.
  - **Error:** Disk full / write error.
  - **Priority:** High | **Scope:** MVP
* **MON-002: Filtered Log Inspection (`porta logs`)**
  - **Actor:** Developer
  - **Input:** Flag `--service=<name>`, `--level=error`, `--follow/-f`.
  - **Processing:** Membaca file log runtime dan melakukan filtering real-time sesuai kriteria.
  - **Output:** Streamed filtered log.
  - **Error:** Log file not found / unreadable.
  - **Priority:** Medium | **Scope:** MVP

---

## 5. PARAMETER ANALYSIS & OPTIMIZATION MATRIX

Tabel analisis parameter mengevaluasi setiap atribut konfigurasi untuk memastikan **zero redundancy**, pemanfaatan **smart defaults**, dan meminimalkan beban kognitif developer.

| Parameter Key | Tipe Data | Req / Opt | Default Value | Valid Values | Validation Rule | Dependency | Modul Konsumen | Evaluasi Redundansi & Optimasi |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `project.name` | `string` | **Req** | Auto-detected dari folder name | Regex `^[a-z0-9-_]+$` | Max 40 chars, lowercase alphanumeric & hyphen | None | Project Manager, Tunnel Manager | **Esensial.** Menentukan namespace session dan identitas tunnel. |
| `project.environment`| `string` | Opt | `development` | `development`, `testing`, `staging` | String enum | None | Project Manager | **Redundant untuk Core MVP.** Tidak mengubah routing logic. Disederhanakan menjadi metadata opsional. |
| `services.<id>.port` | `integer` | **Req** | None | `1 - 65535` | Integer > 0 and <= 65535 | None | Service Registry, Proxy Manager, Health Check | **Esensial.** Parameter inti penentu target upstream. |
| `services.<id>.host` | `string` | Opt | `127.0.0.1` | Valid IPv4 / hostname | Non-empty string | None | Service Registry, Proxy Manager | **Bisa di-defaultkan.** 99% developer menembak `localhost`. Default `127.0.0.1`. Tidak wajib ditulis di YAML. |
| `services.<id>.route`| `string` | Opt | `/` jika single service, `/<id>` jika multi | Path string (e.g. `/`, `/api`) | Must start with `/` | None | Proxy Manager | **Esensial.** Penentu path routing pada reverse proxy gateway. |
| `services.<id>.strip_path` | `boolean` | Opt | `false` | `true`, `false` | Boolean | `route` != `/` | Proxy Manager | **Esensial.** Sangat dibutuhkan ketika backend API tidak memiliki prefix `/api` di router internalnya. |
| `services.<id>.websocket` | `boolean` | Opt | `true` (auto-detect via header) | `true`, `false` | Boolean | None | Proxy Manager | **Bisa di-auto-detect.** Gateway dapat mendeteksi header `Upgrade: websocket` secara dinamis. Flag ini hanya sebagai manual override jika perlu disable. |
| `services.<id>.type` | `string` | Opt | `auto` | `web`, `api`, `static`, `auto` | String enum | None | None | **REDUNDANT.** Tidak mempengaruhi routing reverse proxy (keduanya hanyalah HTTP/TCP target). **Dieliminasi dari MVP schema**. |
| `services.<id>.health_check.path` | `string` | Opt | `""` (TCP ping if empty) | Valid URI path | Starts with `/` | `port` | Health Checker | **Esensial.** Jika kosong, gunakan TCP connection ping. Jika ada path, gunakan HTTP GET 2xx/3xx. |
| `services.<id>.health_check.interval` | `duration` | Opt | `5s` | `1s` - `60s` | Valid Go duration format | None | Health Checker | **Bisa di-defaultkan.** Developer jarang mengubahnya. |
| `proxy.engine` | `string` | Opt | `embedded` | `embedded` | String enum | None | Proxy Manager | **REDUNDANT untuk User Config.** PORTA menggunakan embedded gateway engine secara native. Dieliminasi dari user-facing config. |
| `proxy.timeout` | `duration` | Opt | `30s` | `1s` - `300s` | Valid duration | None | Proxy Manager | **Bisa di-defaultkan.** Standard 30s timeout untuk HTTP requests. |
| `tunnel.provider` | `string` | Opt | `cloudflare` | `cloudflare`, `ngrok` | Enum: `cloudflare`, `ngrok` | None | Tunnel Manager | **Esensial.** Menentukan backend tunneling adapter. Default `cloudflare` (zero setup cost). |
| `tunnel.subdomain` | `string` | Opt | `""` (random ephemeral) | Regex `^[a-z0-9-]+$` | Lowercase alphanumeric | Tunnel provider support | Tunnel Manager | **Opsional.** Request static subdomain jika developer menggunakan named tunnel terautentikasi. |
| `security.mode` | `string` | Opt | `public` | `public`, `password`, `token` | Enum | None | Security Manager | **Esensial.** Mengatur level proteksi public gateway. |
| `security.password` | `string` | Opt | `""` | Any string min 6 chars | Required if `mode=password` | `security.mode` | Security Manager | **Esensial untuk Protected Mode.** Mendukung format `user:pass` atau plaintext pass (default user: `porta`). Mendukung syntax env: `${PORTA_AUTH_PASS}`. |
| `security.token` | `string` | Opt | `""` | String min 16 chars | Required if `mode=token` | `security.mode` | Security Manager | **Esensial untuk Token Mode.** Mendukung syntax env: `${PORTA_AUTH_TOKEN}`. |
| `security.allowed_ips` | `list` | Opt | `[]` (allow all) | CIDR / IPv4 / IPv6 strings | Valid IP format | None | Security Manager | **Opsional.** Filter IP public visitor. |

---

## 6. CONFIGURATION DESIGN & MVP SCHEMA SPECIFICATION

### 6.1 Desain Format YAML Minimalis vs Lengkap

#### A. Kasus 1: Minimalist Single Service (Zero Boilerplate)
Developer hanya punya 1 web app di port 3000. Cukup 3 baris:
```yaml
# porta.yaml
project: my-app
services:
  app: 3000
```
*Auto-resolved oleh PORTA:*
- `host` -> `127.0.0.1`
- `route` -> `/`
- `tunnel.provider` -> `cloudflare`
- `security.mode` -> `public`

---

#### B. Kasus 2: Multi-Service Standard (Frontend + Backend API)
```yaml
# porta.yaml
project: qulineria

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
```

---

#### C. Kasus 3: Full Feature Production-Ready Setup (Protected + Webhook + Token)
```yaml
# porta.yaml
version: "1"

project:
  name: store-platform

services:
  web:
    host: 127.0.0.1
    port: 3000
    route: /

  api:
    host: 127.0.0.1
    port: 8000
    route: /api
    strip_path: true
    health_check:
      path: /api/v1/ping
      interval: 3s

  realtime:
    port: 9001
    route: /socket.io

tunnel:
  provider: cloudflare # cloudflare | ngrok

security:
  mode: password # public | password | token
  password: ${PORTA_PASSWORD:-admin:SecretPass123!}
  allowed_ips:
    - 0.0.0.0/0
```

---

## 7. FUNCTIONAL DEPENDENCY GRAPH

```text
[ Project Manager ]
        │ (reads workspace)
        ▼
[ Config Engine ] ◄── (validates YAML schema & rules)
        │
        ├────────────────────────────────┐
        ▼                                ▼
[ Service Registry ]             [ Security Manager ]
        │ (maps routes & ports)          │ (injects auth interceptor)
        ▼                                │
[ Health Checker ]                       │
        │ (monitors targets)             │
        ▼                                ▼
[ Reverse Proxy Gateway ] ◄──────────────┘
        │ (aggregates into local listener)
        ▼
[ Tunnel Manager ]
        │ (establishes public egress)
        ▼
[ Public URL Manager ]
        │ (exposes HTTPS endpoint)
        ▼
[ Observability & CLI ]
```

---
*End of Functional Requirement Map.*
