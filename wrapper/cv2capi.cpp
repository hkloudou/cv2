// Implementation of the C ABI declared in cv2capi.h.
//
// This file is compiled ahead of time (by build/build-wrapper.sh, normally
// inside GitHub Actions) into libcv2wrapper.a, so packages importing
// github.com/hkloudou/cv2 never need OpenCV headers or a C++ compiler
// beyond what cgo already requires.

#include "cv2capi.h"

#include <opencv2/imgproc.hpp>

#include <cmath>
#include <cstdlib>
#include <cstring>
#include <exception>
#include <vector>

namespace
{
  // The Go side speaks a fixed "wire" Mat type encoding - the classic
  // OpenCV 4.x layout (depth in bits 0-2, channels-1 from bit 3) - so the
  // same Go binaries work against every OpenCV line. OpenCV 5 changed
  // CV_CN_SHIFT from 3 to 5, so translate at the boundary.
  int type_to_native(int wire)
  {
#if CV_VERSION_MAJOR >= 5
    return CV_MAKETYPE(wire & 7, (wire >> 3) + 1);
#else
    return wire;
#endif
  }

  int type_from_native(int native)
  {
#if CV_VERSION_MAJOR >= 5
    return CV_MAT_DEPTH(native) + ((CV_MAT_CN(native) - 1) << 3);
#else
    return native;
#endif
  }

  // Static fallback used when malloc fails; the error channel must never
  // alias success (NULL). Never passed to free().
  const char kErrorAllocFailed[] = "C++ exception (message lost: allocation failure)";

  // Returns a malloc'd copy of msg so it can cross the cgo boundary and be
  // freed by Cv2_FreeString.
  char *copy_error(const char *msg)
  {
    const size_t len = std::strlen(msg) + 1;
    char *out = static_cast<char *>(std::malloc(len));
    if (out == nullptr)
    {
      return const_cast<char *>(kErrorAllocFailed);
    }
    std::memcpy(out, msg, len);
    return out;
  }

  char *current_exception_message()
  {
    try
    {
      throw;
    }
    catch (const std::exception &e)
    {
      return copy_error(e.what());
    }
    catch (...)
    {
      return copy_error("unknown C++ exception");
    }
  }
} // namespace

Cv2Mat Cv2_Mat_New(void)
{
  try
  {
    return new cv::Mat();
  }
  catch (...)
  {
    // No error channel in this signature; NULL signals failure to Go.
    return nullptr;
  }
}

Cv2Mat Cv2_Mat_NewFromBytes(int rows, int cols, int type, Cv2ByteArray buf)
{
  try
  {
    // Wrap the caller's buffer without copying, then clone so the returned
    // Mat owns its pixels. The Go garbage collector is therefore free to
    // move or collect the source slice after this call returns. The Go side
    // guarantees buf holds exactly rows*cols*elemSize bytes.
    const cv::Mat borrowed(rows, cols, type_to_native(type), buf.data);
    return new cv::Mat(borrowed.clone());
  }
  catch (...)
  {
    return nullptr;
  }
}

void Cv2_Mat_Close(Cv2Mat m)
{
  delete m;
}

char *Cv2_Mat_DataCopy(Cv2Mat m, char *dst, int dstLen)
{
  if (m == nullptr || dst == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    const cv::Mat continuous = m->isContinuous() ? *m : m->clone();
    const size_t total = continuous.total() * continuous.elemSize();
    if ((size_t)dstLen != total)
    {
      return copy_error("destination length does not match Mat data size");
    }
    std::memcpy(dst, continuous.ptr(), total);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

int Cv2_Mat_Rows(Cv2Mat m)
{
  return m == nullptr ? -1 : m->rows;
}

int Cv2_Mat_Cols(Cv2Mat m)
{
  return m == nullptr ? -1 : m->cols;
}

int Cv2_Mat_Channels(Cv2Mat m)
{
  return m == nullptr ? -1 : m->channels();
}

int Cv2_Mat_Type(Cv2Mat m)
{
  return m == nullptr ? -1 : type_from_native(m->type());
}

char *Cv2_Resize(Cv2Mat src, Cv2Mat dst, int width, int height, int interpolation)
{
  if (src == nullptr || dst == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::resize(*src, *dst, cv::Size(width, height), 0, 0, interpolation);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_CvtColor(Cv2Mat src, Cv2Mat dst, int code)
{
  if (src == nullptr || dst == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::cvtColor(*src, *dst, code);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_GaussianBlur(Cv2Mat src, Cv2Mat dst, int ksizeW, int ksizeH,
                       double sigmaX, double sigmaY)
{
  if (src == nullptr || dst == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::GaussianBlur(*src, *dst, cv::Size(ksizeW, ksizeH), sigmaX, sigmaY);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_Threshold(Cv2Mat src, Cv2Mat dst, double thresh, double maxval,
                    int type, double *computed)
{
  if (src == nullptr || dst == nullptr || computed == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    *computed = cv::threshold(*src, *dst, thresh, maxval, type);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_MatchTemplate(Cv2Mat image, Cv2Mat templ, Cv2Mat result, int method, Cv2Mat mask)
{
  // A NULL dereference is a hardware fault, not a C++ exception; check
  // before the try block can pretend to help.
  if (image == nullptr || templ == nullptr || result == nullptr || mask == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::matchTemplate(*image, *templ, *result, method, *mask);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_MinMaxLoc(Cv2Mat m, double *minVal, double *maxVal, Cv2Point *minLoc, Cv2Point *maxLoc)
{
  if (m == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::Point cMinLoc;
    cv::Point cMaxLoc;
    cv::minMaxLoc(*m, minVal, maxVal, &cMinLoc, &cMaxLoc);

    minLoc->x = cMinLoc.x;
    minLoc->y = cMinLoc.y;
    maxLoc->x = cMaxLoc.x;
    maxLoc->y = cMaxLoc.y;
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_Canny(Cv2Mat src, Cv2Mat dst, double threshold1, double threshold2,
                int apertureSize, int l2gradient)
{
  if (src == nullptr || dst == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::Canny(*src, *dst, threshold1, threshold2, apertureSize, l2gradient != 0);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_GetStructuringElement(int shape, int ksizeW, int ksizeH, Cv2Mat out)
{
  if (out == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    *out = cv::getStructuringElement(shape, cv::Size(ksizeW, ksizeH));
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_Erode(Cv2Mat src, Cv2Mat dst, Cv2Mat kernel, int iterations)
{
  if (src == nullptr || dst == nullptr || kernel == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::erode(*src, *dst, *kernel, cv::Point(-1, -1), iterations);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_Dilate(Cv2Mat src, Cv2Mat dst, Cv2Mat kernel, int iterations)
{
  if (src == nullptr || dst == nullptr || kernel == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::dilate(*src, *dst, *kernel, cv::Point(-1, -1), iterations);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_GetRotationMatrix2D(double centerX, double centerY, double angle,
                              double scale, Cv2Mat out)
{
  if (out == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    // The documented cv::getRotationMatrix2D closed form (OpenCV 5 moved
    // the helper into the geometry module, which the base library set
    // deliberately excludes):
    //   [  alpha  beta  (1-alpha)*cx - beta*cy ]
    //   [ -beta   alpha  beta*cx + (1-alpha)*cy ]
    const double rad = angle * CV_PI / 180.0;
    const double alpha = std::cos(rad) * scale;
    const double beta = std::sin(rad) * scale;
    out->create(2, 3, CV_64F);
    double *m = out->ptr<double>();
    m[0] = alpha;
    m[1] = beta;
    m[2] = (1 - alpha) * centerX - beta * centerY;
    m[3] = -beta;
    m[4] = alpha;
    m[5] = beta * centerX + (1 - alpha) * centerY;
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_WarpAffine(Cv2Mat src, Cv2Mat dst, Cv2Mat m, int width, int height, int flags)
{
  if (src == nullptr || dst == nullptr || m == nullptr)
  {
    return copy_error("null Mat handle");
  }
  try
  {
    cv::warpAffine(*src, *dst, *m, cv::Size(width, height), flags);
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

char *Cv2_FindExternalContourRects(Cv2Mat src, int **rects, int *count)
{
  if (src == nullptr || rects == nullptr || count == nullptr)
  {
    return copy_error("null Mat handle");
  }
  *rects = nullptr;
  *count = 0;
  try
  {
    std::vector<std::vector<cv::Point>> contours;
    cv::findContours(*src, contours, cv::RETR_EXTERNAL, cv::CHAIN_APPROX_SIMPLE);
    if (contours.empty())
    {
      return nullptr;
    }
    int *out = static_cast<int *>(std::malloc(contours.size() * 4 * sizeof(int)));
    if (out == nullptr)
    {
      return copy_error("allocation failure");
    }
    for (size_t i = 0; i < contours.size(); i++)
    {
      // Tight integer bounding box, the documented cv::boundingRect
      // semantics, computed by hand: OpenCV 5 moved boundingRect into the
      // geometry module, which the base library set deliberately excludes.
      int minX = contours[i][0].x, maxX = minX;
      int minY = contours[i][0].y, maxY = minY;
      for (size_t p = 1; p < contours[i].size(); p++)
      {
        const cv::Point pt = contours[i][p];
        if (pt.x < minX) minX = pt.x;
        if (pt.x > maxX) maxX = pt.x;
        if (pt.y < minY) minY = pt.y;
        if (pt.y > maxY) maxY = pt.y;
      }
      out[i * 4] = minX;
      out[i * 4 + 1] = minY;
      out[i * 4 + 2] = maxX - minX + 1;
      out[i * 4 + 3] = maxY - minY + 1;
    }
    *rects = out;
    *count = (int)contours.size();
    return nullptr;
  }
  catch (...)
  {
    return current_exception_message();
  }
}

void Cv2_FreeIntArray(int *arr)
{
  std::free(arr);
}

void Cv2_FreeString(char *s)
{
  if (s != kErrorAllocFailed)
  {
    std::free(s);
  }
}

char *Cv2_CopyError(const char *msg)
{
  // The sentinel lives in this translation unit, so other wrapper sources
  // must allocate through here for Cv2_FreeString's identity check to hold.
  return copy_error(msg);
}
