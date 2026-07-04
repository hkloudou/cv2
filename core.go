//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

package cv2

/*
#include <stdlib.h>

typedef struct Cv2ByteArray
{
	char *data;
	int length;
} Cv2ByteArray;

typedef struct Cv2Point
{
	int x;
	int y;
} Cv2Point;

typedef void *Cv2Mat;

extern Cv2Mat Cv2_Mat_New(void);
extern Cv2Mat Cv2_Mat_NewFromBytes(int rows, int cols, int type, Cv2ByteArray buf);
extern void Cv2_Mat_Close(Cv2Mat m);
extern char *Cv2_MatchTemplate(Cv2Mat image, Cv2Mat templ, Cv2Mat result, int method, Cv2Mat mask);
extern char *Cv2_MinMaxLoc(Cv2Mat m, double *minVal, double *maxVal, Cv2Point *minLoc, Cv2Point *maxLoc);
extern void Cv2_FreeString(char *s);
*/
import "C"

import (
	"errors"
	"image"
	"unsafe"
)

// ErrEmptyByteSlice is returned when a Mat is created from an empty byte slice.
var ErrEmptyByteSlice = errors.New("cv2: empty byte slice")

// Mat is an opaque handle to a native OpenCV cv::Mat.
//
// A Mat owns native memory: call Close when done with it.
type Mat struct {
	p C.Cv2Mat
}

// NewMat returns a new empty Mat.
func NewMat() Mat {
	return Mat{p: C.Cv2_Mat_New()}
}

// NewMatFromBytes returns a new Mat with the given size and type, initialized
// with a copy of data. The slice can be modified or garbage collected freely
// after the call returns.
func NewMatFromBytes(rows int, cols int, mt MatType, data []byte) (Mat, error) {
	if len(data) == 0 {
		return Mat{}, ErrEmptyByteSlice
	}
	buf := C.Cv2ByteArray{
		data:   (*C.char)(unsafe.Pointer(&data[0])),
		length: C.int(len(data)),
	}
	return Mat{p: C.Cv2_Mat_NewFromBytes(C.int(rows), C.int(cols), C.int(mt), buf)}, nil
}

// Close releases the native memory held by the Mat.
func (m *Mat) Close() error {
	if m.p != nil {
		C.Cv2_Mat_Close(m.p)
		m.p = nil
	}
	return nil
}

// takeError converts a C error message (or NULL) into a Go error,
// releasing the native string.
func takeError(msg *C.char) error {
	if msg == nil {
		return nil
	}
	err := errors.New(C.GoString(msg))
	C.Cv2_FreeString(msg)
	return err
}

// MatchTemplate compares a template against overlapped regions of image and
// writes the comparison map into result.
//
// It panics if OpenCV rejects the inputs (e.g. the template is larger than
// the image, or the types differ).
//
// See https://docs.opencv.org/4.x/df/dfb/group__imgproc__object.html#ga586ebfb0a7fb604b35a23d85391329be
func MatchTemplate(image Mat, templ Mat, result *Mat, method TemplateMatchMode, mask Mat) {
	if err := takeError(C.Cv2_MatchTemplate(image.p, templ.p, result.p, C.int(method), mask.p)); err != nil {
		panic("cv2: MatchTemplate: " + err.Error())
	}
}

// MinMaxLoc finds the global minimum and maximum values and their positions
// in a single-channel Mat (such as a MatchTemplate result).
//
// See https://docs.opencv.org/4.x/d2/de8/group__core__array.html#gab473bf2eb6d14ff97e89b355dac20707
func MinMaxLoc(input Mat) (minVal, maxVal float32, minLoc, maxLoc image.Point) {
	var cMinVal, cMaxVal C.double
	var cMinLoc, cMaxLoc C.Cv2Point

	if err := takeError(C.Cv2_MinMaxLoc(input.p, &cMinVal, &cMaxVal, &cMinLoc, &cMaxLoc)); err != nil {
		panic("cv2: MinMaxLoc: " + err.Error())
	}

	minLoc = image.Pt(int(cMinLoc.x), int(cMinLoc.y))
	maxLoc = image.Pt(int(cMaxLoc.x), int(cMaxLoc.y))
	return float32(cMinVal), float32(cMaxVal), minLoc, maxLoc
}
