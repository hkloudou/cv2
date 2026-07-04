//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386))

package cv2

import (
	"image"
)

// Match looks for sub inside parent using normalized correlation coefficient
// template matching (TM_CCOEFF_NORMED).
//
// It returns, in order: the minimum match value and its X/Y position, then
// the maximum match value and its X/Y position. For TM_CCOEFF_NORMED the
// best candidate is the maximum: a maxVal close to 1.0 means a confident
// match whose top-left corner is at (maxX, maxY) in parent coordinates.
//
// Match panics if an image is empty or if OpenCV rejects the inputs
// (for example when sub is larger than parent).
func Match(parent, sub image.Image) (float32, int, int, float32, int, int) {
	parentMat, err := ImageToMatRGBA(parent)
	if err != nil {
		panic(err)
	}
	defer parentMat.Close()

	subMat, err := ImageToMatRGBA(sub)
	if err != nil {
		panic(err)
	}
	defer subMat.Close()

	result := NewMat()
	defer result.Close()
	mask := NewMat()
	defer mask.Close()

	MatchTemplate(parentMat, subMat, &result, TmCcoeffNormed, mask)

	minVal, maxVal, minLoc, maxLoc := MinMaxLoc(result)
	return minVal, minLoc.X, minLoc.Y, maxVal, maxLoc.X, maxLoc.Y
}
