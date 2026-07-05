//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package f2d

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"testing"
)

// newSceneImage builds a deterministic scene that behaves like a natural
// image for ORB: a band-limited sinusoid background makes every location
// locally unique (so the ratio test keeps true matches), and blended
// rectangles add strong corners. Being band-limited, the structure
// survives rescaling, unlike pixel noise.
func newSceneImage(w, h int, seed int64) *image.RGBA {
	rng := rand.New(rand.NewSource(seed))
	type wave struct{ ax, ay, phase, amp float64 }
	waves := make([]wave, 6)
	for i := range waves {
		wavelength := 20 + rng.Float64()*60
		angle := rng.Float64() * 2 * math.Pi
		k := 2 * math.Pi / wavelength
		waves[i] = wave{k * math.Cos(angle), k * math.Sin(angle), rng.Float64() * 2 * math.Pi, 0.4 + rng.Float64()}
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := 0.0
			for _, wv := range waves {
				v += wv.amp * math.Sin(wv.ax*float64(x)+wv.ay*float64(y)+wv.phase)
			}
			g := uint8(128 + 80*v/4)
			img.SetRGBA(x, y, color.RGBA{g, g, g, 255})
		}
	}
	for i := 0; i < 50; i++ {
		x := rng.Intn(w - 20)
		y := rng.Intn(h - 20)
		rw := 8 + rng.Intn(40)
		rh := 8 + rng.Intn(30)
		c := color.RGBA{uint8(rng.Intn(200)), uint8(rng.Intn(200)), uint8(rng.Intn(200)), 170}
		draw.Draw(img, image.Rect(x, y, min(x+rw, w), min(y+rh, h)), &image.Uniform{c}, image.Point{}, draw.Over)
	}
	return img
}

// scaleNearest rescales by factor with nearest-neighbor sampling.
func scaleNearest(src *image.RGBA, factor float64) *image.RGBA {
	b := src.Bounds()
	w := int(float64(b.Dx()) * factor)
	h := int(float64(b.Dy()) * factor)
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := b.Min.Y + int(float64(y)/factor)
		for x := 0; x < w; x++ {
			sx := b.Min.X + int(float64(x)/factor)
			dst.SetRGBA(x, y, src.RGBAAt(sx, sy))
		}
	}
	return dst
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func TestLocateExactCrop(t *testing.T) {
	scene := newSceneImage(480, 360, 31)
	// The crop must be comfortably larger than ORB's 31px border margin on
	// each axis so enough interior keypoints survive.
	const x0, y0, x1, y1 = 140, 90, 340, 250
	sub := scene.SubImage(image.Rect(x0, y0, x1, y1))

	res := LocateWithOptions(scene, sub, Options{MaxFeatures: 1000})
	if !res.Found {
		t.Fatalf("exact crop not found (matched %d, inliers %d)", res.Matched, res.Inliers)
	}
	wantCenter := image.Pt((x0+x1)/2, (y0+y1)/2)
	if abs(res.Center.X-wantCenter.X) > 10 || abs(res.Center.Y-wantCenter.Y) > 10 {
		t.Fatalf("center %v, want within 10px of %v (corners %v)", res.Center, wantCenter, res.Corners)
	}
}

// TestLocateScaledTemplate is the scenario template matching cannot handle:
// the template is a 1.5x rescale of the region it must be found in.
func TestLocateScaledTemplate(t *testing.T) {
	scene := newSceneImage(480, 360, 32)
	const x0, y0, x1, y1 = 120, 80, 280, 220
	crop := scene.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
	// Tight copy, then upscale 1.5x.
	tight := image.NewRGBA(image.Rect(0, 0, x1-x0, y1-y0))
	draw.Draw(tight, tight.Bounds(), crop, crop.Bounds().Min, draw.Src)
	scaled := scaleNearest(tight, 1.5)

	res := LocateWithOptions(scene, scaled, Options{MaxFeatures: 1000})
	if !res.Found {
		t.Fatalf("scaled template not found (matched %d, inliers %d)", res.Matched, res.Inliers)
	}
	wantCenter := image.Pt((x0+x1)/2, (y0+y1)/2)
	if abs(res.Center.X-wantCenter.X) > 10 || abs(res.Center.Y-wantCenter.Y) > 10 {
		t.Fatalf("center %v, want within 10px of %v (corners %v)", res.Center, wantCenter, res.Corners)
	}
	if res.Inliers < 8 {
		t.Fatalf("only %d inliers", res.Inliers)
	}
}

// TestLocateAbsentTemplate must report not-found, not a false positive,
// when the template comes from a completely different scene.
func TestLocateAbsentTemplate(t *testing.T) {
	scene := newSceneImage(480, 360, 33)
	other := newSceneImage(200, 150, 99)

	res := Locate(scene, other)
	if res.Found {
		t.Fatalf("template from an unrelated scene reported found at %v (inliers %d)", res.Center, res.Inliers)
	}
}

func TestLocateNilImagePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Locate(nil, nil) did not panic")
		}
	}()
	Locate(nil, nil)
}

// TestLocateFeaturelessTemplate: a flat single-color template has no
// keypoints; the call must cleanly report not-found instead of erroring.
func TestLocateFeaturelessTemplate(t *testing.T) {
	scene := newSceneImage(320, 240, 34)
	flat := image.NewRGBA(image.Rect(0, 0, 60, 60))
	draw.Draw(flat, flat.Bounds(), &image.Uniform{color.RGBA{128, 128, 128, 255}}, image.Point{}, draw.Src)

	res := Locate(scene, flat)
	if res.Found {
		t.Fatal("featureless template reported found")
	}
}

