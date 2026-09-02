# Changelog

All notable changes to the **PORTA** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0] - 2026-09-02

### Added
- **Automated Installers:** One-line PowerShell installer (`scripts/install-windows.ps1`) and POSIX shell installer (`scripts/install-posix.sh`) with dynamic OS/architecture detection and SHA-256 integrity verification.
- **Uninstaller Scripts:** Automated uninstallers (`scripts/uninstall-windows.ps1`, `scripts/uninstall-posix.sh`) with PATH cleanup and optional runtime data purge.
- **First Run Setup (`porta setup`):** Initial setup engine that creates `~/.porta/`, generates `~/.porta/config/global.yaml`, verifies loopback socket bindability, and pre-caches tunnel drivers.
- **Self-Upgrade Engine (`porta upgrade`):** In-place binary self-upgrade checking GitHub Releases API with SHA-256 verification and atomic executable swap.
- **Version Command (`porta version`):** Version metadata inspector with commit hash, build date, Go runtime, and `--json` support.
- **CI/CD Release Automation:** Automatic SHA-256 `checksums.txt` computation and release asset uploading via GitHub Actions.

---

## [1.0.0] - 2026-09-02

### Added
- **Core CLI Engine:** Implemented Cobra CLI suite (`init`, `start`, `status`, `logs`, `config`, `doctor`).
- **Unified Service Cardinality (DECISION SERVICE-001):** Seamless support for single-service (`localhost:3000`) and multi-service (`:3000` + `:8000`) architectures through a single pipeline.
- **Embedded Reverse Proxy:** High-performance in-process gateway on ephemeral loopback ports (`127.0.0.1:0`).
- **Longest Prefix Matching (LPM):** Intelligent path routing with optional prefix stripping (`strip_path: true`).
- **WebSocket Hijacking:** Transparent full-duplex TCP piping for WebSockets (`http.Hijacker`).
- **Unbuffered Streaming:** Zero-delay buffer flushing for Server-Sent Events (SSE) and LLM streaming completions.
- **Cloudflare Quick Tunnel Driver:** Zero-auth public HTTPS tunnel driver with on-demand binary downloader and local cache (`~/.porta/bin/`).
- **Strict SSRF Guard:** Two-layer loopback firewall blocking RFC 1918 subnets, cloud metadata (`169.254.169.254`), and DNS rebinding attacks.
- **Ingress Access Gatekeeper:** HTTP Basic Auth and Bearer/Query Token authentication with constant-time string comparison.
- **Safe Structured Logging:** Disk logging (`.porta/logs/access.log`) with automated token redaction (`?porta_token=[REDACTED]`).
- **Foreground Terminal UI (TUI):** Colorized ANSI status table and live request log feed.
- **System Doctor Diagnostics:** `porta doctor` for pre-flight networking, connectivity, and driver checks.

---
