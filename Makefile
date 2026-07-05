# Developer entry points. The published binaries are produced by GitHub
# Actions (see .github/workflows/build-libs.yml); these targets exist for
# local hacking and verification.
#
# OPENCV selects the version line (defaults to the primary in
# build/build.conf), e.g.: make libs TARGET=linux-amd64 OPENCV=4.12.0

TARGET ?= linux-amd64
OPENCV ?=

.PHONY: libs dev test vet clean

# Build the static libraries for one target from scratch (OpenCV + wrappers).
libs:
	CV2_OPENCV_VERSION=$(OPENCV) build/build-opencv.sh $(TARGET)
	CV2_OPENCV_VERSION=$(OPENCV) build/build-wrapper.sh $(TARGET)
	CV2_OPENCV_VERSION=$(OPENCV) build/package-libs.sh $(TARGET)

# Fetch the CI-built libraries for TARGET and wire them up with go.work.
dev:
	CV2_OPENCV_VERSION=$(OPENCV) build/fetch-prebuilt.sh $(TARGET)
	build/setup-gowork.sh

test:
	go test -v ./...

vet:
	go vet ./...

clean:
	rm -rf .work .prebuilt go.work go.work.sum
