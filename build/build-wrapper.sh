#!/usr/bin/env bash
# Usage: build/build-wrapper.sh <target>
#
# Builds the "moving layer": compiles wrapper/cv2capi.cpp against the OpenCV
# SDK in .work/dist/<target>/opencv into libcv2wrapper.a. This is the only
# thing that needs rebuilding when the binding code changes.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

cv2_load_target "${1:?usage: build-wrapper.sh <target>}"

core_hpp=$(find "$CV2_DIST_DIR" -path '*opencv2/core.hpp' | head -1)
if [ -z "$core_hpp" ]; then
  echo "error: OpenCV SDK not found in $CV2_DIST_DIR (run build-opencv.sh first)" >&2
  exit 1
fi
include_root=$(dirname "$(dirname "$core_hpp")")

mkdir -p "$CV2_OBJ_DIR"
echo "==> Compiling base wrapper with $WRAPPER_CXX"
# shellcheck disable=SC2086
"$WRAPPER_CXX" -std=c++11 -O2 $WRAPPER_CXXFLAGS \
  -I"$include_root" \
  -c "$CV2_ROOT/wrapper/cv2capi.cpp" \
  -o "$CV2_OBJ_DIR/cv2capi.o"

rm -f "$CV2_OBJ_DIR/libcv2wrapper.a"
"$WRAPPER_AR" rcs "$CV2_OBJ_DIR/libcv2wrapper.a" "$CV2_OBJ_DIR/cv2capi.o"
echo "==> Built $CV2_OBJ_DIR/libcv2wrapper.a"

# Feature-set wrappers: each set gets its own archive so importers that skip
# the subpackage never link (or download) any of it.
for set in ${CV2_FEATURE_SETS:-}; do
  set_upper=$(printf '%s' "$set" | tr '[:lower:]' '[:upper:]')
  sources_var="CV2_${set_upper}_WRAPPER_SOURCES"
  sources=${!sources_var:-}
  [ -n "$sources" ] || continue
  objs=()
  for src in $sources; do
    obj="$CV2_OBJ_DIR/${set}_$(basename "$src" .cpp).o"
    echo "==> Compiling feature wrapper $src ($set)"
    # shellcheck disable=SC2086
    "$WRAPPER_CXX" -std=c++11 -O2 $WRAPPER_CXXFLAGS \
      -I"$include_root" \
      -c "$CV2_ROOT/wrapper/$src" \
      -o "$obj"
    objs+=("$obj")
  done
  rm -f "$CV2_OBJ_DIR/libcv2${set}wrapper.a"
  "$WRAPPER_AR" rcs "$CV2_OBJ_DIR/libcv2${set}wrapper.a" "${objs[@]}"
  echo "==> Built $CV2_OBJ_DIR/libcv2${set}wrapper.a"
done
