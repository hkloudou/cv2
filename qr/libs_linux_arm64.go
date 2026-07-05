//go:build linux && arm64

package qr

// Pull in the optional qr feature-set static libraries for this platform.
import _ "github.com/hkloudou/cv2/libs/linux_arm64_qr"
