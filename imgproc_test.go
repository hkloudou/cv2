//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

import (
	"image"
	"strings"
	"testing"
)

// TestResizeMultiScaleMatch is the multi-scale matching scenario Resize
// exists for: both the scene and the template are upscaled 2x with
// nearest-neighbor (pixel-exact duplication), so the match must land at
// exactly twice the original coordinates.
func TestResizeMultiScaleMatch(t *testing.T) {
	const wantX, wantY = 60, 45
	parentImg := newNoiseImage(200, 150, 21)
	subImg := parentImg.SubImage(image.Rect(wantX, wantY, wantX+40, wantY+30))

	parent, err := ImageToMatRGBA(parentImg)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	sub, err := ImageToMatRGBA(subImg)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	parent2x := NewMat()
	defer parent2x.Close()
	sub2x := NewMat()
	defer sub2x.Close()
	Resize(parent, &parent2x, parent.Cols()*2, parent.Rows()*2, InterpolationNearestNeighbor)
	Resize(sub, &sub2x, sub.Cols()*2, sub.Rows()*2, InterpolationNearestNeighbor)

	if parent2x.Cols() != 400 || parent2x.Rows() != 300 {
		t.Fatalf("resized parent is %dx%d, want 400x300", parent2x.Cols(), parent2x.Rows())
	}

	result := NewMat()
	defer result.Close()
	mask := NewMat()
	defer mask.Close()
	MatchTemplate(parent2x, sub2x, &result, TmCcoeffNormed, mask)
	_, maxVal, _, maxLoc := MinMaxLoc(result)

	if maxLoc.X != wantX*2 || maxLoc.Y != wantY*2 {
		t.Fatalf("match at (%d, %d), want (%d, %d)", maxLoc.X, maxLoc.Y, wantX*2, wantY*2)
	}
	if maxVal < 0.999 {
		t.Fatalf("maxVal = %f, want >= 0.999", maxVal)
	}
}

func TestCvtColorGrayMatch(t *testing.T) {
	const wantX, wantY = 33, 27
	parentImg := newNoiseImage(160, 120, 22)
	subImg := parentImg.SubImage(image.Rect(wantX, wantY, wantX+24, wantY+18))

	parent, _ := ImageToMatRGBA(parentImg)
	defer parent.Close()
	sub, _ := ImageToMatRGBA(subImg)
	defer sub.Close()

	parentGray := NewMat()
	defer parentGray.Close()
	subGray := NewMat()
	defer subGray.Close()
	CvtColor(parent, &parentGray, ColorRGBAToGray)
	CvtColor(sub, &subGray, ColorRGBAToGray)

	if parentGray.Channels() != 1 || parentGray.Type() != MatTypeCV8UC1 {
		t.Fatalf("gray mat has %d channels type %d, want 1 channel CV8UC1", parentGray.Channels(), parentGray.Type())
	}
	if parentGray.Rows() != 120 || parentGray.Cols() != 160 {
		t.Fatalf("gray mat is %dx%d, want 160x120", parentGray.Cols(), parentGray.Rows())
	}

	result := NewMat()
	defer result.Close()
	mask := NewMat()
	defer mask.Close()
	MatchTemplate(parentGray, subGray, &result, TmCcoeffNormed, mask)
	_, maxVal, _, maxLoc := MinMaxLoc(result)

	if maxLoc.X != wantX || maxLoc.Y != wantY || maxVal < 0.999 {
		t.Fatalf("gray match at (%d, %d) val %f, want (%d, %d) >= 0.999", maxLoc.X, maxLoc.Y, maxVal, wantX, wantY)
	}
}

func TestThresholdOtsu(t *testing.T) {
	img, _ := ImageToMatRGBA(newNoiseImage(64, 48, 23))
	defer img.Close()
	gray := NewMat()
	defer gray.Close()
	CvtColor(img, &gray, ColorRGBAToGray)

	binary := NewMat()
	defer binary.Close()
	computed := Threshold(gray, &binary, 0, 255, ThresholdBinary|ThresholdOtsu)

	if computed <= 0 || computed >= 255 {
		t.Fatalf("OTSU computed threshold %f, want within (0, 255)", computed)
	}
	if binary.Rows() != 48 || binary.Cols() != 64 || binary.Channels() != 1 {
		t.Fatalf("unexpected binary mat shape %dx%dx%d", binary.Cols(), binary.Rows(), binary.Channels())
	}
}

func TestGaussianBlur(t *testing.T) {
	img, _ := ImageToMatRGBA(newNoiseImage(80, 60, 24))
	defer img.Close()
	blurred := NewMat()
	defer blurred.Close()
	GaussianBlur(img, &blurred, 5, 5, 0, 0)

	if blurred.Rows() != 60 || blurred.Cols() != 80 || blurred.Channels() != 4 {
		t.Fatalf("unexpected blurred mat shape %dx%dx%d", blurred.Cols(), blurred.Rows(), blurred.Channels())
	}
}

func TestImgprocErrorPaths(t *testing.T) {
	// Closed source Mat must panic with a message, not crash.
	func() {
		defer func() {
			r := recover()
			if r == nil || !strings.Contains(r.(string), "Resize") {
				t.Fatalf("Resize on closed Mat: unexpected panic value %v", r)
			}
		}()
		m := NewMat()
		m.Close()
		dst := NewMat()
		defer dst.Close()
		Resize(m, &dst, 10, 10, InterpolationLinear)
	}()

	// An OpenCV-rejected argument (invalid conversion source) must surface
	// as a panic carrying the OpenCV message.
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("CvtColor gray->gray on RGBA input did not panic")
			}
		}()
		img, _ := ImageToMatRGBA(newNoiseImage(8, 8, 25))
		defer img.Close()
		dst := NewMat()
		defer dst.Close()
		// GRAY2BGR on a 4-channel input is invalid.
		CvtColor(img, &dst, ColorGrayToBGR)
	}()

	// Shape accessors on closed Mats report -1 instead of crashing.
	m := NewMat()
	m.Close()
	if m.Rows() != -1 || m.Cols() != -1 || m.Channels() != -1 || m.Type() != -1 {
		t.Fatalf("closed Mat accessors = %d/%d/%d/%d, want all -1", m.Rows(), m.Cols(), m.Channels(), m.Type())
	}
}
