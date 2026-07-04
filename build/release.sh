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

version=${1:?usage: release.sh <version-without-v>}
if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-.+][0-9A-Za-z.+-]+)?$ ]]; then
  echo "error: '$version' is not a semantic version (expected e.g. 0.1.0)" >&2
  exit 1
fi
if [[ "$version" =~ ^[2-9]|^1[0-9] ]]; then
  echo "error: major versions >= 2 need a /vN module path suffix; adjust tooling first" >&2
  exit 1
fi

cd "$CV2_ROOT"
git fetch origin 'refs/heads/prebuilt/*:refs/cv2-prebuilt/*'

tags=()
for env_file in build/targets/*.env; do
  target=$(basename "$env_file" .env)
  ref=refs/cv2-prebuilt/$target
  if ! git rev-parse --verify -q "$ref" >/dev/null; then
    echo "error: branch prebuilt/$target is missing; run the build-libs workflow first" >&2
    exit 1
  fi

  manifest=$(git show "$ref:MANIFEST")
  branch_wrapper_key=$(sed -n 's/^WRAPPER_KEY=//p' <<<"$manifest")
  expected_wrapper_key=$(cv2_build_key "$target" wrapper)
  if [ "$branch_wrapper_key" != "$expected_wrapper_key" ]; then
    echo "error: prebuilt/$target is stale (wrapper key mismatch); re-run build-libs first" >&2
    exit 1
  fi

  goos=$(sed -n 's/^GOOS=//p' <<<"$manifest")
  goarch=$(sed -n 's/^GOARCH=//p' <<<"$manifest")
  tag=libs/${goos}_${goarch}/v$version
  if git rev-parse --verify -q "refs/tags/$tag" >/dev/null; then
    echo "error: tag $tag already exists; module versions are immutable, bump the version" >&2
    exit 1
  fi
  git tag "$tag" "$ref"
  tags+=("$tag")
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

push_ref=${GITHUB_REF_NAME:-$(git rev-parse --abbrev-ref HEAD)}
echo "==> Pushing release commit and tag to $push_ref"
git push origin "HEAD:refs/heads/$push_ref"
git push origin "refs/tags/v$version"

echo "==> Released v$version"
