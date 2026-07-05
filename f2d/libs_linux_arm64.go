//go:build linux && arm64

package f2d

// Pull in the optional f2d feature-set static libraries for this platform.
import _ "github.com/hkloudou/cv2/libs/linux_arm64_f2d"
