//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package qr

// Testing principle (applies to every C/C++-backed test in this repo): the
// official OpenCV implementation is assumed correct and serves as ground
// truth; these tests verify that the BINDING layer marshals data through it
// faithfully. The round trip below is OpenCV-vs-OpenCV: the official
// cv::QRCodeEncoder produces the image, the official cv::QRCodeDetector
// must read it back through our Go surface.

import (
	"image"
	"image/draw"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	const payload = "https://github.com/hkloudou/cv2?check=roundtrip-1"

	img, err := Encode(payload, 4, 4)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() < 21*4 {
		t.Fatalf("encoded image suspiciously small: %v", img.Bounds())
	}

	codes := Decode(img)
	if len(codes) != 1 {
		t.Fatalf("decoded %d codes, want 1", len(codes))
	}
	if codes[0].Text != payload {
		t.Fatalf("decoded %q, want %q", codes[0].Text, payload)
	}
}

// TestDecodeInScene places an encoded QR inside a larger scene at a known
// offset and checks both the payload and the reported corner geometry.
func TestDecodeInScene(t *testing.T) {
	const payload = "scene-embed-42"
	qrImg, err := Encode(payload, 4, 4)
	if err != nil {
		t.Fatal(err)
	}
	qb := qrImg.Bounds()

	scene := image.NewRGBA(image.Rect(0, 0, 400, 300))
	draw.Draw(scene, scene.Bounds(), image.White, image.Point{}, draw.Src)
	const offX, offY = 120, 80
	draw.Draw(scene, image.Rect(offX, offY, offX+qb.Dx(), offY+qb.Dy()), qrImg, qb.Min, draw.Src)

	codes := Decode(scene)
	if len(codes) != 1 || codes[0].Text != payload {
		t.Fatalf("decode in scene failed: %+v", codes)
	}
	for _, c := range codes[0].Corners {
		if c.X < offX-8 || c.X > offX+qb.Dx()+8 || c.Y < offY-8 || c.Y > offY+qb.Dy()+8 {
			t.Fatalf("corner %v outside the embedded QR region", c)
		}
	}
}

func TestDecodeNoQR(t *testing.T) {
	blank := image.NewRGBA(image.Rect(0, 0, 120, 90))
	draw.Draw(blank, blank.Bounds(), image.White, image.Point{}, draw.Src)
	if codes := Decode(blank); len(codes) != 0 {
		t.Fatalf("blank image decoded to %+v", codes)
	}
}

// TestEncodeGeometry checks the Go-side raster layer only (module scaling
// and quiet zone are ours; the QR matrix itself comes from the official
// encoder and is covered by the round-trip test).
func TestEncodeGeometry(t *testing.T) {
	const payload = "geometry-check"
	raw, err := Encode(payload, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	w := raw.Bounds().Dx()
	if w < 21 || w != raw.Bounds().Dy() {
		t.Fatalf("raw QR matrix %dx%d, want square >= 21", w, raw.Bounds().Dy())
	}

	scaled, err := Encode(payload, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	want := (w + 8) * 3
	if scaled.Bounds().Dx() != want || scaled.Bounds().Dy() != want {
		t.Fatalf("scaled QR is %dx%d, want %dx%d", scaled.Bounds().Dx(), scaled.Bounds().Dy(), want, want)
	}
}

func TestDecodeNilPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Decode(nil) did not panic")
		}
	}()
	Decode(nil)
}
