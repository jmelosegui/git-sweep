param(
  [switch]$Fix
)

$ErrorActionPreference = 'Stop'

$exe = 'golangci-lint'
if (-not (Get-Command $exe -ErrorAction SilentlyContinue)) {
  Write-Host "golangci-lint not found. Install from https://golangci-lint.run/" -ForegroundColor Yellow
  exit 1
}

$args = @('run')
if ($Fix) { $args += '--fix' }

& $exe @args
