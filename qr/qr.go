//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

// Package qr encodes and decodes QR codes using OpenCV's official
// cv::QRCodeEncoder and cv::QRCodeDetector (the objdetect module).
//
// This is an optional feature set: importing this package makes `go build`
// download the extra static libraries for your platform only; programs that
// do not import it stay on the base libraries.
package qr

/*
#include <stdlib.h>

typedef void *Cv2Mat;

extern char *Cv2_QRDetectAndDecodeMulti(Cv2Mat img, char ***texts, double **corners, int *count);
extern char *Cv2_QREncode(const char *text, Cv2Mat out);
extern void Cv2_FreeStringArray(char **arr, int n);
extern void Cv2_FreeDoubleArray(double *arr);
extern void Cv2_FreeString(char *s);
*/
import "C"

import (
	"errors"
	"image"
	"unsafe"

	"github.com/hkloudou/cv2"
)

// Code is one decoded QR code.
type Code struct {
	// Text is the decoded payload.
	Text string
	// Corners is the code's quadrilateral in image coordinates.
	Corners [4]image.Point
}

func takeError(msg *C.char) error {
	if msg == nil {
		return nil
	}
	err := errors.New(C.GoString(msg))
	C.Cv2_FreeString(msg)
	return err
}

// Decode finds and decodes every QR code in img. An image without QR codes
// yields an empty slice; it panics only if the image is empty or OpenCV
// rejects the input.
func Decode(img image.Image) []Code {
	m, err := cv2.ImageToMatRGBA(img)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	var texts **C.char
	var corners *C.double
	var count C.int
	msg := C.Cv2_QRDetectAndDecodeMulti(C.Cv2Mat(m.Ptr()), &texts, &corners, &count)
	if err := takeError(msg); err != nil {
		panic("cv2/qr: Decode: " + err.Error())
	}
	n := int(count)
	if n == 0 {
		return nil
	}
	defer C.Cv2_FreeStringArray(texts, count)
	defer C.Cv2_FreeDoubleArray(corners)

	textSlice := unsafe.Slice(texts, n)
	cornerSlice := unsafe.Slice(corners, n*8)
	out := make([]Code, n)
	for i := 0; i < n; i++ {
		out[i].Text = C.GoString(textSlice[i])
		for c := 0; c < 4; c++ {
			out[i].Corners[c] = image.Pt(
				int(cornerSlice[i*8+c*2]+0.5),
				int(cornerSlice[i*8+c*2+1]+0.5),
			)
		}
	}
	return out
}

// Encode renders text as a QR code image. moduleSize is the side length in
// pixels of one QR module (1 yields the raw matrix); quietZone adds that
// many modules of white border around the code, as the QR specification
// recommends (4 is the standard value).
func Encode(text string, moduleSize, quietZone int) (image.Image, error) {
	if moduleSize < 1 {
		moduleSize = 1
	}
	if quietZone < 0 {
		quietZone = 0
	}
	raw := cv2.NewMat()
	defer raw.Close()

	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	if err := takeError(C.Cv2_QREncode(ctext, C.Cv2Mat(raw.Ptr()))); err != nil {
		return nil, errors.New("cv2/qr: Encode: " + err.Error())
	}

	data, err := raw.ToBytes()
	if err != nil {
		return nil, err
	}
	w, h := raw.Cols(), raw.Rows()

	outW := (w + 2*quietZone) * moduleSize
	outH := (h + 2*quietZone) * moduleSize
	img := image.NewGray(image.Rect(0, 0, outW, outH))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if data[y*w+x] != 0 {
				continue // white module, background already white
			}
			for dy := 0; dy < moduleSize; dy++ {
				row := ((y+quietZone)*moduleSize + dy) * outW
				col := (x + quietZone) * moduleSize
				for dx := 0; dx < moduleSize; dx++ {
					img.Pix[row+col+dx] = 0
				}
			}
		}
	}
	return img, nil
}
