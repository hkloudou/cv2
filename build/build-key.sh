#!/usr/bin/env bash
# Usage: build/build-key.sh <target> opencv|wrapper
# Prints the content hash for one build layer of one target.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

target=${1:?usage: build-key.sh <target> opencv|wrapper}
layer=${2:?usage: build-key.sh <target> opencv|wrapper}
cv2_load_target "$target"
cv2_build_key "$target" "$layer"
