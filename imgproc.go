//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

/*
typedef void *Cv2Mat;

extern char *Cv2_Resize(Cv2Mat src, Cv2Mat dst, int width, int height, int interpolation);
extern char *Cv2_CvtColor(Cv2Mat src, Cv2Mat dst, int code);
extern char *Cv2_GaussianBlur(Cv2Mat src, Cv2Mat dst, int ksizeW, int ksizeH, double sigmaX, double sigmaY);
extern char *Cv2_Threshold(Cv2Mat src, Cv2Mat dst, double thresh, double maxval, int type, double *computed);
extern char *Cv2_Canny(Cv2Mat src, Cv2Mat dst, double threshold1, double threshold2, int apertureSize, int l2gradient);
extern char *Cv2_GetStructuringElement(int shape, int ksizeW, int ksizeH, Cv2Mat out);
extern char *Cv2_Erode(Cv2Mat src, Cv2Mat dst, Cv2Mat kernel, int iterations);
extern char *Cv2_Dilate(Cv2Mat src, Cv2Mat dst, Cv2Mat kernel, int iterations);
extern char *Cv2_GetRotationMatrix2D(double centerX, double centerY, double angle, double scale, Cv2Mat out);
extern char *Cv2_WarpAffine(Cv2Mat src, Cv2Mat dst, Cv2Mat m, int width, int height, int flags);
extern char *Cv2_FindExternalContourRects(Cv2Mat src, int **rects, int *count);
extern void Cv2_FreeIntArray(int *arr);
*/
import "C"

import (
	"image"
	"unsafe"
)

// InterpolationFlags are the resampling methods for Resize.
type InterpolationFlags int

const (
	// InterpolationNearestNeighbor maps to INTER_NEAREST.
	InterpolationNearestNeighbor InterpolationFlags = 0
	// InterpolationLinear maps to INTER_LINEAR.
	InterpolationLinear InterpolationFlags = 1
	// InterpolationCubic maps to INTER_CUBIC.
	InterpolationCubic InterpolationFlags = 2
	// InterpolationArea maps to INTER_AREA (best for shrinking).
	InterpolationArea InterpolationFlags = 3
	// InterpolationLanczos4 maps to INTER_LANCZOS4.
	InterpolationLanczos4 InterpolationFlags = 4
)

// ColorConversionCode is the conversion selector for CvtColor.
//
// Mats produced by ImageToMatRGBA hold RGBA byte order, so the RGBA-prefixed
// conversions are the ones that apply to them.
type ColorConversionCode int

const (
	// ColorBGRToBGRA maps to COLOR_BGR2BGRA.
	ColorBGRToBGRA ColorConversionCode = 0
	// ColorBGRAToBGR maps to COLOR_BGRA2BGR.
	ColorBGRAToBGR ColorConversionCode = 1
	// ColorBGRToRGBA maps to COLOR_BGR2RGBA.
	ColorBGRToRGBA ColorConversionCode = 2
	// ColorRGBAToBGR maps to COLOR_RGBA2BGR.
	ColorRGBAToBGR ColorConversionCode = 3
	// ColorBGRToRGB maps to COLOR_BGR2RGB.
	ColorBGRToRGB ColorConversionCode = 4
	// ColorBGRAToRGBA maps to COLOR_BGRA2RGBA.
	ColorBGRAToRGBA ColorConversionCode = 5
	// ColorBGRToGray maps to COLOR_BGR2GRAY.
	ColorBGRToGray ColorConversionCode = 6
	// ColorRGBToGray maps to COLOR_RGB2GRAY.
	ColorRGBToGray ColorConversionCode = 7
	// ColorGrayToBGR maps to COLOR_GRAY2BGR.
	ColorGrayToBGR ColorConversionCode = 8
	// ColorGrayToBGRA maps to COLOR_GRAY2BGRA.
	ColorGrayToBGRA ColorConversionCode = 9
	// ColorBGRAToGray maps to COLOR_BGRA2GRAY.
	ColorBGRAToGray ColorConversionCode = 10
	// ColorRGBAToGray maps to COLOR_RGBA2GRAY.
	ColorRGBAToGray ColorConversionCode = 11
)

// ThresholdType is the operation selector for Threshold.
type ThresholdType int

const (
	// ThresholdBinary maps to THRESH_BINARY.
	ThresholdBinary ThresholdType = 0
	// ThresholdBinaryInv maps to THRESH_BINARY_INV.
	ThresholdBinaryInv ThresholdType = 1
	// ThresholdTrunc maps to THRESH_TRUNC.
	ThresholdTrunc ThresholdType = 2
	// ThresholdToZero maps to THRESH_TOZERO.
	ThresholdToZero ThresholdType = 3
	// ThresholdToZeroInv maps to THRESH_TOZERO_INV.
	ThresholdToZeroInv ThresholdType = 4
	// ThresholdOtsu maps to THRESH_OTSU (combine with a base type).
	ThresholdOtsu ThresholdType = 8
	// ThresholdTriangle maps to THRESH_TRIANGLE (combine with a base type).
	ThresholdTriangle ThresholdType = 16
)

func requireMats(op string, dst *Mat, srcs ...Mat) {
	for _, m := range srcs {
		if m.p == nil {
			panic("cv2: " + op + ": nil Mat handle (create Mats with NewMat or NewMatFromBytes; a Mat is invalid after Close)")
		}
	}
	if dst == nil || dst.p == nil {
		panic("cv2: " + op + ": nil destination Mat handle")
	}
}

// Resize scales src into dst at width x height pixels.
//
// Combined with MatchTemplate this enables multi-scale matching: resize the
// template across a range of scales and keep the best MinMaxLoc result.
//
// See https://docs.opencv.org/4.x/da/d54/group__imgproc__transform.html#ga47a974309e9102f5f08231edc7e7529d
func Resize(src Mat, dst *Mat, width, height int, interpolation InterpolationFlags) {
	requireMats("Resize", dst, src)
	if err := takeError(C.Cv2_Resize(src.p, dst.p, C.int(width), C.int(height), C.int(interpolation))); err != nil {
		panic("cv2: Resize: " + err.Error())
	}
}

// CvtColor converts src into dst using the given color conversion code.
//
// See https://docs.opencv.org/4.x/d8/d01/group__imgproc__color__conversions.html#ga397ae87e1288a81d2363b61574eb8cab
func CvtColor(src Mat, dst *Mat, code ColorConversionCode) {
	requireMats("CvtColor", dst, src)
	if err := takeError(C.Cv2_CvtColor(src.p, dst.p, C.int(code))); err != nil {
		panic("cv2: CvtColor: " + err.Error())
	}
}

// GaussianBlur smooths src into dst with a ksizeW x ksizeH kernel (both must
// be odd or zero; zero means computed from sigma).
//
// See https://docs.opencv.org/4.x/d4/d86/group__imgproc__filter.html#gaabe8c836e97159a9193fb0b11ac52cf1
func GaussianBlur(src Mat, dst *Mat, ksizeW, ksizeH int, sigmaX, sigmaY float64) {
	requireMats("GaussianBlur", dst, src)
	if err := takeError(C.Cv2_GaussianBlur(src.p, dst.p, C.int(ksizeW), C.int(ksizeH), C.double(sigmaX), C.double(sigmaY))); err != nil {
		panic("cv2: GaussianBlur: " + err.Error())
	}
}

// Threshold applies a fixed-level (or OTSU/TRIANGLE computed) threshold to
// src, writing the result to dst, and returns the effective threshold value.
//
// See https://docs.opencv.org/4.x/d7/d1b/group__imgproc__misc.html#gae8a4a146d1ca78c626a53577199e9c57
func Threshold(src Mat, dst *Mat, thresh, maxval float64, typ ThresholdType) float64 {
	requireMats("Threshold", dst, src)
	var computed C.double
	if err := takeError(C.Cv2_Threshold(src.p, dst.p, C.double(thresh), C.double(maxval), C.int(typ), &computed)); err != nil {
		panic("cv2: Threshold: " + err.Error())
	}
	return float64(computed)
}

// MorphShape selects the kernel shape for GetStructuringElement.
type MorphShape int

const (
	// MorphRect maps to MORPH_RECT.
	MorphRect MorphShape = 0
	// MorphCross maps to MORPH_CROSS.
	MorphCross MorphShape = 1
	// MorphEllipse maps to MORPH_ELLIPSE.
	MorphEllipse MorphShape = 2
)

// Canny runs the Canny edge detector with OpenCV's default aperture (3)
// and L1 gradient, writing the 8-bit edge map to dst.
//
// See https://docs.opencv.org/4.x/dd/d1a/group__imgproc__feature.html#ga04723e007ed888ddf11d9ba04e2232de
func Canny(src Mat, dst *Mat, threshold1, threshold2 float64) {
	requireMats("Canny", dst, src)
	if err := takeError(C.Cv2_Canny(src.p, dst.p, C.double(threshold1), C.double(threshold2), 3, 0)); err != nil {
		panic("cv2: Canny: " + err.Error())
	}
}

// GetStructuringElement returns a morphology kernel of the given shape and
// size. The caller owns the returned Mat.
//
// See https://docs.opencv.org/4.x/d4/d86/group__imgproc__filter.html#gac342a1bb6eabf6f55c803b09268e36dc
func GetStructuringElement(shape MorphShape, ksizeW, ksizeH int) Mat {
	out := NewMat()
	if err := takeError(C.Cv2_GetStructuringElement(C.int(shape), C.int(ksizeW), C.int(ksizeH), out.p)); err != nil {
		out.Close()
		panic("cv2: GetStructuringElement: " + err.Error())
	}
	return out
}

// Erode erodes src into dst with the given kernel, default anchor and the
// given number of iterations.
//
// See https://docs.opencv.org/4.x/d4/d86/group__imgproc__filter.html#gaeb1e0c1033e3f6b891a25d0511362aeb
func Erode(src Mat, dst *Mat, kernel Mat, iterations int) {
	requireMats("Erode", dst, src, kernel)
	if err := takeError(C.Cv2_Erode(src.p, dst.p, kernel.p, C.int(iterations))); err != nil {
		panic("cv2: Erode: " + err.Error())
	}
}

// Dilate dilates src into dst with the given kernel, default anchor and the
// given number of iterations.
//
// See https://docs.opencv.org/4.x/d4/d86/group__imgproc__filter.html#ga4ff0f3318642c4f469d0e11f242f3b6c
func Dilate(src Mat, dst *Mat, kernel Mat, iterations int) {
	requireMats("Dilate", dst, src, kernel)
	if err := takeError(C.Cv2_Dilate(src.p, dst.p, kernel.p, C.int(iterations))); err != nil {
		panic("cv2: Dilate: " + err.Error())
	}
}

// GetRotationMatrix2D returns the 2x3 affine matrix (CV_64F) for rotating
// by angle degrees around (centerX, centerY) with the given scale, for use
// with WarpAffine. The caller owns the returned Mat.
//
// See https://docs.opencv.org/4.x/da/d54/group__imgproc__transform.html#gafbbc470ce83812914a70abfb604f4326
func GetRotationMatrix2D(centerX, centerY, angle, scale float64) Mat {
	out := NewMat()
	if err := takeError(C.Cv2_GetRotationMatrix2D(C.double(centerX), C.double(centerY), C.double(angle), C.double(scale), out.p)); err != nil {
		out.Close()
		panic("cv2: GetRotationMatrix2D: " + err.Error())
	}
	return out
}

// WarpAffine applies the 2x3 matrix m to src, producing a width x height
// dst with the given interpolation.
//
// See https://docs.opencv.org/4.x/da/d54/group__imgproc__transform.html#ga0203d9ee5fcd28d40dbc4a1ea4451983
func WarpAffine(src Mat, dst *Mat, m Mat, width, height int, interpolation InterpolationFlags) {
	requireMats("WarpAffine", dst, src, m)
	if err := takeError(C.Cv2_WarpAffine(src.p, dst.p, m.p, C.int(width), C.int(height), C.int(interpolation))); err != nil {
		panic("cv2: WarpAffine: " + err.Error())
	}
}

// FindExternalContourRects runs findContours(RETR_EXTERNAL,
// CHAIN_APPROX_SIMPLE) on a binary 8-bit image and returns the bounding
// rectangle of every external contour - a pragmatic primitive for locating
// solid regions (e.g. UI elements after thresholding).
//
// See https://docs.opencv.org/4.x/d3/dc0/group__imgproc__shape.html#gadf1ad6a0b82947fa1fe3c3d497f260e0
func FindExternalContourRects(src Mat) []image.Rectangle {
	if src.p == nil {
		panic("cv2: FindExternalContourRects: nil Mat handle (create Mats with NewMat or NewMatFromBytes; a Mat is invalid after Close)")
	}
	var rects *C.int
	var count C.int
	if err := takeError(C.Cv2_FindExternalContourRects(src.p, &rects, &count)); err != nil {
		panic("cv2: FindExternalContourRects: " + err.Error())
	}
	n := int(count)
	if n == 0 {
		return nil
	}
	defer C.Cv2_FreeIntArray(rects)
	vals := unsafe.Slice(rects, n*4)
	out := make([]image.Rectangle, n)
	for i := 0; i < n; i++ {
		x, y := int(vals[i*4]), int(vals[i*4+1])
		w, h := int(vals[i*4+2]), int(vals[i*4+3])
		out[i] = image.Rect(x, y, x+w, y+h)
	}
	return out
}
