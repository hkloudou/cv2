# cv2

[![test](https://github.com/hkloudou/cv2/actions/workflows/test.yml/badge.svg)](https://github.com/hkloudou/cv2/actions/workflows/test.yml)
[![build-libs](https://github.com/hkloudou/cv2/actions/workflows/build-libs.yml/badge.svg)](https://github.com/hkloudou/cv2/actions/workflows/build-libs.yml)

Minimal Go bindings for OpenCV **template matching** ("find a small image
inside a big image") with batteries included:

- **No OpenCV installation required.** Prebuilt static libraries (OpenCV
  `core` + `imgproc` only) are linked into your binary via cgo.
- **You only download your platform.** Each platform's libraries live in
  their own Go module; `go build` fetches just the one your target needs
  (~5 MB compressed), never all of them.
- **All binaries are built by GitHub Actions** from the official OpenCV
  source tarball (checksum-pinned) — nothing is hand-built.

```go
package main

import (
	"fmt"
	"image"

	"github.com/hkloudou/cv2"
)

func main() {
	var screenshot, button image.Image // e.g. decoded PNGs

	minVal, minX, minY, maxVal, maxX, maxY := cv2.Match(screenshot, button)
	fmt.Println(minVal, minX, minY)
	// TM_CCOEFF_NORMED: maxVal close to 1.0 means a confident match,
	// whose top-left corner is at (maxX, maxY).
	fmt.Println(maxVal, maxX, maxY)
}
```

## Supported platforms

| GOOS/GOARCH | libraries built with | notes |
| --- | --- | --- |
| `linux/amd64` | gcc | tested in CI |
| `linux/386` | gcc -m32 | test binary executed in CI |
| `linux/arm64` | aarch64-linux-gnu-gcc | test binary executed under qemu in CI |
| `windows/amd64` | MinGW-w64 (POSIX threads) | tested in CI on a Windows runner |
| `windows/386` | MinGW-w64 i686 (POSIX threads) | link-checked in CI |
| `darwin/arm64` | Apple clang (macos-14 runner) | tested in CI on Apple Silicon |

Requirements: Go with `CGO_ENABLED=1` and a matching C toolchain. On Windows
use an MSYS2/MinGW-w64 gcc (POSIX threads variant, which is the MSYS2
default); the produced `.exe` is fully static (no extra DLLs needed).

## How it works (architecture)

The repository is split so that **source stays small and binaries stay out of
history**, while `go build` still gets everything it needs:

```
default branch          source only (Go + C++ wrapper + build scripts) — small forever
prebuilt/<target>       one branch per platform, force-pushed as a SINGLE commit by CI:
                          libs/<goos>_<goarch>/   Go module with the static .a files
                          sdk/                    OpenCV install tree (headers) for
                                                  incremental wrapper rebuilds
                          MANIFEST                provenance + cache keys
tags v0.X.Y             release of the root module
tags libs/<goos>_<goarch>/v0.X.Y
                        release of each platform's libs module, pointing into
                        its prebuilt branch
```

The root package selects the right libs module per platform with build tags:

```go
//go:build linux && amd64
import _ "github.com/hkloudou/cv2/libs/linux_amd64"
```

so the Go module system downloads **only** the static libraries for the
platform being compiled. cgo link flags propagate from the libs module; the
Go code itself declares the C prototypes inline, so users never need OpenCV
headers or a C++ compile — the C++ wrapper is precompiled into
`libcv2wrapper.a` by CI.

### Two build layers (why rebuilds stay cheap)

| layer | contents | rebuilt when |
| --- | --- | --- |
| fixed | `libopencv_core.a`, `libopencv_imgproc.a`, `libzlib.a` | `build/build.conf` changes (OpenCV version / module list / flags) |
| moving | `libcv2wrapper.a` (the whole C++ binding) | wrapper source changes — seconds, not minutes |

The `build-libs` workflow hashes everything that influences the fixed layer
into `OPENCV_BUILD_KEY`. If the key on the `prebuilt/<target>` branch still
matches, CI skips the OpenCV compile entirely and rebuilds only the wrapper
against the SDK cached in that same branch. Adding a new binding function
therefore costs seconds of CI and replaces a 7 KB archive — the big OpenCV
libraries are reused byte-for-byte.

### Adding functionality later

1. Need another OpenCV module (e.g. `features2d`)? Add it to
   `OPENCV_MODULES` in `build/build.conf` — the key changes, CI rebuilds the
   fixed layer once per platform.
2. Add C functions in `wrapper/` (keep the `Cv2_` prefix), mirror the
   prototypes in the cgo preamble, add the Go API.
3. Push — CI rebuilds the wrapper, refreshes `prebuilt/*`, tests run against
   the new binaries via `go.work` overlays (no release needed to test).
4. Run the `release` workflow to cut versioned tags when ready.

Because prebuilt branches are **orphan single commits that get force-pushed**,
repository history never accumulates binary blobs; only release tags keep
snapshots alive, and each user downloads a single platform's ~5 MB module.

## Workflows

| workflow | trigger | what it does |
| --- | --- | --- |
| `build-libs` | push touching `build/**` or `wrapper/**`; manual | builds all 6 targets: Linux runners cross-build the Linux and Windows targets (MinGW-w64), a macos-14 runner builds darwin/arm64; force-pushes `prebuilt/*` branches |
| `test` | every push/PR; after `build-libs` | linux/amd64, windows/amd64 and darwin/arm64 native test runs, linux/386 native run, linux/arm64 run under qemu, windows/386 link check |
| `release` | manual (`version` input, e.g. `0.1.0`) | tags all `libs/.../vX.Y.Z` modules, pins them in `go.mod`, tags `vX.Y.Z` |

## Local development

```sh
# Use the CI-built libraries (recommended):
make dev TARGET=linux-amd64   # fetch prebuilt libs + set up go.work
make test

# Or build the static libraries yourself (cmake + ninja + toolchain needed):
make libs TARGET=linux-amd64
```

## API

```go
func Match(parent, sub image.Image) (float32, int, int, float32, int, int)
```

Runs `TM_CCOEFF_NORMED` template matching and returns
`(minVal, minX, minY, maxVal, maxX, maxY)`. The best match is at
`(maxX, maxY)`; `maxVal` close to `1.0` means high confidence. `Match`
panics if an image is empty or OpenCV rejects the inputs (e.g. the template
is wider but shorter than the image).

Lower-level pieces are exported too (`Mat`, `NewMatFromBytes`,
`ImageToMatRGBA`, `MatchTemplate`, `MinMaxLoc`, `TemplateMatchMode`) for
custom pipelines.

## License

The Go and C++ code in this repository is provided by the repository owner.
The prebuilt branches redistribute OpenCV binaries under the
[Apache 2.0 license](https://github.com/opencv/opencv/blob/4.x/LICENSE)
(a copy ships in every libs module as `LICENSE-OPENCV.txt`) and zlib under
the zlib license.
