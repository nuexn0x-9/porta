
# PORTA Windows Uninstaller (PowerShell)
param (
    [string]$InstallDir = "$env:LOCALAPPDATA\PORTA\bin",
    [switch]$KeepData
)

$ErrorActionPreference = "SilentlyContinue"

Write-Host ""
Write-Host "[i] Uninstalling PORTA from your system..." -ForegroundColor Cyan

# 1. Remove Executable & Install Folder
$targetExe = Join-Path $InstallDir "porta.exe"
if (Test-Path $targetExe) {
    Remove-Item -Path $targetExe -Force -ErrorAction SilentlyContinue
    Write-Host "[+] Removed binary $targetExe" -ForegroundColor Green
}

$parentDir = Split-Path -Parent $InstallDir
if (Test-Path $parentDir) {
    Remove-Item -Path $parentDir -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "[+] Cleaned installation folder $parentDir" -ForegroundColor Green
}

# 2. Clean User PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -split ";" -contains $InstallDir) {
    $newPath = ($userPath -split ";" | Where-Object { $_ -ne $InstallDir -and $_ -ne "" }) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "[+] Removed $InstallDir from User PATH." -ForegroundColor Green
}

# 3. Clean Global Runtime Data (~/.porta)
$portaHome = Join-Path $env:USERPROFILE ".porta"
if (-not $KeepData -and (Test-Path $portaHome)) {
    Remove-Item -Path $portaHome -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "[+] Removed global configuration, logs, and drivers in $portaHome" -ForegroundColor Green
}

Write-Host ""
Write-Host "[+] PORTA and all its components have been uninstalled successfully!" -ForegroundColor Green
Write-Host ""

