# PORTA Windows Installer (PowerShell)
# Installs PORTA into %LOCALAPPDATA%\PORTA\bin\ (No Admin required)
param (
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\PORTA\bin",
    [switch]$NoSetup,
    [switch]$Silent
)

$ErrorActionPreference = "Stop"

function Write-Info($msg) {
    if (-not $Silent) {
        Write-Host "[i] $msg" -ForegroundColor Cyan
    }
}

function Write-Success($msg) {
    if (-not $Silent) {
        Write-Host "[✓] $msg" -ForegroundColor Green
    }
}

function Write-Err($msg) {
    Write-Host "[✗] $msg" -ForegroundColor Red
}

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Blue
Write-Host "║              PORTA Installer for Windows                   ║" -ForegroundColor Cyan
Write-Host "║     Local Multi-Service Public Exposure Platform           ║" -ForegroundColor Blue
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Blue
Write-Host ""

# 1. Determine Architecture
$arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
}
$assetName = "porta-windows-$arch.exe"
Write-Info "Target system: Windows ($arch)"

# 2. Resolve Release Version
$repo = "nuexn0x-9/porta"
$downloadVersion = $Version

if ($Version -eq "latest") {
    Write-Info "Resolving latest stable release from GitHub..."
    try {
        $releaseJson = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -Headers @{"User-Agent"="PORTA-Installer"}
        $downloadVersion = $releaseJson.tag_name
    } catch {
        $downloadVersion = "v1.0.0"
    }
}

Write-Info "Installing PORTA version: $downloadVersion"

# 3. Prepare Target Directories
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$targetExe = Join-Path $InstallDir "porta.exe"
$tempExe = Join-Path $InstallDir "porta.tmp.exe"
$checksumFile = Join-Path $InstallDir "checksums.txt"

# 4. Download Binary & Checksums
$downloadUrl = "https://github.com/$repo/releases/download/$downloadVersion/$assetName"
$checksumUrl = "https://github.com/$repo/releases/download/$downloadVersion/checksums.txt"

Write-Info "Downloading $assetName..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempExe -UseBasicParsing
} catch {
    Write-Err "Failed to download $assetName from $downloadUrl: $_"
    exit 1
}

# Checksum Verification
$checksumVerified = $false
try {
    Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumFile -UseBasicParsing
    if (Test-Path $checksumFile) {
        $actualHash = (Get-FileHash -Path $tempExe -Algorithm SHA256).Hash.ToLower()
        $expectedLine = Get-Content $checksumFile | Where-Object { $_ -match $assetName }
        if ($expectedLine) {
            $expectedHash = ($expectedLine -split '\s+')[0].ToLower()
            if ($actualHash -eq $expectedHash) {
                $checksumVerified = $true
                Write-Success "SHA-256 integrity verified ($actualHash)"
            } else {
                Write-Err "Checksum mismatch! Expected: $expectedHash, got: $actualHash"
                Remove-Item -Force $tempExe, $checksumFile -ErrorAction SilentlyContinue
                exit 1
            }
        }
    }
} catch {
    Write-Info "Release checksums.txt not published; continuing with downloaded binary."
} finally {
    Remove-Item -Force $checksumFile -ErrorAction SilentlyContinue
}

# 5. Place Binary
Move-Item -Path $tempExe -Destination $targetExe -Force
Write-Success "PORTA binary installed to $targetExe"

# 6. Configure User PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -split ';' -notcontains $InstallDir) {
    Write-Info "Adding $InstallDir to User PATH..."
    $newUserPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Success "User PATH updated successfully."
} else {
    Write-Success "$InstallDir is already in User PATH."
}

# 7. Run Setup
if (-not $NoSetup) {
    Write-Host ""
    & $targetExe setup
} else {
    Write-Success "Installation complete! Run 'porta doctor' to verify."
}
