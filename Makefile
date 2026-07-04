# Developer entry points. The published binaries are produced by GitHub
# Actions (see .github/workflows/build-libs.yml); these targets exist for
# local hacking and verification.

TARGET ?= linux-amd64

.PHONY: libs dev test vet clean

# Build the static libraries for one target from scratch (OpenCV + wrapper).
libs:
	build/build-opencv.sh $(TARGET)
	build/build-wrapper.sh $(TARGET)
	build/package-libs.sh $(TARGET)

# Fetch the CI-built libraries for TARGET and wire them up with go.work.
dev:
	build/fetch-prebuilt.sh $(TARGET)
	build/setup-gowork.sh

test:
	go test -v ./...

vet:
	go vet ./...

clean:
	rm -rf .work .prebuilt go.work go.work.sum
