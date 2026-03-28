#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> Running go vet..."
go vet ./...

echo "==> Running golangci-lint..."
golangci-lint run

echo "==> Running tests..."
go test ./...

echo "==> All checks passed."
