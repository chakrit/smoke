#!/bin/sh

set -e

go install -v .
smoke="$(go env GOPATH)/bin/smoke"

# Real suites live at the top of test/; fixtures they drive live in test/testdata/
# (intentionally RED/NEW/malformed, so never run them directly). Globs don't recurse,
# so testdata/ is excluded for free. Skip lockfiles — they're outputs, not suites.
for spec in test/*.yml test/*.cue test/*.json test/*.jsonl; do
  [ -e "$spec" ] || continue
  case "$spec" in
  *.lock.yml) continue ;;
  esac

  "$smoke" "$@" "$spec"
done
