#!/bin/sh

LOGLEVEL="${LOGLEVEL:-info}"
DATA_DIR="${DATA_DIR:-/var/lib/cloudflare-warp}"

run() {
  # execute extra commands
  if [ -n "$EXTRA_COMMANDS" ]; then
    sh -c "$EXTRA_COMMANDS"
  fi

  exec warp run \
    --data-dir "$DATA_DIR" \
    --loglevel "$LOGLEVEL" \
    "$@"
}

run "$@" || exit 1
