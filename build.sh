#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="${BINARY_NAME:-termigram}"

if [[ -n "${VERSION:-}" ]]; then
  APP_VERSION="$VERSION"
elif APP_VERSION="$(git describe --tags --dirty 2>/dev/null)"; then
  :
else
  APP_VERSION="dev"
fi

go build -ldflags "-X main.appVersion=${APP_VERSION}" -o "${BINARY_NAME}" .
echo "Built ${BINARY_NAME} (${APP_VERSION})"
