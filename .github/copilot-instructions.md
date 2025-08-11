# git-sweep
git-sweep is a cross-platform Go CLI that removes local branches whose upstream has been removed (e.g., branches marked as "[gone]"). It's safe by default, supports dry-run mode, and is designed for Windows/macOS/Linux.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Prerequisites and Setup
- Go 1.24.x or later is REQUIRED. Check with `go version`.
- Install required tools:
  - `go install mvdan.cc/gofumpt@latest` -- for code formatting
  - `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest` -- for linting (takes ~3 seconds)
  - Add `$(go env GOPATH)/bin` to your PATH if not already there

### Build and Development Commands
- Download dependencies: `go mod download` -- takes ~2 seconds
- Build: `make build` -- takes ~12 seconds. NEVER CANCEL. Use timeout of 60+ seconds.
  - Alternative: `./scripts/build.sh` (Linux/macOS) or `./scripts/build.ps1` (Windows) -- takes <1 second
  - Binary outputs to `bin/git-sweep` (or `bin/git-sweep.exe` on Windows)
- Format code: `make fmt` -- takes <1 second. ALWAYS run before committing.
- Lint: `make lint` -- takes ~3 seconds. NEVER CANCEL. Use timeout of 300+ seconds.
- Clean: `make clean` -- removes bin/ directory

### Testing
- Run all tests: `make test` -- takes ~3 seconds. NEVER CANCEL. Use timeout of 300+ seconds.
- Run with race detection: `go test -race ./...` -- takes ~11 seconds. NEVER CANCEL. Use timeout of 300+ seconds.
- Run integration tests only: `go test -v ./test/integration` -- takes <1 second. Tests real git scenarios.
- Alternative: `./scripts/test.ps1` (Windows) -- automatically handles race detection based on CGO_ENABLED

### Validation Scenarios
After making changes, ALWAYS validate functionality:
1. Build the application: `make build`
2. Test basic functionality: `./bin/git-sweep --version` and `./bin/git-sweep --help`
3. Test in a git repository: `./bin/git-sweep --json` (shows current branch status)
4. Run full test suite: `make test`
5. Run linting: `make lint`
6. Format code: `make fmt`

## Application Usage
- Dry-run (default): `git sweep` or `./bin/git-sweep`
- Execute deletions: `git sweep -y` or `./bin/git-sweep -y`
- JSON output: `git sweep --json` or `./bin/git-sweep --json`
- Filter branches: `git sweep -i "pattern"` (include) or `git sweep -x "pattern"` (exclude)
- Use different remote: `git sweep -r upstream`

## Validation Requirements
- ALWAYS run `make fmt` and `make lint` before committing or the CI (.github/workflows/ci.yml) will fail
- ALWAYS test the built binary with `--version`, `--help`, and `--json` flags
- Integration tests validate real git scenarios with remote repositories
- The application requires a valid git repository with a remote to function properly

## Common Tasks and Directory Structure

### Repository Root Files
```
.
├── README.md           -- main documentation
├── LICENSE            -- MIT license
├── Makefile           -- primary build commands
├── go.mod             -- Go module definition (requires go 1.24)
├── go.sum             -- dependency checksums
├── .golangci.yml      -- linter configuration
├── .goreleaser.yaml   -- release configuration
└── .github/           -- GitHub workflows and templates
    └── workflows/
        ├── ci.yml         -- main CI pipeline
        ├── release.yml    -- automated releases
        └── golangci-lint.yml  -- dedicated linting workflow
```

### Source Code Structure
```
├── cmd/
│   └── git-sweep/
│       └── main.go        -- CLI entry point
├── internal/
│   ├── config/           -- configuration handling
│   ├── git/              -- git command execution
│   ├── logging/          -- logging utilities
│   ├── sweep/            -- core sweep logic
│   └── ui/               -- user interface and output
├── scripts/              -- cross-platform helper scripts
│   ├── build.sh         -- Linux/macOS build script
│   ├── build.ps1        -- Windows build script
│   ├── test.ps1         -- Windows test script
│   └── lint.ps1         -- Windows lint script
└── test/
    └── integration/      -- end-to-end integration tests
        └── git_sweep_integration_test.go
```

### Key Build Files Content
go.mod:
```
module github.com/jmelosegui/git-sweep

go 1.24

require github.com/spf13/pflag v1.0.5
```

Makefile key targets:
- `build` -- builds to bin/git-sweep
- `test` -- runs go test ./...
- `lint` -- runs golangci-lint run
- `fmt` -- runs gofumpt -l -w .
- `clean` -- removes bin/ directory

## Troubleshooting
- If `golangci-lint` fails with version mismatch, install latest version with the curl command above
- If `gofumpt` command not found, run `go install mvdan.cc/gofumpt@latest`
- Scripts may need execute permissions on Linux/macOS: `chmod +x scripts/*.sh`
- Git operations require a valid git repository with remote configured
- Integration tests create temporary git repositories and don't require existing remotes

## CI/CD Information
- GitHub Actions runs tests on Ubuntu, macOS, and Windows with Go 1.24.x
- Tests run with race detection on Linux/macOS, without race detection on Windows
- golangci-lint runs with 3-minute timeout
- Releases use goreleaser for cross-platform builds (Linux, macOS, Windows on amd64/arm64)