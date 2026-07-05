//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

import (
	"errors"
	"image"
	"image/draw"
)

// ErrEmptyImage is returned when an image with no pixels is converted.
var ErrEmptyImage = errors.New("cv2: empty image")

// ToImage converts a Mat back to a Go image: MatTypeCV8UC4 (RGBA byte
// order, as produced by ImageToMatRGBA) becomes *image.RGBA, and
// MatTypeCV8UC1 (e.g. a CvtColor gray result) becomes *image.Gray.
func (m Mat) ToImage() (image.Image, error) {
	data, err := m.ToBytes()
	if err != nil {
		return nil, err
	}
	w, h := m.Cols(), m.Rows()
	switch m.Type() {
	case MatTypeCV8UC4:
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		copy(img.Pix, data)
		return img, nil
	case MatTypeCV8UC1:
		img := image.NewGray(image.Rect(0, 0, w, h))
		copy(img.Pix, data)
		return img, nil
	default:
		return nil, ErrInvalidMatParams
	}
}

// ImageToMatRGBA converts an image.Image to a Mat of type MatTypeCV8UC4
// (8-bit RGBA). Every input image kind is normalized to RGBA so that two
// converted images are always directly comparable with MatchTemplate.
//
// The caller owns the returned Mat and must Close it.
func ImageToMatRGBA(img image.Image) (Mat, error) {
	if img == nil {
		return Mat{}, ErrEmptyImage
	}
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
