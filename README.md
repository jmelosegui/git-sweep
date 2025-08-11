# git-sweep

A small, cross-platform Go CLI that removes local branches whose upstream has been removed (e.g., branches marked as "[gone]"). Safe by default, dry-run first, and designed for Windows/macOS/Linux.

Status: Actively developed. See `PLAN.md` for the roadmap.

## Install

- Windows (winget):
  ```powershell
  winget install jmelosegui.git-sweep
  ```

- Direct download (macOS/Linux/Windows):
  - Download the latest archive for your OS/arch from the Releases page
  - Extract and place `git-sweep` (or `git-sweep.exe`) in your PATH

Notes:
- Homebrew Core (macOS) submission is planned; once accepted you’ll be able to:
  ```sh
  brew install git-sweep
  ```
- Linux universal channels (Snap/Nix) are planned; for now, use direct downloads.

### Manual install from GitHub Releases (including pre-releases)
1) Go to the [repository Releases page](https://github.com/jmelosegui/git-sweep/releases) and pick a version. To use a pre-release, select a tag labeled “Pre-release”.
2) Download the archive matching your OS/arch:
   - linux: `git-sweep_<version>_linux_amd64.tar.gz` (or `_arm64`)
   - macOS: `git-sweep_<version>_darwin_amd64.tar.gz` (or `_arm64`)
   - Windows: `git-sweep_<version>_windows_amd64.zip` (or `_arm64`)
3) (Recommended) Verify checksum:
   - Download `checksums.txt` from the same release
   - macOS/Linux:
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