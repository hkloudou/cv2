//go:build windows && amd64

package f2d

// Pull in the optional f2d feature-set static libraries for this platform.
import _ "github.com/hkloudou/cv2/libs/windows_amd64_f2d"
