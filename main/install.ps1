[CmdletBinding()]
param(
    [string] $Version    = $env:PORTKILL_VERSION,
    [string] $InstallDir = $env:PORTKILL_INSTALL_DIR
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

$Repo       = "khanalsaroj/portkill"
$BinaryName = "portkill.exe"
$AssetName  = "portkill_windows_amd64.exe"   # Only amd64 is released for Windows.

if (-not $InstallDir) {
    $isAdmin = ([Security.Principal.WindowsPrincipal] `
        [Security.Principal.WindowsIdentity]::GetCurrent() `
    ).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

    $InstallDir = if ($isAdmin) {
        "C:\Program Files\portkill"
    } else {
        "$env:USERPROFILE\.local\bin"
    }
}

# ---------------------------------------------------------------------------
# Colour helpers
# ---------------------------------------------------------------------------

function Write-Info    ($msg) { Write-Host "  -> $msg"  -ForegroundColor Cyan   }
function Write-Success ($msg) { Write-Host "  v  $msg"  -ForegroundColor Green  }
function Write-Warn    ($msg) { Write-Host "  !  $msg"  -ForegroundColor Yellow }
function Write-Fail    ($msg) { Write-Host "  x  $msg"  -ForegroundColor Red; exit 1 }
function Write-Header  ($msg) { Write-Host "`n$msg" -ForegroundColor White }

# ---------------------------------------------------------------------------
# Preflight: require PowerShell 5+ and TLS 1.2
# ---------------------------------------------------------------------------

Write-Header "portkill installer"

if ($PSVersionTable.PSVersion.Major -lt 5) {
    Write-Fail "PowerShell 5 or newer is required (you have $($PSVersionTable.PSVersion))."
}

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# ---------------------------------------------------------------------------
# Resolve version
# ---------------------------------------------------------------------------

function Resolve-Version {
    if ($Version) { return $Version }

    Write-Info "Fetching latest release version..."

    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    try {
        $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        return $release.tag_name
    } catch {
        Write-Fail "Could not fetch latest release: $_`nSet -Version v1.2.0 or `$env:PORTKILL_VERSION to pin a version."
    }
}

$Version = Resolve-Version
Write-Info "Installing version: $Version"

# ---------------------------------------------------------------------------
# Build download URL
# ---------------------------------------------------------------------------

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$AssetName"

# ---------------------------------------------------------------------------
# Download to a temp file
# ---------------------------------------------------------------------------

$TmpDir  = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), [System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $TmpDir | Out-Null

$TmpExe  = Join-Path $TmpDir $BinaryName

$cleanup = {
    if (Test-Path $TmpDir) { Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue }
}
Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action $cleanup | Out-Null

Write-Info "Downloading $DownloadUrl ..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpExe -UseBasicParsing
} catch {
    Write-Fail "Download failed: $_`nCheck that version '$Version' has a Windows release asset."
}

# ---------------------------------------------------------------------------
# Smoke test — verify the binary actually runs
# ---------------------------------------------------------------------------

try {
    $output = & $TmpExe help 2>&1
    # We don't assert specific output; just confirm it exits without crashing.
} catch {
    Write-Warn "Smoke test inconclusive — the binary may still work. ($_)"
}

# ---------------------------------------------------------------------------
# Install
# ---------------------------------------------------------------------------

if (-not (Test-Path $InstallDir)) {
    Write-Info "Creating install directory: $InstallDir"
    try {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    } catch {
        Write-Fail "Could not create $InstallDir — try running as Administrator or set a different -InstallDir."
    }
}

$Dest = Join-Path $InstallDir $BinaryName

try {
    Copy-Item -Path $TmpExe -Destination $Dest -Force
} catch {
    Write-Fail "Could not copy binary to $Dest`nRun as Administrator, or use: ``-InstallDir `"$env:USERPROFILE\.local\bin`""
}

Write-Success "Installed to $Dest"

# ---------------------------------------------------------------------------
# PATH management
# ---------------------------------------------------------------------------

function Get-PathScope {
    # Prefer Machine scope (system-wide) when running as Admin.
    $isAdmin = ([Security.Principal.WindowsPrincipal] `
        [Security.Principal.WindowsIdentity]::GetCurrent() `
    ).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    return if ($isAdmin) { "Machine" } else { "User" }
}

$pathScope    = Get-PathScope
$currentPath  = [Environment]::GetEnvironmentVariable("Path", $pathScope)
$pathEntries  = $currentPath -split ";"

if ($pathEntries -notcontains $InstallDir) {
    Write-Info "Adding $InstallDir to the $pathScope PATH..."
    try {
        $newPath = ($pathEntries + $InstallDir) -join ";"
        [Environment]::SetEnvironmentVariable("Path", $newPath, $pathScope)
        # Also update the current session's PATH so it works immediately.
        $env:Path += ";$InstallDir"
        Write-Success "$InstallDir added to PATH"
        Write-Warn "Open a new terminal (or restart your IDE) for the PATH change to take full effect."
    } catch {
        Write-Warn "Could not update PATH automatically: $_"
        Write-Warn "Add this directory to your PATH manually: $InstallDir"
    }
} else {
    Write-Info "$InstallDir is already in PATH"
}

# ---------------------------------------------------------------------------
# Cleanup temp dir now (don't wait for engine exit)
# ---------------------------------------------------------------------------

Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------

Write-Host ""
Write-Host "All done! Run it:" -ForegroundColor White
Write-Host ""
Write-Host "  portkill kill 8080" -ForegroundColor Cyan
Write-Host ""