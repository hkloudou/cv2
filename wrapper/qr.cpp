// QR code encode/detect/decode (github.com/hkloudou/cv2/qr).
//
// Compiled into libcv2qrwrapper.a, shipped in the optional
// libs/<goos>_<goarch>_qr modules together with opencv_objdetect and its
// dependency closure; only builds importing the qr subpackage link any of
// it. The implementation defers entirely to the official cv::QRCodeDetector
// and cv::QRCodeEncoder - this layer only marshals data.

#include "cv2internal.h"

#include <opencv2/imgproc.hpp>
#include <opencv2/objdetect.hpp>

#include <cstdlib>
#include <cstring>
#include <string>
#include <vector>

extern "C"
{
  char *Cv2_QRDetectAndDecodeMulti(Cv2Mat img, char ***texts, double **corners, int *count);
  char *Cv2_QREncode(const char *text, Cv2Mat out);
  void Cv2_FreeStringArray(char **arr, int n);
  void Cv2_FreeDoubleArray(double *arr);
}

char *Cv2_QRDetectAndDecodeMulti(Cv2Mat img, char ***texts, double **corners, int *count)
{
  if (img == nullptr || texts == nullptr || corners == nullptr || count == nullptr)
  {
    return Cv2_CopyError("null argument");
  }
  *texts = nullptr;
  *corners = nullptr;
  *count = 0;
  try
  {
    cv::Mat gray;
    if (img->channels() == 4)
      cv::cvtColor(*img, gray, cv::COLOR_RGBA2GRAY);
    else if (img->channels() == 3)
      cv::cvtColor(*img, gray, cv::COLOR_RGB2GRAY);
    else
      gray = *img;

    cv::QRCodeDetector detector;
    std::vector<std::string> decoded;
    std::vector<cv::Point2f> points;
    if (!detector.detectAndDecodeMulti(gray, decoded, points) || decoded.empty())
    {
      return nullptr; // none found
    }

    const int n = (int)decoded.size();
    char **outTexts = static_cast<char **>(std::calloc(n, sizeof(char *)));
    double *outCorners = static_cast<double *>(std::calloc((size_t)n * 8, sizeof(double)));
    if (outTexts == nullptr || outCorners == nullptr)
    {
      std::free(outTexts);
      std::free(outCorners);
      return Cv2_CopyError("allocation failure");
    }
    for (int i = 0; i < n; i++)
    {
      const size_t len = decoded[i].size() + 1;
      outTexts[i] = static_cast<char *>(std::malloc(len));
      if (outTexts[i] == nullptr)
      {
        Cv2_FreeStringArray(outTexts, n);
        std::free(outCorners);
        return Cv2_CopyError("allocation failure");
      }
      std::memcpy(outTexts[i], decoded[i].c_str(), len);
      for (int c = 0; c < 4; c++)
      {
        outCorners[i * 8 + c * 2] = points[(size_t)i * 4 + c].x;
        outCorners[i * 8 + c * 2 + 1] = points[(size_t)i * 4 + c].y;
      }
    }
    *texts = outTexts;
    *corners = outCorners;
    *count = n;
    return nullptr;
  }
  catch (...)
  {
    return cv2_current_exception_message();
  }
}

char *Cv2_QREncode(const char *text, Cv2Mat out)
{
  if (text == nullptr || out == nullptr)
  {
    return Cv2_CopyError("null argument");
  }
  try
  {
    cv::Ptr<cv::QRCodeEncoder> encoder = cv::QRCodeEncoder::create();
    encoder->encode(text, *out); // CV_8UC1, one pixel per module
    return nullptr;
  }
  catch (...)
  {
    return cv2_current_exception_message();
  }
}

void Cv2_FreeStringArray(char **arr, int n)
{
  if (arr == nullptr)
  {
    return;
  }
  for (int i = 0; i < n; i++)
  {
    std::free(arr[i]);
  }
  std::free(arr);
}

void Cv2_FreeDoubleArray(double *arr)
{
  std::free(arr);
}
