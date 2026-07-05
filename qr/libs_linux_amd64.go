//go:build linux && amd64

package qr

// Pull in the optional qr feature-set static libraries for this platform.
import _ "github.com/hkloudou/cv2/libs/linux_amd64_qr"
