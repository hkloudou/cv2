#!/usr/bin/env bash
# Usage: build/ci-apt-packages.sh <target>
# Prints the extra apt packages the CI runner must install for this target.
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

cv2_load_target "${1:?usage: ci-apt-packages.sh <target>}"
echo "$CI_APT_PACKAGES"
