#!/usr/bin/env bash
# Usage: build/build-opencv.sh <target>
#
# Builds the "fixed layer": static OpenCV limited to the modules listed in
# build/build.conf (BUILD_LIST), plus the bundled static zlib. The result is
# installed into .work/dist/<target>/opencv.
#
# If .work/src/opencv-<version>/ already exists it is used as-is, so an
# alternative source mirror can be pre-extracted there.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

cv2_load_target "${1:?usage: build-opencv.sh <target>}"

mkdir -p "$CV2_SRC_DIR"
src_tree=$CV2_SRC_DIR/opencv-$OPENCV_VERSION

if [ ! -d "$src_tree" ]; then
  tarball=$CV2_SRC_DIR/opencv-$OPENCV_VERSION.tar.gz
  if [ ! -f "$tarball" ]; then
    echo "==> Downloading OpenCV $OPENCV_VERSION"
    curl -fL --retry 3 --retry-delay 2 -o "$tarball.tmp" "$OPENCV_URL_BASE/$OPENCV_VERSION.tar.gz"
    mv "$tarball.tmp" "$tarball"
  fi
  echo "==> Verifying source checksum"
  if ! cv2_verify_sha256 "$tarball" "$OPENCV_SHA256"; then
    # Drop the bad tarball so a rerun redownloads instead of wedging.
    rm -f "$tarball"
    exit 1
  fi
  echo "==> Extracting"
  # Extract into a staging directory and move atomically, so an interrupted
  # extraction can never masquerade as a complete source tree on rerun.
  rm -rf "$CV2_SRC_DIR/.extract"
  mkdir -p "$CV2_SRC_DIR/.extract"
  tar -xzf "$tarball" -C "$CV2_SRC_DIR/.extract"
  mv "$CV2_SRC_DIR/.extract/opencv-$OPENCV_VERSION" "$src_tree"
  rm -rf "$CV2_SRC_DIR/.extract"
fi

echo "==> Configuring OpenCV for $CV2_TARGET (modules: $OPENCV_MODULES)"
toolchain_arg=()
if [ -n "$CMAKE_TOOLCHAIN" ]; then
  toolchain_arg=(-DCMAKE_TOOLCHAIN_FILE="$CV2_ROOT/build/toolchains/$CMAKE_TOOLCHAIN")
fi

# The ${arr[@]+...} idiom keeps empty-array expansion safe under set -u on
# bash 3.2 (macOS runners).
# shellcheck disable=SC2086
cmake -G Ninja -S "$src_tree" -B "$CV2_BUILD_DIR" \
  ${toolchain_arg[@]+"${toolchain_arg[@]}"} \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$CV2_DIST_DIR" \
  -DBUILD_SHARED_LIBS=OFF \
  -DBUILD_LIST="$OPENCV_MODULES" \
  -DBUILD_opencv_world=OFF \
  -DENABLE_PIC=ON \
  -DCPU_DISPATCH= \
  -DWITH_IPP=OFF \
  -DWITH_ITT=OFF \
  -DWITH_OPENCL=OFF \
  -DWITH_OPENCLAMDFFT=OFF \
  -DWITH_OPENCLAMDBLAS=OFF \
  -DWITH_VA=OFF \
  -DWITH_VA_INTEL=OFF \
  -DWITH_TBB=OFF \
  -DWITH_OPENMP=OFF \
  -DWITH_EIGEN=OFF \
  -DWITH_LAPACK=OFF \
  -DWITH_PROTOBUF=OFF \
  -DBUILD_PROTOBUF=OFF \
  -DWITH_ADE=OFF \
  -DWITH_ZLIB=ON \
  -DBUILD_ZLIB=ON \
  -DWITH_ZLIB_NG=OFF \
  -DWITH_JPEG=OFF \
  -DWITH_PNG=OFF \
  -DWITH_TIFF=OFF \
  -DWITH_WEBP=OFF \
  -DWITH_OPENJPEG=OFF \
  -DWITH_JASPER=OFF \
  -DWITH_OPENEXR=OFF \
  -DWITH_V4L=OFF \
  -DWITH_FFMPEG=OFF \
  -DWITH_GSTREAMER=OFF \
  -DWITH_GTK=OFF \
  -DWITH_WIN32UI=OFF \
  -DWITH_MSMF=OFF \
  -DWITH_DSHOW=OFF \
  -DBUILD_TESTS=OFF \
  -DBUILD_PERF_TESTS=OFF \
  -DBUILD_EXAMPLES=OFF \
  -DBUILD_DOCS=OFF \
  -DBUILD_opencv_apps=OFF \
  -DBUILD_JAVA=OFF \
  -DBUILD_opencv_python2=OFF \
  -DBUILD_opencv_python3=OFF \
  -DBUILD_opencv_js=OFF \
  -DENABLE_PRECOMPILED_HEADERS=OFF \
  -DOPENCV_GENERATE_PKGCONFIG=OFF \
  -DOPENCV_ALLOCATOR_STATS_COUNTER_TYPE=int64_t \
  $CMAKE_EXTRA \
  -Wno-dev

echo "==> Building"
cmake --build "$CV2_BUILD_DIR" --parallel

echo "==> Installing to $CV2_DIST_DIR"
rm -rf "$CV2_DIST_DIR"
cmake --install "$CV2_BUILD_DIR"

# The bundled zlib static library is not installed by OpenCV; copy it (and any
# other 3rdparty archives) into the dist tree so packaging can pick them up.
mkdir -p "$CV2_DIST_DIR/3rdparty-lib"
find "$CV2_BUILD_DIR/3rdparty" -name '*.a' -exec cp -v {} "$CV2_DIST_DIR/3rdparty-lib/" \;

# Slim the SDK stored on prebuilt branches: the cascade/model data files and
# sample assets are useless for linking and would add ~10 MB per branch.
rm -rf "$CV2_DIST_DIR/share/opencv4/haarcascades" \
  "$CV2_DIST_DIR/share/opencv4/lbpcascades" \
  "$CV2_DIST_DIR/share/opencv4/quality" \
  "$CV2_DIST_DIR/share/opencv4/samples" \
  "$CV2_DIST_DIR/etc/haarcascades" \
  "$CV2_DIST_DIR/etc/lbpcascades" \
  "$CV2_DIST_DIR/samples"

cp "$src_tree/LICENSE" "$CV2_DIST_DIR/OPENCV-LICENSE.txt"

echo "==> Done: $CV2_TARGET"
find "$CV2_DIST_DIR" -name '*.a' -exec ls -la {} \;
