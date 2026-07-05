#!/usr/bin/env bash
# Usage: build/release.sh <version-without-v>
#
# Cuts a release:
#   1. verifies every prebuilt/<target> branch exists and was built from the
#      current wrapper sources (WRAPPER_KEY match);
#   2. tags each libs module (libs/<goos>_<goarch>/v<version>) at the head of
#      its prebuilt branch and pushes the tags;
#   3. pins those versions in the root go.mod, refreshes go.sum, commits,
#      tags v<version> and pushes.
#
# Meant to run inside the release GitHub Actions workflow, but works locally
# too as long as `git push` is authorized.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

version=${1:?usage: release.sh <module-version-without-v, e.g. 0.40801.0>}
# Module versions encode the OpenCV line: 0.<code>.<revision> where
# code = major*10000 + minor*100 + patch (4.8.1 -> 40801, 4.12.0 -> 41200).
if ! [[ "$version" =~ ^0\.([1-9][0-9]{4,5})\.(0|[1-9][0-9]*)(-[0-9A-Za-z][0-9A-Za-z.-]*)?$ ]]; then
  echo "error: '$version' is not a valid release version; expected 0.<opencv-code>.<revision>, e.g. 0.40801.0 for OpenCV 4.8.1" >&2
  exit 1
fi
opencv_code=${BASH_REMATCH[1]}
opencv_version=$(cv2_opencv_version_from_code "$opencv_code")
# shellcheck disable=SC1091
source "$CV2_ROOT/build/build.conf"
case " $OPENCV_VERSIONS " in
*" $opencv_version "*) ;;
*)
  echo "error: version code $opencv_code decodes to OpenCV $opencv_version, which is not in OPENCV_VERSIONS ($OPENCV_VERSIONS)" >&2
  exit 1
  ;;
esac
echo "==> Releasing v$version (OpenCV $opencv_version line)"

# Refuse to release from anything but a branch: in CI a tag/SHA dispatch, or
# locally a detached HEAD, would push the go.mod pin commit to a junk ref.
if [ -n "${GITHUB_REF_TYPE:-}" ] && [ "$GITHUB_REF_TYPE" != "branch" ]; then
  echo "error: release must be dispatched from a branch, not a $GITHUB_REF_TYPE" >&2
  exit 1
fi
push_ref=${GITHUB_REF_NAME:-}
if [ -z "$push_ref" ]; then
  push_ref=$(git -C "$CV2_ROOT" symbolic-ref --quiet --short HEAD || true)
fi
if [ -z "$push_ref" ]; then
  echo "error: detached HEAD and no GITHUB_REF_NAME; check out a branch before releasing" >&2
  exit 1
fi

cd "$CV2_ROOT"
# The + prefix forces the update: prebuilt branches are force-pushed orphan
# commits by design, so they never fast-forward from a previous fetch.
git fetch origin '+refs/heads/prebuilt/*:refs/cv2-prebuilt/*'

tags=()
for env_file in build/targets/*.env; do
  target=$(basename "$env_file" .env)
  ref=refs/cv2-prebuilt/$opencv_version/$target
  if ! git rev-parse --verify -q "$ref" >/dev/null; then
    echo "error: branch prebuilt/$opencv_version/$target is missing; run the build-libs workflow first" >&2
    exit 1
  fi

  manifest=$(git show "$ref:MANIFEST")
  branch_wrapper_key=$(sed -n 's/^WRAPPER_KEY=//p' <<<"$manifest")
  CV2_OPENCV_VERSION=$opencv_version cv2_load_target "$target"
  expected_wrapper_key=$(cv2_build_key "$target" wrapper)
  if [ "$branch_wrapper_key" != "$expected_wrapper_key" ]; then
    echo "error: prebuilt/$opencv_version/$target is stale (wrapper key mismatch); re-run build-libs first" >&2
    exit 1
  fi

  # Tag every libs module the branch carries (base plus feature sets).
  while IFS= read -r module_dir; do
    [ -n "$module_dir" ] || continue
    tag=libs/$module_dir/v$version
    if git rev-parse --verify -q "refs/tags/$tag" >/dev/null; then
      echo "error: tag $tag already exists; module versions are immutable, bump the version" >&2
      exit 1
    fi
    git tag "$tag" "$ref"
    tags+=("$tag")
  done < <(git ls-tree --name-only "$ref:libs")
done

echo "==> Pushing libs tags: ${tags[*]}"
git push origin "${tags[@]/#/refs/tags/}"

echo "==> Pinning libs modules in go.mod"
for tag in "${tags[@]}"; do
  module=github.com/hkloudou/cv2/${tag%/v$version}
  go mod edit -require="$module@v$version"
done

echo "==> Refreshing go.sum (waits for the module proxy to index the new tags)"
ok=0
for attempt in $(seq 1 30); do
  if go mod tidy; then
    ok=1
    break
  fi
  echo "go mod tidy failed (attempt $attempt); the proxy may not have indexed the tags yet, retrying in 10s"
  sleep 10
done
if [ "$ok" != 1 ]; then
  echo "error: go mod tidy did not succeed; releasing aborted before the root tag" >&2
  exit 1
fi

git add go.mod go.sum
export GIT_AUTHOR_NAME=${GIT_AUTHOR_NAME:-github-actions[bot]}
export GIT_AUTHOR_EMAIL=${GIT_AUTHOR_EMAIL:-41898282+github-actions[bot]@users.noreply.github.com}
export GIT_COMMITTER_NAME=$GIT_AUTHOR_NAME
export GIT_COMMITTER_EMAIL=$GIT_AUTHOR_EMAIL
git commit -m "release: v$version"
git tag "v$version"

echo "==> Pushing release commit and tag to $push_ref"
git push origin "HEAD:refs/heads/$push_ref"
git push origin "refs/tags/v$version"

echo "==> Released v$version"
