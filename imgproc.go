//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

/*
typedef void *Cv2Mat;

extern char *Cv2_Resize(Cv2Mat src, Cv2Mat dst, int width, int height, int interpolation);
extern char *Cv2_CvtColor(Cv2Mat src, Cv2Mat dst, int code);
extern char *Cv2_GaussianBlur(Cv2Mat src, Cv2Mat dst, int ksizeW, int ksizeH, double sigmaX, double sigmaY);
extern char *Cv2_Threshold(Cv2Mat src, Cv2Mat dst, double thresh, double maxval, int type, double *computed);
*/
import "C"

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
