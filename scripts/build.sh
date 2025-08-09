#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="$(pwd)/bin"
mkdir -p "$BIN_DIR"

go build -o "$BIN_DIR/git-sweep" ./cmd/git-sweep

echo "Built $BIN_DIR/git-sweep"
