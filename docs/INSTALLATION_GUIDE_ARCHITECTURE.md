# INSTALLATION GUIDE & DISTRIBUTION ARCHITECTURE

**Product:** PORTA  
**Target Release:** v1.1.0  
**Document Status:** OFFICIAL USER DOCUMENTATION ARCHITECTURE  

---

## 1. Overview of Installation Channels

PORTA provides **three flexible installation channels** tailored to developer preferences:
1. **Automated One-Line Script (Recommended):** Fast, automated OS/architecture detection, checksum verification, and PATH injection.
2. **Package Managers (Community Channels):** Integration with Homebrew (macOS/Linux), Winget/Scoop (Windows), and AUR (Arch Linux).
3. **Manual Standalone Binary:** Direct download from GitHub Releases for air-gapped or custom infrastructure.

---

## 2. Platform Installation Workflows

### 2.1 Windows Installation

#### Method A: Automated PowerShell One-Liner (Recommended)
Open Windows PowerShell (Run as Administrator or standard user):
```powershell
irm https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/install.ps1 | iex
```

#### Method B: Winget Package Manager (Upcoming)
```powershell
winget install PORTA.porta
```

#### Method C: Manual Binary Placement
1. Download `porta-windows-amd64.exe` from [GitHub Releases](https://github.com/nuexn0x-9/porta/releases).
2. Rename to `porta.exe` and move to `C:\Program Files\PORTA\` (or `%USERPROFILE%\bin\`).
3. Add directory to system PATH.

---

### 2.2 Linux Installation

#### Method A: Automated Shell One-Liner (Recommended)
Open Terminal (Bash/Zsh):
```bash
curl -fsSL https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/install.sh | bash
```

#### Method B: Manual Binary Installation
```bash
# 1. Download appropriate architecture (amd64 or arm64)
curl -fsSL -o porta https://github.com/nuexn0x-9/porta/releases/latest/download/porta-linux-amd64

# 2. Make executable and move to system PATH
chmod +x porta
sudo mv porta /usr/local/bin/porta

# 3. Initialize setup
porta setup
```

---

### 2.3 macOS Installation

#### Method A: Automated Shell One-Liner (Recommended)
Open Terminal:
```bash
curl -fsSL https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/install.sh | bash
```

#### Method B: Homebrew (Upcoming)
```bash
brew install nuexn0x-9/tap/porta
```

---

## 3. Post-Installation Verification & Upgrades

### Verification:
```bash
porta doctor
```

### Upgrading to Latest Version:
```bash
# Re-run installer script or run built-in upgrade:
porta upgrade
```

### Uninstallation:
```bash
# Windows (PowerShell)
irm https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/uninstall.ps1 | iex

# Linux / macOS (Shell)
curl -fsSL https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/uninstall.sh | bash
```

---
*End of Installation Guide Architecture.*
