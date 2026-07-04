// C ABI exposed to Go (cgo). The Go side re-declares these prototypes in its
// cgo preamble; keep both in sync. Everything is prefixed with Cv2_ so the
// symbols cannot collide with other C libraries linked into the same binary.
#ifndef CV2_CAPI_H_
#define CV2_CAPI_H_

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

#ifdef __cplusplus
#include <opencv2/core.hpp>
extern "C"
{
  typedef cv::Mat *Cv2Mat;
#else
typedef void *Cv2Mat;
#endif

  // Creates a new empty Mat. Returns NULL on allocation failure.
  Cv2Mat Cv2_Mat_New(void);

  // Creates a Mat that owns a private copy of buf. The caller keeps
  // ownership of buf and may free it as soon as the call returns. buf must
  // hold exactly rows*cols*elemSize(type) bytes. Returns NULL if OpenCV
  // rejects the parameters or allocation fails.
  Cv2Mat Cv2_Mat_NewFromBytes(int rows, int cols, int type, Cv2ByteArray buf);

  // Releases a Mat created by Cv2_Mat_New or Cv2_Mat_NewFromBytes.
  void Cv2_Mat_Close(Cv2Mat m);

  // Runs cv::matchTemplate. Returns NULL on success, otherwise a malloc'd
  // error message that the caller must release with Cv2_FreeString.
  char *Cv2_MatchTemplate(Cv2Mat image, Cv2Mat templ, Cv2Mat result, int method, Cv2Mat mask);

  // Runs cv::minMaxLoc. Returns NULL on success, otherwise a malloc'd
  // error message that the caller must release with Cv2_FreeString.
  char *Cv2_MinMaxLoc(Cv2Mat m, double *minVal, double *maxVal, Cv2Point *minLoc, Cv2Point *maxLoc);

  // Releases an error string returned by the functions above.
  void Cv2_FreeString(char *s);

#ifdef __cplusplus
}
#endif

#endif // CV2_CAPI_H_
