# INSTALLER FUNCTIONAL REQUIREMENT MAP (FRM)

**Product:** PORTA Installation System & First Run Experience  
**Document Status:** OFFICIAL FUNCTIONAL SPECIFICATION  

---

## 1. Requirement Specification Matrix

```text
Requirement Taxonomy:
INSTALL-xxx: Installer & Distribution Lifecycle Requirements
FRE-xxx:     First Run Experience Requirements
```

---

### INSTALL-001: Automatic OS & Architecture Detection
- **Feature:** Client-Side Platform Identification
- **Description:** The installer script detects the host operating system (`Windows`, `Linux`, `Darwin`) and hardware architecture (`amd64`, `arm64`, `x86_64`, `aarch64`) dynamically and maps them to the appropriate release asset filename.
- **Priority:** P0 (Must Have)
- **Acceptance Criteria:**
  - On Windows x64: Selects `porta-windows-amd64.exe`.
  - On Linux x64: Selects `porta-linux-amd64`.
  - On Linux ARM64: Selects `porta-linux-arm64`.
  - On macOS Apple Silicon (M1/M2/M3/M4): Selects `porta-darwin-arm64`.
  - On macOS Intel: Selects `porta-darwin-amd64`.
  - If OS or architecture is unsupported: Exits gracefully with a human-readable explanation.

---

### INSTALL-002: Secure Binary Download & SHA-256 Checksum Validation
- **Feature:** Cryptographic Integrity Verification
- **Description:** The installer downloads both the binary asset and `checksums.txt` from the official GitHub Release over HTTPS. It calculates the SHA-256 hash of the downloaded binary and verifies it matches the official checksum prior to placement in system directories.
- **Priority:** P0 (Must Have)
- **Acceptance Criteria:**
  - Download occurs strictly over TLS 1.3/1.2 HTTPS.
  - Matches binary hash against official checksum.
  - If checksum mismatch occurs, execution terminates immediately with exit code `1` and removes temporary files.

---

### INSTALL-003: Platform-Compliant Installation Directory Selection
- **Feature:** Smart Directory Hierarchy Selection
- **Description:** Installs the binary into standard operating system directories, supporting both elevated (administrator/root) and non-elevated (user-level) permissions.
- **Priority:** P0 (Must Have)
- **Acceptance Criteria:**
  - **Windows Elevated:** Installs to `C:\Program Files\PORTA\porta.exe`.
  - **Windows Non-Elevated:** Installs to `%LOCALAPPDATA%\PORTA\bin\porta.exe`.
  - **Linux / macOS Elevated:** Installs to `/usr/local/bin/porta` (executable permissions `0755`).
  - **Linux / macOS Non-Elevated:** Installs to `~/.local/bin/porta` (executable permissions `0755`).

---

### INSTALL-004: Automated PATH Configuration & Session Refresh
- **Feature:** PATH Environment Configuration
- **Description:** Checks if the target installation directory exists in the user or system `PATH`. If missing, appends the path persistently to the user profile or Windows registry and refreshes the active session environment so `porta` is immediately executable.
- **Priority:** P0 (Must Have)
- **Acceptance Criteria:**
  - Windows: Updates User Environment registry (`HKCU\Environment\Path`) and sets `$env:Path` in the running PowerShell session.
  - Linux / macOS: Appends export statement to `~/.bashrc`, `~/.zshrc`, or `~/.profile` if not already in system `/usr/local/bin`.
  - Running `porta --version` or `porta doctor` immediately after installation succeeds in the same terminal.

---

### INSTALL-005: First-Time Setup Engine (`porta setup`)
- **Feature:** First Run Experience (FRE) Command
- **Description:** Provides a dedicated setup command (`porta setup`) triggered automatically at the end of installation (or manually by user) to initialize directory trees, verify loopback interfaces, download tunnel drivers, and run diagnostics.
- **Priority:** P0 (Must Have)
- **Acceptance Criteria:**
  - Creates `~/.porta/` structure (`bin/`, `config/`, `logs/`, `cache/`).
  - Verifies loopback interface `127.0.0.1` bindability.
  - Probes outbound internet connectivity.
  - Pre-caches `cloudflared` into `~/.porta/bin/`.
  - Emits colorized confirmation: `PORTA is ready! Run 'porta init' to start.`

---

### INSTALL-006: Automated Dependency Pre-Caching
- **Feature:** Driver Lifecycle Management
- **Description:** Ensures developers never need to manually download, install, or configure Cloudflare tunnel drivers.
- **Priority:** P0 (Must Have)
- **Acceptance Criteria:**
  - `porta setup` downloads `cloudflared` driver during installation.
  - Verifies executable permissions (`0755`).

---

### INSTALL-007: Clean Uninstallation Support
- **Feature:** Complete System De-provisioning
- **Description:** Provides an automated uninstallation command (`porta uninstall`) and script (`uninstall.sh` / `uninstall.ps1`) to remove binaries, PATH entries, and runtime data.
- **Priority:** P1 (Should Have)
- **Acceptance Criteria:**
  - Removes binary file from installation folder.
  - Removes directory from user/system `PATH`.
  - Prompts user whether to keep or remove `~/.porta/` configuration and logs.

---

### INSTALL-008: In-Place Self-Upgrade Engine (`porta upgrade`)
- **Feature:** Automated Version Update
- **Description:** Inspects GitHub Releases API for new version tags, compares with running binary version, downloads updated binary with checksum validation, and atomically replaces the executable in-place.
- **Priority:** P1 (Should Have)
- **Acceptance Criteria:**
  - Running `porta upgrade` detects if current version is up to date.
  - If a newer version exists, downloads and replaces the binary.
  - Windows handles executable renaming (`.old` swap) to allow in-place replacement while running.

---
*End of Installer FRM.*
