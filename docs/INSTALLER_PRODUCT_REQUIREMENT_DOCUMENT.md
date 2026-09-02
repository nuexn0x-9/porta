# INSTALLER PRODUCT REQUIREMENT DOCUMENT (PRD)

**Product:** PORTA Installation System & First Run Experience (FRE)  
**Target Release:** PORTA v1.1.0  
**Document Status:** OFFICIAL PRODUCT SPECIFICATION  
**Author:** Senior Product Manager + DX Lead + Release Architect  

---

## 1. Executive Summary & Problem Statement

### 1.1 Current Friction (Manual Binary Distribution)
In PORTA v1.0.0 (MVP), developers download compiled binaries directly from GitHub Releases. While technically functional, this manual distribution introduces significant friction for average developers:
1. **Manual OS/Architecture Selection:** Developers must manually identify whether their machine requires `amd64`, `arm64`, `windows-amd64.exe`, or `darwin-arm64`. Choosing the wrong binary leads to cryptic OS execution errors (e.g., `exec format error` or `bad CPU type in executable`).
2. **Manual PATH Configuration:** Placing the binary in a working directory does not make `porta` globally executable. Developers must manually edit system environment variables on Windows or copy to `/usr/local/bin` on POSIX systems with `sudo`.
3. **Absence of Integrity Verification:** Manual downloads lack automatic SHA-256 checksum verification, exposing users to corrupted downloads or potential tampering.
4. **First-Run Uncertainty:** A new user does not know whether their workstation has working loopback permissions, outbound internet connectivity, or required tunnel drivers until they attempt a live session.

### 1.2 Product Vision
> **"PORTA should be installable and verified on any developer workstation with a single terminal command in under 60 seconds."**

Target One-Line Installers:
- **Windows (PowerShell):**
  ```powershell
  irm https://porta.dev/install.ps1 | iex
  # or GitHub raw fallback:
  irm https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/install.ps1 | iex
  ```
- **Linux & macOS (Shell):**
  ```bash
  curl -fsSL https://porta.dev/install.sh | bash
  # or GitHub raw fallback:
  curl -fsSL https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/install.sh | bash
  ```

---

## 2. User Personas & Target Journeys

### 2.1 Personas
1. **Frontend Engineer (Alex):** Works on React/Next.js on macOS (Apple Silicon M2). Wants a one-command install without dealing with Homebrew formulas or PATH exports.
2. **Backend / API Engineer (Budi):** Works on FastAPI/Django on Ubuntu Linux. Wants a script that handles permissions, verifies architecture (`x86_64`), and puts `porta` in `/usr/local/bin`.
3. **Full-Stack Developer (Charlie):** Works on Windows 11 with PowerShell. Wants an automated installer that places PORTA in `Program Files\PORTA`, adds it to user PATH, and verifies execution without manual reboots.
4. **QA / Automation Tester (Diana):** Running integration tests in CI/CD and local environments. Needs headless silent installation (`--silent`) and version pinning (`--version v1.0.0`).

### 2.2 User Journey Comparison

```text
BEFORE (Manual Binary Model):
[Visit GitHub] ──► [Find Releases] ──► [Select Asset] ──► [Download] ──► [Move to Folder] ──► [Edit PATH] ──► [Test porta]
Friction: ~6-8 steps | Duration: ~5-10 minutes | Drop-off Rate: ~25%

AFTER (Automated One-Line Installer):
[Copy One-Liner] ──► [Paste in Terminal] ──► [Auto-Detect OS/Arch] ──► [Auto-Verify SHA256] ──► [PORTA Ready]
Friction: 1 step | Duration: < 45 seconds | Drop-off Rate: < 2%
```

---

## 3. Success Metrics & Objectives

| Objective | Target Metric | Measurement Method |
| :--- | :--- | :--- |
| **Installation Speed** | $< 60$ seconds on broadband | Script execution to command prompt readiness. |
| **Zero Manual Dependencies** | 0 external packages required | Portable shell/PowerShell scripts with native HTTPS clients. |
| **First Public URL Time** | $< 3$ minutes from zero | One-liner install $\rightarrow$ `porta init` $\rightarrow$ `porta start`. |
| **Checksum Integrity** | 100% verified downloads | Mandatory SHA-256 validation against official `checksums.txt`. |
| **Clean Uninstallation** | 100% complete removal | `porta uninstall` removes binary, PATH entries, and runtime cache cleanly. |

---

## 4. Supported Platform Matrix

| Platform | Architecture | Target Installation Path | Privilege Level |
| :--- | :--- | :--- | :--- |
| **Windows 10 / 11** | `amd64` (x86_64) | `C:\Program Files\PORTA\porta.exe` (or `%LOCALAPPDATA%\PORTA\bin\`) | Admin or User |
| **Linux (Ubuntu/Debian/RHEL)** | `amd64`, `arm64` | `/usr/local/bin/porta` (or `~/.local/bin/porta`) | Sudo or User |
| **macOS (Intel & Apple Silicon)** | `amd64`, `arm64` | `/usr/local/bin/porta` (or `/opt/homebrew/bin/porta`) | Sudo or User |

---
*End of Installer PRD.*
