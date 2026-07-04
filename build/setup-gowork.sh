#!/usr/bin/env bash
# Creates a go.work that overlays the fetched .prebuilt/libs modules onto the
# root module, so `go build` / `go test` work before (or regardless of) the
# release tags referenced in go.mod.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

cd "$CV2_ROOT"
rm -f go.work go.work.sum
go work init .
found=0
for d in .prebuilt/libs/*/; do
  if [ -f "$d/go.mod" ]; then
    go work use "./${d%/}"
    found=1
  fi
done
if [ "$found" = 0 ]; then
  echo "warning: no modules under .prebuilt/libs; run build/fetch-prebuilt.sh first" >&2
fi
echo "==> go.work ready:"
cat go.work
