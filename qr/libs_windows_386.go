//go:build windows && 386

package qr

// Pull in the optional qr feature-set static libraries for this platform.
import _ "github.com/hkloudou/cv2/libs/windows_386_qr"
