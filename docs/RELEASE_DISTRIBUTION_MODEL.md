# PORTA RELEASE & BINARY DISTRIBUTION MODEL

**Target Repository:** `https://github.com/nuexn0x-9/porta.git`  
**Product:** PORTA v1.0.0 (MVP)  

---

## 1. Why Compiled Binaries are Never Committed to Git

In professional open-source software engineering, compiled executable binaries (`.exe`, `.so`, `.dylib`, ELF binaries) are strictly **excluded from the Git source code repository tree**.

### Rationale:
1. **Repository Bloat:** Git is designed to track text deltas. Every new 15MB binary commit causes permanent repository size inflation that cannot be compressed.
2. **Security & Anti-Virus False Positives:** Automated package scanners and security tools flag binary executables in source trees.
3. **Cross-Platform Reproducibility:** A binary compiled on one developer's machine might contain OS-specific paths or local build artifacts. Binaries must be built reproducibly via CI or attached to official release tags.

---

## 2. Professional Binary Distribution Architecture

PORTA follows the standard GitHub Release Distribution Model:

```text
                                  ┌───────────────────────────┐
                                  │     GitHub Repository     │
                                  │ (Pure Source Code & Docs) │
                                  └─────────────┬─────────────┘
                                                │
                                                ▼  git tag v1.0.0
                                  ┌───────────────────────────┐
                                  │  GitHub Actions CI / CD   │
                                  │   (.github/workflows/)    │
                                  └─────────────┬─────────────┘
                                                │
                                                ▼
                                  ┌───────────────────────────┐
                                  │  Official GitHub Release  │
                                  │ (github.com/.../releases) │
                                  └───────┬─────┬─────┬───────┘
                                          │     │     │
                 ┌────────────────────────┘     │     └────────────────────────┐
                 ▼                              ▼                              ▼
    ┌───────────────────────────┐ ┌───────────────────────────┐ ┌───────────────────────────┐
    │  porta-windows-amd64.exe  │ │    porta-linux-amd64      │ │    porta-darwin-arm64     │
    │      (Windows 64-bit)     │ │     (Linux 64-bit)        │ │  (macOS Apple Silicon)    │
    └───────────────────────────┘ └───────────────────────────┘ └───────────────────────────┘
```

---

## 3. How Users Download PORTA

Users download pre-compiled standalone executables directly from:
👉 **[https://github.com/nuexn0x-9/porta/releases](https://github.com/nuexn0x-9/porta/releases)**

### Target Assets:
- `porta-windows-amd64.exe` — Windows x86_64
- `porta-linux-amd64` — Linux x86_64
- `porta-darwin-arm64` — macOS Apple Silicon (M1/M2/M3)
- `porta-darwin-amd64` — macOS Intel

---

## 4. Automated Release Pipeline (`.github/workflows/release.yml`)

Whenever a new version tag (e.g. `v1.0.0`, `v1.1.0`) is pushed to GitHub, GitHub Actions automatically:
1. Compiles optimized cross-platform binaries using `CGO_ENABLED=0 go build -ldflags="-s -w"`.
2. Generates the GitHub Release with release notes.
3. Attaches all standalone executable binaries to the Release page.

---
*End of Release Distribution Model.*
