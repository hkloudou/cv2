// Implementation of the C ABI declared in cv2capi.h.
//
// This file is compiled ahead of time (by build/build-wrapper.sh, normally
// inside GitHub Actions) into libcv2wrapper.a, so packages importing
// github.com/hkloudou/cv2 never need OpenCV headers or a C++ compiler
// beyond what cgo already requires.

#include "cv2capi.h"

#include <opencv2/imgproc.hpp>

#include <cstdlib>
#include <cstring>
#include <exception>

namespace
{
  // Returns a malloc'd copy of msg so it can cross the cgo boundary and be
  // freed by Cv2_FreeString.
  char *copy_error(const char *msg)
  {
    const size_t len = std::strlen(msg) + 1;
    char *out = static_cast<char *>(std::malloc(len));
    if (out != nullptr)
    {
      std::memcpy(out, msg, len);
    }
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
  return new cv::Mat();
}

Cv2Mat Cv2_Mat_NewFromBytes(int rows, int cols, int type, Cv2ByteArray buf)
{
  // Wrap the caller's buffer without copying, then clone so the returned Mat
  // owns its pixels. The Go garbage collector is therefore free to move or
  // collect the source slice after this call returns.
  const cv::Mat borrowed(rows, cols, type, buf.data);
  return new cv::Mat(borrowed.clone());
}

void Cv2_Mat_Close(Cv2Mat m)
{
  delete m;
}

char *Cv2_MatchTemplate(Cv2Mat image, Cv2Mat templ, Cv2Mat result, int method, Cv2Mat mask)
{
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

void Cv2_FreeString(char *s)
{
  std::free(s);
}
