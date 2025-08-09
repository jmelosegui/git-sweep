$ErrorActionPreference = 'Stop'

$bin = Join-Path (Get-Location) 'bin'
if (-not (Test-Path $bin)) { New-Item -ItemType Directory -Path $bin | Out-Null }

$exe = Join-Path $bin 'git-sweep.exe'

go build -o $exe ./cmd/git-sweep

Write-Host "Built $exe"
