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

  // Shape accessors. Return -1 when m is NULL.
  int Cv2_Mat_Rows(Cv2Mat m);
  int Cv2_Mat_Cols(Cv2Mat m);
  int Cv2_Mat_Channels(Cv2Mat m);
  int Cv2_Mat_Type(Cv2Mat m);

  // cv::resize to width x height. Returns NULL on success, otherwise a
  // malloc'd error message to release with Cv2_FreeString.
  char *Cv2_Resize(Cv2Mat src, Cv2Mat dst, int width, int height, int interpolation);

  // cv::cvtColor. Same error contract as Cv2_Resize.
  char *Cv2_CvtColor(Cv2Mat src, Cv2Mat dst, int code);

  // cv::GaussianBlur with a ksizeW x ksizeH kernel. Same error contract.
  char *Cv2_GaussianBlur(Cv2Mat src, Cv2Mat dst, int ksizeW, int ksizeH,
                         double sigmaX, double sigmaY);

  // cv::threshold; the effective threshold (useful with OTSU/TRIANGLE) is
  // written to computed. Same error contract.
  char *Cv2_Threshold(Cv2Mat src, Cv2Mat dst, double thresh, double maxval,
                      int type, double *computed);

  // Runs cv::matchTemplate. Returns NULL on success, otherwise a malloc'd
  // error message that the caller must release with Cv2_FreeString.
  char *Cv2_MatchTemplate(Cv2Mat image, Cv2Mat templ, Cv2Mat result, int method, Cv2Mat mask);

  // Runs cv::minMaxLoc. Returns NULL on success, otherwise a malloc'd
  // error message that the caller must release with Cv2_FreeString.
  char *Cv2_MinMaxLoc(Cv2Mat m, double *minVal, double *maxVal, Cv2Point *minLoc, Cv2Point *maxLoc);

  // Releases an error string returned by the functions above.
  void Cv2_FreeString(char *s);

  // Internal helper shared by all wrapper translation units: returns a
  // malloc'd copy of msg (or a static sentinel when allocation fails) that
  // must be released with Cv2_FreeString. Not intended for Go callers.
  char *Cv2_CopyError(const char *msg);

#ifdef __cplusplus
}
#endif

#endif // CV2_CAPI_H_
