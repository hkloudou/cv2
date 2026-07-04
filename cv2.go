// Package cv2 provides minimal, dependency-free Go bindings for OpenCV
// template matching ("find a small image inside a big image").
//
// Unlike gocv, this package does not require OpenCV to be installed on the
// machine that builds or runs your program. It links prebuilt static
// libraries (OpenCV core + imgproc only) that are published as per-platform
// Go modules under github.com/hkloudou/cv2/libs/... — `go build` downloads
// only the libraries for your target platform.
//
// Supported platforms:
//
//	linux/amd64, linux/386, linux/arm64
//	windows/amd64, windows/386 (MinGW-w64 toolchain, POSIX threads variant)
//
// CGO is required (CGO_ENABLED=1 and a C/C++ toolchain for your target).
//
// Basic usage:
//
//	minVal, minX, minY, maxVal, maxX, maxY := cv2.Match(screenshot, button)
//	// For TM_CCOEFF_NORMED the best match is at (maxX, maxY);
//	// maxVal close to 1.0 means a confident match.
package cv2
