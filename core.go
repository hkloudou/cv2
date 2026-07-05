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
extern char *Cv2_Mat_DataCopy(Cv2Mat m, char *dst, int dstLen);
extern int Cv2_Mat_Rows(Cv2Mat m);
extern int Cv2_Mat_Cols(Cv2Mat m);
extern int Cv2_Mat_Channels(Cv2Mat m);
extern int Cv2_Mat_Type(Cv2Mat m);
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

// ErrBadByteSliceLength is returned when the byte slice does not hold exactly
// rows*cols*elemSize bytes.
var ErrBadByteSliceLength = errors.New("cv2: byte slice length does not match rows*cols*element size")

// ErrInvalidMatParams is returned when OpenCV rejects the Mat parameters or
// cannot allocate the matrix.
var ErrInvalidMatParams = errors.New("cv2: invalid Mat dimensions, type, or allocation failure")

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
//
// data must hold exactly rows*cols*elemSize bytes: the C++ side copies that
// many bytes, so a shorter slice would let native code read past the Go
// allocation.
func NewMatFromBytes(rows int, cols int, mt MatType, data []byte) (Mat, error) {
	if len(data) == 0 {
		return Mat{}, ErrEmptyByteSlice
	}
	if rows <= 0 || cols <= 0 || mt.elemSize() <= 0 {
		return Mat{}, ErrInvalidMatParams
	}
	if len(data) != rows*cols*mt.elemSize() {
		return Mat{}, ErrBadByteSliceLength
	}
	buf := C.Cv2ByteArray{
		data:   (*C.char)(unsafe.Pointer(&data[0])),
		length: C.int(len(data)),
	}
	p := C.Cv2_Mat_NewFromBytes(C.int(rows), C.int(cols), C.int(mt), buf)
	if p == nil {
		return Mat{}, ErrInvalidMatParams
	}
	return Mat{p: p}, nil
}

// Close releases the native memory held by the Mat.
func (m *Mat) Close() error {
	if m.p != nil {
		C.Cv2_Mat_Close(m.p)
		m.p = nil
	}
	return nil
}

// ToBytes returns a copy of the Mat's pixel data in row-major order
// (rows*cols*elemSize bytes) — the inverse of NewMatFromBytes. Multi-byte
// elements (e.g. the float32 values of a MatchTemplate result) use the
// platform byte order; every supported platform is little-endian.
func (m Mat) ToBytes() ([]byte, error) {
	if m.p == nil {
		return nil, ErrInvalidMatParams
	}
	rows, cols := m.Rows(), m.Cols()
	size := rows * cols * m.Type().elemSize()
	if rows <= 0 || cols <= 0 || size <= 0 {
		return nil, ErrInvalidMatParams
	}
	buf := make([]byte, size)
	if err := takeError(C.Cv2_Mat_DataCopy(m.p, (*C.char)(unsafe.Pointer(&buf[0])), C.int(size))); err != nil {
		return nil, err
	}
	return buf, nil
}

// Ptr exposes the native cv::Mat handle for subpackages (such as
// github.com/hkloudou/cv2/f2d) and advanced interop. It is nil for a closed
// or zero-value Mat. Treat it as opaque.
func (m Mat) Ptr() unsafe.Pointer {
	return unsafe.Pointer(m.p)
}

// Rows returns the number of rows, or -1 for a closed or zero-value Mat.
func (m Mat) Rows() int {
	return int(C.Cv2_Mat_Rows(m.p))
}

// Cols returns the number of columns, or -1 for a closed or zero-value Mat.
func (m Mat) Cols() int {
	return int(C.Cv2_Mat_Cols(m.p))
}

// Channels returns the channel count, or -1 for a closed or zero-value Mat.
func (m Mat) Channels() int {
	return int(C.Cv2_Mat_Channels(m.p))
}

// Type returns the OpenCV type code, or -1 for a closed or zero-value Mat.
func (m Mat) Type() MatType {
	return MatType(C.Cv2_Mat_Type(m.p))
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
	if image.p == nil || templ.p == nil || result == nil || result.p == nil || mask.p == nil {
		panic("cv2: MatchTemplate: nil Mat handle (create Mats with NewMat or NewMatFromBytes; a Mat is invalid after Close)")
	}
	if err := takeError(C.Cv2_MatchTemplate(image.p, templ.p, result.p, C.int(method), mask.p)); err != nil {
		panic("cv2: MatchTemplate: " + err.Error())
	}
}

// MinMaxLoc finds the global minimum and maximum values and their positions
// in a single-channel Mat (such as a MatchTemplate result).
//
// See https://docs.opencv.org/4.x/d2/de8/group__core__array.html#gab473bf2eb6d14ff97e89b355dac20707
func MinMaxLoc(input Mat) (minVal, maxVal float32, minLoc, maxLoc image.Point) {
	if input.p == nil {
		panic("cv2: MinMaxLoc: nil Mat handle (create Mats with NewMat or NewMatFromBytes; a Mat is invalid after Close)")
	}
	var cMinVal, cMaxVal C.double
	var cMinLoc, cMaxLoc C.Cv2Point

	if err := takeError(C.Cv2_MinMaxLoc(input.p, &cMinVal, &cMaxVal, &cMinLoc, &cMaxLoc)); err != nil {
		panic("cv2: MinMaxLoc: " + err.Error())
	}

	minLoc = image.Pt(int(cMinLoc.x), int(cMinLoc.y))
	maxLoc = image.Pt(int(cMaxLoc.x), int(cMaxLoc.y))
	return float32(cMinVal), float32(cMaxVal), minLoc, maxLoc
}
