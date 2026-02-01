#Requires -Version 5.1
<#
.SYNOPSIS
    Install Pact CLI on Windows
.DESCRIPTION
    Downloads and installs the latest Pact CLI binary from GitHub releases
.EXAMPLE
    iwr -useb https://pact-dev.com/install.ps1 | iex
#>

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "cloudboy-jh/pact"
$BinaryName = "pact.exe"

# Colors for output
function Write-Info($message) {
    Write-Host "→ $message" -ForegroundColor Green
}

function Write-Warn($message) {
    Write-Host "⚠ $message" -ForegroundColor Yellow
}

function Write-Error($message) {
    Write-Host "✗ $message" -ForegroundColor Red
}

function Write-Success($message) {
    Write-Host "✓ $message" -ForegroundColor Green
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { 
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Get latest version from GitHub
function Get-LatestVersion {
    Write-Info "Checking for latest version..."
    
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Method Get
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to get latest version: $_"
        exit 1
    }
}

# Download file with progress
function Download-File($url, $output) {
    Write-Info "Downloading from $url..."
    
    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $url -OutFile $output -UseBasicParsing
        $ProgressPreference = 'Continue'
    }
    catch {
        Write-Error "Download failed: $_"
        exit 1
    }
}

# Main installation
function Install-Pact {
    Write-Host ""
    Write-Host "Installing Pact CLI..." -ForegroundColor Cyan
    Write-Host ""

    # Detect system info
    $arch = Get-Architecture
    $version = Get-LatestVersion
    $versionNum = $version -replace '^v', ''

    Write-Info "Architecture: $arch"
    Write-Info "Version: $version"
    Write-Host ""

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "pact-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        # Download
        $filename = "pact_${versionNum}_windows_${arch}.zip"
        $url = "https://github.com/$Repo/releases/download/$version/$filename"
        $zipPath = Join-Path $tempDir $filename

        Download-File $url $zipPath

        # Extract
        Write-Info "Extracting..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

        # Find the binary
        $binaryPath = Join-Path $tempDir $BinaryName
        if (-not (Test-Path $binaryPath)) {
            Write-Error "Binary not found in archive"
            exit 1
        }

        # Determine install location
        $installDir = $null
        
        # Check if Scoop is installed
        $scoopPath = Get-Command scoop -ErrorAction SilentlyContinue
        if ($scoopPath) {
            Write-Info "Scoop detected - installing via Scoop..."
            
            # Check if pact bucket is added
            $buckets = scoop bucket list 2>$null
            if ($buckets -notmatch "pact-bucket") {
                Write-Info "Adding pact bucket..."
                scoop bucket add pact-bucket https://github.com/cloudboy-jh/pact-bucket
            }
            
            # Install/update via scoop
            scoop install pact
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Pact installed successfully via Scoop!"
                Write-Host ""
                Write-Host "Run 'pact --help' to get started" -ForegroundColor Cyan
                return
            }
            else {
                Write-Warn "Scoop install failed, falling back to direct install..."
            }
        }

        # Direct install to user's bin directory
        $userBin = Join-Path $env:USERPROFILE "bin"
        if (-not (Test-Path $userBin)) {
            Write-Info "Creating $userBin directory..."
            New-Item -ItemType Directory -Path $userBin -Force | Out-Null
        }

        $installPath = Join-Path $userBin $BinaryName

        # Copy binary
        Write-Info "Installing to $installPath..."
        Copy-Item -Path $binaryPath -Destination $installPath -Force

        # Verify installation
        $installedVersion = & $installPath --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Pact installed successfully!"
            Write-Host ""
            Write-Host "Version: $installedVersion" -ForegroundColor Gray
            Write-Host ""
            
            # Check if bin directory is in PATH
            $pathDirs = $env:PATH -split ';'
            $binInPath = $pathDirs | Where-Object { $_ -eq $userBin }
            
            if (-not $binInPath) {
                Write-Warn "$userBin is not in your PATH"
                Write-Host ""
                Write-Host "Add the following to your PowerShell profile:" -ForegroundColor Yellow
                Write-Host '$env:PATH = "$env:USERPROFILE\bin;$env:PATH"' -ForegroundColor Cyan
                Write-Host ""
                Write-Host "Or run this command to add it now:" -ForegroundColor Yellow
                Write-Host "[Environment]::SetEnvironmentVariable('PATH', `"`$env:USERPROFILE\bin;`$env:PATH`", 'User')" -ForegroundColor Cyan
                Write-Host ""
            }

            Write-Host "Run 'pact --help' to get started" -ForegroundColor Cyan
        }
        else {
            Write-Error "Installation verification failed"
            exit 1
        }
    }
    finally {
        # Cleanup
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Run installation
Install-Pact
