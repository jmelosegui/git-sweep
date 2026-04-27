# git-sweep installer for Windows
# Usage:
#   irm https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.ps1 | iex
#   & ([scriptblock]::Create((irm https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.ps1))) -Prerelease

[CmdletBinding()]
param(
    [switch]$Prerelease
)

$ErrorActionPreference = "Stop"

$Repo = "jmelosegui/git-sweep"
$InstallDir = if ($env:GIT_SWEEP_INSTALL_DIR) { $env:GIT_SWEEP_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "git-sweep" }

function Write-Info { param($msg) Write-Host "==> " -ForegroundColor Green -NoNewline; Write-Host $msg }
function Write-Warn { param($msg) Write-Host "warning: " -ForegroundColor Yellow -NoNewline; Write-Host $msg }
function Write-Err  { param($msg) Write-Host "error: " -ForegroundColor Red -NoNewline; Write-Host $msg; exit 1 }

if ($Prerelease) {
    Write-Info "Installing git-sweep (including pre-releases)..."
} else {
    Write-Info "Installing git-sweep..."
}

# Detect architecture
$Arch = switch -Wildcard ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { Write-Err "unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)" }
}
Write-Info "Detected: windows-$Arch"

# Resolve release tag
$Headers = @{ "User-Agent" = "git-sweep-installer" }
try {
    if ($Prerelease) {
        $Releases = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases?per_page=1" -Headers $Headers
        if (-not $Releases -or $Releases.Count -eq 0) {
            Write-Err "no releases found. See https://github.com/$Repo/releases"
        }
        $Tag = $Releases[0].tag_name
    } else {
        $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers $Headers
        $Tag = $Release.tag_name
    }
} catch {
    Write-Err "could not determine release tag. See https://github.com/$Repo/releases"
}

if (-not $Tag) {
    Write-Err "could not determine release tag. See https://github.com/$Repo/releases"
}

Write-Info "Release: $Tag"

# Goreleaser archive name: git-sweep_<version-without-v>_windows_<arch>.zip
$Version = if ($Tag.StartsWith("v")) { $Tag.Substring(1) } else { $Tag }
$Filename = "git-sweep_${Version}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Filename"

$TempBase = (Get-Item $env:TEMP).FullName
$TempDir = Join-Path $TempBase "git-sweep-install-$PID"
$ZipPath = Join-Path $TempDir $Filename
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null

try {
    Write-Info "Downloading $Url..."
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing -Headers $Headers

    # Optional checksum verification
    $ChecksumsUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
    $ChecksumsPath = Join-Path $TempDir "checksums.txt"
    try {
        Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $ChecksumsPath -UseBasicParsing -Headers $Headers -ErrorAction Stop
        Write-Info "Verifying checksum..."
        $expectedLine = (Get-Content $ChecksumsPath | Where-Object { $_ -match "\s$([regex]::Escape($Filename))\s*$" } | Select-Object -First 1)
        if ($expectedLine) {
            $expected = ($expectedLine -split '\s+')[0]
            $actual = (Get-FileHash $ZipPath -Algorithm SHA256).Hash.ToLowerInvariant()
            if ($expected.ToLowerInvariant() -ne $actual) {
                Write-Err "checksum mismatch for $Filename"
            }
        } else {
            Write-Warn "no checksum entry for $Filename; skipping verification"
        }
    } catch {
        # No checksums published; skip silently
    }

    Write-Info "Extracting..."
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    # Stop any running instance before replacing
    Get-Process -Name "git-sweep" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

    Write-Info "Installing to $InstallDir..."
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path (Join-Path $TempDir "git-sweep.exe") -Destination (Join-Path $InstallDir "git-sweep.exe") -Force
    Unblock-File -Path (Join-Path $InstallDir "git-sweep.exe")
} finally {
    if (Test-Path $TempDir) {
        Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
    }
}

# Add to user PATH if missing
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $UserPath) { $UserPath = "" }
$pathSegments = $UserPath -split ';' | Where-Object { $_ -ne "" }
if ($pathSegments -notcontains $InstallDir) {
    Write-Info "Adding $InstallDir to user PATH..."
    $newPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Info "Restart your shell to pick up the updated PATH."
}

Write-Info "Successfully installed git-sweep!"
& (Join-Path $InstallDir "git-sweep.exe") --version
