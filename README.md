# prebuilt/windows-386

Machine-generated branch. Do not edit or commit here by hand.

This branch is force-pushed as a single commit by the `build-libs` GitHub
Actions workflow of [github.com/hkloudou/cv2](https://github.com/hkloudou/cv2).
It carries the prebuilt static libraries for windows/386:

- `libs/windows_386/` - the Go module `github.com/hkloudou/cv2/libs/windows_386`
  (release tags `libs/windows_386/vX.Y.Z` point into this branch)
- `sdk/` - the OpenCV install tree, kept so wrapper-only rebuilds can skip
  compiling OpenCV
- `MANIFEST` - build provenance and layer cache keys
