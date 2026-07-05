//go:build windows && 386

package f2d

// Pull in the optional f2d feature-set static libraries for this platform.
import _ "github.com/hkloudou/cv2/libs/windows_386_f2d"
