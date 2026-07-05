//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

// Native-parity tests: each test mirrors the reference computation used by
// OpenCV's own accuracy tests, so the Go bindings are checked against the
// documented native semantics rather than against themselves.
//
// References into the OpenCV source tree (same versions we build):
//   - modules/imgproc/test/test_templmatch.cpp   (naive TM_CCOEFF_NORMED)
//   - modules/imgproc/src/color.simd_helpers.hpp (fixed-point RGB->gray)
//   - modules/imgproc/test/test_imgwarp.cpp      (INTER_NEAREST mapping)
//   - modules/imgproc/test/test_thresh.cpp       (THRESH_BINARY semantics)
//   - modules/core/test/test_arithm.cpp          (minMaxLoc semantics)

import (
	"encoding/binary"
	"math"
	"math/rand"
	"testing"
)

func noiseBytes(n int, seed int64) []byte {
	rng := rand.New(rand.NewSource(seed))
	b := make([]byte, n)
	for i := range b {
		b[i] = uint8(rng.Intn(256))
	}
	return b
}

func float32sOf(t *testing.T, m Mat) []float32 {
	t.Helper()
	raw, err := m.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes: %v", err)
	}
	out := make([]float32, len(raw)/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[i*4:]))
	}
	return out
}

// refCcoeffNormed is the naive TM_CCOEFF_NORMED from OpenCV's accuracy test:
// per-channel means are subtracted (matchTemplate docs), the correlation is
// summed over all channels, and the result is normalized by the two
// zero-mean energies.
func refCcoeffNormed(img []byte, iw, ih int, tpl []byte, tw, th, ch int) []float64 {
	rw, rh := iw-tw+1, ih-th+1
	out := make([]float64, rw*rh)

	tplMean := make([]float64, ch)
	for c := 0; c < ch; c++ {
		sum := 0.0
		for y := 0; y < th; y++ {
			for x := 0; x < tw; x++ {
				sum += float64(tpl[(y*tw+x)*ch+c])
			}
		}
		tplMean[c] = sum / float64(tw*th)
	}

	for ry := 0; ry < rh; ry++ {
		for rx := 0; rx < rw; rx++ {
			imgMean := make([]float64, ch)
			for c := 0; c < ch; c++ {
				sum := 0.0
				for y := 0; y < th; y++ {
					for x := 0; x < tw; x++ {
						sum += float64(img[((ry+y)*iw+rx+x)*ch+c])
					}
				}
				imgMean[c] = sum / float64(tw*th)
			}
			var num, dT, dI float64
			for c := 0; c < ch; c++ {
				for y := 0; y < th; y++ {
					for x := 0; x < tw; x++ {
						tv := float64(tpl[(y*tw+x)*ch+c]) - tplMean[c]
						iv := float64(img[((ry+y)*iw+rx+x)*ch+c]) - imgMean[c]
						num += tv * iv
						dT += tv * tv
						dI += iv * iv
					}
				}
			}
			denom := math.Sqrt(dT * dI)
			if denom != 0 {
				out[ry*rw+rx] = num / denom
			}
		}
	}
	return out
}

func TestParityMatchTemplateCcoeffNormed(t *testing.T) {
	const iw, ih, tw, th, ch = 20, 15, 6, 5, 4
	imgBytes := noiseBytes(iw*ih*ch, 41)
	tplBytes := noiseBytes(tw*th*ch, 42)
	// Constant alpha, as ImageToMatRGBA produces: contributes zero variance.
	for i := 3; i < len(imgBytes); i += 4 {
		imgBytes[i] = 0xff
	}
	for i := 3; i < len(tplBytes); i += 4 {
		tplBytes[i] = 0xff
	}

	img, err := NewMatFromBytes(ih, iw, MatTypeCV8UC4, imgBytes)
	if err != nil {
		t.Fatal(err)
	}
	defer img.Close()
	tpl, err := NewMatFromBytes(th, tw, MatTypeCV8UC4, tplBytes)
	if err != nil {
		t.Fatal(err)
	}
	defer tpl.Close()

	result := NewMat()
	defer result.Close()
	mask := NewMat()
	defer mask.Close()
	MatchTemplate(img, tpl, &result, TmCcoeffNormed, mask)

	if result.Rows() != ih-th+1 || result.Cols() != iw-tw+1 || result.Type() != MatTypeCV32FC1 {
		t.Fatalf("result shape %dx%d type %d, want %dx%d type %d",
			result.Cols(), result.Rows(), result.Type(), iw-tw+1, ih-th+1, MatTypeCV32FC1)
	}

	got := float32sOf(t, result)
	want := refCcoeffNormed(imgBytes, iw, ih, tplBytes, tw, th, ch)
	for i := range want {
		if math.Abs(float64(got[i])-want[i]) > 1e-3 {
			t.Fatalf("result[%d] = %v, reference %v (diff %g)", i, got[i], want[i], math.Abs(float64(got[i])-want[i]))
		}
	}
}

// TestParityCvtColorRGBAToGray checks OpenCV's fixed-point luma transform:
// gray = (R*4899 + G*9617 + B*1868 + 2^13) >> 14, coefficients from
// modules/imgproc/src/color.simd_helpers.hpp (yuv_shift = 14).
func TestParityCvtColorRGBAToGray(t *testing.T) {
	const w, h = 16, 9
	src := noiseBytes(w*h*4, 43)
	m, err := NewMatFromBytes(h, w, MatTypeCV8UC4, src)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	gray := NewMat()
	defer gray.Close()
	CvtColor(m, &gray, ColorRGBAToGray)

	got, err := gray.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < w*h; i++ {
		r := int(src[i*4])
		g := int(src[i*4+1])
		b := int(src[i*4+2])
		want := (r*4899 + g*9617 + b*1868 + (1 << 13)) >> 14
		diff := int(got[i]) - want
		if diff < -1 || diff > 1 {
			t.Fatalf("gray[%d] = %d, reference %d (RGB %d,%d,%d)", i, got[i], want, r, g, b)
		}
	}
}

// TestParityResizeNearest2x: for an exact 2x upscale INTER_NEAREST maps
// dst(x, y) = src(x/2, y/2); the output must be byte-identical.
func TestParityResizeNearest2x(t *testing.T) {
	const w, h = 9, 7
	src := noiseBytes(w*h*4, 44)
	m, err := NewMatFromBytes(h, w, MatTypeCV8UC4, src)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	dst := NewMat()
	defer dst.Close()
	Resize(m, &dst, w*2, h*2, InterpolationNearestNeighbor)

	got, err := dst.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < h*2; y++ {
		for x := 0; x < w*2; x++ {
			for c := 0; c < 4; c++ {
				want := src[((y/2)*w+x/2)*4+c]
				if got[(y*w*2+x)*4+c] != want {
					t.Fatalf("dst(%d,%d)[%d] = %d, want %d", x, y, c, got[(y*w*2+x)*4+c], want)
				}
			}
		}
	}
}

// TestParityThresholdBinary: THRESH_BINARY is dst = (src > thresh) ? maxval
// : 0, strictly greater, byte-exact.
func TestParityThresholdBinary(t *testing.T) {
	src := make([]byte, 256)
	for i := range src {
		src[i] = uint8(i)
	}
	m, err := NewMatFromBytes(16, 16, MatTypeCV8UC1, src)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	dst := NewMat()
	defer dst.Close()
	const thresh, maxval = 127, 200
	Threshold(m, &dst, thresh, maxval, ThresholdBinary)

	got, err := dst.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	for i := range src {
		want := byte(0)
		if src[i] > thresh {
			want = maxval
		}
		if got[i] != want {
			t.Fatalf("dst[%d] = %d, want %d (src %d)", i, got[i], want, src[i])
		}
	}
}

// TestParityMinMaxLoc: unique extrema must be reported at their exact
// positions with their exact values.
func TestParityMinMaxLoc(t *testing.T) {
	const w, h = 13, 8
	src := make([]byte, w*h)
	for i := range src {
		src[i] = 100
	}
	src[3*w+7] = 2   // unique minimum at (7, 3)
	src[6*w+11] = 250 // unique maximum at (11, 6)

	m, err := NewMatFromBytes(h, w, MatTypeCV8UC1, src)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	minVal, maxVal, minLoc, maxLoc := MinMaxLoc(m)
	if minVal != 2 || minLoc.X != 7 || minLoc.Y != 3 {
		t.Fatalf("min %v at %v, want 2 at (7,3)", minVal, minLoc)
	}
	if maxVal != 250 || maxLoc.X != 11 || maxLoc.Y != 6 {
		t.Fatalf("max %v at %v, want 250 at (11,6)", maxVal, maxLoc)
	}
}

// TestParityToBytesRoundTrip: NewMatFromBytes -> ToBytes must be the
// identity for continuous mats.
func TestParityToBytesRoundTrip(t *testing.T) {
	src := noiseBytes(12*10*4, 45)
	m, err := NewMatFromBytes(10, 12, MatTypeCV8UC4, src)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	got, err := m.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(src) {
		t.Fatalf("length %d, want %d", len(got), len(src))
	}
	for i := range src {
		if got[i] != src[i] {
			t.Fatalf("byte %d differs: %d vs %d", i, got[i], src[i])
		}
	}
}
