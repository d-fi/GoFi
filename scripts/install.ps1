# GoFi Installation Script for Windows
# This script downloads and installs the latest GoFi binary for Windows

$ErrorActionPreference = "Stop"

# Configuration
$repo = "d-fi/GoFi"
$binaryName = "gofi.exe"
$installDir = "$env:LOCALAPPDATA\Programs\GoFi"

# Colors for output
function Write-Success { param($Message) Write-Host $Message -ForegroundColor Green }
function Write-Info { param($Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Error { param($Message) Write-Host $Message -ForegroundColor Red }

Write-Success "GoFi Installation Script for Windows"
Write-Host "===================================="

# Get the latest release
function Get-LatestRelease {
    Write-Info "Fetching latest release..."
    
    try {
        $apiUrl = "https://api.github.com/repos/$repo/releases/latest"
        $release = Invoke-RestMethod -Uri $apiUrl -Headers @{"Accept"="application/vnd.github.v3+json"}
        
        # Find the Windows binary
        $asset = $release.assets | Where-Object { $_.name -eq "gofi-windows-amd64.zip" }
        
        if (-not $asset) {
            throw "Could not find Windows binary in latest release"
        }
        
        return $asset.browser_download_url
    }
    catch {
        Write-Error "Failed to get latest release: $_"
        exit 1
    }
}

# Download and install
function Install-GoFi {
    param($DownloadUrl)
    
    # Create temp directory
    $tempDir = New-TemporaryFile | %{ Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $zipPath = Join-Path $tempDir "gofi.zip"
    
    try {
        # Download
        Write-Info "Downloading GoFi..."
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $zipPath
        
        # Extract
        Write-Info "Extracting archive..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
        
        # Create install directory
        if (-not (Test-Path $installDir)) {
            Write-Info "Creating install directory: $installDir"
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }
        
        # Find the executable
        $exePath = Get-ChildItem -Path $tempDir -Filter "*.exe" -Recurse | Select-Object -First 1
        
        if (-not $exePath) {
            throw "Could not find executable in archive"
        }
        
        # Copy to install directory
        $targetPath = Join-Path $installDir $binaryName
        Write-Info "Installing to: $targetPath"
        Copy-Item -Path $exePath.FullName -Destination $targetPath -Force
        
        # Add to PATH if not already there
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$installDir*") {
            Write-Info "Adding GoFi to PATH..."
            [Environment]::SetEnvironmentVariable(
                "Path",
                "$userPath;$installDir",
                "User"
            )
            $env:Path = "$env:Path;$installDir"
            Write-Success "Added $installDir to PATH"
            Write-Info "You may need to restart your terminal for PATH changes to take effect"
        }
        
        Write-Success "✓ GoFi has been installed successfully!"
        Write-Success "Installation location: $targetPath"
        
        # Verify installation
        if (Get-Command gofi -ErrorAction SilentlyContinue) {
            $version = & gofi --version 2>$null
            Write-Success "GoFi is available in PATH"
            if ($version) {
                Write-Success "Version: $version"
            }
        } else {
            Write-Info "Run 'gofi --help' to get started (you may need to restart your terminal first)"
        }
    }
    finally {
        # Clean up
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Main
try {
    # Check if running as administrator (optional, but recommended)
    $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
    if (-not $isAdmin) {
        Write-Info "Note: Running without administrator privileges. Installation will be for current user only."
    }
    
    # Allow custom install directory
    if ($args.Count -gt 0) {
        $installDir = $args[0]
        Write-Info "Using custom install directory: $installDir"
    }
    
    $downloadUrl = Get-LatestRelease
    Write-Success "Download URL: $downloadUrl"
    
    Install-GoFi -DownloadUrl $downloadUrl
}
catch {
    Write-Error "Installation failed: $_"
    exit 1
}