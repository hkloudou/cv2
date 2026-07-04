//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

import (
	"image"
	"math/rand"
	"strings"
	"testing"
)

// newNoiseImage builds a deterministic random RGBA image. Random noise makes
// template matching unambiguous: the template matches exactly one place.
func newNoiseImage(w, h int, seed int64) *image.RGBA {
	rng := rand.New(rand.NewSource(seed))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = uint8(rng.Intn(256))
		img.Pix[i+1] = uint8(rng.Intn(256))
		img.Pix[i+2] = uint8(rng.Intn(256))
		img.Pix[i+3] = 0xff
	}
	return img
}

func TestMatchFindsSubImage(t *testing.T) {
	const wantX, wantY = 73, 41
	parent := newNoiseImage(320, 240, 1)
	sub := parent.SubImage(image.Rect(wantX, wantY, wantX+48, wantY+36))

	minVal, _, _, maxVal, maxX, maxY := Match(parent, sub)

	if maxX != wantX || maxY != wantY {
		t.Fatalf("Match located (%d, %d), want (%d, %d)", maxX, maxY, wantX, wantY)
	}
	if maxVal < 0.999 {
		t.Fatalf("maxVal = %f, want >= 0.999 for an exact copy", maxVal)
	}
	if minVal < -1.001 || minVal > maxVal {
		t.Fatalf("minVal = %f out of range [-1, maxVal]", minVal)
	}
}

func TestMatchWholeImage(t *testing.T) {
	parent := newNoiseImage(120, 90, 2)

	_, _, _, maxVal, maxX, maxY := Match(parent, parent)

	if maxX != 0 || maxY != 0 {
		t.Fatalf("Match located (%d, %d), want (0, 0)", maxX, maxY)
	}
	if maxVal < 0.999 {
		t.Fatalf("maxVal = %f, want >= 0.999", maxVal)
	}
}

// TestMatchNonRGBAInput exercises the draw-based conversion path: the
// template is an *image.NRGBA copy of a parent region, so both images go
// through different conversion paths yet must still match exactly.
func TestMatchNonRGBAInput(t *testing.T) {
	const wantX, wantY = 17, 63
	parent := newNoiseImage(200, 160, 3)

	sub := image.NewNRGBA(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			sub.Set(x, y, parent.RGBAAt(wantX+x, wantY+y))
		}
	}

	_, _, _, maxVal, maxX, maxY := Match(parent, sub)

	if maxX != wantX || maxY != wantY {
		t.Fatalf("Match located (%d, %d), want (%d, %d)", maxX, maxY, wantX, wantY)
	}
	if maxVal < 0.999 {
		t.Fatalf("maxVal = %f, want >= 0.999", maxVal)
	}
}

// TestMatchGrayTemplate checks that mixing color models cannot crash: a
// grayscale template is normalized to RGBA, so OpenCV always sees matching
// types. The gray template correlates weakly but the call must succeed.
func TestMatchGrayTemplate(t *testing.T) {
	parent := newNoiseImage(100, 80, 4)
	sub := image.NewGray(image.Rect(0, 0, 20, 20))
	for i := range sub.Pix {
		sub.Pix[i] = uint8(i % 251)
	}

	_, _, _, maxVal, _, _ := Match(parent, sub)

	if maxVal < -1.001 || maxVal > 1.001 {
		t.Fatalf("maxVal = %f, want within [-1, 1]", maxVal)
	}
}

// TestMatchIncompatibleSizesPanics verifies that OpenCV errors are caught in
// the C++ wrapper and surface as a Go panic with a message, instead of
// aborting the process. A template that is wider but shorter than the image
// fails OpenCV's size assertion. (A template larger in BOTH dimensions does
// not error: matchTemplate silently swaps the two images.)
func TestMatchIncompatibleSizesPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Match with incompatible sizes did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "MatchTemplate") {
			t.Fatalf("unexpected panic value: %v", r)
		}
	}()

	parent := newNoiseImage(50, 50, 5)
	sub := newNoiseImage(100, 10, 6)
	Match(parent, sub)
}

func TestImageToMatRGBAEmptyImage(t *testing.T) {
	empty := image.NewRGBA(image.Rect(0, 0, 0, 0))
	if _, err := ImageToMatRGBA(empty); err == nil {
		t.Fatal("ImageToMatRGBA(empty) returned nil error")
	}
}

func TestMatLifecycle(t *testing.T) {
	m := NewMat()
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Closing twice must be safe.
	if err := m.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}

	if _, err := NewMatFromBytes(1, 1, MatTypeCV8UC4, nil); err == nil {
		t.Fatal("NewMatFromBytes with empty data returned nil error")
	}
}

func BenchmarkMatch(b *testing.B) {
	parent := newNoiseImage(1280, 720, 7)
	subRect := image.Rect(400, 300, 460, 340)
	sub := parent.SubImage(subRect)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Match(parent, sub)
	}
}

func TestMatchNilImagePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Match(nil, nil) did not panic")
		}
	}()
	Match(nil, nil)
}
