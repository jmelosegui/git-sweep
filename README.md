# git-sweep

A small, cross-platform Go CLI that removes local branches whose upstream has been removed (e.g., branches marked as "[gone]"). Safe by default, dry-run first, and designed for Windows/macOS/Linux.

## Installation

### One-line install

**Linux / macOS**
```sh
curl -fsSL https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.sh | bash
```
Pass `-s -- --prerelease` to opt into the latest pre-release tag:
```sh
curl -fsSL https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.sh | bash -s -- --prerelease
```
The script installs to `~/.git-sweep/bin` (override with `INSTALL_DIR=...`) and adds that directory to your shell rc (`bash`, `zsh`, `fish`, or `~/.profile` as a fallback). Reload your shell or `source` the rc file afterwards.

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.ps1 | iex
```
For pre-releases, run the script with the `-Prerelease` switch:
```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/jmelosegui/git-sweep/main/scripts/install.ps1))) -Prerelease
```
The script installs to `%LOCALAPPDATA%\git-sweep` (override with `$env:GIT_SWEEP_INSTALL_DIR`) and updates your user `Path`. Restart your terminal afterwards.

### Manual install from GitHub Releases (including pre-releases)
1) Go to the [repository Releases page](https://github.com/jmelosegui/git-sweep/releases) and pick a version. To use a pre-release, select a tag labeled “Pre-release”.
2) Download the archive matching your OS/arch:
   - linux: `git-sweep_<version>_linux_amd64.tar.gz` (or `_arm64`)
   - macOS: `git-sweep_<version>_darwin_amd64.tar.gz` (or `_arm64`)
   - Windows: `git-sweep_<version>_windows_amd64.zip` (or `_arm64`)
3) (Recommended) Verify checksum:
   - Download `checksums.txt` from the same release
   - macOS/Linux:
     ```
     # Replace <version>, <os>, and <arch> with the actual values for your downloaded file.
     # Example for version 1.2.3 on Linux amd64:
     shasum -a 256 -c checksums.txt | grep git-sweep_1.2.3_linux_amd64
     # Or, for your specific file:
     shasum -a 256 -c checksums.txt | grep git-sweep_<version>_<os>_<arch>
     ```
   - Windows PowerShell:
     ```powershell
     Get-FileHash .\git-sweep_<version>_windows_amd64.zip -Algorithm SHA256
     # Compare against the corresponding line in checksums.txt
     ```
4) Extract:
   - macOS/Linux:
     ```sh
     tar -xzf git-sweep_<version>_<os>_<arch>.tar.gz
     ```
   - Windows PowerShell:
     ```powershell
     Expand-Archive .\git-sweep_<version>_windows_amd64.zip -DestinationPath .\git-sweep
     ```
5) Move the binary into your PATH:
   - macOS/Linux (example):
     ```sh
     sudo mv git-sweep /usr/local/bin/
     ```
   - Windows (example): add the extracted folder to the "Path" user environment variable, or move `git-sweep.exe` into a directory already on PATH.
6) Verify:
   ```sh
   git sweep -V
   ```

## Usage

Dry-run (default):
```sh
git sweep
```

Execute deletions (interactive confirm if TTY):
```sh
git sweep -y
```

JSON plan output:
```sh
git sweep --json
```

### Update notifications

`git-sweep` checks the GitHub Releases API at most once every 24 hours and prints a one-line notice on `stderr` when a newer version is available. The check is skipped automatically when `--json` is set, when `stderr` is not a terminal (CI, redirects), and for the development build. To opt out entirely, set `GIT_SWEEP_NO_UPDATE_CHECK=1`.
