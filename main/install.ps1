$ErrorActionPreference = "Stop"

$Repo = "khanalsaroj/portkill"
$InstallDir = "$env:USERPROFILE\.portkill\bin"

Write-Host "Installing portkill for Windows..."

# Create install directory
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Detect OS Arch
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

# Get latest release
$Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$Asset = $Release.assets | Where-Object { $_.name -match "windows" -and $_.name -match $Arch }

if (-not $Asset) {
    Write-Error "No Windows binary found in latest release."
}

$ZipUrl = $Asset.browser_download_url
$ZipFile = "$env:TEMP\portkill.zip"

Write-Host "Downloading $ZipUrl"
Invoke-WebRequest -Uri $ZipUrl -OutFile $ZipFile

# Extract
Expand-Archive -Force $ZipFile $InstallDir

# Add to PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    Write-Host "Added to PATH. Restart your terminal."
}

Write-Host "Portkill installed successfully!"
Write-Host "Run: portkill --version"