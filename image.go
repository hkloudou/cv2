//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386))

package cv2

import (
	"errors"
	"image"
	"image/draw"
)

// ErrEmptyImage is returned when an image with no pixels is converted.
var ErrEmptyImage = errors.New("cv2: empty image")

// ImageToMatRGBA converts an image.Image to a Mat of type MatTypeCV8UC4
// (8-bit RGBA). Every input image kind is normalized to RGBA so that two
// converted images are always directly comparable with MatchTemplate.
//
// The caller owns the returned Mat and must Close it.
func ImageToMatRGBA(img image.Image) (Mat, error) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return Mat{}, ErrEmptyImage
	}

	// Fast path: a full-frame *image.RGBA can be handed over as-is.
	if m, ok := img.(*image.RGBA); ok && bounds.Min.X == 0 && bounds.Min.Y == 0 && m.Stride == w*4 {
		return NewMatFromBytes(h, w, MatTypeCV8UC4, m.Pix)
	}

	// Everything else (sub-images, NRGBA, YCbCr/JPEG, Gray, Paletted, ...)
	// is redrawn into a tightly packed RGBA buffer.
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)
	return NewMatFromBytes(h, w, MatTypeCV8UC4, rgba.Pix)
}
