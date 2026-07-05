# cv2

[![test](https://github.com/hkloudou/cv2/actions/workflows/test.yml/badge.svg)](https://github.com/hkloudou/cv2/actions/workflows/test.yml)
[![build-libs](https://github.com/hkloudou/cv2/actions/workflows/build-libs.yml/badge.svg)](https://github.com/hkloudou/cv2/actions/workflows/build-libs.yml)

Minimal Go bindings for OpenCV **template matching** ("find a small image
inside a big image") with batteries included:

- **No OpenCV installation required.** Prebuilt static libraries (OpenCV
  `core` + `imgproc` only) are linked into your binary via cgo.
- **You only link your platform.** Each platform's libraries live in their
  own Go module. `go build` extracts and compiles against just the one your
  target needs (~5 MB compressed). The first `go get`/`go mod tidy` also
  downloads the other platforms' base modules once to record their go.sum
  checksums (a ~30 MB one-time cost per module cache); optional feature-set
  modules (like f2d) are never downloaded unless you import the matching
  subpackage.
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

## Choosing your OpenCV version

Module versions encode the OpenCV line they were built from:

```
v0.<code>.<revision>        code = major*10000 + minor*100 + patch
```

| OpenCV | module versions | pin the line with | status |
| --- | --- | --- | --- |
| 5.0.0 | `v0.50000.N` | `go get github.com/hkloudou/cv2@v0.50000` | active |
| 4.12.0 | `v0.41200.N` | `go get github.com/hkloudou/cv2@v0.41200` | active |
| 4.8.1 | `v0.40801.0` | `go get github.com/hkloudou/cv2@v0.40801` | frozen (tags stay consumable; no further builds or revisions) |

`@v0.40801` is a Go version-prefix query: it resolves to the newest binding
revision of that OpenCV line, and Go's minimal version selection keeps you
on it. Switching OpenCV versions is exactly one `go get` — the matching
prebuilt static libraries come along automatically. Plain `@latest` follows
the newest OpenCV line. Each line has its own
`prebuilt/<opencv-version>/<target>` branches, is built from the
checksum-pinned official source tarball, and passes the same full test
matrix.

## How it works (architecture)

The repository is split so that **source stays small and binaries stay out of
history**, while `go build` still gets everything it needs:

```
default branch          source only (Go + C++ wrapper + build scripts) — small forever
prebuilt/<ocv>/<target> one branch per OpenCV version line and platform,
                        force-pushed as a SINGLE commit by CI:
                          libs/<goos>_<goarch>/       base Go module (static .a files)
                          libs/<goos>_<goarch>_f2d/   optional feature-set module
                          sdk/                        OpenCV install tree (headers) for
                                                      incremental wrapper rebuilds
                          MANIFEST                    provenance + cache keys
tags v0.<code>.N        release of the root module for one OpenCV line
tags libs/<module>/v0.<code>.N
                        release of each libs module, pointing into its
                        prebuilt branch
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
| `release` | manual (`version` input) or committing the version to `RELEASE_VERSION` | takes `0.<code>.N` (e.g. `0.40801.3` = OpenCV 4.8.1 line), tags every libs module at its `prebuilt/<ocv>/*` head, pins them in `go.mod`, tags `v0.<code>.N` |

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

Lower-level pieces are exported too for custom pipelines: `Mat` (with
`Rows`/`Cols`/`Channels`/`Type`, `ToBytes`, `ToImage`), `NewMatFromBytes`,
`ImageToMatRGBA`, `MatchTemplate`, `MinMaxLoc`, plus imgproc primitives:
`Resize` (the building block for multi-scale matching), `CvtColor`,
`GaussianBlur`, `Threshold`, `Canny`, `Erode`/`Dilate` (with
`GetStructuringElement`), `WarpAffine`/`GetRotationMatrix2D`, and
`FindExternalContourRects`. All of these run against the same prebuilt
base libraries — no extra downloads.

### Testing principle

The official OpenCV implementation is assumed correct and serves as ground
truth. Tests exist to prove the BINDING layer is a faithful pass-through
(parameter order, type mapping, memory marshalling), never to re-verify
OpenCV itself. Concretely: parity tests reproduce the reference
computations OpenCV's own accuracy tests use (with citations into the
OpenCV tree), and where OpenCV offers both directions of an operation the
tests go OpenCV-vs-OpenCV (the qr package round-trips
`cv::QRCodeEncoder` -> `cv::QRCodeDetector` through the Go surface).
Every new binding must come with such a test.

### API conventions: primitives vs pipelines

The root package and the feature subpackages share one contract language —
English godoc, panics with `cv2:`-prefixed messages for invalid inputs
(nil/closed Mats, OpenCV rejections), "not present" is a result rather
than an error (a low `maxVal` for Match, `Found: false` for f2d), and every
binding cites the OpenCV documentation entry it wraps. They differ in
altitude on purpose:

- the root package exposes 1:1 OpenCV primitives (`MatchTemplate`,
  `Resize`, `CvtColor`, ...) plus the `Match` convenience;
- feature subpackages expose goal-level pipelines (`f2d.Locate` runs
  ORB + ratio test + RANSAC homography in a single native call), because
  shuttling intermediate state (keypoints, matches) across cgo would be
  slow and error-prone. Tunables travel in an Options struct instead.

### Optional feature sets (import-driven)

Extra OpenCV capability ships as subpackages backed by their own libs
modules. You opt in by importing — programs that skip the import never
download or link the extra binaries:

```go
import "github.com/hkloudou/cv2/f2d"

// Scale/rotation-tolerant localization via ORB features + RANSAC
// homography (needs OpenCV features2d + calib3d, ~4 MB extra, fetched
// automatically for your platform only):
res := f2d.Locate(screenshot, button)
if res.Found {
	fmt.Println("center:", res.Center, "corners:", res.Corners)
}
```

```go
import "github.com/hkloudou/cv2/qr"

// QR encode/decode via the official cv::QRCodeEncoder/cv::QRCodeDetector
// (OpenCV objdetect + its dependency closure, fetched only when imported):
img, _ := qr.Encode("https://example.com", 4, 4)
codes := qr.Decode(screenshot) // every QR in the image, with corners
```

Not importing a feature subpackage isolates you from it at three levels,
so it can never break or bloat a base-only build:

1. **Source**: `wrapper/*.cpp` is never compiled on user machines at all —
   CI precompiles it; the `wrapper/` directory contains no Go files, so
   `go build` ignores it.
2. **Download**: the feature wrapper archive (`libcv2f2dwrapper.a`) and its
   OpenCV modules ship only inside the `libs/<goos>_<goarch>_f2d` modules,
   which are fetched only when the subpackage is imported.
3. **Link**: without the import there are no `-l` flags for those archives,
   and static linking additionally prunes at archive-member granularity.
   Measured on linux/amd64: a Match-only binary is ~7.7 MB; adding
   `f2d.Locate` brings ~2.7 MB of reachable code (not the full ~9 MB of
   feature archives).

## Maintainer guide (for humans and future AIs)

Everything below is the complete operating model of this repository. Read it
before changing anything; every rule here is load-bearing.

### Topology

```
source of truth      the default branch: Go + wrapper C++ + build/ scripts + workflows.
                     No binaries, ever.

prebuilt/<ocv>/<target>
                     machine-generated, one per (OpenCV line x platform), e.g.
                     prebuilt/4.8.1/linux-amd64. Always a SINGLE orphan commit,
                     force-pushed by CI. Content:
                       libs/<goos>_<goarch>/        base libs Go module
                       libs/<goos>_<goarch>_f2d/    feature-set libs Go module(s)
                       sdk/                         full OpenCV install tree (headers+libs)
                       MANIFEST                     provenance + cache keys
                     Never commit to these by hand; never expect history on them.

tags                 v0.<code>.N            root module release of one OpenCV line
                     libs/<dir>/v0.<code>.N one per libs module, pointing INTO the
                                            prebuilt branch commit that was released.
                     Tags are immutable: never delete or move a published tag
                     (the Go module proxy has already mirrored it). Fixes = new N.

                     Expected GitHub UI note: libs tag commits show "This
                     commit does not belong to any branch" - correct by
                     design. Prebuilt branches are rolling force-pushed
                     pointers; released snapshots are anchored by their tags
                     alone, and Go tooling resolves versions via tags only.
```

### Version scheme

`v0.<code>.<revision>` where `code = major*10000 + minor*100 + patch` of the
OpenCV version (4.8.1 → 40801, 4.12.0 → 41200). Properties this must keep:
codes order monotonically with OpenCV versions; `@v0.<code>` prefix queries
select the newest revision of a line; `@latest` follows the biggest code.
The set of buildable lines is `OPENCV_VERSIONS` in `build/build.conf`, each
with a pinned tarball sha256 (`OPENCV_SHA256_<code>`). `release.sh` decodes
the code from the requested module version and refuses unknown lines.

### The two-layer cache (why CI is fast)

Two content hashes, computed by `build/lib.sh:cv2_build_key` and stored in
each prebuilt branch's MANIFEST:

- `OPENCV_BUILD_KEY` = sha256(build.conf + targets/<t>.env + toolchain file
  + build-opencv.sh + "opencv=<version>"). If the branch's key matches, CI
  restores `sdk/` from the branch and skips the OpenCV compile entirely.
- `WRAPPER_KEY` = sha256(build.conf + targets/<t>.env + build-wrapper.sh +
  package-libs.sh + every file under wrapper/ + "opencv=<version>"). Used
  by `fetch-prebuilt.sh --require-fresh` (CI tests skip stale binaries and
  rely on the workflow_run re-trigger) and by `release.sh` (refuses to tag
  stale binaries).

Consequence: editing wrapper/ or Go code costs seconds of CI per job;
editing build.conf/envs/toolchains rebuilds the fixed layer for the
affected lines/targets only.

### Script contracts (build/)

| script | contract |
| --- | --- |
| `lib.sh` | shared: loads build.conf + target env; `CV2_OPENCV_VERSION` env selects the line (default `OPENCV_VERSION`); computes versioned paths and `CV2_PREBUILT_BRANCH`; portable sha256 |
| `build-opencv.sh <t>` | downloads (sha256-verified, atomic extract) and builds static OpenCV per BUILD_LIST into `.work/dist/<ocv>/<t>/opencv`; a pre-extracted `.work/src/opencv-<ver>/` is used as-is (offline mirror hook) |
| `build-wrapper.sh <t>` | compiles wrapper/cv2capi.cpp -> libcv2wrapper.a and each feature set's sources -> libcv2<set>wrapper.a |
| `package-libs.sh <t>` | emits `.work/out/<ocv>/<t>/`: MANIFEST, README, sdk/, and one Go module per lib set (base + feature sets) with generated go.mod/libs.go (build tag + cgo LDFLAGS from the env); normalizes archive names |
| `push-prebuilt.sh <t>` | stages that out dir with a throwaway index, writes an orphan commit inside the main repo's object DB (reuses its credentials), force-pushes to `CV2_PREBUILT_BRANCH` |
| `fetch-prebuilt.sh [--require-fresh] <t>` | pulls the branch's libs/ into `.prebuilt/`; exit 3 = branch missing, exit 4 = stale wrapper key (only with the flag), other = real error (never masked) |
| `setup-gowork.sh` | go.work overlaying the root module with every fetched `.prebuilt/libs/*` module |
| `release.sh <0.code.N>` | validates + decodes the line; verifies every target branch exists and is fresh; tags every module dir found in each branch's `libs/`; pins them in go.mod (go mod tidy with proxy-indexing retries); commits, tags `v0.code.N`, pushes. Refuses detached HEAD/tag dispatch |
| `build-key.sh <t> opencv\|wrapper` | prints the layer key (used by CI plan steps) |
| `ci-apt-packages.sh <t>` | prints the target's cross-toolchain apt packages |

### Workflows

- `build-libs`: push (default branch, paths build/ wrapper/ or itself) or
  manual. `setup` job reads OPENCV_VERSIONS -> matrix (lines x 6 targets,
  Linux runners cross-compile Windows via MinGW-w64; darwin on macos-14).
  Per job: plan (key check) -> restore-or-build OpenCV -> wrappers ->
  package -> force-push branch. Global concurrency group, newest push wins.
  GITHUB_TOKEN pushes deliberately do not retrigger workflows, so branch
  publishing cannot loop.
- `test`: every push/PR + workflow_run after build-libs. Same version
  matrix. Native full runs on linux/amd64, windows/amd64 (MSYS2 MINGW64
  shell, path-type inherit), darwin/arm64; linux/386 executes natively;
  linux/arm64 executes under qemu-user; windows link-checks. Fetch step maps
  exit 3/4 to skip-with-warning so bootstrap and coordinated wrapper+Go
  pushes stay green (the workflow_run pass tests for real).
- `release`: manual (version input) or push changing RELEASE_VERSION on the
  default branch; already-released versions no-op. Runs build/release.sh.

### Releasing / recovery runbook

1. Ensure build-libs and test are green for the target line.
2. Commit the module version (e.g. `0.40801.1`) to RELEASE_VERSION (or
   dispatch the workflow). Watch the release run.
3. If release fails after libs tags were pushed but before the root tag:
   those libs tags are consumed only via go.mod pins, so simply fix and
   re-run with a BUMPED revision (tags are immutable; never force-move).
4. `go mod tidy` inside release retries while proxy.golang.org indexes the
   fresh tags; a persistent failure aborts before the root tag on purpose.
5. Bootstrap-from-zero: push source -> build-libs populates branches ->
   release. Nothing else is stateful.

### Version-line divergence policy (single branch until proven otherwise)

All OpenCV lines build from ONE source branch. The Go sources cannot vary
per line even in principle: users select a line via module versions, the
import paths are identical, so build tags cannot see the OpenCV version.
When the lines need different treatment, escalate in this order:

1. Build-layer difference (flags, modules): per-line knobs in build.conf
   (`OPENCV_MODULES_<code>` overrides the BUILD_LIST — OpenCV 5 renamed
   features2d to features and moved findHomography into the new geometry
   module) and feature-set candidate lists that simply carry both
   generations' archive names side by side.
2. C++ API difference: preprocessor guards in wrapper/*.cpp; the wrapper
   is compiled once per line against that line's headers, so one source
   file yields per-line binaries. First real use: f2d.cpp selects
   `<opencv2/features.hpp>` + `<opencv2/geometry.hpp>` on
   `CV_VERSION_MAJOR >= 5` and the classic headers otherwise.
3. Capability only newer lines can offer: older lines' wrapper returns a
   "not supported on this OpenCV line" error string; the Go API stays
   uniform and reports it at runtime.
4. Only if the GO API ITSELF must fork (realistically: OpenCV 5.x breaking
   changes): open a maintenance branch per line, add it to the workflow
   branch filters, and accept the cherry-pick cost. Do not do this
   preemptively - a single branch keeps every fix landing once, tested on
   all lines by the same matrix, with WRAPPER_KEY anchoring both lines'
   binaries to one source commit.

### Extension recipes

- New exported function from an already-built module: add `Cv2_*` to
  wrapper (error-string contract below), mirror the prototype in the cgo
  preamble, add the Go API + a parity test mirroring OpenCV's reference
  semantics (see parity_test.go). Cost: wrapper-layer rebuild only.
- New OpenCV module / feature set: extend OPENCV_MODULES (superset build),
  declare `CV2_<SET>_LIBS` + `CV2_<SET>_WRAPPER_SOURCES` in build.conf, add
  wrapper/<set>.cpp, a subpackage with per-platform blank imports of
  `libs/<goos>_<goarch>_<set>`, tests. Import-driven: non-importers never
  download it.
- New OpenCV version line: append to OPENCV_VERSIONS + pin
  `OPENCV_SHA256_<code>` (use an authoritative source, e.g. the easybuild
  archives); CI builds the new branches; release `0.<newcode>.0`.
- New platform: `build/targets/<os>-<arch>.env` (+ toolchain file), a
  matrix entry in both workflows, `libs_<goos>_<goarch>.go` link files in
  the root package and each subpackage, extend the build tags.

### Root-libs pairing: the link-time ABI handshake

Go's MVS guarantees libs versions are AT LEAST the go.mod pins, but has no
"exactly equal" constraint. The remaining skew direction (manually
upgrading a libs module past the root) is closed at link time: the root
package's generated `zz_abi_link.go` calls `cv2_abi_<hash>()` from an
init, the f2d subpackage calls `cv2_f2d_abi_<hash>()`, and only libs
modules generated from the same wrapper sources define those symbols
(`<hash>` = first 12 hex of sha256 over `wrapper/`). Mixing generations
fails the build with `undefined reference to cv2_abi_<expected-hash>`.
After any change under `wrapper/`, run `build/gen-abi.sh` and commit the
regenerated files; CI enforces freshness via `build/gen-abi.sh --check`.
Note the hash is deliberately line-independent: one source commit serves
every OpenCV line with one ABI generation.

`go.mod` additionally carries `retract` directives for the bootstrap-era
v0.1.0/v0.2.0 (their tags are gone; the module proxy still mirrors them,
and retraction is the mechanism that tells `go` tooling not to pick them).

### Invariants — do not break these

1. The cgo preambles re-declare the C prototypes: any wrapper ABI change
   must update wrapper/cv2capi.h AND every Go preamble in the same commit.
2. Error strings cross the ABI as malloc'd `char*` allocated ONLY through
   `Cv2_CopyError`/`copy_error` in cv2capi.cpp (the malloc-failure sentinel
   is compared by pointer identity in Cv2_FreeString — a second allocator
   in another translation unit would break it). NULL always means success.
3. C++ exceptions must never unwind across the ABI: every wrapper entry
   point that can throw wraps its body in try/catch; constructors signal
   failure with NULL instead.
4. Nil/closed Mats must be rejected on both sides (Go guard + C NULL
   check): a NULL dereference is a hardware fault a try/catch cannot stop.
5. libs modules and prebuilt branches are generated artifacts: regenerate
   via CI, never edit; local hand-built binaries must never be pushed to
   prebuilt branches (CI is the only publisher).
6. Everything in this repository is English-only.
7. Module zips must stay per-platform-small: never add binaries to the
   default branch or cross-platform requires outside the build-tag-guarded
   link files.

## License

MIT (see `LICENSE`). The prebuilt branches redistribute OpenCV binaries
under the
[Apache 2.0 license](https://github.com/opencv/opencv/blob/4.x/LICENSE)
(a copy ships in every libs module as `LICENSE-OPENCV.txt`) and zlib under
the zlib license.
