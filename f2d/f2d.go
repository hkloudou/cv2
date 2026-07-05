//go:build (linux && (amd64 || 386 || arm64)) || (windows && (amd64 || 386)) || (darwin && arm64)

// Package f2d locates a small image inside a big one using ORB feature
// matching and a RANSAC-estimated homography. Unlike cv2.Match (template
// matching), it tolerates scaling, rotation, and moderate perspective
// change of the target.
//
// This is an optional feature set: importing this package makes `go build`
// download the extra static libraries (OpenCV features2d + calib3d, ~4 MB)
// for your platform only. Programs that do not import it stay on the small
// base libraries.
package f2d

/*
typedef void *Cv2Mat;

extern char *Cv2_FeatureLocate(Cv2Mat parent, Cv2Mat sub, int maxFeatures,
	double ratio, double ransacThreshold, int minInliers,
	double *corners, int *inliers, int *matched, int *found);
extern void Cv2_FreeString(char *s);
*/
import "C"

import (
	"errors"
	"image"

	"github.com/hkloudou/cv2"
)

// Result describes where (and how confidently) the template was located.
type Result struct {
	// Found reports whether a geometrically consistent location survived
	// RANSAC with at least MinInliers supporting matches.
	Found bool
	// Corners is the template's quadrilateral in parent coordinates
	// (top-left, top-right, bottom-right, bottom-left) — a rotated or
	// scaled target yields a non-axis-aligned quad.
	Corners [4]image.Point
	// Center is the centroid of Corners.
	Center image.Point
	// Inliers is the number of matches consistent with the homography.
	Inliers int
	// Matched is the number of matches that survived the ratio test.
	Matched int
}

// Options tunes Locate. The zero value selects the defaults documented on
// each field.
type Options struct {
	// MaxFeatures caps the ORB keypoints detected per image. Default 500.
	MaxFeatures int
	// Ratio is Lowe's ratio-test threshold. Default 0.75.
	Ratio float64
	// RansacThreshold is the RANSAC reprojection error in pixels. Default 5.
	RansacThreshold float64
	// MinInliers is the minimum consensus size to accept a location.
	// Default 8.
	MinInliers int
}

func (o Options) withDefaults() Options {
	if o.MaxFeatures <= 0 {
		o.MaxFeatures = 500
	}
	if o.Ratio <= 0 {
		o.Ratio = 0.75
	}
	if o.RansacThreshold <= 0 {
		o.RansacThreshold = 5
	}
	if o.MinInliers <= 0 {
		o.MinInliers = 8
	}
	return o
}

func takeError(msg *C.char) error {
	if msg == nil {
		return nil
	}
	err := errors.New(C.GoString(msg))
	C.Cv2_FreeString(msg)
	return err
}

// Locate finds sub inside parent with default Options.
//
// It panics if an image is empty or OpenCV rejects the inputs; a target
// that simply is not present yields Result{Found: false}.
func Locate(parent, sub image.Image) Result {
	return LocateWithOptions(parent, sub, Options{})
}

// LocateWithOptions is Locate with explicit tuning.
func LocateWithOptions(parent, sub image.Image, opts Options) Result {
	opts = opts.withDefaults()

	parentMat, err := cv2.ImageToMatRGBA(parent)
	if err != nil {
		panic(err)
	}
	defer parentMat.Close()
	subMat, err := cv2.ImageToMatRGBA(sub)
	if err != nil {
		panic(err)
	}
	defer subMat.Close()

	var corners [8]C.double
	var cInliers, cMatched, cFound C.int
	msg := C.Cv2_FeatureLocate(
		C.Cv2Mat(parentMat.Ptr()), C.Cv2Mat(subMat.Ptr()),
		C.int(opts.MaxFeatures), C.double(opts.Ratio),
		C.double(opts.RansacThreshold), C.int(opts.MinInliers),
		&corners[0], &cInliers, &cMatched, &cFound,
	)
	if err := takeError(msg); err != nil {
		panic("cv2/f2d: Locate: " + err.Error())
	}

	res := Result{
		Found:   cFound != 0,
		Inliers: int(cInliers),
		Matched: int(cMatched),
	}
	if res.Found {
		sumX, sumY := 0, 0
		for i := 0; i < 4; i++ {
			x := int(corners[i*2] + 0.5)
			y := int(corners[i*2+1] + 0.5)
			res.Corners[i] = image.Pt(x, y)
			sumX += x
			sumY += y
		}
		res.Center = image.Pt(sumX/4, sumY/4)
	}
	return res
}
