#!/usr/bin/env bash
# Usage: build/fetch-prebuilt.sh <target>
#
# Downloads the libs module published on the prebuilt/<target> branch into
# .prebuilt/libs/<goos>_<goarch>, for local development and CI testing
# (combined with build/setup-gowork.sh).
#
# Exit code 3 means the prebuilt branch does not exist yet.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

cv2_load_target "${1:?usage: fetch-prebuilt.sh <target>}"
branch=prebuilt/$CV2_TARGET

if ! git -C "$CV2_ROOT" ls-remote --exit-code origin "refs/heads/$branch" >/dev/null 2>&1; then
  echo "prebuilt branch '$branch' does not exist yet (run the build-libs workflow first)" >&2
  exit 3
fi

git -C "$CV2_ROOT" fetch --depth 1 origin "refs/heads/$branch"
mkdir -p "$CV2_ROOT/.prebuilt"
rm -rf "$CV2_ROOT/.prebuilt/libs/${TARGET_GOOS}_${TARGET_GOARCH}"
git -C "$CV2_ROOT" archive FETCH_HEAD libs | tar -x -C "$CV2_ROOT/.prebuilt"
echo "==> Fetched $branch into .prebuilt/libs/${TARGET_GOOS}_${TARGET_GOARCH}"
