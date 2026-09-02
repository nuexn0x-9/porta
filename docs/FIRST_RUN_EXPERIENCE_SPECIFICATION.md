# FIRST RUN EXPERIENCE (FRE) SPECIFICATION

**Product:** PORTA  
**Feature:** First-Time Setup & Onboarding Engine (`porta setup`)  
**Document Status:** OFFICIAL SPECIFICATION  

---

## 1. Objectives & First-Run Philosophy

The **First Run Experience (FRE)** ensures that a newly installed PORTA instance is completely configured, verified, and pre-cached before the developer ever attempts to expose their first application.

### Key Goals:
1. **Zero Runtime Delay on First `porta start`:** Pre-downloads and verifies the Cloudflare tunnel driver during setup so `porta start` establishes public ingress instantly ($< 2$ seconds).
2. **Pre-Emptive Failure Detection:** Diagnoses firewall blocks, loopback permission issues, or internet disconnection during setup with clear actionable guidance.
3. **Frictionless Developer Handoff:** Guides the developer directly to the next command (`porta init` $\rightarrow$ `porta start`).

---

## 2. The `porta setup` Command Workflow

```text
[ Developer Runs: 'porta setup' ]
                │
                ▼
      1. System Detection (OS & Arch)
                │
                ▼
      2. Directory Tree Creation (~/.porta/)
                │
                ▼
      3. Loopback Interface Probe (127.0.0.1:0)
                │
                ▼
      4. Outbound HTTPS Connectivity Probe (1.1.1.1)
                │
                ▼
      5. Tunnel Driver Pre-Cache (cloudflared in ~/.porta/bin/)
                │
                ▼
      6. Diagnostic Verification Complete
                │
                ▼
[ Ready Banner: "Run 'porta init' in your project directory" ]
```

---

## 3. Terminal Output Specification

```text
$ porta setup

╔════════════════════════════════════════════════════════════╗
║                   PORTA Initial Setup                      ║
║     Local Multi-Service Public Exposure Platform           ║
╚════════════════════════════════════════════════════════════╝

[i] Initializing PORTA environment...

  [✓] System detected       : windows/amd64
  [✓] Directory structure   : ~/.porta/ initialized
  [✓] Loopback interface    : 127.0.0.1 bindable
  [✓] Internet connectivity : DNS & HTTPS connection OK
  [✓] Tunnel driver         : Cloudflare driver pre-cached in ~/.porta/bin/
  [✓] Diagnostics passed    : All subsystems ready

──────────────────────────────────────────────────────────────
🎉 PORTA is successfully configured and ready to use!

Next Steps:
  1. Open your project folder : cd my-web-app
  2. Initialize configuration : porta init
  3. Start public exposure    : porta start
──────────────────────────────────────────────────────────────
```

---

## 4. First-Run Filesystem State (`~/.porta/`)

The setup command initializes the user directory structure with strict permissions (`0700` for directories, `0600` for files):

```text
~/.porta/
├── bin/
│   └── cloudflared.exe         # Pre-cached official tunnel provider binary
├── config/
│   └── global.yaml             # Global user preferences & default security modes
├── logs/
│   └── access.log              # Persistent structured access logs
└── cache/
    └── metadata.json           # Cached version check and system diagnostic stamps
```

---

## 5. Failure Handling & Recovery during First Run

| Check Failure | User-Facing Error Message | Automated Remediation / Solution |
| :--- | :--- | :--- |
| **Loopback Bind Failure** | `[✗] Loopback interface: Cannot bind on 127.0.0.1` | Explains that local socket permissions or third-party firewall is blocking localhost. |
| **Outbound Internet Blocked** | `[✗] Internet connectivity: Unreachable (1.1.1.1)` | Checks corporate proxy settings or network connection. |
| **Driver Download Failure** | `[✗] Tunnel driver: Failed to download cloudflared` | Retries download from fallback mirror or prompts user to install via package manager (`brew`/`winget`). |

---
*End of First Run Experience Specification.*
