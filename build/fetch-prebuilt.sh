#!/usr/bin/env bash
# Usage: build/fetch-prebuilt.sh [--require-fresh] <target>
#
# Downloads the libs module published on the prebuilt/<target> branch into
# .prebuilt/libs/<goos>_<goarch>, for local development and CI testing
# (combined with build/setup-gowork.sh).
#
# Exit codes:
#   0  fetched
#   3  the prebuilt branch does not exist yet
#   4  (--require-fresh only) the branch exists but was built from different
#      wrapper sources than this checkout (WRAPPER_KEY mismatch)
#   other  a real git/network failure - never masked as "missing"
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

require_fresh=0
if [ "${1:-}" = "--require-fresh" ]; then
  require_fresh=1
  shift
fi

cv2_load_target "${1:?usage: fetch-prebuilt.sh [--require-fresh] <target>}"
branch=prebuilt/$CV2_TARGET

# git ls-remote --exit-code: 2 = ref absent, anything else non-zero = real
# failure (network, auth, ...). Only the former may be treated as "missing",
# otherwise CI would silently skip tests on transient errors.
status=0
git -C "$CV2_ROOT" ls-remote --exit-code origin "refs/heads/$branch" >/dev/null || status=$?
if [ "$status" -eq 2 ]; then
  echo "prebuilt branch '$branch' does not exist yet (run the build-libs workflow first)" >&2
  exit 3
elif [ "$status" -ne 0 ]; then
  echo "error: git ls-remote failed (exit $status) while checking for '$branch'" >&2
  exit "$status"
fi

git -C "$CV2_ROOT" fetch --depth 1 origin "refs/heads/$branch"

if [ "$require_fresh" = 1 ]; then
  branch_key=$(git -C "$CV2_ROOT" show FETCH_HEAD:MANIFEST | sed -n 's/^WRAPPER_KEY=//p')
  local_key=$(cv2_build_key "$CV2_TARGET" wrapper)
  if [ "$branch_key" != "$local_key" ]; then
    echo "prebuilt branch '$branch' is stale: wrapper key $branch_key does not match local $local_key" >&2
    exit 4
  fi
fi

mkdir -p "$CV2_ROOT/.prebuilt"
rm -rf "$CV2_ROOT/.prebuilt/libs/${TARGET_GOOS}_${TARGET_GOARCH}"
git -C "$CV2_ROOT" archive FETCH_HEAD libs | tar -x -C "$CV2_ROOT/.prebuilt"
echo "==> Fetched $branch into .prebuilt/libs/${TARGET_GOOS}_${TARGET_GOARCH}"
