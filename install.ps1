# CuRe Code - Windows Installer Script
# Usage: iex (irm https://raw.githubusercontent.com/broman0x/cure-code/main/install.ps1)

$ErrorActionPreference = "Stop"

$Repo = "broman0x/cure-code"
$BinaryName = "curecode.exe"

Write-Host "✦ Checking for latest version..." -ForegroundColor Cyan
$ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $ReleaseInfo.tag_name

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/curecode-windows-amd64.exe"

$InstallDir = Join-Path $env:LOCALAPPDATA "CuReCode"
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$TargetPath = Join-Path $InstallDir $BinaryName

Write-Host "✦ Downloading CuRe Code $Version..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetPath

# Add to PATH for current session
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    Write-Host "✦ Adding to User PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
    $env:Path += ";$InstallDir"
}

Write-Host "✦ Installation complete!" -ForegroundColor Green
Write-Host "✦ Running initial setup..." -ForegroundColor Cyan

& $TargetPath --install

Write-Host "══════════════════════════════════════════" -ForegroundColor White
Write-Host "  CuRe Code installed successfully!" -ForegroundColor Green
Write-Host "  Type 'curecode' to start." -ForegroundColor White
Write-Host "  Note: You might need to restart your terminal." -ForegroundColor Yellow
Write-Host "══════════════════════════════════════════" -ForegroundColor White
