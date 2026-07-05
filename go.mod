module github.com/hkloudou/cv2

go 1.24.0

require (
	github.com/hkloudou/cv2/libs/darwin_arm64 v0.50000.0
	github.com/hkloudou/cv2/libs/darwin_arm64_f2d v0.50000.0
	github.com/hkloudou/cv2/libs/darwin_arm64_qr v0.50000.0
	github.com/hkloudou/cv2/libs/linux_386 v0.50000.0
	github.com/hkloudou/cv2/libs/linux_386_f2d v0.50000.0
	github.com/hkloudou/cv2/libs/linux_386_qr v0.50000.0
	github.com/hkloudou/cv2/libs/linux_amd64 v0.50000.0
	github.com/hkloudou/cv2/libs/linux_amd64_f2d v0.50000.0
	github.com/hkloudou/cv2/libs/linux_amd64_qr v0.50000.0
	github.com/hkloudou/cv2/libs/linux_arm64 v0.50000.0
	github.com/hkloudou/cv2/libs/linux_arm64_f2d v0.50000.0
	github.com/hkloudou/cv2/libs/linux_arm64_qr v0.50000.0
	github.com/hkloudou/cv2/libs/windows_386 v0.50000.0
	github.com/hkloudou/cv2/libs/windows_386_f2d v0.50000.0
	github.com/hkloudou/cv2/libs/windows_386_qr v0.50000.0
	github.com/hkloudou/cv2/libs/windows_amd64 v0.50000.0
	github.com/hkloudou/cv2/libs/windows_amd64_f2d v0.50000.0
	github.com/hkloudou/cv2/libs/windows_amd64_qr v0.50000.0
)

retract (
	v0.2.0
	// Bootstrap-era versions predating the OpenCV-encoded version scheme;
	// their tags were removed. Use the v0.40801.x or v0.41200.x lines.
	v0.1.0
)
