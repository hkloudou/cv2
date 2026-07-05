# prebuilt/linux-arm64

Machine-generated branch. Do not edit or commit here by hand.

This branch is force-pushed as a single commit by the `build-libs` GitHub
Actions workflow of [github.com/hkloudou/cv2](https://github.com/hkloudou/cv2).
It carries the prebuilt static libraries for linux/arm64:

- `libs/linux_arm64/` - the Go module `github.com/hkloudou/cv2/libs/linux_arm64`
  (release tags `libs/linux_arm64/vX.Y.Z` point into this branch)
- `sdk/` - the OpenCV install tree, kept so wrapper-only rebuilds can skip
  compiling OpenCV
- `MANIFEST` - build provenance and layer cache keys
