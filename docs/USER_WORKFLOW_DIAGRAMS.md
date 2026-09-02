# DIAGRAM & PANDUAN LENGKAP PENGGUNAAN PORTA (DARI AWAL DOWNLOAD)

Dokumen ini menyediakan diagram visual lengkap mulai dari **mengunduh PORTA pertama kali**, inisialisasi proyek, hingga menjalankan terowongan publik di **Windows** dan **Linux**.

![PORTA Usage Infographic](images/porta-usage-infographic.png)

---

## 1. Diagram Alur Lengkap dari Awal (Download -> Run -> Share)

```mermaid
flowchart TD
    subgraph FASE_1["FASE 1: Download & Install PORTA (Hanya 1x di Awal)"]
        A1["🌐 Buka GitHub Releases:<br/>github.com/nuexn0x-9/porta/releases"] --> A2["📥 Unduh Binary Sesuai OS:<br/>• Windows: porta.exe<br/>• Linux/macOS: porta"]
        A2 --> A3["📂 Pindahkan ke Folder PATH<br/>(misal: C:\Windows\System32 atau /usr/local/bin)"]
        A3 --> A4["🔍 Verifikasi di Terminal:<br/>porta doctor"]
    end

    subgraph FASE_2["FASE 2: Persiapan Aplikasi Lokal"]
        B1["💻 Developer Menjalankan App Lokal<br/>(contoh: npm run dev di port 3000 / 8000)"]
    end

    subgraph FASE_3["FASE 3: Inisialisasi Proyek"]
        C1["📁 Buka Terminal di Folder Proyek"] --> C2["⚙️ Ketik: porta init"]
        C2 --> C3["🔍 PORTA Scan Port Lokal Secara Offline"]
        C3 --> C4["📝 Terbentuk File: porta.yaml"]
    end

    subgraph FASE_4["FASE 4: Menjalankan PORTA"]
        D1["🚀 Ketik: porta start"] --> D2{"Apakah driver cloudflared<br/>sudah ada di ~/.porta/bin/?"}
        D2 -- Belum Ada --> D3["⬇️ PORTA Otomatis Download Driver Tunnel"]
        D2 -- Sudah Ada --> D4["⚡ PORTA Langsung Buka Gateway 127.0.0.1"]
        D3 --> D4
        D4 --> D5["🔒 Cloudflare Quick Tunnel Terhubung"]
        D5 --> D6["🌐 Muncul Public URL HTTPS<br/>(https://xxxx.trycloudflare.com)"]
    end

    subgraph FASE_5["FASE 5: Penggunaan & Selesai"]
        E1["📱 Bagikan URL ke Klien / QA / Webhook"] --> E2["🛑 Tekan Ctrl+C di Terminal untuk Berhenti"]
    end

    FASE_1 --> FASE_2
    FASE_2 --> FASE_3
    FASE_3 --> FASE_4
    FASE_4 --> FASE_5
```

---

## 2. Diagram Penggunaan Pengguna WINDOWS

### 2.1 Alur Kerja Pengguna Windows (PowerShell / CMD)

```mermaid
sequenceDiagram
    autonumber
    actor Dev as Developer (Windows)
    participant PS as PowerShell / CMD
    participant Porta as PORTA (porta.exe)
    participant LocalApp as App Lokal (:3000 / :8000)
    participant Cloud as Cloudflare Quick Tunnel
    actor User as Klien / QA / Smartphone

    Note over Dev,PS: Tahap 1: Persiapan Aplikasi
    Dev->>PS: Jalankan Vite / React / API (npm run dev)
    PS->>LocalApp: App Aktif di localhost:3000

    Note over Dev,Porta: Tahap 2: Inisialisasi & Start PORTA
    Dev->>PS: porta init
    Porta-->>PS: File porta.yaml terbuat otomatis
    Dev->>PS: porta start
    Porta->>Porta: Buka Reverse Proxy di 127.0.0.1:0 (Ephemeral)
    Porta->>Cloud: Hubungkan HTTPS Tunnel
    Cloud-->>Porta: Public URL Dialokasikan
    Porta-->>PS: Tampilkan TUI Live Status & URL Publik

    Note over User,LocalApp: Tahap 3: Akses Publik
    User->>Cloud: Buka https://slug.trycloudflare.com
    Cloud->>Porta: Teruskan Request ke PORTA Gateway
    Porta->>LocalApp: Reverse Proxy ke localhost:3000
    LocalApp-->>Porta: Response HTML/JSON
    Porta-->>User: Halaman Web Tampil di Perangkat Pengguna

    Note over Dev,Porta: Tahap 4: Selesai Sesi
    Dev->>PS: Tekan Ctrl+C
    Porta->>Cloud: Putus Tunnel & Bersihkan Resource (< 500ms)
```

### 2.2 Langkah Praktis di Windows:

```powershell
# 1. Pastikan aplikasi Anda sudah berjalan di localhost
npm run dev

# 2. Buka jendela PowerShell baru di folder proyek Anda
cd C:\Users\nama-user\my-project

# 3. Cek kesiapan sistem & driver tunnel
porta doctor

# 4. Inisialisasi konfigurasi otomatis
porta init

# 5. Jalankan PORTA (Akan muncul TUI interaktif dengan URL HTTPS publik)
porta start

# 6. Selesai: Tekan Ctrl + C untuk keluar
```

---

## 3. Diagram Penggunaan Pengguna LINUX

### 3.1 Alur Kerja Pengguna Linux (Bash / Zsh)

```mermaid
sequenceDiagram
    autonumber
    actor Dev as Developer (Linux)
    participant Bash as Terminal (Bash/Zsh)
    participant Porta as Binary PORTA (/usr/local/bin/porta)
    participant Service as Node.js/Python/Go Server
    participant Edge as Cloudflare Anycast Edge
    actor Tester as External Tester / Webhook

    Dev->>Bash: python3 main.py (Port 8000)
    Bash->>Service: Service listening on 127.0.0.1:8000

    Dev->>Bash: porta init
    Porta-->>Bash: porta.yaml created

    Dev->>Bash: porta start
    Porta->>Porta: Ephemeral Loopback Bind (127.0.0.1:0)
    Porta->>Edge: Establish Encrypted QUIC/mTLS Session
    Edge-->>Porta: Assigned Public URL (https://*.trycloudflare.com)
    Porta-->>Bash: Render Live ANSI Terminal UI

    Tester->>Edge: GET /api/v1/data
    Edge->>Porta: Proxy request through encrypted tunnel
    Porta->>Porta: SSRF Check & Path Routing (/api -> :8000)
    Porta->>Service: Forward request to 127.0.0.1:8000
    Service-->>Porta: Response 200 OK
    Porta-->>Tester: Stream Response to Tester

    Dev->>Bash: Ctrl + C (SIGINT)
    Porta->>Edge: Close Tunnel Connection Gracefully
```

### 3.2 Langkah Praktis di Linux:

```bash
# 1. Jalankan aplikasi lokal Anda
python3 -m http.server 8000

# 2. Buka terminal di direktori proyek
cd ~/projects/my-app

# 3. Inisialisasi porta.yaml
porta init

# 4. Jalankan PORTA
porta start

# 5. Cek log secara terpisah jika diperlukan (di terminal lain):
porta logs -f

# 6. Hentikan dengan Ctrl + C
```

---

## 4. Diagram Arsitektur Routing Multi-Service

Diagram ini menunjukkan bagaimana PORTA membagi satu URL publik ke beberapa aplikasi lokal yang berbeda:

```mermaid
graph LR
    subgraph Internet["Public Internet"]
        Client["📱 Browser / Klien Luar"]
    end

    subgraph TunnelEdge["Cloudflare Edge"]
        Edge["https://demo.trycloudflare.com"]
    end

    subgraph Workstation["Komputer Developer (Windows / Linux)"]
        Gateway["PORTA Gateway<br/>(127.0.0.1:0)"]
        Router["LPM Router"]
        
        Frontend["Frontend App<br/>localhost:3000<br/>(Route: /)"]
        Backend["Backend API<br/>localhost:8000<br/>(Route: /api)"]
        WS["WebSocket Server<br/>localhost:9001<br/>(Route: /ws)"]
    end

    Client -->|HTTPS Request| Edge
    Edge -->|Encrypted Tunnel| Gateway
    Gateway --> Router
    Router -->|Path: /| Frontend
    Router -->|Path: /api/*| Backend
    Router -->|Path: /ws| WS
```

---

## 5. Diagram Setup & Konfigurasi: Cara Otomatis vs Cara Manual

Berikut adalah alur perbandingan lengkap antara **Konfigurasi Otomatis (Auto)** dan **Konfigurasi Manual**:

```mermaid
flowchart TD
    Start(["🚀 Mulai Projek Baru / Ganti Service"]) --> SetupCheck{"Apakah PORTA<br/>Sudah Terpasang?"}

    %% Tahap Setup
    subgraph SETUP["1. Setup & Instalasi (Hanya 1x di Awal)"]
        SetupCheck -- Belum --> InstallChoice["Pilih Sistem Operasi:"]
        InstallChoice --> WinInst["Windows PowerShell:<br/><code>irm https://.../install-windows.ps1 | iex</code>"]
        InstallChoice --> PosixInst["Linux / macOS:<br/><code>curl -fsSL https://.../install-posix.sh | sh</code>"]
        WinInst --> AutoSetup["Otomatis Menjalankan:<br/><code>porta setup</code>"]
        PosixInst --> AutoSetup
    end

    SetupCheck -- Sudah Terpasang --> RunApps
    AutoSetup --> RunApps["💻 Jalankan Aplikasi / Server Lokal Anda<br/>(Contoh: Vite di :5173, Django di :8000)"]

    RunApps --> ConfigChoice{"Pilih Metode<br/>Konfigurasi:"}

    %% Cara Otomatis
    subgraph AUTO["2. CARA OTOMATIS (Auto Detection)"]
        ConfigChoice -- Opsi A: Auto --> AutoInit["Ketik di Terminal:<br/><code>porta init</code><br/><i>(atau <code>porta init --force</code>)</i>"]
        AutoInit --> Scanner["🔍 PORTA Scan Port Lokal yang Aktif<br/>(Deteksi port 3000, 5173, 8000, dll)"]
        Scanner --> GenYAML["📝 Otomatis Menghasilkan <code>porta.yaml</code><br/>dengan routing standar (/)"]
    end

    %% Cara Manual
    subgraph MANUAL["3. CARA MANUAL (Custom Config)"]
        ConfigChoice -- Opsi B: Manual --> EditFile["Buka / Buat file <code>porta.yaml</code><br/>di Text Editor (VS Code / Notepad)"]
        EditFile --> SetService["Tentukan Service & Port:<br/>• Single Service (:5173 -> /)<br/>• Multi-Service (:5173 -> /, :8000 -> /api)<br/>• Proteksi Password / Token"]
        SetService --> Validate["Cek Validasi Konfigurasi:<br/><code>porta config</code>"]
    end

    GenYAML --> StartPorta
    Validate --> StartPorta

    %% Menjalankan PORTA
    subgraph RUN["4. Menjalankan & Menguji"]
        StartPorta["🚀 Ketik: <code>porta start</code>"] --> CheckTunnel["🌐 Terhubung ke Cloudflare Tunnel"]
        CheckTunnel --> Ready["✅ Public HTTPS URL Aktif!<br/>Siap diakses dari luar / HP / Klien"]
    end
```

---

### Perbandingan Langkah Praktis:

#### A. Cara Otomatis (Auto)
Cocok untuk pemula atau saat berpindah ke projek baru:
```powershell
# 1. Masuk ke folder projek Anda
cd G:\projek-anda

# 2. Jalankan aplikasi lokal Anda terlebih dahulu
npm run dev

# 3. Jalankan inisialisasi otomatis
porta init

# 4. Langsung jalankan PORTA
porta start
```

#### B. Cara Manual (Custom Multi-Service & Password)
Cocok jika Anda memiliki arsitektur Frontend + Backend + WebSocket atau ingin menambahkan proteksi password:
```yaml
# Simpan sebagai: porta.yaml di root folder projek Anda
version: "1"

project:
  name: my-custom-app

services:
  # Service 1: Frontend (Next.js / Vite / React)
  frontend:
    port: 3000
    route: /

  # Service 2: Backend REST API
  backend:
    port: 8000
    route: /api
    strip_path: false

# Opsional: Keamanan Password
security:
  mode: password
  password: "admin:Demo12345"
```

Setelah file disimpan:
```powershell
# 1. Validasi sintaks file konfigurasi
porta config

# 2. Jalankan PORTA
porta start
```

---

## 6. Ringkasan Perintah CLI Lengkap

| Perintah | Fungsi Utama | Kapan Digunakan? |
| :--- | :--- | :--- |
| `porta setup` | Inisialisasi folder global `~/.porta/` & cek dependensi | Dipanggil otomatis saat instalasi |
| `porta doctor` | Cek koneksi internet, loopback bind, dan driver tunnel | Jika terjadi kendala koneksi |
| `porta init` | Scan port lokal secara otomatis dan buat `porta.yaml` | **Metode Otomatis** |
| `porta init --force` | Timpa konfigurasi `porta.yaml` lama dengan hasil scan baru | Saat ganti projek / port |
| `porta config` | Memvalidasi sintaks dan menampilkan struktur routing `porta.yaml` | **Metode Manual** sebelum start |
| `porta start` | Membuka gateway reverse proxy dan membuat URL publik | Saat ingin membagikan aplikasi ke internet |
| `porta status` | Menampilkan ringkasan status service yang terhubung | Memeriksa service yang aktif |
| `porta logs -f` | Menampilkan live streaming log request yang masuk | Debugging request yang gagal/404/500 |
| `porta upgrade` | Memperbarui binary PORTA ke versi terbaru | Pembaruan versi |
| `porta version` | Menampilkan versi PORTA yang terpasang | Pengecekan versi |
