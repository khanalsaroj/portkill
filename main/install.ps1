# portkill installer for Windows (PowerShell).
#
#   iwr -useb https://raw.githubusercontent.com/khanalsaroj/portkill/main/main/install.ps1 | iex
#
# Environment overrides:
#   $env:PORTKILL_VERSION       install a specific version (e.g. v1.2.3), default: latest
#   $env:PORTKILL_INSTALL_DIR   install location, default: %USERPROFILE%\.portkill\bin
$ErrorActionPreference = "Stop"

$Repo = "khanalsaroj/portkill"
$BinName = "portkill"
$InstallDir = if ($env:PORTKILL_INSTALL_DIR) { $env:PORTKILL_INSTALL_DIR } else { "$env:USERPROFILE\.portkill\bin" }

function Info($m) { Write-Host "  $m" }
function Ok($m)   { Write-Host "  $m" -ForegroundColor Green }
function Warn($m) { Write-Host "  $m" -ForegroundColor Yellow }
function Die($m)  { Write-Host "  $m" -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "  portkill installer" -ForegroundColor Cyan
Write-Host "  kill the process holding a port - instantly"
Write-Host ""

# ---------- Architecture ----------
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    "x86"   { "386" }
    default { if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" } }
}

# ---------- Version ----------
if ($env:PORTKILL_VERSION) {
    $Version = $env:PORTKILL_VERSION.TrimStart("v")
} else {
    try {
        $Version = (Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest").tag_name.TrimStart("v")
    } catch {
        Die "Could not resolve the latest release: $_"
    }
}
Info "Installing $BinName v$Version for windows/$Arch"

$Asset = "$BinName-windows-$Arch.zip"
$Url = "https://github.com/$Repo/releases/download/v$Version/$Asset"

$Tmp = Join-Path $env:TEMP ("portkill-" + [System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $Tmp | Out-Null
$Zip = Join-Path $Tmp $Asset

Info "Downloading $Url"
try {
    Invoke-WebRequest -Uri $Url -OutFile $Zip -UseBasicParsing
} catch {
    Die "Download failed - does a release exist for windows/$Arch? ($_)"
}

# ---------- Checksum verification (best effort) ----------
try {
    $SumsUrl = "https://github.com/$Repo/releases/download/v$Version/checksums.txt"
    $Sums = (Invoke-WebRequest -Uri $SumsUrl -UseBasicParsing).Content
    $Line = ($Sums -split "`n") | Where-Object { $_ -match [regex]::Escape($Asset) } | Select-Object -First 1
    if ($Line) {
        $Expected = ($Line.Trim() -split "\s+")[0]
        $Actual = (Get-FileHash -Algorithm SHA256 -Path $Zip).Hash
        if ($Expected -and ($Expected.ToLower() -ne $Actual.ToLower())) {
            Die "Checksum mismatch for $Asset"
        }
        Ok "Checksum verified"
    } else {
        Warn "No checksum entry for $Asset - skipping verification"
    }
} catch {
    Warn "Skipping checksum verification ($_)"
}

# ---------- Extract ----------
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Expand-Archive -Force -Path $Zip -DestinationPath $InstallDir

# ---------- PATH ----------
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $UserPath) { $UserPath = "" }
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", ($UserPath.TrimEnd(";") + ";" + $InstallDir), "User")
    Warn "Added $InstallDir to your PATH. Restart your terminal to pick it up."
}
$env:Path = "$env:Path;$InstallDir"

Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue

# ---------- Verify ----------
$Exe = Join-Path $InstallDir "$BinName.exe"
if (Test-Path $Exe) {
    Ok "Installed: $Exe"
    & $Exe version
    Write-Host ""
    Ok "Done! Try:  portkill kill 8080"
} else {
    Die "Installation failed - $Exe was not found"
}
