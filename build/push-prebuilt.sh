#!/usr/bin/env bash
# Usage: build/push-prebuilt.sh <target>
#
# Publishes .work/out/<target> as a single orphan commit, force-pushed to the
# branch prebuilt/<target>. History is intentionally discarded: prebuilt
# branches never grow, so the repository stays small. Release tags
# (libs/<goos>_<goarch>/vX.Y.Z) pin the snapshots that must stay reachable.
#
# The commit is created inside the main repository's object database so the
# push reuses whatever credentials the current checkout already has (both
# locally and inside GitHub Actions).
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

cv2_load_target "${1:?usage: push-prebuilt.sh <target>}"

[ -f "$CV2_OUT_DIR/MANIFEST" ] || { echo "error: $CV2_OUT_DIR not packaged yet" >&2; exit 1; }

branch=$CV2_PREBUILT_BRANCH

export GIT_AUTHOR_NAME=${GIT_AUTHOR_NAME:-github-actions[bot]}
export GIT_AUTHOR_EMAIL=${GIT_AUTHOR_EMAIL:-41898282+github-actions[bot]@users.noreply.github.com}
export GIT_COMMITTER_NAME=$GIT_AUTHOR_NAME
export GIT_COMMITTER_EMAIL=$GIT_AUTHOR_EMAIL

# Stage the out dir with a throwaway index, write a tree, commit it parentless.
export GIT_INDEX_FILE=$CV2_WORK/prebuilt-index-$OPENCV_VERSION-$CV2_TARGET
rm -f "$GIT_INDEX_FILE"
git -C "$CV2_ROOT" --work-tree="$CV2_OUT_DIR" add -Af .
tree=$(git -C "$CV2_ROOT" write-tree)
unset GIT_INDEX_FILE

msg="prebuilt: $CV2_TARGET (opencv $OPENCV_VERSION, modules $OPENCV_MODULES)

$(cat "$CV2_OUT_DIR/MANIFEST")"

commit=$(git -C "$CV2_ROOT" commit-tree "$tree" -m "$msg")

echo "==> Force-pushing $commit to $branch"
git -C "$CV2_ROOT" push --force origin "$commit:refs/heads/$branch"
