#!/usr/bin/env bash
# Shared helpers sourced by the other build scripts.
set -euo pipefail

CV2_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
CV2_WORK=${CV2_WORK_DIR:-$CV2_ROOT/.work}

# Portable sha256 (GNU coreutils on Linux, shasum on macOS runners).
cv2_sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum | cut -d' ' -f1
  else
    shasum -a 256 | cut -d' ' -f1
  fi
}

# cv2_verify_sha256 <file> <expected-hash>
cv2_verify_sha256() {
  local file=$1 expected=$2 actual
  actual=$(cv2_sha256 <"$file")
  if [ "$actual" != "$expected" ]; then
    echo "error: checksum mismatch for $file" >&2
    echo "  expected: $expected" >&2
    echo "  actual:   $actual" >&2
    return 1
  fi
}

cv2_load_target() {
  local target=$1
  local env_file="$CV2_ROOT/build/targets/$target.env"
  if [ ! -f "$env_file" ]; then
    echo "error: unknown target '$target' (no $env_file)" >&2
    echo "known targets:" >&2
    ls "$CV2_ROOT/build/targets" | sed 's/\.env$//; s/^/  /' >&2
    exit 1
  fi
  # shellcheck disable=SC1090
  source "$CV2_ROOT/build/build.conf"
  # shellcheck disable=SC1090
  source "$env_file"

  CV2_TARGET=$target
  CV2_SRC_DIR=$CV2_WORK/src
  CV2_BUILD_DIR=$CV2_WORK/build/$target
  CV2_DIST_DIR=$CV2_WORK/dist/$target/opencv
  CV2_OBJ_DIR=$CV2_WORK/obj/$target
  CV2_OUT_DIR=$CV2_WORK/out/$target
}

# cv2_build_key <target> opencv|wrapper
# Prints a content hash of everything that influences the corresponding
# build layer. CI skips the expensive OpenCV rebuild when the opencv key
# stored in the prebuilt branch MANIFEST still matches.
cv2_build_key() {
  local target=$1 layer=$2
  local files=()
  case "$layer" in
  opencv)
    files=("$CV2_ROOT/build/build.conf" "$CV2_ROOT/build/targets/$target.env" "$CV2_ROOT/build/build-opencv.sh")
    # shellcheck disable=SC1090
    local toolchain
    toolchain=$(sed -n 's/^CMAKE_TOOLCHAIN=//p' "$CV2_ROOT/build/targets/$target.env" | tr -d '"')
    if [ -n "$toolchain" ]; then
      files+=("$CV2_ROOT/build/toolchains/$toolchain")
    fi
    ;;
  wrapper)
    files=("$CV2_ROOT/build/build.conf" "$CV2_ROOT/build/targets/$target.env"
      "$CV2_ROOT/build/build-wrapper.sh" "$CV2_ROOT/build/package-libs.sh")
    while IFS= read -r f; do files+=("$f"); done < <(find "$CV2_ROOT/wrapper" -type f | LC_ALL=C sort)
    ;;
  *)
    echo "error: unknown layer '$layer'" >&2
    return 1
    ;;
  esac
  cat "${files[@]}" | cv2_sha256
}
