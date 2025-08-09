$ErrorActionPreference = 'Stop'

# Only use -race when CGO is enabled; otherwise run without it
$useRace = $false
try {
  $cgo = & go env CGO_ENABLED
  if ($cgo -eq '1') { $useRace = $true }
} catch {
  $useRace = $false
}

if ($useRace) {
  Write-Host "Running tests with race detector (CGO_ENABLED=1)"
  go test -race ./...
} else {
  Write-Host "Running tests without race detector (CGO not enabled)"
  go test ./...
}
