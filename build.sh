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

LDFLAGS=("-X main.appVersion=${APP_VERSION}")

if [[ -n "${TELEGRAM_APP_ID_BAKED:-}" ]]; then
  LDFLAGS+=("-X main.telegramAppIDBaked=${TELEGRAM_APP_ID_BAKED}")
fi

if [[ -n "${TELEGRAM_APP_HASH_BAKED:-}" ]]; then
  LDFLAGS+=("-X main.telegramAppHashBaked=${TELEGRAM_APP_HASH_BAKED}")
fi

go build -ldflags "${LDFLAGS[*]}" -o "${BINARY_NAME}" .
echo "Built ${BINARY_NAME} (${APP_VERSION})"

if [[ -n "${TELEGRAM_APP_ID_BAKED:-}" || -n "${TELEGRAM_APP_HASH_BAKED:-}" ]]; then
  echo "Included baked Telegram credentials via ldflags"
fi
