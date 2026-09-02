# INSTALLER SYSTEM REQUIREMENT SPECIFICATION (SRS)

**Product:** PORTA Installation System & First Run Experience  
**Document Status:** OFFICIAL TECHNICAL SPECIFICATION  

---

## 1. System Architecture & Component Layering

```text
                                  ┌─────────────────────────────┐
                                  │   GitHub Releases API       │
                                  │  (Binaries & Checksums)     │
                                  └──────────────┬──────────────┘
                                                 │ HTTPS (TLS 1.3)
═════════════════════════════════════════════════╪═════════════════════════════════════════════
 DEVELOPER WORKSTATION                           ▼
                                  ┌─────────────────────────────┐
                                  │       INSTALLER LAYER       │
                                  │ (install.ps1 / install.sh)  │
                                  └──────────────┬──────────────┘
                                                 │
                   ┌─────────────────────────────┼─────────────────────────────┐
                   ▼                             ▼                             ▼
        ┌──────────────────────┐      ┌──────────────────────┐      ┌──────────────────────┐
        │  Platform Detector   │      │  Release Downloader  │      │  Integrity Verifier  │
        │  (OS & CPU Arch)     │      │   (GitHub Releases)  │      │  (SHA-256 Checksums) │
        └──────────┬───────────┘      └──────────┬───────────┘      └──────────┬───────────┘
                   │                             │                             │
                   └─────────────────────────────┼─────────────────────────────┘
                                                 ▼
                                  ┌─────────────────────────────┐
                                  │       BINARY INSTALLER      │
                                  │  (Placement & Permissions)  │
                                  └──────────────┬──────────────┘
                                                 ▼
                                  ┌─────────────────────────────┐
                                  │   ENVIRONMENT CONFIGURATOR  │
                                  │  (PATH Injection & Refresh) │
                                  └──────────────┬──────────────┘
                                                 ▼
                                  ┌─────────────────────────────┐
                                  │      PORTA RUNTIME (CLI)    │
                                  │   (porta setup / doctor)    │
                                  └─────────────────────────────┘
```

---

## 2. Installer Component Specifications

### 2.1 Platform Detector
- **Input:** Native OS commands (`$PSVersionTable`, `uname -s`, `uname -m`, `arch`).
- **Logic:**
  ```text
  OS:
    Windows / MINGW / CYGWIN -> "windows"
    Linux -> "linux"
    Darwin -> "darwin"

  Architecture:
    x86_64 / amd64 -> "amd64"
    arm64 / aarch64 / armv8* -> "arm64"
  ```
- **Output:** Canonical asset name (e.g., `porta-windows-amd64.exe`, `porta-darwin-arm64`).

### 2.2 Release Downloader
- **Protocol:** HTTPS (TLS 1.3 / 1.2).
- **Endpoint Structure:**
  ```text
  Binary URL:
  https://github.com/nuexn0x-9/porta/releases/download/{VERSION}/{ASSET_NAME}

  Checksum URL:
  https://github.com/nuexn0x-9/porta/releases/download/{VERSION}/checksums.txt
  ```
- **Fallback:** Uses GitHub API (`/repos/nuexn0x-9/porta/releases/latest`) to resolve the latest tag dynamically when `--version` is omitted.

### 2.3 Integrity Verifier
- **Windows Implementation:** `Get-FileHash -Algorithm SHA256`
- **Linux Implementation:** `sha256sum` (or `shasum -a 256`)
- **macOS Implementation:** `shasum -a 256`
- **Validation Routine:** Computes the hash of the downloaded temporary file and compares it in a constant-time string check with the entry in `checksums.txt`. If mismatch, script exits immediately and deletes temporary files.

### 2.4 Binary Installer & Permissions Manager
- **Atomic File Operations:** Downloads to a `.tmp` file, validates checksum, then performs an atomic move/rename to the target binary location.
- **Permission Assignment:**
  - POSIX (`Linux`/`macOS`): `chmod 0755 <target_path>`
  - Windows: Preserves user access control list (ACL).

### 2.5 Environment Configurator (PATH Manager)
- **Windows Path Manipulation:**
  - Elevated: Appends to System PATH via `[Environment]::SetEnvironmentVariable("Path", ..., "Machine")`.
  - User-level: Appends to User PATH via `[Environment]::SetEnvironmentVariable("Path", ..., "User")`.
  - In-session refresh: Updates `$env:Path` in the calling PowerShell process.
- **POSIX Path Manipulation:**
  - If installed to `/usr/local/bin` (which is in standard PATH on 99% of Unix systems), no profile edit required.
  - If installed to `~/.local/bin`, inspects `$SHELL` and appends `export PATH="$HOME/.local/bin:$PATH"` to `~/.zshrc`, `~/.bashrc`, or `~/.profile`.

---

## 3. Installer Script Specifications

### 3.1 Windows PowerShell Script (`scripts/install.ps1`)
- **Execution Policy:** Compatible with `Restricted` policy via `powershell -ExecutionPolicy Bypass -File ...` or standard pipe `irm ... | iex`.
- **Arguments:**
  - `-Version <string>`: Target release version (default: `latest`).
  - `-InstallDir <string>`: Custom installation directory.
  - `-NoSetup`: Skip running `porta setup` after install.
  - `-Silent`: Suppress interactive prompt and animations.

### 3.2 Linux & macOS Shell Script (`scripts/install.sh`)
- **Shell Compatibility:** POSIX compliant (`sh`, `bash`, `zsh`).
- **Dependencies:** Uses standard utilities (`curl` or `wget`, `tar` or `gzip`, `sha256sum` or `shasum`).
- **Arguments / Flags:**
  - `--version <vX.Y.Z>`: Pin specific version.
  - `--prefix <path>`: Custom installation prefix.
  - `--no-setup`: Skip first-run initialization.
  - `--silent`: Unattended mode for CI/CD pipelines.

### 3.3 Uninstaller Scripts (`scripts/uninstall.ps1` & `scripts/uninstall.sh`)
- **Action:**
  1. Identifies binary location and deletes executable.
  2. Cleans up PATH entries from profile/registry.
  3. Prompts user: `Do you want to delete runtime data in ~/.porta/? [y/N]`.

---

## 4. Security Architecture & Threat Mitigations

| Threat Vector | Mitigation Strategy |
| :--- | :--- |
| **Man-in-the-Middle (MitM) Tampering** | Strict HTTPS transport with TLS 1.3 enforcement; zero fallback to plain HTTP. |
| **Asset Corruption / Supply Chain Injection** | Mandatory SHA-256 checksum matching against signed release `checksums.txt`. |
| **Arbitrary Code Execution via URL** | Scripts hosted exclusively on official project GitHub repository (`github.com/nuexn0x-9/porta`). |
| **Privilege Escalation Risks** | Scripts default to user-level permissions (`~/.local/bin`, `%LOCALAPPDATA%`) and only request elevation when writing to system directories. |

---
*End of Installer SRS.*
