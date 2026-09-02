# PORTA Windows Uninstaller (PowerShell)
param (
    [string]$InstallDir = "$env:LOCALAPPDATA\PORTA\bin",
    [switch]$PurgeData
)

$ErrorActionPreference = "SilentlyContinue"

Write-Host ""
Write-Host "[i] Uninstalling PORTA from your system..." -ForegroundColor Cyan

# 1. Remove Executable
$targetExe = Join-Path $InstallDir "porta.exe"
if (Test-Path $targetExe) {
    Remove-Item -Path $targetExe -Force
    Write-Host "[+] Removed binary $targetExe" -ForegroundColor Green
}

# 2. Clean User PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -split ';' -contains $InstallDir) {
    $newPath = ($userPath -split ';' | Where-Object { $_ -ne $InstallDir -and $_ -ne "" }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "[+] Removed $InstallDir from User PATH." -ForegroundColor Green
}

# 3. Clean Runtime Data if requested
$portaHome = Join-Path $env:USERPROFILE ".porta"
if ($PurgeData -and (Test-Path $portaHome)) {
    Remove-Item -Path $portaHome -Recurse -Force
    Write-Host "[+] Removed $portaHome" -ForegroundColor Green
}

Write-Host "[+] PORTA has been uninstalled successfully." -ForegroundColor Green
