//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

const (
	// MatChannels1 is a single channel Mat.
	MatChannels1 = 0

	// MatChannels2 is a 2 channel Mat.
	MatChannels2 = 8

	// MatChannels3 is a 3 channel Mat.
	MatChannels3 = 16

	// MatChannels4 is a 4 channel Mat.
	MatChannels4 = 24
)

// MatType is the type for the various different kinds of Mat you can create.
type MatType int

const (
	// MatTypeCV8U is a Mat of 8-bit unsigned int
	MatTypeCV8U MatType = 0

	// MatTypeCV8S is a Mat of 8-bit signed int
	MatTypeCV8S MatType = 1

	// MatTypeCV16U is a Mat of 16-bit unsigned int
	MatTypeCV16U MatType = 2

	// MatTypeCV16S is a Mat of 16-bit signed int
	MatTypeCV16S MatType = 3

	// MatTypeCV32S is a Mat of 32-bit signed int
	MatTypeCV32S MatType = 4

	// MatTypeCV32F is a Mat of 32-bit float
	MatTypeCV32F MatType = 5

	// MatTypeCV64F is a Mat of 64-bit float
	MatTypeCV64F MatType = 6

	// MatTypeCV8UC1 is a Mat of 8-bit unsigned int with a single channel
	MatTypeCV8UC1 = MatTypeCV8U + MatChannels1

	// MatTypeCV8UC2 is a Mat of 8-bit unsigned int with 2 channels
	MatTypeCV8UC2 = MatTypeCV8U + MatChannels2

	// MatTypeCV8UC3 is a Mat of 8-bit unsigned int with 3 channels
	MatTypeCV8UC3 = MatTypeCV8U + MatChannels3

	// MatTypeCV8UC4 is a Mat of 8-bit unsigned int with 4 channels
	MatTypeCV8UC4 = MatTypeCV8U + MatChannels4

	// MatTypeCV32FC1 is a Mat of 32-bit float with a single channel
	MatTypeCV32FC1 = MatTypeCV32F + MatChannels1

	// MatTypeCV32FC2 is a Mat of 32-bit float with 2 channels
	MatTypeCV32FC2 = MatTypeCV32F + MatChannels2

	// MatTypeCV32FC3 is a Mat of 32-bit float with 3 channels
	MatTypeCV32FC3 = MatTypeCV32F + MatChannels3

	// MatTypeCV32FC4 is a Mat of 32-bit float with 4 channels
	MatTypeCV32FC4 = MatTypeCV32F + MatChannels4
)

// TemplateMatchMode is the comparison method for MatchTemplate.
type TemplateMatchMode int

const (
	// TmSqdiff maps to TM_SQDIFF
	TmSqdiff TemplateMatchMode = 0
	// TmSqdiffNormed maps to TM_SQDIFF_NORMED
	TmSqdiffNormed TemplateMatchMode = 1
	// TmCcorr maps to TM_CCORR
	TmCcorr TemplateMatchMode = 2
	// TmCcorrNormed maps to TM_CCORR_NORMED
	TmCcorrNormed TemplateMatchMode = 3
	// TmCcoeff maps to TM_CCOEFF
	TmCcoeff TemplateMatchMode = 4
	// TmCcoeffNormed maps to TM_CCOEFF_NORMED
	TmCcoeffNormed TemplateMatchMode = 5
)
